package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

// annotatedSource renders existing results beside their source declarations or
// checkpoints. It does not add assertions or change the report's differences.
func annotatedSource(source []byte, cases []layoutCase, report runReport) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	declarations := make(map[string]int)
	fields := make(map[string]int)
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		declarations[spec.Name.Name] = fset.Position(spec.Pos()).Line
		if st, ok := spec.Type.(*ast.StructType); ok {
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					fields[spec.Name.Name+"."+name.Name] = fset.Position(field.Pos()).Line
				}
			}
		}
		return false
	})
	checkpoints := make(map[string]int)
	for _, group := range file.Comments {
		for _, comment := range group.List {
			label, ok := strings.CutPrefix(comment.Text, "// LLDB_BREAK: ")
			if !ok {
				continue
			}
			line := fset.Position(comment.Pos()).Line
			// A marker may end a multiline call, e.g. interface_values. Put
			// the annotation before the statement, not among its arguments.
			ast.Inspect(file, func(node ast.Node) bool {
				switch node.(type) {
				case *ast.ExprStmt, *ast.AssignStmt, *ast.ReturnStmt, *ast.IncDecStmt:
					start, end := fset.Position(node.Pos()).Line, fset.Position(node.End()).Line
					if start <= line && line <= end {
						checkpoints[label] = start
					}
				}
				return true
			})
		}
	}
	type annotationKey struct {
		line    int
		message string
	}
	type annotation struct {
		annotationKey
		levels []string
	}
	var annotations []annotation
	groups := make(map[annotationKey]int)
	total := 0
	for _, result := range report.Results {
		for _, diff := range result.Differences {
			line := 0
			if diff.Stage != "" {
				line = checkpoints[diff.Stage]
			} else {
				for _, c := range cases {
					if strings.HasPrefix(diff.Item, c.name+".") {
						line = declarations[c.name]
					}
					for i, field := range c.fields {
						fieldMatch := strings.HasPrefix(diff.Item, c.name+"."+field.path+" ")
						readMatch := strings.HasPrefix(diff.Item, fmt.Sprintf("main.layoutRead_%s_%d ", c.name, i))
						writeMatch := strings.HasPrefix(diff.Item, fmt.Sprintf("main.layoutWrite_%s_%d ", c.name, i))
						if fieldMatch || readMatch || writeMatch {
							// Nested offsets are relative to the outer struct. For
							// StructWithAllTypeFields.arr2[0].i, annotate arr2.
							name := field.path
							if end := strings.IndexAny(name, ".["); end >= 0 {
								name = name[:end]
							}
							line = fields[c.name+"."+name]
						}
					}
				}
			}
			if line == 0 {
				return nil, fmt.Errorf("cannot locate source for %s/%s: %s", diff.Check, diff.Stage, diff.Item)
			}
			want, err := json.Marshal(diff.Want)
			if err != nil {
				return nil, err
			}
			got, err := json.Marshal(diff.Got)
			if err != nil {
				return nil, err
			}
			check := diff.Check
			if diff.Stage != "" {
				check += "/" + diff.Stage
			}
			key := annotationKey{line, fmt.Sprintf("%s: %s; want=%s; got=%s", check, diff.Item, want, got)}
			index, ok := groups[key]
			if !ok {
				index = len(annotations)
				groups[key] = index
				annotations = append(annotations, annotation{annotationKey: key})
			}
			level := "O" + result.Level
			if !slices.Contains(annotations[index].levels, level) {
				annotations[index].levels = append(annotations[index].levels, level)
			}
			total++
		}
	}
	slices.SortStableFunc(annotations, func(a, b annotation) int { return a.line - b.line })
	var output strings.Builder
	fmt.Fprintf(&output, "// Layout observations for %s/%s from report.json.\n", report.GOOS, report.GOARCH)
	output.WriteString("// Source: test/debug/runtime/main.go; source line numbers below refer to that file.\n")
	output.WriteString("// Annotations locate failed observations, not confirmed compiler root causes.\n")
	fmt.Fprintf(&output, "// %d differences, %d annotations; identical results across optimization levels are grouped.\n", total, len(annotations))
	for _, result := range report.Results {
		fmt.Fprintf(&output, "// %s O%s: %s (%d differences)", result.Case, result.Level, result.Status, len(result.Differences))
		if result.Reason != "" {
			fmt.Fprintf(&output, "; reason=%q", result.Reason)
		}
		output.WriteByte('\n')
	}
	output.WriteByte('\n')
	index := 0
	for i, line := range strings.SplitAfter(string(source), "\n") {
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		for index < len(annotations) && annotations[index].line == i+1 {
			item := annotations[index]
			fmt.Fprintf(&output, "%s// LAYOUT-FAIL [%s] source:%d %s\n", indent, strings.Join(item.levels, ","), item.line, item.message)
			index++
		}
		output.WriteString(line)
	}
	return []byte(output.String()), nil
}
