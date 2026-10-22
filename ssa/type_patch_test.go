package ssa

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/goplus/gogen/packages"
	gossa "golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func TestFieldOutOfRangePanicsWithTypeString(t *testing.T) {
	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
	}, nil)
	typ := &aType{raw: rawType{Type: st}}

	var p Program
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic")
		}
		msg := r.(string)
		if !strings.Contains(msg, "Field: struct index out of range") {
			t.Fatalf("panic = %q, want out-of-range message", msg)
		}
	}()
	_ = p.Field(typ, 1)
}

func TestSyntheticFieldOutOfRangePanics(t *testing.T) {
	prog := NewProgram(nil)
	t.Cleanup(prog.Dispose)
	empty := types.NewInterfaceType(nil, nil)
	empty.Complete()
	for _, test := range []struct {
		name string
		typ  types.Type
		kind valueKind
		idx  int
		want string
	}{
		{name: "string", typ: types.Typ[types.String], kind: vkString, idx: 2, want: "Field: string index out of range"},
		{name: "slice", typ: types.NewSlice(types.Typ[types.Byte]), kind: vkSlice, idx: 3, want: "Field: slice index out of range"},
		{name: "interface", typ: empty, kind: vkEface, idx: 2, want: "Field: interface index out of range"},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				got := recover()
				if got == nil || got != test.want {
					t.Fatalf("panic = %v, want %q", got, test.want)
				}
			}()
			prog.Field(&aType{raw: rawType{Type: test.typ}, kind: test.kind}, test.idx)
		})
	}
}

func TestTypeStringWithPkgAndIsPkgScope(t *testing.T) {
	obj := types.NewTypeName(token.NoPos, nil, "Local", nil)
	local := types.NewNamed(obj, types.Typ[types.Int], nil)
	if got := typeStringWithPkg(local); got != "Local" {
		t.Fatalf("typeStringWithPkg(local) = %q, want %q", got, "Local")
	}

	pkg1 := types.NewPackage("example.com/p", "p")
	if !isPkgScope(pkg1.Scope(), pkg1.Scope()) {
		t.Fatalf("isPkgScope(scope, scope) should be true")
	}
	if isPkgScope(nil, pkg1.Scope()) {
		t.Fatalf("isPkgScope(nil, scope) should be false")
	}
}

func TestNamedTypeEquivalent(t *testing.T) {
	pkg1 := types.NewPackage("example.com/p", "p")
	pkg2 := types.NewPackage("example.com/p", "p")
	obj1 := types.NewTypeName(token.NoPos, pkg1, "T", nil)
	obj2 := types.NewTypeName(token.NoPos, pkg2, "T", nil)
	a := types.NewNamed(obj1, types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
	}, nil), nil)
	b := types.NewNamed(obj2, types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
	}, nil), nil)
	if !namedTypeEquivalent(a, b) {
		t.Fatalf("namedTypeEquivalent should be true for equivalent named structs")
	}

	c := types.NewNamed(types.NewTypeName(token.NoPos, pkg2, "T", nil), types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.String], false),
	}, nil), nil)
	if namedTypeEquivalent(a, c) {
		t.Fatalf("namedTypeEquivalent should be false for different underlying types")
	}
}

func TestNamedTypeEquivalentRecursiveSignature(t *testing.T) {
	pkg1 := types.NewPackage("example.com/p", "p")
	pkg2 := types.NewPackage("example.com/p", "p")
	a := types.NewNamed(types.NewTypeName(token.NoPos, pkg1, "F", nil), nil, nil)
	b := types.NewNamed(types.NewTypeName(token.NoPos, pkg2, "F", nil), nil, nil)
	a.SetUnderlying(types.NewSignatureType(nil, nil, nil,
		types.NewTuple(types.NewParam(token.NoPos, nil, "", a)),
		types.NewTuple(types.NewParam(token.NoPos, nil, "", a)),
		false))
	b.SetUnderlying(types.NewSignatureType(nil, nil, nil,
		types.NewTuple(types.NewParam(token.NoPos, nil, "", b)),
		types.NewTuple(types.NewParam(token.NoPos, nil, "", b)),
		false))
	if namedTypeEquivalent(a, b) {
		t.Fatalf("namedTypeEquivalent should be false for recursive function signatures")
	}
}

