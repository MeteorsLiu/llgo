package main

import (
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func dwarfData(executable, goos, toolDir string) (*dwarf.Data, error) {
	switch goos {
	case "darwin":
		// Mach-O executables carry a debug map; assemble the referenced object
		// DWARF into a private dSYM before using Go's DWARF reader.
		bundle := executable + ".dSYM"
		if _, err := runCommand(filepath.Dir(executable), executable+".dsym.log", filepath.Join(toolDir, "dsymutil"), executable, "-o", bundle); err != nil {
			return nil, err
		}
		file, err := macho.Open(filepath.Join(bundle, "Contents", "Resources", "DWARF", filepath.Base(executable)))
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return file.DWARF()
	case "windows":
		file, err := pe.Open(executable)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return file.DWARF()
	default:
		file, err := elf.Open(executable)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return file.DWARF()
	}
}

func underlyingDWARF(t dwarf.Type) dwarf.Type {
	for {
		switch value := t.(type) {
		case *dwarf.TypedefType:
			t = value.Type
		case *dwarf.QualType:
			t = value.Type
		default:
			return t
		}
	}
}

func checkDWARF(c layoutCase, data *dwarf.Data, reference sample) ([]difference, error) {
	var subject *dwarf.StructType
	reader := data.Reader()
	observer := fmt.Sprintf("%s.layoutType_%d", c.irPrefix, reference.Words[0])
	inside := false
	depth := 0
	for {
		entry, err := reader.Next()
		if err != nil {
			return nil, err
		}
		if entry == nil {
			break
		}
		if entry.Tag == 0 {
			if inside {
				depth--
				if depth == 0 {
					break
				}
			}
			continue
		}
		name, _ := entry.Val(dwarf.AttrName).(string)
		if !inside {
			if entry.Tag == dwarf.TagSubprogram && name == observer {
				inside = true
				depth = 1
			}
			continue
		}
		if entry.Children {
			depth++
		}
		if entry.Tag != dwarf.TagVariable || name != "items" {
			continue
		}
		offset, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			return nil, fmt.Errorf("DWARF %s.items has no type reference", observer)
		}
		typ, err := data.Type(offset)
		if err != nil {
			return nil, err
		}
		array, ok := underlyingDWARF(typ).(*dwarf.ArrayType)
		if !ok || array.Count != elementCount {
			return nil, fmt.Errorf("DWARF %s.items is not a [%d] array", observer, elementCount)
		}
		st, ok := underlyingDWARF(array.Type).(*dwarf.StructType)
		if !ok || st.Incomplete {
			return nil, fmt.Errorf("DWARF %s.items has no complete struct element", observer)
		}
		if subject != nil && subject.String() != st.String() {
			return nil, fmt.Errorf("conflicting DWARF %s definitions", c.name)
		}
		subject = st
	}
	if subject == nil {
		return nil, fmt.Errorf("%s.items struct is missing from DWARF", observer)
	}
	var differences []difference
	location := c.sourceLocation
	check := func(name string, want uint64, got int64) {
		if got < 0 || uint64(got) != want {
			differences = append(differences, difference{"dwarf", "", c.displayName + "." + name, want, got, location.file, location.line, false})
		}
	}
	check("struct size", reference.Words[6], subject.Size())
	for i, field := range c.fields {
		location = field.sourceLocation
		steps, err := selectorSteps(field.path)
		if err != nil {
			return nil, err
		}
		var typ dwarf.Type = subject
		var offset int64
		for _, step := range steps {
			typ = underlyingDWARF(typ)
			if step.name != "" {
				st, ok := typ.(*dwarf.StructType)
				if !ok {
					return nil, fmt.Errorf("DWARF %s: expected struct, got %s", field.path, typ)
				}
				var found *dwarf.StructField
				for _, member := range st.Field {
					if member.Name == step.name {
						found = member
						break
					}
				}
				if found == nil {
					return nil, fmt.Errorf("DWARF field %s not found", field.path)
				}
				offset += found.ByteOffset
				typ = found.Type
			} else {
				array, ok := typ.(*dwarf.ArrayType)
				if !ok || int64(step.index) >= array.Count {
					return nil, fmt.Errorf("DWARF array index missing: %s", field.path)
				}
				offset += int64(step.index) * array.Type.Size()
				typ = array.Type
			}
		}
		index := headerWords + i*fieldWords
		check(field.path+" offset", reference.Words[index+1], offset)
		check(field.path+" size", reference.Words[index+3], typ.Size())
	}
	return differences, nil
}

func checkIR(c layoutCase, path string, reference sample) ([]difference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ir := string(data)
	var differences []difference
	alignmentPattern := regexp.MustCompile(`\balign ([0-9]+)\b`)
	for i, field := range c.fields {
		if field.typ == "" {
			continue
		}
		for _, operation := range []string{"Read", "Write"} {
			name := fmt.Sprintf("%s.layout%s_%s_%d", c.irPrefix, operation, c.name, i)
			function := regexp.MustCompile(`(?m)^define [^\n]*@"?` + regexp.QuoteMeta(name) + `"?\([^\n]*\{\n([\s\S]*?)^}`)
			match := function.FindStringSubmatch(ir)
			if match == nil {
				return nil, fmt.Errorf("IR function %s missing", name)
			}
			mnemonic := "load"
			if operation == "Write" {
				mnemonic = "store"
			}
			access := regexp.MustCompile(`(?m)^\s*(?:%[^=]+ = )?`+mnemonic+`\s+([^\s,]+)[^\n]*`).FindAllStringSubmatch(match[1], -1)
			if len(access) != 1 {
				return nil, fmt.Errorf("%s has %d %s instructions; expected one directly generated field access", name, len(access), mnemonic)
			}
			index := headerWords + i*fieldWords
			alignment := alignmentPattern.FindStringSubmatch(access[0][0])
			if alignment == nil {
				return nil, fmt.Errorf("%s has no explicit alignment in serialized IR", name)
			}
			got, err := strconv.ParseUint(alignment[1], 10, 64)
			if err != nil {
				return nil, err
			}
			if got > reference.Words[index+4] {
				differences = append(differences, difference{"ir", "", name + " alignment", reference.Words[index+4], got, field.file, field.line, false})
			}
			wantType := fmt.Sprintf("i%d", reference.Words[index+3]*8)
			if field.typ == "float64" {
				wantType = "double"
			} else if field.typ == "float32" {
				wantType = "float"
			} else if field.typ == "bool" {
				wantType = "i1"
			} else if strings.HasPrefix(field.typ, "*") || field.typ == "unsafe.Pointer" {
				wantType = "ptr"
			}
			if access[0][1] != wantType {
				differences = append(differences, difference{"ir", "", name + " access type", wantType, access[0][1], field.file, field.line, false})
			}
		}
	}
	return differences, nil
}
