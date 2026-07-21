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

func TestNoBoundsSlicesAndConversion(t *testing.T) {
	build := func(noBounds bool) string {
		prog := newBoundsTestProgram(t)
		defer prog.Dispose()
		prog.SetNoBounds(noBounds)
		pkg := prog.NewPackage("p", "p")

		intParam := func(name string) *types.Var {
			return types.NewVar(0, nil, name, types.Typ[types.Int])
		}
		buildSlice := func(name string, valueType types.Type, threeIndex bool) {
			params := []*types.Var{types.NewVar(0, nil, "v", valueType), intParam("low"), intParam("high")}
			if threeIndex {
				params = append(params, intParam("max"))
			}
			results := types.NewTuple(types.NewVar(0, nil, "", valueType))
			fn := pkg.NewFunc(name, types.NewSignatureType(nil, nil, nil, types.NewTuple(params...), results, false), InGo)
			b := fn.MakeBody(1)
			max := Nil
			if threeIndex {
				max = fn.Param(3)
			}
			b.Return(b.Slice(fn.Param(0), fn.Param(1), fn.Param(2), max))
			b.EndBuild()
		}

		byteSlice := types.NewSlice(types.Typ[types.Byte])
		buildSlice("String2", types.Typ[types.String], false)
		buildSlice("Slice2", byteSlice, false)
		buildSlice("Slice3", byteSlice, true)

		array := types.NewArray(types.Typ[types.Byte], 4)
		arrayPtr := types.NewPointer(array)
		arrayFn := pkg.NewFunc("Array2", types.NewSignatureType(nil, nil, nil,
			types.NewTuple(types.NewVar(0, nil, "v", arrayPtr), intParam("low"), intParam("high")),
			types.NewTuple(types.NewVar(0, nil, "", byteSlice)), false), InGo)
		arrayBuilder := arrayFn.MakeBody(1)
		arrayBuilder.Return(arrayBuilder.Slice(arrayFn.Param(0), arrayFn.Param(1), arrayFn.Param(2), Nil))
		arrayBuilder.EndBuild()

		convertFn := pkg.NewFunc("Convert", types.NewSignatureType(nil, nil, nil,
			types.NewTuple(types.NewVar(0, nil, "v", byteSlice)),
			types.NewTuple(types.NewVar(0, nil, "", arrayPtr)), false), InGo)
		convertBuilder := convertFn.MakeBody(1)
		convertBuilder.Return(convertBuilder.SliceToArrayPointer(convertFn.Param(0), prog.rawType(arrayPtr)))
		convertBuilder.EndBuild()

		if err := llvm.VerifyModule(pkg.Module(), llvm.ReturnStatusAction); err != nil {
			t.Fatal(err)
		}
		return pkg.String()
	}

	checked := build(false)
	for _, helper := range []string{"StringSlice2", "NewSlice2", "NewSlice3Bounds", "PanicSliceConvert"} {
		if !strings.Contains(checked, helper) {
			t.Fatalf("default build did not emit %s:\n%s", helper, checked)
		}
	}

	unchecked := build(true)
	for _, helper := range []string{"StringSlice2", "NewSlice2", "NewSlice3Bounds", "PanicSliceConvert"} {
		if strings.Contains(unchecked, helper) {
			t.Fatalf("-B emitted checked helper %s:\n%s", helper, unchecked)
		}
	}
	if !strings.Contains(unchecked, "NewSliceNoBounds") {
		t.Fatalf("-B did not emit unchecked slice helper:\n%s", unchecked)
	}
	if !strings.Contains(unchecked, "StringSliceNoBounds") {
		t.Fatalf("-B did not emit unchecked string helper:\n%s", unchecked)
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
