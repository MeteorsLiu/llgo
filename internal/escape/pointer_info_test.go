//go:build !llgo

package escape

import (
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestPotentialCopiesOfRoot(t *testing.T) {
	input := filepath.Join("testdata", "pointer-info", "in.txt")
	buffer, err := llvm.NewMemoryBufferFromFile(input)
	if err != nil {
		t.Fatal(err)
	}
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	mod, err := ctx.ParseIR(buffer)
	if err != nil {
		t.Fatalf("parse %s: %v", input, err)
	}
	defer mod.Dispose()

	analysis := newCopyAnalysis(mod)
	defer analysis.dispose()
	tests := []struct {
		function   string
		wantOK     bool
		wantCopies []string
	}{
		{function: "exact_field", wantOK: true, wantCopies: []string{"q"}},
		{function: "disjoint_field", wantOK: true},
		{function: "overlapping_whole_read"},
		{function: "overwritten_before_read", wantOK: true},
		{function: "unknown_offset"},
		{function: "undef_destination", wantOK: true},
		{function: "null_destination", wantOK: true},
		{function: "null_pointer_valid"},
		{function: "unsupported_destination"},
		{function: "invalid_pointer_info"},
	}
	for _, test := range tests {
		t.Run(test.function, func(t *testing.T) {
			fn := mod.NamedFunction(test.function)
			if fn.IsNil() {
				t.Fatalf("function %s not found", test.function)
			}
			var root llvm.Value
			for block := fn.FirstBasicBlock(); !block.IsNil(); block = llvm.NextBasicBlock(block) {
				for instr := block.FirstInstruction(); !instr.IsNil(); instr = llvm.NextInstruction(instr) {
					if instr.Name() == "root" {
						root = instr
					}
				}
			}
			if root.IsNil() {
				t.Fatalf("root in %s not found", test.function)
			}

			ok := true
			var copies []llvm.Value
			stores := 0
			for use := root.FirstUse(); !use.IsNil(); use = use.NextUse() {
				user := use.User()
				if user.InstructionOpcode() != llvm.Store || user.Operand(0) != root {
					continue
				}
				stores++
				storeCopies, complete := analysis.getPotentialCopiesOfStoredValue(user)
				if !complete {
					ok = false
					break
				}
				copies = append(copies, storeCopies...)
			}
			if stores == 0 {
				t.Fatalf("root %s has no stored-value use", root.Name())
			}
			if ok != test.wantOK {
				t.Fatalf("potentialCopies(%s) ok = %v, want %v", root.Name(), ok, test.wantOK)
			}
			got := make([]string, 0, len(copies))
			for _, copy := range copies {
				got = append(got, copy.Name())
			}
			sort.Strings(got)
			if !slices.Equal(got, test.wantCopies) {
				t.Fatalf("potentialCopies(%s) = %v, want %v", root.Name(), got, test.wantCopies)
			}
		})
	}
}
