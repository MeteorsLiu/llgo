package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
)

// expectedType owns only storage facts. Pointer referents and function
// signatures do not participate in the containing value's physical layout.
type expectedType struct {
	size, align, ptrBytes int64
	closure               bool
	offsets               []int64
	fields                []*expectedType
	element               *expectedType
}

func expectedLayout(typ types.Type, sizes types.Sizes) *expectedType {
	word := sizes.Sizeof(types.Typ[types.Uintptr])
	layout := &expectedType{size: sizes.Sizeof(typ), align: sizes.Alignof(typ)}
	switch t := typ.Underlying().(type) {
	case *types.Signature:
		// LLGo stores the code pointer and environment inline. Go stores a
		// single pointer to a funcval; that representation is not our oracle.
		layout.size, layout.ptrBytes, layout.closure = 2*word, 2*word, true
	case *types.Struct:
		var end int64
		for i := 0; i < t.NumFields(); i++ {
			field := expectedLayout(t.Field(i).Type(), sizes)
			offset := (end + field.align - 1) &^ (field.align - 1)
			layout.offsets = append(layout.offsets, offset)
			layout.fields = append(layout.fields, field)
			layout.closure = layout.closure || field.closure
			if field.ptrBytes != 0 {
				layout.ptrBytes = offset + field.ptrBytes
			}
			end = offset + field.size
			// A trailing zero-sized field in a nonempty Go struct must not
			// point immediately past the allocation.
			if i == t.NumFields()-1 && field.size == 0 && end != 0 {
				end++
			}
		}
		layout.size = (end + layout.align - 1) &^ (layout.align - 1)
	case *types.Array:
		layout.element = expectedLayout(t.Elem(), sizes)
		layout.closure = layout.element.closure
		layout.size = t.Len() * layout.element.size
		if t.Len() != 0 && layout.element.ptrBytes != 0 {
			layout.ptrBytes = (t.Len()-1)*layout.element.size + layout.element.ptrBytes
		}
	case *types.Pointer, *types.Map, *types.Chan, *types.Slice:
		layout.ptrBytes = word
	case *types.Interface:
		layout.ptrBytes = 2 * word
	case *types.Basic:
		if t.Kind() == types.String || t.Kind() == types.UnsafePointer {
			layout.ptrBytes = word
		}
	}
	return layout
}

func expectedReference(c layoutCase, goarch string, reference []sample) ([]sample, error) {
	fset := token.NewFileSet()
	source, err := parser.ParseFile(fset, "declarations.go", "package main\n"+c.declaration, 0)
	if err != nil {
		return nil, err
	}
	sizes := types.SizesFor("gc", goarch)
	if sizes == nil {
		return nil, fmt.Errorf("no Go layout rules for %s", goarch)
	}
	config := types.Config{Importer: importer.Default(), Sizes: sizes, IgnoreFuncBodies: true}
	pkg, err := config.Check("layoutcase", fset, []*ast.File{source}, nil)
	if err != nil {
		return nil, err
	}
	obj := pkg.Scope().Lookup("subject")
	if obj == nil {
		return nil, fmt.Errorf("subject declaration is missing")
	}
	typ := obj.Type()
	layout := expectedLayout(typ, sizes)
	if !layout.closure {
		return reference, nil
	}
	// Preserve baseline.json exactly as observed. All downstream layout,
	// metadata, DWARF and IR checks use this separate LLGo expectation.
	expected := make([]sample, len(reference))
	for stage, current := range reference {
		expected[stage] = current
		words := append([]uint64(nil), current.Words...)
		expected[stage].Words = words
		if words[6] != uint64(sizes.Sizeof(typ)) || words[8] != uint64(sizes.Alignof(typ)) {
			return nil, fmt.Errorf("Go reference layout for %s disagrees with gc sizes", c.name)
		}
		for _, index := range []int{2, 3, 6, 11} {
			words[index] = uint64(layout.size)
		}
		words[7] = uint64(layout.ptrBytes)
		for i, field := range c.fields {
			steps, err := selectorSteps(field.path)
			if err != nil {
				return nil, err
			}
			currentType, currentLayout := typ, layout
			var offset int64
			for _, step := range steps {
				if step.name != "" {
					st, ok := currentType.Underlying().(*types.Struct)
					if !ok {
						return nil, fmt.Errorf("%s: %s is not a struct", field.path, currentType)
					}
					index := -1
					for j := 0; j < st.NumFields(); j++ {
						if st.Field(j).Name() == step.name {
							index = j
							break
						}
					}
					if index < 0 {
						return nil, fmt.Errorf("%s: field %s is missing", field.path, step.name)
					}
					offset += currentLayout.offsets[index]
					currentType, currentLayout = st.Field(index).Type(), currentLayout.fields[index]
				} else {
					array, ok := currentType.Underlying().(*types.Array)
					if !ok || int64(step.index) >= array.Len() {
						return nil, fmt.Errorf("%s: invalid array index %d", field.path, step.index)
					}
					offset += int64(step.index) * currentLayout.element.size
					currentType, currentLayout = array.Elem(), currentLayout.element
				}
			}
			index := headerWords + i*fieldWords
			words[index], words[index+1], words[index+2] = uint64(offset), uint64(offset), uint64(offset)
			words[index+3] = uint64(currentLayout.size)
		}
	}
	return expected, nil
}
