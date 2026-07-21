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
	prog := newBoundsTestProgram(t)
	defer prog.Dispose()
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

func TestNoBoundsSlices(t *testing.T) {
	prog := newBoundsTestProgram(t)
	defer prog.Dispose()
	prog.SetNoBounds(true)
	pkg := prog.NewPackage("p", "p")

	intParam := func(name string) *types.Var {
		return types.NewVar(0, nil, name, types.Typ[types.Int])
	}
	build := func(name string, valueType types.Type, threeIndex bool) {
		params := []*types.Var{types.NewVar(0, nil, "v", valueType), intParam("low"), intParam("high")}
		if threeIndex {
			params = append(params, intParam("max"))
		}
		sig := types.NewSignatureType(nil, nil, nil, types.NewTuple(params...), types.NewTuple(types.NewVar(0, nil, "", valueType)), false)
		fn := pkg.NewFunc(name, sig, InGo)
		b := fn.MakeBody(1)
		max := Nil
		if threeIndex {
			max = fn.Param(3)
		}
		b.Return(b.Slice(fn.Param(0), fn.Param(1), fn.Param(2), max))
		b.EndBuild()
	}

	build("String2", types.Typ[types.String], false)
	build("Slice2", types.NewSlice(types.Typ[types.Byte]), false)
	build("Slice3", types.NewSlice(types.Typ[types.Byte]), true)

	if err := llvm.VerifyModule(pkg.Module(), llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	ir := pkg.String()
	for _, helper := range []string{"StringSlice2", "NewSlice2", "NewSlice3Bounds"} {
		if strings.Contains(ir, helper) {
			t.Fatalf("-B emitted checked slice helper %s:\n%s", helper, ir)
		}
	}
}

func newBoundsTestProgram(t *testing.T) Program {
	t.Helper()
	prog := NewProgram(nil)
	prog.TypeSizes(types.SizesFor("gc", runtime.GOARCH))
	prog.SetRuntime(func() *types.Package {
		pkg, err := importer.For("source", nil).Import(PkgRuntime)
		if err != nil {
			t.Fatal(err)
		}
		return pkg
	})
	return prog
}
