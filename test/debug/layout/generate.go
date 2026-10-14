package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

//go:embed fixture.go.tmpl
var fixtureTemplate string

const headerWords = 19
const fieldWords = 6
const elementCount = 2

func skipReason(c layoutCase) (string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "declarations.go", "package main\n"+c.declaration, parser.ParseComments)
	if err != nil {
		return "", err
	}
	for _, group := range f.Comments {
		for _, comment := range group.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
			if text == "llgo:cdecl" || text == "llgo:stdcall" || strings.HasPrefix(text, "llgo:type C") {
				return "special ABI declaration: " + text, nil
			}
		}
	}
	return "", nil
}

type selectorStep struct {
	name  string
	index int
}

func selectorSteps(path string) ([]selectorStep, error) {
	expression, err := parser.ParseExpr("root." + path)
	if err != nil {
		return nil, err
	}
	var reversed []selectorStep
	for {
		switch e := expression.(type) {
		case *ast.SelectorExpr:
			reversed = append(reversed, selectorStep{name: e.Sel.Name})
			expression = e.X
		case *ast.IndexExpr:
			literal, ok := e.Index.(*ast.BasicLit)
			if !ok {
				return nil, fmt.Errorf("non-constant field index: %s", path)
			}
			n, err := strconv.Atoi(literal.Value)
			if err != nil || n < 0 {
				return nil, fmt.Errorf("invalid field index: %s", path)
			}
			reversed = append(reversed, selectorStep{index: n})
			expression = e.X
		case *ast.Ident:
			steps := make([]selectorStep, len(reversed))
			for i := range reversed {
				steps[len(steps)-1-i] = reversed[i]
			}
			return steps, nil
		default:
			return nil, fmt.Errorf("unsupported field path: %s", path)
		}
	}
}

func generate(c layoutCase, id int) (string, error) {
	var body strings.Builder
	fmt.Fprintf(&body, "func layoutType_%d() {\n var items [2]%s\n layoutRoots=append(layoutRoots,unsafe.Pointer(&items))\n", id, c.name)
	fmt.Fprintf(&body, "observation:=make([]uint64,%d)\n", headerWords+len(c.fields)*fieldWords)
	fmt.Fprintln(&body, `var boxed any=items[0]
 typeAddress:=(*[2]unsafe.Pointer)(unsafe.Pointer(&boxed))[0]
 header:=(*layoutTypePrefix)(typeAddress)
 typ:=reflect.TypeOf(items[0])
 base:=uintptr(unsafe.Pointer(&items[0]))`)
	fmt.Fprintf(&body, "observation[0]=%d\n", id)
	fmt.Fprintln(&body, `observation[1]=uint64(base)
 observation[2]=uint64(uintptr(unsafe.Pointer(&items[1]))-base)
 observation[3]=uint64(unsafe.Sizeof(items[0]))
 observation[4]=uint64(unsafe.Alignof(items[0]))
 observation[5]=uint64(uintptr(typeAddress))
 observation[6]=uint64(header.size)
 observation[7]=uint64(header.ptrBytes)
 observation[8]=uint64(header.align)
 observation[9]=uint64(header.fieldAlign)
 observation[10]=uint64(header.kind & 31)
 observation[11]=uint64(typ.Size())
 observation[12]=uint64(typ.Align())
 observation[13]=uint64(typ.FieldAlign())
 observation[17]=uint64(unsafe.Sizeof(uintptr(0)))
 one:=uint16(1)
 observation[18]=uint64(*(*byte)(unsafe.Pointer(&one)))`)
	for i, field := range c.fields {
		steps, err := selectorSteps(field.path)
		if err != nil {
			return "", err
		}
		fmt.Fprintln(&body, "{ ft:=typ; var metadataOffset uintptr")
		prefix := "items[0]"
		var offsets []string
		for _, step := range steps {
			if step.name != "" {
				fmt.Fprintf(&body, "{ f,_:=ft.FieldByName(%q); metadataOffset+=f.Offset; ft=f.Type }\n", step.name)
				prefix += "." + step.name
				offsets = append(offsets, "unsafe.Offsetof("+prefix+")")
			} else {
				fmt.Fprintf(&body, "metadataOffset+=uintptr(%d)*ft.Elem().Size(); ft=ft.Elem()\n", step.index)
				offsets = append(offsets, fmt.Sprintf("uintptr(%d)*unsafe.Sizeof(%s[0])", step.index, prefix))
				prefix += fmt.Sprintf("[%d]", step.index)
			}
		}
		index := headerWords + i*fieldWords
		fmt.Fprintf(&body, "observation[%d]=uint64(%s)\nobservation[%d]=uint64(uintptr(unsafe.Pointer(&%s))-base)\n", index, strings.Join(offsets, "+"), index+1, prefix)
		fmt.Fprintf(&body, "observation[%d]=uint64(metadataOffset)\nobservation[%d]=uint64(ft.Size())\nobservation[%d]=uint64(ft.Align())\nobservation[%d]=uint64(ft.Kind())\n}\n", index+2, index+3, index+4, index+5)
	}
	fmt.Fprintln(&body, "layoutSnapshot(observation)\n}")
	// These private accessors only expose IR for existing fields; they do not
	// introduce a runtime write or another test scenario.
	for i, field := range c.fields {
		if field.typ == "" {
			continue
		}
		fmt.Fprintf(&body, "//go:noinline\nfunc layoutRead_%s_%d(p *%s) %s {return p.%s}\n", c.name, i, c.name, field.typ, field.path)
		fmt.Fprintf(&body, "//go:noinline\nfunc layoutWrite_%s_%d(p *%s,value %s) {p.%s=value}\n", c.name, i, c.name, field.typ, field.path)
	}
	return body.String(), nil
}

func formatSource(source string) (string, error) {
	data, err := format.Source([]byte(source))
	return string(data), err
}
