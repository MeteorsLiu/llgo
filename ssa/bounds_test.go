//go:build !llgo

package ssa

import (
	"go/importer"
	"go/types"
	"runtime"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestNoBoundsIndexAddr(t *testing.T) {
	prog := NewProgram(nil)
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", runtime.GOARCH))
	prog.SetRuntime(func() *types.Package {
		pkg, err := importer.For("source", nil).Import(PkgRuntime)
		if err != nil {
			t.Fatal(err)
		}
		return pkg
	})
	prog.SetNoBounds(true)

	pkg := prog.NewPackage("p", "p")
	array := types.NewArray(types.Typ[types.Byte], 0)
	params := types.NewTuple(
		types.NewVar(0, nil, "a", types.NewPointer(array)),
		types.NewVar(0, nil, "i", types.Typ[types.Int]),
	)
	results := types.NewTuple(types.NewVar(0, nil, "", types.Typ[types.Byte]))
	fn := pkg.NewFunc("Get", types.NewSignatureType(nil, nil, nil, params, results, false), InGo)
	b := fn.MakeBody(1)
	b.Return(b.Load(b.IndexAddr(fn.Param(0), fn.Param(1))))
	b.EndBuild()

	if err := llvm.VerifyModule(pkg.Module(), llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	ir := pkg.String()
	if strings.Contains(ir, "CheckIndexRange") {
		t.Fatalf("-B emitted an index bounds check:\n%s", ir)
	}
	if !strings.Contains(ir, "AssertNilDeref") {
		t.Fatalf("-B must retain nil pointer checks:\n%s", ir)
	}
}
