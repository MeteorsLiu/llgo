//go:build !llgo

package escape

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goplus/llgo/internal/littest"
	"github.com/xgo-dev/llvm"
)

func TestTransformModule(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			dir := filepath.Join("testdata", entry.Name())
			input := filepath.Join(dir, "in.txt")
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
			if err := TransformModule(mod); err != nil {
				t.Fatal(err)
			}

			output := filepath.Join(dir, "out.txt")
			want, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			spec := littest.Spec{Path: output, Text: string(want), Mode: littest.ModeLiteral}
			if err := littest.Check(spec, mod.String()); err != nil {
				t.Fatalf("%v\n%s", err, mod.String())
			}
		})
	}
}