func TestNamedTypeEquivalentRejectsAnySignature(t *testing.T) {
	pkg1 := types.NewPackage("example.com/p", "p")
	pkg2 := types.NewPackage("example.com/p", "p")
	sigNamed := types.NewNamed(types.NewTypeName(token.NoPos, pkg1, "F", nil),
		types.NewSignatureType(nil, nil, nil, nil, nil, false), nil)
	structNamed := types.NewNamed(types.NewTypeName(token.NoPos, pkg2, "F", nil),
		types.NewStruct([]*types.Var{
			types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
		}, nil), nil)
	if namedTypeEquivalent(sigNamed, structNamed) {
		t.Fatalf("namedTypeEquivalent should be false when only one side is a signature")
	}
	if namedTypeEquivalent(structNamed, sigNamed) {
		t.Fatalf("namedTypeEquivalent should be false when only one side is a signature")
	}
}

func TestToLLVMFuncPtrUsesVoidPtr(t *testing.T) {
	prog := NewProgram(nil)
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	if got, want := prog.toLLVMFuncPtr(sig).String(), prog.tyVoidPtr().String(); got != want {
		t.Fatalf("toLLVMFuncPtr() = %q, want %q", got, want)
	}
}

func TestPointerTypeDoesNotExpandRecursiveNamedElement(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "recursive.go", `package p

type T18 *[10]T19
type T19 T18
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := types.NewPackage("example.com/p", "p")
	if err := types.NewChecker(&types.Config{}, fset, pkg, nil).Files([]*ast.File{f}); err != nil {
		t.Fatal(err)
	}

	prog := NewProgram(nil)
	prog.TypeSizes(types.SizesFor("gc", runtime.GOARCH))
	t18 := pkg.Scope().Lookup("T18").Type()
	if got, want := prog.Type(t18, GoDecl).ll.String(), prog.tyVoidPtr().String(); got != want {
		t.Fatalf("recursive named pointer LLVM type = %q, want %q", got, want)
	}
}

func TestGoProgramSizesUnaliasFunctionFieldStruct(t *testing.T) {
	prog := NewProgram(nil)
	sizes := prog.TypeSizes(types.SizesFor("gc", runtime.GOARCH))

	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	fnStruct := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "Fn", sig, false),
	}, nil)
	alias := types.NewAlias(types.NewTypeName(token.NoPos, nil, "FnStruct", nil), fnStruct)
	fields := []*types.Var{
		types.NewField(token.NoPos, nil, "A", alias, false),
		types.NewField(token.NoPos, nil, "B", types.Typ[types.Int], false),
	}

	want := int64(prog.PointerSize() * 2)
	if got := sizes.Sizeof(alias); got != want {
		t.Fatalf("Sizeof(alias to func-field struct) = %d, want %d", got, want)
	}
	if got := sizes.Offsetsof(fields)[1]; got != want {
		t.Fatalf("Offsetsof(field after alias to func-field struct) = %d, want %d", got, want)
	}

	// Query unsafe while checking declarations, before any runtime or LLVM
	// named types exist. Array bounds must already use the final storage size.
	const source = `package p
import "unsafe"
type Native struct { X float64; Z struct{} }
type GoInner struct { X float64; Z struct{} }
type NativeGo struct { V GoInner; Z struct{} }
type GoNative struct { V Native; F func(); Z struct{} }
type Recursive struct { Next *Recursive; F func(Recursive) }
// TODO(#2907): restore cdecl function layouts after representation is consistent.
// type Callback func()
// type CallbackAlias = Callback
// type WithCallback struct { F CallbackAlias; B byte }
// type GoFunction struct { F func() }
// type NativeGoFunction struct { V GoFunction }
type Buffer [unsafe.Sizeof(GoNative{})]byte
type RecursiveBuffer [unsafe.Sizeof(Recursive{})]byte
const NativeSize = unsafe.Sizeof(Native{})
const NestedOffset = unsafe.Offsetof(GoNative{}.F)
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "layout.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	pkg := types.NewPackage("example.com/earlylayout", "p")
	for _, name := range []string{"Native", "NativeGo" /*, "Callback", "NativeGoFunction" */} {
		prog.SetTypeCallConv(pkg.Path()+"."+name, CDecl)
	}
	conf := &types.Config{Sizes: sizes, Importer: importer.Default()}
	if err := types.NewChecker(conf, fset, pkg, nil).Files([]*ast.File{f}); err != nil {
		t.Fatal(err)
	}
	if prog.rt != nil || len(prog.named) != 0 {
		t.Fatal("early layout loaded runtime or cached LLVM named types")
	}
	for _, name := range []string{"Native", "GoInner", "NativeGo", "GoNative", "Recursive" /*, "NativeGoFunction", "WithCallback" */} {
		typ := pkg.Scope().Lookup(name).Type()
		storage := prog.Type(typ, GoDecl)
		if size, alignment := sizes.Sizeof(typ), sizes.Alignof(typ); size != int64(prog.SizeOf(storage)) || alignment != int64(prog.AlignOf(storage)) {
			t.Errorf("%s early size/alignment=%d/%d, LLVM=%d/%d", name, size, alignment, prog.SizeOf(storage), prog.AlignOf(storage))
		}
	}
	for _, test := range []struct{ array, element string }{{"Buffer", "GoNative"}, {"RecursiveBuffer", "Recursive"}} {
		bound := pkg.Scope().Lookup(test.array).Type().Underlying().(*types.Array).Len()
		storage := prog.Type(pkg.Scope().Lookup(test.element).Type(), GoDecl)
		if bound != int64(prog.SizeOf(storage)) {
			t.Errorf("%s bound=%d, LLVM size=%d", test.array, bound, prog.SizeOf(storage))
		}
	}
	if size, _ := constant.Int64Val(pkg.Scope().Lookup("NativeSize").(*types.Const).Val()); size != 8 {
		t.Errorf("unsafe.Sizeof(Native{})=%d, want 8", size)
	}
	if offset, _ := constant.Int64Val(pkg.Scope().Lookup("NestedOffset").(*types.Const).Val()); offset != 8 {
		t.Errorf("unsafe.Offsetof(GoNative{}.F)=%d, want 8", offset)
	}

	// Exercise actual unsafe expressions in the type checker. Every storable
	// Go type category appears as a value and across three struct levels. Scalar
	// and data aggregates retain all eight Go/cdecl owner combinations; function
	// values retain Go owners only until #2907 is resolved. LLVM storage is the
	// independent observation.
	var matrix strings.Builder
	matrix.WriteString(`package p
import "unsafe"
type Word uint64
type WordAlias = Word
type Function func(int) int
type FunctionAlias = Function
// type Callback func(int) int
// type CallbackAlias = Callback
// type CallbackAlias2 = CallbackAlias
// type DerivedCallback Callback
// type NativeAlias = Function
type RecursiveFunction func(RecursiveFunction) RecursiveFunction
type GenericFunction[T any] func(T) T
type GenericFunctionAlias[T any] = GenericFunction[T]
// type NativeGenericFunction[T any] func(T) T
// type NativeVariadic func(marker int32, __llgo_va_list ...any) int32
// type FunctionPair struct { Go Function; C Callback }
// type CallbackPair struct { C Callback; Go Function }
// type NativeFunctionPair FunctionPair
// type NativeCallbackPair CallbackPair
type Node struct { Next *Node; Visit func(*Node) }
type Empty struct{}
type GoRecord struct { X float64; End struct{} }
type NativeRecord struct { X float64; End struct{} }
type Generic[T any] struct { V T; End [0]T }
`)
	leaves := []struct{ name, typ string }{
		{"bool", "bool"},
		{"int", "int"},
		{"int8", "int8"},
		{"int16", "int16"},
		{"int32", "int32"},
		{"int64", "int64"},
		{"uint", "uint"},
		{"uint8", "uint8"},
		{"uint16", "uint16"},
		{"uint32", "uint32"},
		{"uint64", "uint64"},
		{"uintptr", "uintptr"},
		{"byte", "byte"},
		{"rune", "rune"},
		{"float32", "float32"},
		{"float64", "float64"},
		{"complex64", "complex64"},
		{"complex128", "complex128"},
		{"string", "string"},
		{"pointer", "*byte"},
		{"unsafe-pointer", "unsafe.Pointer"},
		{"recursive-pointer", "*Node"},
		{"slice", "[]int"},
		{"map", "map[string]int"},
		{"channel", "chan int"},
		{"receive-channel", "<-chan int"},
		{"send-channel", "chan<- int"},
		{"empty-interface", "any"},
		{"interface", "error"},
		{"named-scalar", "Word"},
		{"scalar-alias", "WordAlias"},
		{"empty-struct", "Empty"},
		{"anonymous-struct", "struct { I int32; U uint64 }"},
		{"blank-field", "struct { _ uint64; B byte; End struct{} }"},
		{"embedded-fields", "struct { Word; *Node }"},
		{"go-struct", "GoRecord"},
		{"native-struct", "NativeRecord"},
		{"array-zero", "[0]uint64"},
		{"array-one", "[1]uint64"},
		{"array-two", "[2]uint64"},
		{"array-matrix", "[2][3]uint64"},
		{"array-go-struct", "[2]GoRecord"},
		{"array-native-struct", "[2]NativeRecord"},
		{"array-zero-native", "[0]NativeRecord"},
		{"generic-scalar", "Generic[uint64]"},
		{"generic-native", "Generic[NativeRecord]"},
	}
	nativeLeaves := len(leaves)
	leaves = append(leaves, []struct{ name, typ string }{
		{"function", "func()"},
		{"variadic-function", "func(int, ...any) (int, error)"},
		{"named-function", "Function"},
		{"function-alias", "FunctionAlias"},
		{"array-functions", "[2]Function"},
		{"generic-function", "Generic[Function]"},
		{"recursive-struct", "Node"},
		{"recursive-function", "RecursiveFunction"},
		{"generic-function-type", "GenericFunction[int]"},
		{"generic-function-alias", "GenericFunctionAlias[int]"},
		{"parenthesized-function", "(func(int) int)"},
		{"named-parameters", "func(x int) (result int)"},
		{"blank-parameters", "func(_ int) (_ int)"},
		{"multiple-results", "func() (int, string, error)"},
		{"returns-function", "func() Function"},
		{"accepts-function", "func(Function)"},
		{"function-pointer", "*Function"},
		{"function-slice", "[]Function"},
		{"function-map-value", "map[string]Function"},
		{"function-channel", "chan Function"},
		{"function-array-zero", "[0]Function"},
		{"function-array-one", "[1]Function"},
		{"function-first", "struct { F Function; B byte }"},
		{"function-last", "struct { B byte; F Function }"},
		{"function-zero-tail", "struct { F Function; Z struct{} }"},
		{"blank-function", "struct { B byte; _ Function; Tail byte }"},
		{"embedded-function", "struct { Function; B byte }"},
		// TODO(#2907): cdecl function assignments and mixed containers.
		// {"native-callback", "Callback"},
		// {"native-callback-alias", "CallbackAlias"},
		// {"array-callbacks", "[2]Callback"},
		// {"callback-alias-chain", "CallbackAlias2"},
		// {"derived-callback", "DerivedCallback"},
		// {"native-function-alias", "NativeAlias"},
		// {"native-generic-function", "NativeGenericFunction[int]"},
		// {"native-variadic", "NativeVariadic"},
		// {"callback-pointer", "*Callback"},
		// {"callback-slice", "[]Callback"},
		// {"callback-map-value", "map[string]Callback"},
		// {"callback-channel", "chan Callback"},
		// {"callback-array-zero", "[0]Callback"},
		// {"callback-array-one", "[1]Callback"},
		// {"callback-array-matrix", "[2][3]Callback"},
		// {"go-callback-pair", "FunctionPair"},
		// {"callback-go-pair", "CallbackPair"},
		// {"native-go-callback-pair", "NativeFunctionPair"},
		// {"native-callback-go-pair", "NativeCallbackPair"},
		// {"pair-array", "[2]FunctionPair"},
		// {"native-pair-array", "[2]NativeFunctionPair"},
		// {"generic-callback", "Generic[Callback]"},
		// {"generic-mixed-functions", "Generic[FunctionPair]"},
		// {"callback-last", "struct { B byte; F Callback }"},
		// {"callback-zero-tail", "struct { F Callback; Z struct{} }"},
		// {"embedded-callback", "struct { Callback; B byte }"},
	}...)
	type observation struct {
		name, label string
		fields      bool
	}
	var observations []observation
	nativeNames := []string{"NativeRecord" /*, "Callback", "NativeAlias", "NativeGenericFunction", "NativeVariadic", "NativeFunctionPair", "NativeCallbackPair" */}
	add := func(name, label, body string, native, fields bool) {
		fmt.Fprintf(&matrix, "type %s %s\n", name, body)
		fmt.Fprintf(&matrix, "const %sSize = unsafe.Sizeof(*new(%s))\n", name, name)
		fmt.Fprintf(&matrix, "const %sAlign = unsafe.Alignof(*new(%s))\n", name, name)
		fmt.Fprintf(&matrix, "type %sBound [unsafe.Sizeof(*new(%s))]byte\n", name, name)
		if fields {
			for _, field := range []string{"B", "V", "Tail", "End"} {
				fmt.Fprintf(&matrix, "const %s%s = unsafe.Offsetof((*new(%s)).%s)\n", name, field, name, field)
			}
		}
		if native {
			nativeNames = append(nativeNames, name)
		}
		observations = append(observations, observation{name, label, fields})
	}
	for i, leaf := range leaves {
		add(fmt.Sprintf("Value%d", i), leaf.name, "= "+leaf.typ, false, false)
		masks := []int{0, 1, 2, 3, 4, 5, 6, 7}
		if i >= nativeLeaves {
			// TODO(#2907): func fields in cdecl containers also use the disputed
			// conversion, e.g. Native{G struct{F func()}}. Keep Go-only nesting.
			masks = []int{0 /*, 1, 2, 3, 4, 5, 6, 7 */}
		}
		for _, mask := range masks {
			child := leaf.typ
			for level := 0; level < 3; level++ {
				name := fmt.Sprintf("S%d_%d_%d", i, mask, level)
				add(name, fmt.Sprintf("%s/mask=%03b/level=%d", leaf.name, mask, level+1),
					"struct { B byte; V "+child+"; Tail byte; End struct{} }", mask&(1<<level) != 0, true)
				child = name
			}
		}
	}
	Initialize(InitAllTargets | InitAllTargetInfos | InitAllTargetMCs)
	runtimeTypes, err := importer.For("source", nil).Import(PkgRuntime)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []*Target{
		{GOOS: "linux", GOARCH: "amd64"}, {GOOS: "linux", GOARCH: "arm64"},
		{GOOS: "darwin", GOARCH: "amd64"}, {GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "windows", GOARCH: "amd64"}, {GOOS: "windows", GOARCH: "arm64"},
		{GOOS: "linux", GOARCH: "386"}, {GOOS: "linux", GOARCH: "arm"},
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-pc-windows-msvc"},
		{GOOS: "windows", GOARCH: "386", LLVMTarget: "i686-w64-windows-gnu"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j32", LLVMTarget: "wasm32-unknown-emscripten"},
		{GOOS: "wasip1", GOARCH: "wasm", WasmProfile: "w32", LLVMTarget: "wasm32-unknown-wasip1"},
		{GOOS: "js", GOARCH: "wasm", WasmProfile: "j64", LLVMTarget: "wasm64-unknown-emscripten"},
	} {
		// A fresh Program per order exposes reuse of a Go representation in a
		// later native use (and the reverse), rather than inheriting one cache.
		for _, reverse := range []bool{false, true} {
			t.Run(fmt.Sprintf("unsafe/%s/%s/%s/reverse=%t", target.GOOS, target.GOARCH, target.LLVMTarget, reverse), func(t *testing.T) {
				p := NewProgram(target)
				defer p.Dispose()
				baseSizes := types.SizesFor("gc", target.GOARCH)
				early := p.TypeSizes(baseSizes)
				const path = "example.com/unsafematrix"
				for _, name := range nativeNames {
					p.SetTypeCallConv(path+"."+name, CDecl)
				}
				fset := token.NewFileSet()
				file, err := parser.ParseFile(fset, "matrix.go", matrix.String(), 0)
				if err != nil {
					t.Fatal(err)
				}
				config := &types.Config{Sizes: early, Importer: importer.Default()}
				pkg, err := config.Check(path, fset, []*ast.File{file}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if p.rt != nil || len(p.named) != 0 {
					t.Fatal("unsafe constants loaded runtime or cached LLVM named types")
				}
				p.SetRuntime(runtimeTypes)
				value := func(name string) int64 {
					n, ok := constant.Int64Val(pkg.Scope().Lookup(name).(*types.Const).Val())
					if !ok {
						t.Fatalf("%s is not an int64 constant", name)
					}
					return n
				}
				for i := range observations {
					if reverse {
						i = len(observations) - 1 - i
					}
					check := observations[i]
					typ := pkg.Scope().Lookup(check.name).Type()
					storage := p.Type(typ, GoDecl)
					size, alignment := value(check.name+"Size"), value(check.name+"Align")
					wantSize, wantAlignment := int64(p.SizeOf(storage)), int64(p.AlignOf(storage))
					if _, aggregate := arrayBaseType(typ).Underlying().(*types.Struct); !aggregate {
						// Go may permit weaker scalar alignment than an LLVM allocation
						// provides, e.g. uint64 on Windows/386. Test its language rule.
						wantAlignment = baseSizes.Alignof(typ)
					}
					bound := pkg.Scope().Lookup(check.name + "Bound").Type().Underlying().(*types.Array).Len()
					var gotOffsets, wantOffsets []int64
					if check.fields {
						for i, field := range []string{"B", "V", "Tail", "End"} {
							gotOffsets = append(gotOffsets, value(check.name+field))
							wantOffsets = append(wantOffsets, int64(p.OffsetOf(storage, i)))
						}
					}
					if size != wantSize || alignment != wantAlignment || bound != wantSize || !slices.Equal(gotOffsets, wantOffsets) {
						t.Errorf("%s: unsafe size/align/offsets=%d/%d/%v bound=%d; storage=%d/%d/%v", check.label, size, alignment, gotOffsets, bound, wantSize, wantAlignment, wantOffsets)
					}
				}
				t.Logf("checked %d types across %d leaf categories", len(observations), len(leaves))
			})
		}
	}
}

func TestNamedStructLayoutEquivalent(t *testing.T) {
	prog := NewProgram(nil)
	prog.TypeSizes(types.SizesFor("gc", runtime.GOARCH))

	pkg := types.NewPackage("example.com/p", "p")
	s1 := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
	}, nil)
	n1 := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "T", nil), s1, nil)
	llStruct, _ := prog.toLLVMStruct(s1)
	existing := &aType{
		ll:  llStruct,
		raw: rawType{Type: n1},
	}

	same := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "T", nil), types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
	}, nil), nil)
	if !prog.namedStructLayoutEquivalent(existing, same) {
		t.Fatalf("namedStructLayoutEquivalent should be true for same layout")
	}

	diff := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "T", nil), types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "A", types.Typ[types.Int], false),
		types.NewField(token.NoPos, nil, "B", types.Typ[types.Int], false),
	}, nil), nil)
	if prog.namedStructLayoutEquivalent(existing, diff) {
		t.Fatalf("namedStructLayoutEquivalent should be false for different layout")
	}
}

