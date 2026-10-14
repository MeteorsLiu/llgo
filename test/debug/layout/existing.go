package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type observationSite struct {
	label string
	names []string
	types []types.Type
}
type existingFixture struct {
	cases                              []layoutCase
	sites                              []observationSite
	mainSource, observerSource, module string
}
type sourceEdit struct {
	start, end int
	text       string
}

func loadExisting(root, goarch string) (*existingFixture, error) {
	directory := filepath.Join(root, "test", "debug", "runtime")
	mainData, err := os.ReadFile(filepath.Join(directory, "main.go"))
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	mainFile, err := parser.ParseFile(fset, "main.go", mainData, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	goroutines, err := parser.ParseFile(fset, filepath.Join(directory, "goroutines.go"), nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	sizes := types.SizesFor("gc", goarch)
	info := &types.Info{Scopes: make(map[ast.Node]*types.Scope), Uses: make(map[*ast.Ident]types.Object)}
	config := types.Config{Importer: importer.Default(), Sizes: sizes}
	pkg, err := config.Check("lldbtest", fset, []*ast.File{mainFile, goroutines}, info)
	if err != nil {
		return nil, err
	}
	fixture := &existingFixture{}
	module, err := os.ReadFile(filepath.Join(directory, "go.mod"))
	if err != nil {
		return nil, err
	}
	fixture.module = string(module)
	declarations := strings.TrimPrefix(string(mainData), "package main")
	reason, err := skipReason(layoutCase{declaration: declarations})
	if err != nil {
		return nil, err
	}
	if reason != "" {
		return nil, fmt.Errorf("runtime fixture excluded: %s", reason)
	}
	for _, name := range pkg.Scope().Names() {
		obj, ok := pkg.Scope().Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		st, ok := obj.Type().Underlying().(*types.Struct)
		if !ok {
			continue
		}
		c := layoutCase{name: name, declaration: declarations + "\ntype subject = " + name + "\n"}
		var visit func(types.Type, string)
		visit = func(t types.Type, path string) {
			field := fieldCase{path: path}
			switch t.Underlying().(type) {
			case *types.Basic, *types.Pointer:
				if b, ok := t.Underlying().(*types.Basic); !ok || b.Kind() != types.String && b.Kind() != types.Complex64 && b.Kind() != types.Complex128 {
					field.typ = types.TypeString(t, func(p *types.Package) string {
						if p == pkg {
							return ""
						}
						return p.Name()
					})
				}
			}
			c.fields = append(c.fields, field)
			switch t := t.Underlying().(type) {
			case *types.Struct:
				for i := 0; i < t.NumFields(); i++ {
					visit(t.Field(i).Type(), path+"."+t.Field(i).Name())
				}
			case *types.Array:
				for i := int64(0); i < t.Len(); i++ {
					visit(t.Elem(), fmt.Sprintf("%s[%d]", path, i))
				}
			}
		}
		for i := 0; i < st.NumFields(); i++ {
			visit(st.Field(i).Type(), st.Field(i).Name())
		}
		fixture.cases = append(fixture.cases, c)
	}
	var helper strings.Builder
	helper.WriteString(fixtureTemplate)
	for i, c := range fixture.cases {
		code, err := generate(c, i)
		if err != nil {
			return nil, err
		}
		helper.WriteString(code)
	}
	var edits []sourceEdit
	position := func(p token.Pos) int { return fset.Position(p).Offset }
	edits = append(edits, sourceEdit{position(mainFile.Name.End()), position(mainFile.Name.End()), "\nimport (layoutreflect \"reflect\"; layoutunsafe \"unsafe\")\n"})
	// The native traceback scenario cannot link with Go. This temporary overlay
	// excludes the same scenario from both observations, as approved; the source
	// fixture and its full existing debugger suite remain untouched.
	for _, spec := range mainFile.Imports {
		if spec.Path.Value == `"runtime"` {
			edits = append(edits, sourceEdit{position(spec.Pos()), position(spec.End()), ""})
		}
	}
	for _, decl := range mainFile.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Name.Name == "RuntimeGoroutineValues" {
			edits = append(edits, sourceEdit{position(fn.Pos()), position(fn.End()), ""})
			continue
		}
		if fn.Name.Name == "main" {
			var calls strings.Builder
			for i := range fixture.cases {
				fmt.Fprintf(&calls, "\nlayoutType_%d()", i)
			}
			calls.WriteString("\n")
			edits = append(edits, sourceEdit{position(fn.Body.Lbrace) + 1, position(fn.Body.Lbrace) + 1, calls.String()})
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			statement, ok := node.(*ast.ExprStmt)
			if !ok {
				return true
			}
			call, ok := statement.X.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if ok && name.Name == "RuntimeGoroutineValues" {
				edits = append(edits, sourceEdit{position(statement.Pos()), position(statement.End()), ""})
			}
			return true
		})
		for _, group := range mainFile.Comments {
			for _, comment := range group.List {
				label, ok := strings.CutPrefix(comment.Text, "// LLDB_BREAK: ")
				if !ok || comment.Pos() < fn.Body.Pos() || comment.Pos() > fn.Body.End() {
					continue
				}
				line := fset.Position(comment.Pos()).Line
				var statement ast.Stmt
				ast.Inspect(fn.Body, func(node ast.Node) bool {
					switch n := node.(type) {
					case *ast.ExprStmt, *ast.AssignStmt, *ast.ReturnStmt, *ast.IncDecStmt:
						if fset.Position(n.Pos()).Line <= line && fset.Position(n.End()).Line >= line {
							statement = n.(ast.Stmt)
						}
					}
					return true
				})
				if statement == nil {
					return nil, fmt.Errorf("no statement for marker %s", label)
				}
				scope := info.Scopes[mainFile].Innermost(statement.Pos())
				variables := make(map[string]*types.Var)
				for s := scope; s != nil && s != info.Scopes[mainFile]; s = s.Parent() {
					for _, name := range s.Names() {
						obj, ok := s.Lookup(name).(*types.Var)
						if !ok || obj.Pos() >= statement.Pos() {
							continue
						}
						if _, exists := variables[name]; !exists {
							variables[name] = obj
						}
					}
				}
				// Include globals read at this existing checkpoint, e.g. the map and
				// channel results in InspectContainerValues.
				ast.Inspect(statement, func(node ast.Node) bool {
					if id, ok := node.(*ast.Ident); ok {
						if obj, ok := info.Uses[id].(*types.Var); ok && obj.Parent() == pkg.Scope() {
							variables[obj.Name()] = obj
						}
					}
					return true
				})
				var names []string
				for name := range variables {
					names = append(names, name)
				}
				slices.Sort(names)
				site := observationSite{label: label}
				var capture strings.Builder
				fmt.Fprintf(&capture, "{ layoutWords:=[]uint64{%d}\n", len(fixture.cases)+len(fixture.sites))
				var captureValue func(types.Type, string, bool)
				captureValue = func(t types.Type, expression string, followPointer bool) {
					id := len(site.names)
					site.names = append(site.names, expression)
					site.types = append(site.types, t)
					fmt.Fprintf(&capture, "{ layoutCopy:=%s; layoutWords=append(layoutWords,%d); layoutWords=layoutAppend(layoutWords,layoutreflect.ValueOf(&layoutCopy).Elem(),layoutunsafe.Pointer(&(%s))) }\n", expression, id, expression)
					switch t := t.Underlying().(type) {
					case *types.Struct:
						for i := 0; i < t.NumFields(); i++ {
							captureValue(t.Field(i).Type(), "("+expression+")."+t.Field(i).Name(), false)
						}
					case *types.Array:
						for i := int64(0); i < t.Len(); i++ {
							captureValue(t.Elem(), fmt.Sprintf("(%s)[%d]", expression, i), false)
						}
					case *types.Pointer:
						if _, ok := t.Elem().Underlying().(*types.Struct); ok && followPointer {
							fmt.Fprintf(&capture, "if %s!=nil {\n", expression)
							captureValue(t.Elem(), "(*"+expression+")", false)
							capture.WriteString("}\n")
						}
					}
				}
				for _, name := range names {
					captureValue(variables[name].Type(), name, true)
				}
				capture.WriteString("layoutSnapshot(layoutWords)\n}\n")
				fixture.sites = append(fixture.sites, site)
				edits = append(edits, sourceEdit{position(statement.Pos()), position(statement.Pos()), capture.String()})
			}
		}
	}
	slices.SortStableFunc(edits, func(a, b sourceEdit) int {
		if a.start > b.start {
			return -1
		}
		if a.start < b.start {
			return 1
		}
		return 0
	})
	source := string(mainData)
	for _, edit := range edits {
		source = source[:edit.start] + edit.text + source[edit.end:]
	}
	fixture.mainSource, err = formatSource(source)
	if err != nil {
		return nil, err
	}
	fixture.observerSource, err = formatSource(helper.String())
	if err != nil {
		return nil, err
	}
	return fixture, nil
}
