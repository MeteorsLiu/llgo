package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type observationSite struct {
	label, check string
	names        []string
	types        []types.Type
	locations    []sourceLocation
	sourceLocation
}

type sourceFile struct {
	path, original, instrumented string
	packagePath, packageName     string
	checkpointLine               int
}

type existingFixture struct {
	name, directory, moduleRoot string
	arguments                   []string
	test                        bool
	module                      string
	cases                       []layoutCase
	sites                       []observationSite
	files                       []sourceFile
	excluded                    []string
	loadError                   error
}

type sourceEdit struct {
	start, end int
	text       string
}

func loadFixtures(config runConfig, patterns []string) ([]*existingFixture, error) {
	mode := packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedForTest | packages.NeedModule
	loaded, err := packages.Load(&packages.Config{Dir: config.root, Mode: mode, Tests: config.test}, patterns...)
	if err != nil {
		return nil, err
	}
	var fixtures []*existingFixture
	for _, pkg := range loaded {
		if pkg.ForTest != "" && pkg.ForTest != pkg.PkgPath || strings.HasSuffix(pkg.PkgPath, ".test") {
			continue
		}
		if config.test && pkg.ForTest == "" && slices.ContainsFunc(loaded, func(other *packages.Package) bool {
			return other.ForTest == pkg.PkgPath && other.PkgPath == pkg.PkgPath
		}) {
			continue
		}
		fixture, err := loadFixture(config, pkg, loaded)
		if err != nil {
			fixture = &existingFixture{name: pkg.PkgPath, loadError: err}
		}
		fixtures = append(fixtures, fixture)
	}
	if len(fixtures) == 0 {
		return nil, fmt.Errorf("no input packages matched %q", patterns)
	}
	return fixtures, nil
}

func loadFixture(config runConfig, pkg *packages.Package, loaded []*packages.Package) (*existingFixture, error) {
	var packageErrors []string
	for _, failure := range pkg.Errors {
		packageErrors = append(packageErrors, failure.Error())
	}
	if len(packageErrors) != 0 {
		return nil, fmt.Errorf("load input: %s", strings.Join(packageErrors, "\n"))
	}
	source := pkg
	if config.test && len(source.Syntax) == 0 {
		// A directory containing only package p_test has an empty p package.
		// Keep p as the build input and locate its source through p_test.
		for _, external := range loaded {
			if external.ForTest == pkg.PkgPath && external.PkgPath == pkg.PkgPath+"_test" {
				source = external
				break
			}
		}
	}
	if len(source.Syntax) == 0 {
		return nil, fmt.Errorf("%s has no Go source", pkg.ID)
	}
	fixture := &existingFixture{name: pkg.PkgPath, directory: filepath.Dir(source.GoFiles[0]), arguments: []string{"."}, test: config.test}
	if pkg.PkgPath == "command-line-arguments" {
		fixture.arguments = nil
		for _, file := range source.GoFiles {
			fixture.arguments = append(fixture.arguments, filepath.Base(file))
		}
		fixture.name = filepath.Base(source.GoFiles[0])
	}
	if !config.test && pkg.Name != "main" {
		return nil, fmt.Errorf("%s is a library package; use -test to observe its existing tests", pkg.PkgPath)
	}
	modulePath := pkg.PkgPath
	if modulePath == "command-line-arguments" {
		modulePath = "layout-observation"
	}
	moduleDir := fixture.directory
	var moduleData []byte
	for {
		var readErr error
		moduleData, readErr = os.ReadFile(filepath.Join(moduleDir, "go.mod"))
		if readErr == nil {
			break
		}
		if !os.IsNotExist(readErr) {
			return nil, readErr
		}
		parent := filepath.Dir(moduleDir)
		if parent == moduleDir {
			moduleData = []byte("module " + modulePath + "\n\ngo " + strings.TrimPrefix(config.goversion, "go") + "\n")
			moduleDir = fixture.directory
			break
		}
		moduleDir = parent
	}
	moduleFile, err := modfile.Parse("go.mod", moduleData, nil)
	if err != nil {
		return nil, err
	}
	for _, replace := range moduleFile.Replace {
		if replace.New.Version == "" && !filepath.IsAbs(replace.New.Path) {
			if err := moduleFile.AddReplace(replace.Old.Path, replace.Old.Version, filepath.Join(moduleDir, replace.New.Path), ""); err != nil {
				return nil, err
			}
		}
	}
	moduleFile.Cleanup()
	formatted, err := moduleFile.Format()
	if err != nil {
		return nil, err
	}
	fixture.module = string(formatted)
	fixture.moduleRoot = moduleDir
	if err := discoverPackage(fixture, pkg); err != nil {
		return nil, err
	}
	if config.test {
		for _, external := range loaded {
			if external.ForTest == pkg.PkgPath && external.PkgPath == pkg.PkgPath+"_test" {
				if err := discoverPackage(fixture, external); err != nil {
					return nil, err
				}
			}
		}
	}
	return fixture, nil
}

