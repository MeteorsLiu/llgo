package ssa

import (
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestAbiStructFieldsUseGo386Offsets(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", "386"))

	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "I", types.Typ[types.Int], false),
		types.NewField(token.NoPos, nil, "F", types.Typ[types.Float64], false),
	}, nil)
	got := (&aBuilder{Prog: prog}).abiStructFieldOffsets(st)
	want := []int64{0, 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Windows 386 reflected field offsets = %v, want %v", got, want)
	}
}

func TestGo386AtomicCmpXchgUsesNativeLLVMResult(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", "386"))
	pkg := prog.NewPackage("example.com/p", "p")
	b := pkg.NewFunc("example.com/p.f", NoArgsNoRet, GoDecl).MakeBody(1)
	valueType := prog.Uint64()
	ptr := b.Alloc(valueType, false)
	result := b.AtomicCmpXchg(ptr, prog.IntVal(0, valueType), prog.IntVal(1, valueType))
	_ = b.Extract(result, 1)
	b.Return()
	b.EndBuild()
	if ir := pkg.String(); !strings.Contains(ir, "extractvalue { i64, i1 }") {
		t.Fatalf("Windows/386 cmpxchg result was not extracted from its native LLVM type:\n%s", ir)
	}
}

func TestGo386StructPhysicalLayoutDiffersFromC(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", "386"))

	fields := []*types.Var{
		types.NewField(token.NoPos, nil, "I", types.Typ[types.Int], false),
		types.NewField(token.NoPos, nil, "F", types.Typ[types.Float64], false),
	}
	st := types.NewStruct(fields, nil)
	goType := prog.Type(st, GoDecl)
	if got, want := prog.SizeOf(goType), uint64(12); got != want {
		t.Fatalf("Go Windows/386 struct size = %d, want %d", got, want)
	}
	if got, want := prog.OffsetOf(goType, 1), uint64(4); got != want {
		t.Fatalf("Go Windows/386 float64 field offset = %d, want %d", got, want)
	}
	if zero := prog.Zero(goType); zero.impl.Type() != goType.ll {
		t.Fatalf("Go Windows/386 struct zero has type %s, want %s", zero.impl.Type(), goType.ll)
	}
	other := &aType{
		ll:  goType.ll,
		raw: rawType{Type: types.NewTuple(types.NewVar(token.NoPos, nil, "", types.Typ[types.Int]))},
	}
	if _, ok := prog.structLayout(other); ok {
		t.Fatal("LLVM-identical type inherited layout metadata from a different Go type")
	}

	pkg := types.NewPackage("example.com/c", "c")
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Pair", nil), st, nil)
	prog.SetTypeCallConv("example.com/c.Pair", CDecl)
	cType := prog.Type(named, CDecl)
	if got, want := prog.SizeOf(cType), uint64(16); got != want {
		t.Fatalf("C Windows/386 struct size = %d, want %d", got, want)
	}
	if got, want := prog.OffsetOf(cType, 1), uint64(8); got != want {
		t.Fatalf("C Windows/386 float64 field offset = %d, want %d", got, want)
	}

	Initialize(InitAllTargets | InitAllTargetInfos | InitAllTargetMCs)
	runtimePackage, err := importer.For("source", nil).Import(PkgRuntime)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []*Target{
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-pc-windows-msvc"},
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-w64-windows-gnu"},
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j32", LLVMTarget: "wasm32-unknown-emscripten"},
		{GOOS: "wasip1", GOARCH: "wasm", WasmProfile: "w32", LLVMTarget: "wasm32-unknown-wasip1"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j64", LLVMTarget: "wasm64-unknown-emscripten"},
	} {
		t.Run(target.GOOS+"/"+target.GOARCH+"/"+target.LLVMTarget, func(t *testing.T) {
			p := NewProgram(target)
			defer p.Dispose()
			sizes := types.SizesFor("gc", target.GOARCH)
			p.TypeSizes(sizes)
			p.SetRuntime(runtimePackage)
			owner := types.NewPackage("example.com/nested", "nested")
			module := p.NewPackage(owner.Path(), owner.Name())
			function := module.NewFunc("check", NoArgsNoRet, GoDecl)
			builder := function.MakeBody(1)
			type layout struct {
				size, alignment, prefix uint64
				offsets                 []uint64
			}
			// Expected aggregates are composed from primitive target rules and
			// the selected owner convention, never from the resulting LLVM struct.
			scalar := func(raw types.Type, native bool) layout {
				convention := GoDecl
				if native {
					convention = CDecl
				}
				storage := p.Type(raw, convention)
				value := layout{size: p.SizeOf(storage), alignment: p.AlignOf(storage)}
				if !native {
					value.alignment = uint64(sizes.Alignof(raw))
				}
				if _, ok := raw.(*types.Pointer); ok {
					value.prefix = value.size
				}
				return value
			}
			compose := func(fields []layout) layout {
				value := layout{alignment: 1}
				for _, field := range fields {
					value.size = (value.size + field.alignment - 1) / field.alignment * field.alignment
					value.offsets = append(value.offsets, value.size)
					if field.prefix != 0 {
						value.prefix = value.size + field.prefix
					}
					value.size += field.size
					if field.alignment > value.alignment {
						value.alignment = field.alignment
					}
				}
				value.size = (value.size + value.alignment - 1) / value.alignment * value.alignment
				return value
			}
			verify := func(raw types.Type, want layout) {
				t.Helper()
				storage := p.rawType(raw)
				if p.SizeOf(storage) != want.size || p.AlignOf(storage) != want.alignment {
					t.Errorf("%s storage size/alignment=%d/%d, want %d/%d", raw, p.SizeOf(storage), p.AlignOf(storage), want.size, want.alignment)
				}
				for i, offset := range want.offsets {
					if got := p.OffsetOf(storage, i); got != offset {
						t.Errorf("%s field %d offset=%d, want %d", raw, i, got, offset)
					}
				}
				builder.abiType(raw)
				name, _ := p.abi.TypeName(raw)
				common := module.VarOf(name).impl.Initializer()
				for common.Type() != p.AbiType().ll {
					common = common.Operand(0)
				}
				if common.Operand(0).ZExtValue() != want.size || common.Operand(1).ZExtValue() != want.prefix || common.Operand(4).ZExtValue() != want.alignment || common.Operand(5).ZExtValue() != want.alignment {
					t.Errorf("%s metadata size/prefix/align/fieldAlign=%d/%d/%d/%d, want %d/%d/%d/%d", raw, common.Operand(0).ZExtValue(), common.Operand(1).ZExtValue(), common.Operand(4).ZExtValue(), common.Operand(5).ZExtValue(), want.size, want.prefix, want.alignment, want.alignment)
				}
				if len(want.offsets) != 0 {
					fieldName := name
					if p.nativeMetadataType(raw) == nil {
						fieldName, _ = p.abi.TypeName(raw.Underlying())
					}
					fields := module.VarOf(fieldName + "$fields").impl.Initializer()
					for i, offset := range want.offsets {
						if got := fields.Operand(i).Operand(2).ZExtValue(); got != offset {
							t.Errorf("%s metadata field %d offset=%d, want %d", raw, i, got, offset)
						}
					}
				}
			}
			// Each bit selects Go or cdecl for one of three layers: all eight
			// combinations are checked through direct, alias, generic and array fields.
			for mask := 0; mask < 8; mask++ {
				for _, variant := range []string{"direct", "alias", "generic", "array0", "array1", "array2", "namedarray", "matrix2"} {
					var child types.Type
					var childLayout layout
					for level := 0; level < 3; level++ {
						native := mask&(1<<level) != 0
						name := fmt.Sprintf("S%d_%s_%d", mask, variant, level)
						fieldTypes := []types.Type{types.Typ[types.Byte], types.Typ[types.Uint64], types.NewPointer(types.Typ[types.Byte])}
						fieldLayouts := []layout{scalar(fieldTypes[0], native), scalar(fieldTypes[1], native), scalar(fieldTypes[2], native)}
						if level > 0 {
							wrapped, wrappedLayout := child, childLayout
							switch variant {
							case "alias":
								wrapped = types.NewAlias(types.NewTypeName(token.NoPos, owner, name+"Alias", nil), child)
							case "array0", "array1", "array2":
								length := int64(variant[len(variant)-1] - '0')
								wrapped = types.NewArray(child, length)
								wrappedLayout = layout{size: uint64(length) * childLayout.size, alignment: childLayout.alignment}
								if length != 0 && childLayout.prefix != 0 {
									wrappedLayout.prefix = uint64(length-1)*childLayout.size + childLayout.prefix
								}
							case "namedarray", "matrix2":
								length := uint64(2)
								if variant == "namedarray" {
									wrapped = types.NewNamed(types.NewTypeName(token.NoPos, owner, name+"Array", nil), types.NewArray(child, 2), nil)
								} else {
									wrapped = types.NewArray(types.NewArray(child, 2), 2)
									length = 4
								}
								wrappedLayout = layout{size: length * childLayout.size, alignment: childLayout.alignment}
								if childLayout.prefix != 0 {
									wrappedLayout.prefix = (length-1)*childLayout.size + childLayout.prefix
								}
							}
							fieldTypes = []types.Type{types.Typ[types.Byte], wrapped, types.Typ[types.Uint32]}
							fieldLayouts = []layout{scalar(fieldTypes[0], native), wrappedLayout, scalar(fieldTypes[2], native)}
						}
						fields := make([]*types.Var, len(fieldTypes))
						for i, typ := range fieldTypes {
							fields[i] = types.NewField(token.NoPos, owner, fmt.Sprintf("F%d", i), typ, false)
						}
						named := types.NewNamed(types.NewTypeName(token.NoPos, owner, name, nil), types.NewStruct(fields, nil), nil)
						if variant == "generic" && level > 0 {
							parameter := types.NewTypeParam(types.NewTypeName(token.NoPos, owner, "T", nil), types.NewInterfaceType(nil, nil).Complete())
							fields[1] = types.NewField(token.NoPos, owner, "F1", parameter, false)
							named.SetUnderlying(types.NewStruct(fields, nil))
							named.SetTypeParams([]*types.TypeParam{parameter})
							instance, err := types.Instantiate(nil, named, []types.Type{child}, true)
							if err != nil {
								t.Fatal(err)
							}
							named = instance.(*types.Named)
						}
						if native {
							p.SetTypeCallConv(owner.Path()+"."+name, CDecl)
						}
						child, childLayout = named, compose(fieldLayouts)
						verify(child, childLayout)
						for _, length := range []int64{0, 1, 2} {
							arrayLayout := layout{size: uint64(length) * childLayout.size, alignment: childLayout.alignment}
							if length != 0 && childLayout.prefix != 0 {
								arrayLayout.prefix = uint64(length-1)*childLayout.size + childLayout.prefix
							}
							verify(types.NewArray(child, length), arrayLayout)
						}
						if level > 0 {
							storage := p.rawType(child)
							for element := uint64(0); element < 2; element++ {
								if address := element*p.SizeOf(storage) + p.OffsetOf(storage, 1); address%fieldLayouts[1].alignment != 0 {
									t.Errorf("%s element %d nested field is not aligned to %d", child, element, fieldLayouts[1].alignment)
								}
							}
						}
					}
				}
			}
			builder.Return()
			builder.EndBuild()
			if err := llvm.VerifyModule(module.Module(), llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGo386StructLayoutEdgeCases(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()

	t.Run("single float64 retains Go alignment", func(t *testing.T) {
		st := types.NewStruct([]*types.Var{
			types.NewField(token.NoPos, nil, "F", types.Typ[types.Float64], false),
		}, nil)
		typ := prog.Type(st, GoDecl)
		if got, want := prog.SizeOf(typ), uint64(8); got != want {
			t.Fatalf("Go Windows/386 struct size = %d, want %d", got, want)
		}
		if got, want := prog.AlignOf(typ), uint64(4); got != want {
			t.Fatalf("Go Windows/386 struct alignment = %d, want %d", got, want)
		}
		if got, want := typ.ll.StructElementTypes(), 2; len(got) != want {
			t.Fatalf("LLVM struct element count = %d, want %d", len(got), want)
		}
	})

	t.Run("scalar arrays retain Go field alignment", func(t *testing.T) {
		for _, length := range []int64{0, 1, 2} {
			for _, element := range []types.Type{types.Typ[types.Uint64], types.NewArray(types.Typ[types.Uint64], 2)} {
				array := types.NewArray(element, length)
				st := types.NewStruct([]*types.Var{
					types.NewField(token.NoPos, nil, "B", types.Typ[types.Byte], false),
					types.NewField(token.NoPos, nil, "A", array, false),
					types.NewField(token.NoPos, nil, "End", types.Typ[types.Byte], false),
				}, nil)
				storage := prog.Type(st, GoDecl)
				wantSize := uint64(8 + types.SizesFor("gc", "386").Sizeof(array))
				if prog.OffsetOf(storage, 1) != 4 || prog.AlignOf(storage) != 4 || prog.SizeOf(storage) != wantSize {
					t.Errorf("%s field offset/align/size=%d/%d/%d, want 4/4/%d", array, prog.OffsetOf(storage, 1), prog.AlignOf(storage), prog.SizeOf(storage), wantSize)
				}
			}
		}
	})

	t.Run("final zero-sized field has a distinct address", func(t *testing.T) {
		empty := types.NewStruct(nil, nil)
		st := types.NewStruct([]*types.Var{
			types.NewField(token.NoPos, nil, "I", types.Typ[types.Int32], false),
			types.NewField(token.NoPos, nil, "Z", empty, false),
		}, nil)
		typ := prog.Type(st, GoDecl)
		if got, want := prog.SizeOf(typ), uint64(8); got != want {
			t.Fatalf("Go Windows/386 struct size = %d, want %d", got, want)
		}
		if got, want := prog.OffsetOf(typ, 1), uint64(4); got != want {
			t.Fatalf("zero-sized field offset = %d, want %d", got, want)
		}
	})

	t.Run("named zero constant uses its physical layout", func(t *testing.T) {
		st := types.NewStruct([]*types.Var{
			types.NewField(token.NoPos, nil, "F", types.Typ[types.Float64], false),
		}, nil)
		pkg := types.NewPackage("example.com/p", "p")
		named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Value", nil), st, nil)
		typ := prog.Type(named, GoDecl)
		if zero := prog.Zero(typ); zero.impl.Type() != typ.ll {
			t.Fatalf("named Go Windows/386 struct zero has type %s, want %s", zero.impl.Type(), typ.ll)
		}
	})

	Initialize(InitAllTargets | InitAllTargetInfos | InitAllTargetMCs)
	for _, target := range []*Target{
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-pc-windows-msvc"},
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-w64-windows-gnu"},
		{GOOS: "linux", GOARCH: "386"},
		{GOOS: "linux", GOARCH: "arm"},
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "arm64"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j32", LLVMTarget: "wasm32-unknown-emscripten"},
		{GOOS: "wasip1", GOARCH: "wasm", WasmProfile: "w32", LLVMTarget: "wasm32-unknown-wasip1"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j64", LLVMTarget: "wasm64-unknown-emscripten"},
	} {
		t.Run("atomic/"+target.GOOS+"/"+target.GOARCH+"/"+target.LLVMTarget, func(t *testing.T) {
			p := NewProgram(target)
			defer p.Dispose()
			p.TypeSizes(types.SizesFor("gc", target.GOARCH))
			for _, path := range []string{"sync/atomic", "internal/runtime/atomic", "example.com/other"} {
				owner := types.NewPackage(path, "atomic")
				marker := types.NewNamed(types.NewTypeName(token.NoPos, owner, "align64", nil), types.NewStruct(nil, nil), nil)
				alignment := uint64(8)
				if path == "example.com/other" {
					alignment = 1
				}
				if storage := p.Type(marker, GoDecl); p.SizeOf(storage) != 0 || p.AlignOf(storage) != alignment {
					t.Errorf("%s marker size/alignment=%d/%d, want 0/%d", path, p.SizeOf(storage), p.AlignOf(storage), alignment)
				}
				derived := types.NewNamed(types.NewTypeName(token.NoPos, owner, "Other", nil), marker.Underlying(), nil)
				if got := p.AlignOf(p.Type(derived, GoDecl)); got != 1 {
					t.Errorf("derived empty struct inherited marker alignment: %d", got)
				}
				alias := types.NewAlias(types.NewTypeName(token.NoPos, owner, "Alias", nil), marker)
				for index, fieldType := range []types.Type{marker, alias, types.NewArray(marker, 0), types.NewArray(marker, 2)} {
					for _, native := range []bool{false, true} {
						name := fmt.Sprintf("Owner%d_%t", index, native)
						fields := []*types.Var{
							types.NewField(token.NoPos, owner, "B", types.Typ[types.Byte], false),
							types.NewField(token.NoPos, owner, "M", fieldType, false),
							types.NewField(token.NoPos, owner, "U", types.Typ[types.Uint32], false),
						}
						named := types.NewNamed(types.NewTypeName(token.NoPos, owner, name, nil), types.NewStruct(fields, nil), nil)
						if native {
							p.SetTypeCallConv(owner.Path()+"."+name, CDecl)
						}
						storage := p.Type(named, GoDecl)
						wantOffset, wantSize, wantAlignment := uint64(8), uint64(16), uint64(8)
						if alignment == 1 {
							wantOffset, wantSize, wantAlignment = 1, 8, 4
						}
						if p.OffsetOf(storage, 1) != wantOffset || p.SizeOf(storage) != wantSize || p.AlignOf(storage) != wantAlignment {
							t.Errorf("%s %s field offset/size/alignment=%d/%d/%d, want %d/%d/%d", path, name, p.OffsetOf(storage, 1), p.SizeOf(storage), p.AlignOf(storage), wantOffset, wantSize, wantAlignment)
						}
						if got := p.SizeOf(p.rawType(types.NewArray(named, 2))); got != 2*wantSize {
							t.Errorf("%s %s array size=%d, want %d", path, name, got, 2*wantSize)
						}
					}
				}
			}
		})
	}
}

func TestGo386TupleUsesGoStructLayout(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", "386"))

	results := types.NewTuple(
		types.NewVar(token.NoPos, nil, "", types.Typ[types.Int64]),
		types.NewVar(token.NoPos, nil, "", types.Typ[types.Int32]),
		types.NewVar(token.NoPos, nil, "", types.Typ[types.Int64]),
	)
	tupleType := prog.rawType(results)
	if got, want := prog.SizeOf(tupleType), uint64(20); got != want {
		t.Fatalf("Go Windows/386 tuple size = %d, want %d", got, want)
	}
	if got, want := prog.OffsetOf(tupleType, 2), uint64(12); got != want {
		t.Fatalf("Go Windows/386 tuple third result offset = %d, want %d", got, want)
	}

	sig := types.NewSignatureType(nil, nil, nil, nil, results, false)
	if got, want := prog.toLLVMFunc(sig).ReturnType(), tupleType.ll; got != want {
		t.Fatalf("function result type = %s, want tuple type %s", got, want)
	}
	equivalentStruct := prog.Struct(
		prog.Int64(),
		prog.Int32(),
		prog.Int64(),
	)
	if got, want := equivalentStruct.ll, tupleType.ll; got != want {
		t.Fatalf("multi-result value type = %s, want function tuple type %s", got, want)
	}
}

func TestConstStructValueAsUsesDestinationLLVMType(t *testing.T) {
	prog := NewProgram(&Target{GOOS: "windows", GOARCH: "386"})
	defer prog.Dispose()
	prog.TypeSizes(types.SizesFor("gc", "386"))

	fields := []*types.Var{
		types.NewField(token.NoPos, nil, "B", types.Typ[types.Uint8], false),
		types.NewField(token.NoPos, nil, "U", types.Typ[types.Uint64], false),
	}
	st := types.NewStruct(fields, nil)
	pkg := types.NewPackage("example.com/p", "p")
	named := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Value", nil), st, nil)
	typ := prog.Type(named, GoDecl)
	destination := prog.ctx.StructCreateNamed("example.com/p.destination")
	destination.StructSetBody(typ.ll.StructElementTypes(), typ.ll.IsStructPacked())

	value := prog.constStructValueAs(typ, destination, []llvm.Value{
		prog.IntVal(1, prog.Byte()).impl,
		prog.IntVal(2, prog.Uint64()).impl,
	})
	if value.Type() != destination {
		t.Fatalf("constant type = %s, want exact destination type %s", value.Type(), destination)
	}
}