func buildGoSSAPackageForInternalTypes(t *testing.T, src string) *gossa.Package {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "internaltypes.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	files := []*ast.File{f}
	pkg := types.NewPackage("foo", "foo")
	imp := packages.NewImporter(fset)
	mode := gossa.SanityCheckFunctions | gossa.InstantiateGenerics
	ssapkg, _, err := ssautil.BuildPackage(&types.Config{Importer: imp}, fset, pkg, files, mode)
	if err != nil {
		t.Fatal(err)
	}
	return ssapkg
}

// x/tools represents iterator and defer-stack handles as named unsafe.Pointer
// types. They must use the ordinary named-type path and retain pointer storage.
func TestGoSSAInternalPointerTypes(t *testing.T) {
	ssapkg := buildGoSSAPackageForInternalTypes(t, `package foo
func seq(yield func(int) bool) { _ = yield(1) }
func f(m map[string]int) {
	for v := range seq {
		defer func() { _ = v }()
	}
	for k, v := range m { _, _ = k, v }
}
`)
	internalTypes := make(map[string]types.Type)
	for fn := range ssautil.AllFunctions(ssapkg.Prog) {
		for _, block := range fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *gossa.Defer:
					if instr.DeferStack != nil {
						internalTypes["defer stack"] = instr.DeferStack.Type()
					}
				case *gossa.Range:
					internalTypes["range iterator"] = instr.Type()
				}
			}
		}
	}
	prog := NewProgram(nil)
	t.Cleanup(prog.Dispose)
	for _, name := range []string{"defer stack", "range iterator"} {
		t.Run(name, func(t *testing.T) {
			typ, ok := internalTypes[name].(*types.Named)
			if !ok {
				t.Fatalf("expected named handle type, got %T", internalTypes[name])
			}
			if typ.Underlying() != types.Typ[types.UnsafePointer] {
				t.Fatalf("handle underlying type = %v, want unsafe.Pointer", typ.Underlying())
			}
			if raw, converted := prog.gocvt.cvtType(typ); converted || raw != typ {
				t.Fatalf("cvtType = (%v, %v), want unchanged named handle", raw, converted)
			}
			got := prog.Type(typ, GoDecl)
			if got.RawType() != typ {
				t.Fatalf("raw type = %v, want %v", got.RawType(), typ)
			}
			if got.ll != prog.tyVoidPtr() {
				t.Fatalf("LLVM type = %v, want pointer", got.ll)
			}
			if got := prog.SizeOf(got); got != prog.SizeOf(prog.Type(types.Typ[types.UnsafePointer], GoDecl)) {
				t.Fatalf("handle size = %d, want pointer size", got)
			}
		})
	}
}

func TestCvtStructPreservesTags(t *testing.T) {
	sig := types.NewSignatureType(nil, nil, nil, nil, nil, false)
	st := types.NewStruct([]*types.Var{
		types.NewField(token.NoPos, nil, "Fn", sig, false),
	}, []string{`protobuf:"bytes,1,opt,name=fn" json:"fn,omitempty"`})

	raw, cvt := newGoTypes().cvtType(st)
	if !cvt {
		t.Fatalf("cvtType did not convert function field")
	}
	got := raw.(*types.Struct).Tag(0)
	want := `protobuf:"bytes,1,opt,name=fn" json:"fn,omitempty"`
	if got != want {
		t.Fatalf("converted struct tag = %q, want %q", got, want)
	}
}