func discoverPackage(fixture *existingFixture, pkg *packages.Package) error {
	fset, info := pkg.Fset, pkg.TypesInfo
	special := make(map[*types.TypeName]bool)
	for _, file := range pkg.Syntax {
		for _, declaration := range file.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.TYPE {
				continue
			}
			for _, spec := range group.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				for _, comments := range []*ast.CommentGroup{group.Doc, typeSpec.Doc} {
					if comments == nil {
						continue
					}
					for _, comment := range comments.List {
						text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
						if text == "llgo:cdecl" || text == "llgo:stdcall" || strings.HasPrefix(text, "llgo:type C") {
							object := info.Defs[typeSpec.Name].(*types.TypeName)
							special[object] = true
							fixture.excluded = append(fixture.excluded, pkg.PkgPath+"."+object.Name()+": "+text)
						}
					}
				}
			}
		}
	}
	var excluded func(types.Type, map[types.Type]bool) bool
	excluded = func(typ types.Type, seen map[types.Type]bool) bool {
		if seen[typ] {
			return false
		}
		seen[typ] = true
		if named, ok := typ.(*types.Named); ok && special[named.Obj()] {
			return true
		}
		switch t := typ.Underlying().(type) {
		case *types.Pointer:
			return excluded(t.Elem(), seen)
		case *types.Array:
			return excluded(t.Elem(), seen)
		case *types.Struct:
			for i := 0; i < t.NumFields(); i++ {
				if excluded(t.Field(i).Type(), seen) {
					return true
				}
			}
		}
		return false
	}
	caseStart := len(fixture.cases)
	fileStart := len(fixture.files)
	for _, file := range pkg.Syntax {
		path := fset.PositionFor(file.Pos(), false).Filename
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fixture.files = append(fixture.files, sourceFile{path: path, original: string(data), packagePath: pkg.PkgPath, packageName: pkg.Name})
	}
	// Type identities come from the selected package, including instantiated
	// generic types and anonymous struct literals; no fixture names are selected.
	seenTypes := make(map[types.Type]bool)
	for fileIndex, file := range pkg.Syntax {
		source := &fixture.files[fileStart+fileIndex]
		qualifier := func(p *types.Package) string {
			if p == pkg.Types {
				return ""
			}
			for _, spec := range file.Imports {
				path, _ := strconv.Unquote(spec.Path.Value)
				if path == p.Path() {
					if spec.Name != nil {
						return spec.Name.Name
					}
					return p.Name()
				}
			}
			return p.Name()
		}
		var candidates []struct {
			typ types.Type
			pos token.Pos
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.TypeSpec:
				object := info.Defs[n.Name].(*types.TypeName)
				if object.Parent() == pkg.Types.Scope() {
					typ := object.Type()
					if named, ok := typ.(*types.Named); ok && named.TypeParams().Len() != 0 && named.TypeArgs().Len() == 0 {
						return false
					}
					candidates = append(candidates, struct {
						typ types.Type
						pos token.Pos
					}{typ, n.Pos()})
					return false
				}
			case *ast.StructType:
				candidates = append(candidates, struct {
					typ types.Type
					pos token.Pos
				}{info.TypeOf(n), n.Pos()})
			case *ast.Ident:
				if instance, ok := info.Instances[n]; ok {
					candidates = append(candidates, struct {
						typ types.Type
						pos token.Pos
					}{instance.Type, n.Pos()})
				}
			}
			return true
		})
		for _, candidate := range candidates {
			typ := candidate.typ
			st, ok := typ.Underlying().(*types.Struct)
			if !ok || seenTypes[typ] {
				continue
			}
			seenTypes[typ] = true
			if excluded(typ, make(map[types.Type]bool)) {
				fixture.excluded = append(fixture.excluded, types.TypeString(typ, qualifier)+": contains a special ABI type")
				continue
			}
			if hasTypeParameter(typ, make(map[types.Type]bool)) {
				continue
			}
			name := fmt.Sprintf("layoutSubject_%d", len(fixture.cases))
			display := types.TypeString(typ, qualifier)
			c := layoutCase{name: name, displayName: display, declaration: "type " + name + " = " + display + "\n", typ: typ, sourceLocation: sourceLocation{source.path, fset.Position(candidate.pos).Line}, irPrefix: pkg.PkgPath}
			if pkg.Name == "main" {
				c.irPrefix = "main"
			}
			var visit func(types.Type, string, sourceLocation)
			visit = func(t types.Type, path string, location sourceLocation) {
				field := fieldCase{path: path, sourceLocation: location}
				switch t.Underlying().(type) {
				case *types.Basic, *types.Pointer:
					if basic, ok := t.Underlying().(*types.Basic); !ok || basic.Kind() != types.String && basic.Kind() != types.Complex64 && basic.Kind() != types.Complex128 {
						field.typ = types.TypeString(t, qualifier)
					}
				}
				c.fields = append(c.fields, field)
				switch t := t.Underlying().(type) {
				case *types.Struct:
					for i := 0; i < t.NumFields(); i++ {
						member := t.Field(i)
						if member.Pkg() != nil && member.Pkg() != pkg.Types && !member.Exported() {
							continue
						}
						visit(member.Type(), path+"."+member.Name(), location)
					}
				case *types.Array:
					for i := int64(0); i < t.Len(); i++ {
						visit(t.Elem(), fmt.Sprintf("%s[%d]", path, i), location)
					}
				}
			}
			for i := 0; i < st.NumFields(); i++ {
				member := st.Field(i)
				if member.Name() == "_" {
					continue
				}
				location := c.sourceLocation
				if position := fset.Position(member.Pos()); position.Filename == source.path {
					location.line = position.Line
				}
				visit(member.Type(), member.Name(), location)
			}
			fixture.cases = append(fixture.cases, c)
		}
	}
	for fileIndex, file := range pkg.Syntax {
		source := &fixture.files[fileStart+fileIndex]
		var edits []sourceEdit
		var generated strings.Builder
		position := func(pos token.Pos) int { return fset.PositionFor(pos, false).Offset }
		if fileIndex == 0 {
			_, body, _ := strings.Cut(fixtureTemplate, "var layoutRoots")
			generated.WriteString("\nvar layoutRoots" + body)
		}
		for i := caseStart; i < len(fixture.cases); i++ {
			c := fixture.cases[i]
			if c.file != source.path {
				continue
			}
			code, err := generate(c, i*2)
			if err != nil {
				return err
			}
			generated.WriteString("\n" + c.declaration + code)
			fmt.Fprintf(&generated, "\nfunc init(){layoutType_%d()}\n", i*2)
		}
		capture := func(statement ast.Node, pos token.Pos, at int, after bool, check string) {
			scope := info.Scopes[file].Innermost(pos)
			var literal *ast.FuncLit
			ast.Inspect(file, func(node ast.Node) bool {
				if function, ok := node.(*ast.FuncLit); ok && function.Body.Pos() < pos && pos < function.Body.End() {
					literal = function
				}
				return true
			})
			used := make(map[*types.Var]bool)
			if literal != nil {
				ast.Inspect(literal.Body, func(node ast.Node) bool {
					if identifier, ok := node.(*ast.Ident); ok {
						if object, ok := info.Uses[identifier].(*types.Var); ok {
							used[object] = true
						}
					}
					return true
				})
			}
			variables := make(map[string]*types.Var)
			for s := scope; s != nil && s != pkg.Types.Scope(); s = s.Parent() {
				for _, name := range s.Names() {
					object, ok := s.Lookup(name).(*types.Var)
					if !ok || name == "_" {
						continue
					}
					// A declaration is not visible in its own initializer. For
					// example, var f = func(){...} must not observe f there.
					if _, visible := scope.LookupParent(name, pos); visible != object {
						continue
					}
					if literal != nil && object.Pos() < literal.Pos() && !used[object] {
						continue
					}
					if _, exists := variables[name]; !exists {
						variables[name] = object
					}
				}
			}
			ast.Inspect(statement, func(node ast.Node) bool {
				if identifier, ok := node.(*ast.Ident); ok {
					if object, ok := info.Uses[identifier].(*types.Var); ok && object.Parent() == pkg.Types.Scope() {
						variables[object.Name()] = object
					}
				}
				return true
			})
			var names []string
			for name, object := range variables {
				if excluded(object.Type(), make(map[types.Type]bool)) {
					continue
				}
				if hasTypeParameter(object.Type(), make(map[types.Type]bool)) {
					continue
				}
				names = append(names, name)
			}
			slices.Sort(names)
			if len(names) == 0 {
				return
			}
			location := sourceLocation{source.path, fset.Position(statement.Pos()).Line}
			phase := "before"
			if after {
				phase = "after"
			}
			site := observationSite{label: fmt.Sprintf("%s:%d:%s", filepath.Base(source.path), location.line, phase), check: check, sourceLocation: location}
			var code strings.Builder
			fmt.Fprintf(&code, "\n{ layoutMu.Lock(); layoutWords:=[]uint64{%d}\n", len(fixture.sites)*2+1)
			var value func(types.Type, string, bool, sourceLocation)
			value = func(t types.Type, expression string, followPointer bool, declaration sourceLocation) {
				id := len(site.names)
				site.names = append(site.names, expression)
				site.types = append(site.types, t)
				site.locations = append(site.locations, declaration)
				fmt.Fprintf(&code, "{layoutCopy:=%s;layoutWords=append(layoutWords,%d);layoutWords=layoutAppend(layoutWords,layoutreflect.ValueOf(&layoutCopy).Elem(),layoutunsafe.Pointer(&(%s)))}\n", expression, id, expression)
				switch t := t.Underlying().(type) {
				case *types.Struct:
					for i := 0; i < t.NumFields(); i++ {
						member := t.Field(i)
						if member.Name() == "_" || member.Pkg() != nil && member.Pkg() != pkg.Types && !member.Exported() {
							continue
						}
						fieldLocation := declaration
						if member.Pkg() == pkg.Types {
							position := fset.Position(member.Pos())
							fieldLocation = sourceLocation{position.Filename, position.Line}
						}
						value(member.Type(), "("+expression+")."+member.Name(), false, fieldLocation)
					}
				case *types.Array:
					for i := int64(0); i < t.Len(); i++ {
						value(t.Elem(), fmt.Sprintf("(%s)[%d]", expression, i), false, declaration)
					}
				case *types.Pointer:
					if _, ok := t.Elem().Underlying().(*types.Struct); ok && followPointer {
						fmt.Fprintf(&code, "if %s!=nil {\n", expression)
						value(t.Elem(), "(*"+expression+")", false, declaration)
						code.WriteString("}\n")
					}
				}
			}
			for _, name := range names {
				position := fset.Position(variables[name].Pos())
				value(variables[name].Type(), name, true, sourceLocation{position.Filename, position.Line})
			}
			code.WriteString("layoutSnapshot(layoutWords);layoutMu.Unlock()\n}\n")
			fixture.sites = append(fixture.sites, site)
			edits = append(edits, sourceEdit{at, at, code.String()})
		}
		ast.Inspect(file, func(node ast.Node) bool {
			block, ok := node.(*ast.BlockStmt)
			if !ok {
				return true
			}
			for _, statement := range block.List {
				switch statement.(type) {
				case *ast.ExprStmt, *ast.AssignStmt, *ast.ReturnStmt, *ast.IncDecStmt:
					check := "access"
					if _, ok := statement.(*ast.ReturnStmt); ok {
						check = "call_return"
					}
					capture(statement, statement.Pos(), position(statement.Pos()), false, check)
				}
				switch statement.(type) {
				case *ast.AssignStmt, *ast.IncDecStmt, *ast.DeclStmt:
					capture(statement, statement.End(), position(statement.End()), true, "access")
				}
			}
			return true
		})
		if len(edits) == 0 && generated.Len() == 0 {
			source.instrumented = source.original
			continue
		}
		body := strings.NewReplacer("reflect.", "layoutreflect.", "unsafe.", "layoutunsafe.", "fmt.", "layoutfmt.", "math.", "layoutmath.", "runtime.", "layoutruntime.", "sync.", "layoutsync.").Replace(generated.String())
		edits = append(edits, sourceEdit{len(source.original), len(source.original), body})
		var imports strings.Builder
		imports.WriteString("\nimport (\n")
		for _, name := range []string{"reflect", "unsafe", "fmt", "math", "runtime", "sync"} {
			used := strings.Contains(body, "layout"+name+".")
			for _, edit := range edits {
				used = used || strings.Contains(edit.text, "layout"+name+".")
			}
			if used {
				fmt.Fprintf(&imports, "layout%s %q\n", name, name)
			}
		}
		imports.WriteString(")\n")
		edits = append(edits, sourceEdit{position(file.Name.End()), position(file.Name.End()), imports.String()})
		slices.SortStableFunc(edits, func(a, b sourceEdit) int { return b.start - a.start })
		text := source.original
		for _, edit := range edits {
			text = text[:edit.start] + edit.text + text[edit.end:]
		}
		formatted, err := formatSource(text)
		if err != nil {
			return fmt.Errorf("instrument %s: %w", source.path, err)
		}
		source.instrumented = formatted
		for i, line := range strings.Split(formatted, "\n") {
			if strings.Contains(line, "LLDB_BREAK: layout_snapshot") {
				source.checkpointLine = i + 1
			}
		}
	}
	return nil
}

func hasTypeParameter(typ types.Type, seen map[types.Type]bool) bool {
	if seen[typ] {
		return false
	}
	seen[typ] = true
	switch t := typ.(type) {
	case *types.TypeParam:
		return true
	case *types.Named:
		if t.TypeParams().Len() != 0 && t.TypeArgs().Len() == 0 {
			return true
		}
	}
	switch t := typ.Underlying().(type) {
	case *types.Struct:
		for i := 0; i < t.NumFields(); i++ {
			if hasTypeParameter(t.Field(i).Type(), seen) {
				return true
			}
		}
	case *types.Array:
		return hasTypeParameter(t.Elem(), seen)
	case *types.Pointer:
		return hasTypeParameter(t.Elem(), seen)
	}
	return false
}
