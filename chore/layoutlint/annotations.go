package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func writeAnnotations(directory string, files []sourceFile, report runReport, caseName string) error {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	filtered := report
	filtered.Results = nil
	for _, result := range report.Results {
		if result.Case == caseName {
			filtered.Results = append(filtered.Results, result)
		}
	}
	for _, file := range files {
		annotated, err := annotatedSource([]byte(file.original), file.path, filtered)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(file.path)), annotated, 0644); err != nil {
			return err
		}
	}
	return nil
}

// annotatedSource only renders facts already recorded by the checker. Locations
// come from the input AST, so diagnostics do not depend on fixture names or markers.
func annotatedSource(source []byte, filename string, report runReport) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	fields := make(map[int][]string)
	ast.Inspect(file, func(node ast.Node) bool {
		if field, ok := node.(*ast.Field); ok {
			line := fset.Position(field.Pos()).Line
			for _, name := range field.Names {
				fields[line] = append(fields[line], name.Name)
			}
		}
		return true
	})
	type annotationKey struct {
		line                          int
		check, stage, item, want, got string
		warning                       bool
	}
	type annotation struct {
		key     annotationKey
		message string
		levels  []string
	}
	var annotations []annotation
	groups := make(map[annotationKey]int)
	failures, warnings := 0, 0
	for _, result := range report.Results {
		for _, diff := range result.Differences {
			if diff.File != filename {
				continue
			}
			if diff.Line < 1 || diff.Line > strings.Count(string(source), "\n")+1 {
				return nil, fmt.Errorf("invalid source location %s:%d", filename, diff.Line)
			}
			want, err := json.Marshal(diff.Want)
			if err != nil {
				return nil, err
			}
			got, err := json.Marshal(diff.Got)
			if err != nil {
				return nil, err
			}
			label := diff.Item
			valueSize := diff.Stage != "" && strings.HasSuffix(diff.Item, " size")
			fieldMatch := false
			if diff.Stage == "" || valueSize {
				if dot := strings.LastIndex(label, "."); dot >= 0 {
					label = label[dot+1:]
				}
				for _, name := range fields[diff.Line] {
					if trimmed, ok := strings.CutPrefix(label, name+" "); ok {
						label = trimmed
						fieldMatch = true
					}
				}
				if (diff.Check == "metadata" || valueSize) && label == "size" && fieldMatch {
					label = "reflected size"
				}
			}
			if valueSize && !fieldMatch {
				label = strings.TrimSuffix(label, " size") + " reflected size"
			}
			if diff.Check == "dwarf" {
				label = "DWARF " + label
			}
			if diff.Check == "ir" {
				operation := "load"
				if strings.Contains(diff.Item, ".layoutWrite_") {
					operation = "store"
				}
				property := "access type"
				if strings.HasSuffix(diff.Item, " alignment") {
					property = "alignment"
				}
				label = "LLVM " + operation + " " + property
			}
			unit := ""
			for _, suffix := range []string{"size", "PtrBytes", "offset", "stride", "alignment", "Sizeof", "Alignof"} {
				if strings.HasSuffix(label, suffix) {
					unit = " bytes"
					break
				}
			}
			key := annotationKey{diff.Line, diff.Check, diff.Stage, diff.Item, string(want), string(got), diff.Warning}
			if valueSize || fieldMatch && label == "reflected size" {
				key.check = "metadata"
				key.stage = ""
				key.item = label
			}
			index, ok := groups[key]
			if !ok {
				index = len(annotations)
				groups[key] = index
				message := fmt.Sprintf("%s = %s%s; expected %s", label, got, unit, want)
				if diff.Warning {
					message = fmt.Sprintf("warning: %s = %s; Go reports %s", label, got, want)
				}
				annotations = append(annotations, annotation{key: key, message: message})
			}
			level := "O" + result.Level
			if !slices.Contains(annotations[index].levels, level) {
				annotations[index].levels = append(annotations[index].levels, level)
			}
			if diff.Warning {
				warnings++
			} else {
				failures++
			}
		}
	}
	slices.SortStableFunc(annotations, func(a, b annotation) int { return a.key.line - b.key.line })
	var output strings.Builder
	fmt.Fprintf(&output, "// Layout observations for %s/%s. Source: %s\n", report.GOOS, report.GOARCH, filename)
	fmt.Fprintf(&output, "// %d failed observations, %d warnings; %d annotations after grouping repeated sizes and optimization levels.\n", failures, warnings, len(annotations))
	for _, result := range report.Results {
		fmt.Fprintf(&output, "// %s O%s: %s", result.Case, result.Level, result.Status)
		if result.Reason != "" {
			fmt.Fprintf(&output, "; reason=%q", result.Reason)
		}
		output.WriteByte('\n')
	}
	output.WriteByte('\n')
	index := 0
	for i, line := range strings.SplitAfter(string(source), "\n") {
		text := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		ending := line[len(text):]
		output.WriteString(text)
		separator := " // layoutlint: "
		for index < len(annotations) && annotations[index].key.line == i+1 {
			item := annotations[index]
			fmt.Fprintf(&output, "%s%s (%s)", separator, item.message, strings.Join(item.levels, ", "))
			separator = "; "
			index++
		}
		output.WriteString(ending)
	}
	return []byte(output.String()), nil
}
