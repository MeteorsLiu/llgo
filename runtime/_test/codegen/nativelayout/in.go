// LITTEST linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/386 windows/386
package main

import (
	"reflect"
	"slices"
	"strings"
	"unsafe"

	"github.com/xgo-dev/llgo/runtime/abi"
	// TODO(#2907): restore with the cdecl function cases below.
	// c "github.com/xgo-dev/llgo/runtime/internal/clite"
)

//llgo:cdecl
type record struct {
	Number float64
	End    struct{}
}

type envelope struct {
	Record record
	Stamp  uint64
}

//llgo:cdecl
type nativePair struct {
	Small int32
	Large uint64
}

type goPair struct {
	Small int32
	Large uint64
}

type pointerEnvelope struct {
	Record record
	P      *byte
	Tail   uint64
}

//llgo:cdecl
type pointerRecord struct {
	P   *byte
	End struct{}
}

type pointerRecords [2]pointerRecord
type emptyPointerRecords [0]pointerRecord

//llgo:cdecl
type alignedPointerRecord struct {
	Small int32
	Large uint64
	P     *byte
}

type alignedPointerRecords [2]alignedPointerRecord

type unsafeWord uint64
type unsafeWordAlias = unsafeWord
type unsafeFunction func(int) int
type unsafeFunctionAlias = unsafeFunction

// TODO(#2907): cdecl function representation is outside this layout fix.
// //llgo:cdecl
// type unsafeCallback func(int) int
// type unsafeCallbackAlias = unsafeCallback

// //llgo:cdecl
// type unsafeNativeNullary func() int32

// type unsafeNativeNullaryAlias = unsafeNativeNullary
// type unsafeNativeNullaryAlias2 = unsafeNativeNullaryAlias

// //llgo:cdecl
// type unsafeNativeMethod func() int

// type unsafeCallableMix struct {
// Prefix byte
// Go     func()
// Method func() int
// C      unsafeNativeNullary
// End    struct{}
// }

// //llgo:cdecl
// type unsafeNativeCallableMix unsafeCallableMix

type unsafeNode struct {
	Next  *unsafeNode
	Visit func(*unsafeNode)
}

type unsafeBox[T any] struct {
	Value T
	End   [0]T
}

// The two owners contain the same source fields; their platform alignment
// may differ. Function-bearing fields are deferred under #2907 because this
// container is also used as cdecl. Go-only callable checks remain below.
type unsafeTypes struct {
	Bool             bool
	Int              int
	Int8             int8
	Int16            int16
	Int32            int32
	Int64            int64
	Uint             uint
	Uint8            uint8
	Uint16           uint16
	Uint32           uint32
	Uint64           uint64
	Uintptr          uintptr
	Byte             byte
	Rune             rune
	Float32          float32
	Float64          float64
	Complex64        complex64
	Complex128       complex128
	String           string
	Pointer          *byte
	UnsafePointer    unsafe.Pointer
	RecursivePointer *unsafeNode
	Slice            []int
	Map              map[string]int
	Channel          chan int
	ReceiveChannel   <-chan int
	SendChannel      chan<- int
	Any              any
	Interface        error
	// Function         func()
	// Variadic         func(int, ...any) (int, error)
	// NamedFunction    unsafeFunction
	// FunctionAlias    unsafeFunctionAlias
	// Callback         unsafeCallback
	// CallbackAlias    unsafeCallbackAlias
	Word      unsafeWord
	WordAlias unsafeWordAlias
	Empty     struct{}
	Anonymous struct {
		I int32
		U uint64
	}
	Blank struct {
		_   uint64
		B   byte
		End struct{}
	}
	Embedded struct {
		unsafeWord
		*unsafeNode
	}
	GoRecord     goPair
	NativeRecord record
	Array0       [0]uint64
	Array1       [1]uint64
	Array2       [2]uint64
	Matrix       [2][3]uint64
	GoArray      [2]goPair
	NativeArray  [2]record
	NativeArray0 [0]record
	// Functions       [2]unsafeFunction
	// Callbacks       [2]unsafeCallback
	GenericScalar unsafeBox[uint64]
	GenericNative unsafeBox[record]
	// GenericFunction unsafeBox[unsafeFunction]
	// Node            unsafeNode
	End struct{}
}

//llgo:cdecl
type unsafeNativeTypes unsafeTypes

type unsafeNested0_1 struct {
	Prefix byte
	V      unsafeTypes
	End    struct{}
}

type unsafeNested0_2 struct {
	Prefix byte
	V      unsafeNested0_1
	End    struct{}
}

type unsafeNested1_1 struct {
	Prefix byte
	V      unsafeNativeTypes
	End    struct{}
}

type unsafeNested1_2 struct {
	Prefix byte
	V      unsafeNested1_1
	End    struct{}
}

//llgo:cdecl
type unsafeNested2_1 struct {
	Prefix byte
	V      unsafeTypes
	End    struct{}
}

type unsafeNested2_2 struct {
	Prefix byte
	V      unsafeNested2_1
	End    struct{}
}

//llgo:cdecl
type unsafeNested3_1 struct {
	Prefix byte
	V      unsafeNativeTypes
	End    struct{}
}

type unsafeNested3_2 struct {
	Prefix byte
	V      unsafeNested3_1
	End    struct{}
}

type unsafeNested4_1 struct {
	Prefix byte
	V      unsafeTypes
	End    struct{}
}

//llgo:cdecl
type unsafeNested4_2 struct {
	Prefix byte
	V      unsafeNested4_1
	End    struct{}
}

type unsafeNested5_1 struct {
	Prefix byte
	V      unsafeNativeTypes
	End    struct{}
}

//llgo:cdecl
type unsafeNested5_2 struct {
	Prefix byte
	V      unsafeNested5_1
	End    struct{}
}

//llgo:cdecl
type unsafeNested6_1 struct {
	Prefix byte
	V      unsafeTypes
	End    struct{}
}

//llgo:cdecl
type unsafeNested6_2 struct {
	Prefix byte
	V      unsafeNested6_1
	End    struct{}
}

//llgo:cdecl
type unsafeNested7_1 struct {
	Prefix byte
	V      unsafeNativeTypes
	End    struct{}
}

//llgo:cdecl
type unsafeNested7_2 struct {
	Prefix byte
	V      unsafeNested7_1
	End    struct{}
}

// Metadata keeps the native owner instead of reconstructing a Go struct.
// AMD64-DAG: @_llgo_main.record = {{.*}}i64 8, i64 0,
// ARM64-DAG: @_llgo_main.record = {{.*}}i64 8, i64 0,
// AMD64-DAG: @_llgo_main.envelope = {{.*}}i64 16, i64 0,
// ARM64-DAG: @_llgo_main.envelope = {{.*}}i64 16, i64 0,
// WINDOWS-386-DAG: @_llgo_main.nativePair = {{.*}}i32 16, i32 0, i32 {{[^,]+}}, i8 5, i8 8, i8 8,
// WINDOWS-386-DAG: @"_llgo_main.nativePair$fields" = {{.*}}ptr @_llgo_uint64, i32 8,
// WINDOWS-386-DAG: @_llgo_main.goPair = {{.*}}i32 12, i32 0, i32 {{[^,]+}}, i8 13, i8 4, i8 4,
// AMD64-DAG: @_llgo_main.pointerEnvelope = {{.*}}{ i64 24, i64 16,
// ARM64-DAG: @_llgo_main.pointerEnvelope = {{.*}}{ i64 24, i64 16,
// AMD64-DAG: @_llgo_main.pointerRecords = {{.*}}{ i64 16, i64 16,
// ARM64-DAG: @_llgo_main.pointerRecords = {{.*}}{ i64 16, i64 16,
// WINDOWS-386-DAG: @_llgo_main.pointerRecord = {{.*}}{ i32 4, i32 4,
// WINDOWS-386-DAG: @_llgo_main.pointerRecords = {{.*}}{ i32 8, i32 8,
// WINDOWS-386-DAG: @_llgo_main.alignedPointerRecord = {{.*}}{ i32 24, i32 20,
// WINDOWS-386-DAG: @_llgo_main.alignedPointerRecords = {{.*}}{ i32 48, i32 44,
// WINDOWS-386-DAG: @_llgo_main.emptyPointerRecords = {{.*}}{ i32 0, i32 0,
// CHECK-LABEL: define void @main.main()
func main() {
	values := [3]record{{Number: 10}, {Number: 20}, {Number: 30}}
	step := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
	println("record size", reflect.TypeOf(values[0]).Size() == step)
	println("record array", reflect.TypeOf(values).Size() == 3*step)
	v := reflect.ValueOf(&values).Elem().Index(1).Field(0)
	println("record read", v.Float() == 20)
	v.SetFloat(25)
	println("record write", values[0].Number == 10 && values[1].Number == 25 && values[2].Number == 30)
	count := reflect.Copy(reflect.ValueOf(&values).Elem(), reflect.ValueOf([1]record{{Number: 40}}))
	println("record copy", count == 1 && values[0].Number == 40 && values[1].Number == 25 && values[2].Number == 30)

	packets := [2]envelope{{Record: record{Number: 7}, Stamp: 37}, {Record: record{Number: 9}, Stamp: 99}}
	packetSize := uintptr(unsafe.Pointer(&packets[1])) - uintptr(unsafe.Pointer(&packets[0]))
	stampOffset := uintptr(unsafe.Pointer(&packets[0].Stamp)) - uintptr(unsafe.Pointer(&packets[0]))
	packetType := reflect.TypeOf(packets[0])
	storage := [unsafe.Sizeof(envelope{})]byte{}
	println("unsafe layout", uintptr(len(storage)) == packetSize &&
		unsafe.Sizeof(record{}) == step && unsafe.Sizeof(values) == 3*step &&
		unsafe.Offsetof(envelope{}.Stamp) == stampOffset &&
		unsafe.Alignof(envelope{}) == uintptr(packetType.Align()))
	println("envelope size", packetType.Size() == packetSize)
	println("stamp offset", packetType.Field(1).Offset == stampOffset)
	println("envelope array", reflect.TypeOf(packets).Size() == 2*packetSize)
	field := reflect.ValueOf(&packets[0]).Elem().Field(1)
	println("stamp read", field.Uint() == 37)
	field.SetUint(41)
	println("stamp write", packets[0].Stamp == 41 && packets[1].Record.Number == 9)

	roundTrip := func(p envelope, marker uint64) envelope {
		if marker != 53 {
			panic("argument following envelope was shifted")
		}
		return p
	}
	result := reflect.ValueOf(roundTrip).Call([]reflect.Value{reflect.ValueOf(packets[0]), reflect.ValueOf(uint64(53))})
	println("call", result[0].Interface().(envelope) == packets[0])
	callback := reflect.MakeFunc(reflect.TypeOf(roundTrip), func(args []reflect.Value) []reflect.Value {
		if args[0].Interface().(envelope) != packets[0] || args[1].Uint() != 53 {
			panic("MakeFunc decoded the wrong envelope")
		}
		return args[:1]
	}).Interface().(func(envelope, uint64) envelope)
	println("makefunc", callback(packets[0], 53) == packets[0])

	var native nativePair
	var ordinary goPair
	nativeOffset := uintptr(unsafe.Pointer(&native.Large)) - uintptr(unsafe.Pointer(&native))
	goOffset := uintptr(unsafe.Pointer(&ordinary.Large)) - uintptr(unsafe.Pointer(&ordinary))
	println("separate offsets", reflect.TypeOf(native).Field(1).Offset == nativeOffset && reflect.TypeOf(ordinary).Field(1).Offset == goOffset)
	left, right := [8]nativePair{}, [8]nativePair{}
	for i := uintptr(4); i < nativeOffset; i++ {
		*(*byte)(unsafe.Add(unsafe.Pointer(&left[0]), i)) = 0x11
		*(*byte)(unsafe.Add(unsafe.Pointer(&right[0]), i)) = 0x22
	}
	println("array equality", left == right)

	var pointers pointerRecords
	var nested pointerEnvelope
	var aligned alignedPointerRecords
	pointerWidth := uintptr(unsafe.Pointer(&pointers[1])) - uintptr(unsafe.Pointer(&pointers[0]))
	nestedPrefix := uintptr(unsafe.Pointer(&nested.P)) - uintptr(unsafe.Pointer(&nested)) + pointerWidth
	alignedPrefix := uintptr(unsafe.Pointer(&aligned[0].P)) - uintptr(unsafe.Pointer(&aligned[0])) + pointerWidth
	arrayPrefix := uintptr(unsafe.Pointer(&aligned[1].P)) - uintptr(unsafe.Pointer(&aligned)) + pointerWidth
	prefixes := []struct {
		value any
		want  uintptr
	}{
		{pointerRecord{}, pointerWidth},
		{pointers, 2 * pointerWidth},
		{nested, nestedPrefix},
		{alignedPointerRecord{}, alignedPrefix},
		{aligned, arrayPrefix},
		{emptyPointerRecords{}, 0},
		{envelope{}, 0},
	}
	for _, check := range prefixes {
		typ := (*abi.Type)((*[2]unsafe.Pointer)(unsafe.Pointer(&check.value))[0])
		println("pointer prefix", typ.PtrBytes == check.want)
	}

	// Compare unsafe constants with addresses in real objects. No reflection
	// size or field metadata is used as the expected value.
	checkUnsafe := func(name string, got, want uintptr) bool {
		if got != want {
			println("unsafe mismatch", name, got, want)
			return false
		}
		return true
	}
	goOK, nativeOK, nestedOK := true, true, true
	var goValues [2]unsafeTypes
	var goValuesAlignment struct {
		Prefix byte
		Value  unsafeTypes
	}
	goValuesBase := uintptr(unsafe.Pointer(&goValues[0]))
	goOK = checkUnsafe("unsafeTypes.Sizeof", unsafe.Sizeof(goValues[0]), uintptr(unsafe.Pointer(&goValues[1]))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Alignof", unsafe.Alignof(goValues[0]), uintptr(unsafe.Pointer(&goValuesAlignment.Value))-uintptr(unsafe.Pointer(&goValuesAlignment))) && goOK
	var goValuesBound [unsafe.Sizeof(goValues[0])]byte
	goOK = checkUnsafe("unsafeTypes.array-bound", uintptr(len(goValuesBound)), uintptr(unsafe.Pointer(&goValues[1]))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Bool", unsafe.Offsetof(goValues[0].Bool), uintptr(unsafe.Pointer(&goValues[0].Bool))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Int", unsafe.Offsetof(goValues[0].Int), uintptr(unsafe.Pointer(&goValues[0].Int))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Int8", unsafe.Offsetof(goValues[0].Int8), uintptr(unsafe.Pointer(&goValues[0].Int8))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Int16", unsafe.Offsetof(goValues[0].Int16), uintptr(unsafe.Pointer(&goValues[0].Int16))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Int32", unsafe.Offsetof(goValues[0].Int32), uintptr(unsafe.Pointer(&goValues[0].Int32))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Int64", unsafe.Offsetof(goValues[0].Int64), uintptr(unsafe.Pointer(&goValues[0].Int64))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uint", unsafe.Offsetof(goValues[0].Uint), uintptr(unsafe.Pointer(&goValues[0].Uint))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uint8", unsafe.Offsetof(goValues[0].Uint8), uintptr(unsafe.Pointer(&goValues[0].Uint8))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uint16", unsafe.Offsetof(goValues[0].Uint16), uintptr(unsafe.Pointer(&goValues[0].Uint16))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uint32", unsafe.Offsetof(goValues[0].Uint32), uintptr(unsafe.Pointer(&goValues[0].Uint32))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uint64", unsafe.Offsetof(goValues[0].Uint64), uintptr(unsafe.Pointer(&goValues[0].Uint64))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Uintptr", unsafe.Offsetof(goValues[0].Uintptr), uintptr(unsafe.Pointer(&goValues[0].Uintptr))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Byte", unsafe.Offsetof(goValues[0].Byte), uintptr(unsafe.Pointer(&goValues[0].Byte))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Rune", unsafe.Offsetof(goValues[0].Rune), uintptr(unsafe.Pointer(&goValues[0].Rune))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Float32", unsafe.Offsetof(goValues[0].Float32), uintptr(unsafe.Pointer(&goValues[0].Float32))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Float64", unsafe.Offsetof(goValues[0].Float64), uintptr(unsafe.Pointer(&goValues[0].Float64))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Complex64", unsafe.Offsetof(goValues[0].Complex64), uintptr(unsafe.Pointer(&goValues[0].Complex64))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Complex128", unsafe.Offsetof(goValues[0].Complex128), uintptr(unsafe.Pointer(&goValues[0].Complex128))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.String", unsafe.Offsetof(goValues[0].String), uintptr(unsafe.Pointer(&goValues[0].String))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Pointer", unsafe.Offsetof(goValues[0].Pointer), uintptr(unsafe.Pointer(&goValues[0].Pointer))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.UnsafePointer", unsafe.Offsetof(goValues[0].UnsafePointer), uintptr(unsafe.Pointer(&goValues[0].UnsafePointer))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.RecursivePointer", unsafe.Offsetof(goValues[0].RecursivePointer), uintptr(unsafe.Pointer(&goValues[0].RecursivePointer))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Slice", unsafe.Offsetof(goValues[0].Slice), uintptr(unsafe.Pointer(&goValues[0].Slice))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Map", unsafe.Offsetof(goValues[0].Map), uintptr(unsafe.Pointer(&goValues[0].Map))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Channel", unsafe.Offsetof(goValues[0].Channel), uintptr(unsafe.Pointer(&goValues[0].Channel))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.ReceiveChannel", unsafe.Offsetof(goValues[0].ReceiveChannel), uintptr(unsafe.Pointer(&goValues[0].ReceiveChannel))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.SendChannel", unsafe.Offsetof(goValues[0].SendChannel), uintptr(unsafe.Pointer(&goValues[0].SendChannel))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Any", unsafe.Offsetof(goValues[0].Any), uintptr(unsafe.Pointer(&goValues[0].Any))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Interface", unsafe.Offsetof(goValues[0].Interface), uintptr(unsafe.Pointer(&goValues[0].Interface))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Function", unsafe.Offsetof(goValues[0].Function), uintptr(unsafe.Pointer(&goValues[0].Function))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Variadic", unsafe.Offsetof(goValues[0].Variadic), uintptr(unsafe.Pointer(&goValues[0].Variadic))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.NamedFunction", unsafe.Offsetof(goValues[0].NamedFunction), uintptr(unsafe.Pointer(&goValues[0].NamedFunction))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.FunctionAlias", unsafe.Offsetof(goValues[0].FunctionAlias), uintptr(unsafe.Pointer(&goValues[0].FunctionAlias))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Callback", unsafe.Offsetof(goValues[0].Callback), uintptr(unsafe.Pointer(&goValues[0].Callback))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.CallbackAlias", unsafe.Offsetof(goValues[0].CallbackAlias), uintptr(unsafe.Pointer(&goValues[0].CallbackAlias))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Word", unsafe.Offsetof(goValues[0].Word), uintptr(unsafe.Pointer(&goValues[0].Word))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.WordAlias", unsafe.Offsetof(goValues[0].WordAlias), uintptr(unsafe.Pointer(&goValues[0].WordAlias))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Empty", unsafe.Offsetof(goValues[0].Empty), uintptr(unsafe.Pointer(&goValues[0].Empty))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Anonymous", unsafe.Offsetof(goValues[0].Anonymous), uintptr(unsafe.Pointer(&goValues[0].Anonymous))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Blank", unsafe.Offsetof(goValues[0].Blank), uintptr(unsafe.Pointer(&goValues[0].Blank))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Embedded", unsafe.Offsetof(goValues[0].Embedded), uintptr(unsafe.Pointer(&goValues[0].Embedded))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.GoRecord", unsafe.Offsetof(goValues[0].GoRecord), uintptr(unsafe.Pointer(&goValues[0].GoRecord))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.NativeRecord", unsafe.Offsetof(goValues[0].NativeRecord), uintptr(unsafe.Pointer(&goValues[0].NativeRecord))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Array0", unsafe.Offsetof(goValues[0].Array0), uintptr(unsafe.Pointer(&goValues[0].Array0))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Array1", unsafe.Offsetof(goValues[0].Array1), uintptr(unsafe.Pointer(&goValues[0].Array1))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Array2", unsafe.Offsetof(goValues[0].Array2), uintptr(unsafe.Pointer(&goValues[0].Array2))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.Matrix", unsafe.Offsetof(goValues[0].Matrix), uintptr(unsafe.Pointer(&goValues[0].Matrix))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.GoArray", unsafe.Offsetof(goValues[0].GoArray), uintptr(unsafe.Pointer(&goValues[0].GoArray))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.NativeArray", unsafe.Offsetof(goValues[0].NativeArray), uintptr(unsafe.Pointer(&goValues[0].NativeArray))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.NativeArray0", unsafe.Offsetof(goValues[0].NativeArray0), uintptr(unsafe.Pointer(&goValues[0].NativeArray0))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Functions", unsafe.Offsetof(goValues[0].Functions), uintptr(unsafe.Pointer(&goValues[0].Functions))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Callbacks", unsafe.Offsetof(goValues[0].Callbacks), uintptr(unsafe.Pointer(&goValues[0].Callbacks))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.GenericScalar", unsafe.Offsetof(goValues[0].GenericScalar), uintptr(unsafe.Pointer(&goValues[0].GenericScalar))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.GenericNative", unsafe.Offsetof(goValues[0].GenericNative), uintptr(unsafe.Pointer(&goValues[0].GenericNative))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.GenericFunction", unsafe.Offsetof(goValues[0].GenericFunction), uintptr(unsafe.Pointer(&goValues[0].GenericFunction))-goValuesBase) && goOK
	// goOK = checkUnsafe("unsafeTypes.Node", unsafe.Offsetof(goValues[0].Node), uintptr(unsafe.Pointer(&goValues[0].Node))-goValuesBase) && goOK
	goOK = checkUnsafe("unsafeTypes.End", unsafe.Offsetof(goValues[0].End), uintptr(unsafe.Pointer(&goValues[0].End))-goValuesBase) && goOK
	var nativeValues [2]unsafeNativeTypes
	var nativeValuesAlignment struct {
		Prefix byte
		Value  unsafeNativeTypes
	}
	nativeValuesBase := uintptr(unsafe.Pointer(&nativeValues[0]))
	nativeOK = checkUnsafe("unsafeNativeTypes.Sizeof", unsafe.Sizeof(nativeValues[0]), uintptr(unsafe.Pointer(&nativeValues[1]))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Alignof", unsafe.Alignof(nativeValues[0]), uintptr(unsafe.Pointer(&nativeValuesAlignment.Value))-uintptr(unsafe.Pointer(&nativeValuesAlignment))) && nativeOK
	var nativeValuesBound [unsafe.Sizeof(nativeValues[0])]byte
	nativeOK = checkUnsafe("unsafeNativeTypes.array-bound", uintptr(len(nativeValuesBound)), uintptr(unsafe.Pointer(&nativeValues[1]))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Bool", unsafe.Offsetof(nativeValues[0].Bool), uintptr(unsafe.Pointer(&nativeValues[0].Bool))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Int", unsafe.Offsetof(nativeValues[0].Int), uintptr(unsafe.Pointer(&nativeValues[0].Int))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Int8", unsafe.Offsetof(nativeValues[0].Int8), uintptr(unsafe.Pointer(&nativeValues[0].Int8))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Int16", unsafe.Offsetof(nativeValues[0].Int16), uintptr(unsafe.Pointer(&nativeValues[0].Int16))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Int32", unsafe.Offsetof(nativeValues[0].Int32), uintptr(unsafe.Pointer(&nativeValues[0].Int32))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Int64", unsafe.Offsetof(nativeValues[0].Int64), uintptr(unsafe.Pointer(&nativeValues[0].Int64))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uint", unsafe.Offsetof(nativeValues[0].Uint), uintptr(unsafe.Pointer(&nativeValues[0].Uint))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uint8", unsafe.Offsetof(nativeValues[0].Uint8), uintptr(unsafe.Pointer(&nativeValues[0].Uint8))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uint16", unsafe.Offsetof(nativeValues[0].Uint16), uintptr(unsafe.Pointer(&nativeValues[0].Uint16))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uint32", unsafe.Offsetof(nativeValues[0].Uint32), uintptr(unsafe.Pointer(&nativeValues[0].Uint32))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uint64", unsafe.Offsetof(nativeValues[0].Uint64), uintptr(unsafe.Pointer(&nativeValues[0].Uint64))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Uintptr", unsafe.Offsetof(nativeValues[0].Uintptr), uintptr(unsafe.Pointer(&nativeValues[0].Uintptr))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Byte", unsafe.Offsetof(nativeValues[0].Byte), uintptr(unsafe.Pointer(&nativeValues[0].Byte))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Rune", unsafe.Offsetof(nativeValues[0].Rune), uintptr(unsafe.Pointer(&nativeValues[0].Rune))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Float32", unsafe.Offsetof(nativeValues[0].Float32), uintptr(unsafe.Pointer(&nativeValues[0].Float32))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Float64", unsafe.Offsetof(nativeValues[0].Float64), uintptr(unsafe.Pointer(&nativeValues[0].Float64))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Complex64", unsafe.Offsetof(nativeValues[0].Complex64), uintptr(unsafe.Pointer(&nativeValues[0].Complex64))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Complex128", unsafe.Offsetof(nativeValues[0].Complex128), uintptr(unsafe.Pointer(&nativeValues[0].Complex128))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.String", unsafe.Offsetof(nativeValues[0].String), uintptr(unsafe.Pointer(&nativeValues[0].String))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Pointer", unsafe.Offsetof(nativeValues[0].Pointer), uintptr(unsafe.Pointer(&nativeValues[0].Pointer))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.UnsafePointer", unsafe.Offsetof(nativeValues[0].UnsafePointer), uintptr(unsafe.Pointer(&nativeValues[0].UnsafePointer))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.RecursivePointer", unsafe.Offsetof(nativeValues[0].RecursivePointer), uintptr(unsafe.Pointer(&nativeValues[0].RecursivePointer))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Slice", unsafe.Offsetof(nativeValues[0].Slice), uintptr(unsafe.Pointer(&nativeValues[0].Slice))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Map", unsafe.Offsetof(nativeValues[0].Map), uintptr(unsafe.Pointer(&nativeValues[0].Map))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Channel", unsafe.Offsetof(nativeValues[0].Channel), uintptr(unsafe.Pointer(&nativeValues[0].Channel))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.ReceiveChannel", unsafe.Offsetof(nativeValues[0].ReceiveChannel), uintptr(unsafe.Pointer(&nativeValues[0].ReceiveChannel))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.SendChannel", unsafe.Offsetof(nativeValues[0].SendChannel), uintptr(unsafe.Pointer(&nativeValues[0].SendChannel))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Any", unsafe.Offsetof(nativeValues[0].Any), uintptr(unsafe.Pointer(&nativeValues[0].Any))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Interface", unsafe.Offsetof(nativeValues[0].Interface), uintptr(unsafe.Pointer(&nativeValues[0].Interface))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Function", unsafe.Offsetof(nativeValues[0].Function), uintptr(unsafe.Pointer(&nativeValues[0].Function))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Variadic", unsafe.Offsetof(nativeValues[0].Variadic), uintptr(unsafe.Pointer(&nativeValues[0].Variadic))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.NamedFunction", unsafe.Offsetof(nativeValues[0].NamedFunction), uintptr(unsafe.Pointer(&nativeValues[0].NamedFunction))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.FunctionAlias", unsafe.Offsetof(nativeValues[0].FunctionAlias), uintptr(unsafe.Pointer(&nativeValues[0].FunctionAlias))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Callback", unsafe.Offsetof(nativeValues[0].Callback), uintptr(unsafe.Pointer(&nativeValues[0].Callback))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.CallbackAlias", unsafe.Offsetof(nativeValues[0].CallbackAlias), uintptr(unsafe.Pointer(&nativeValues[0].CallbackAlias))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Word", unsafe.Offsetof(nativeValues[0].Word), uintptr(unsafe.Pointer(&nativeValues[0].Word))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.WordAlias", unsafe.Offsetof(nativeValues[0].WordAlias), uintptr(unsafe.Pointer(&nativeValues[0].WordAlias))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Empty", unsafe.Offsetof(nativeValues[0].Empty), uintptr(unsafe.Pointer(&nativeValues[0].Empty))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Anonymous", unsafe.Offsetof(nativeValues[0].Anonymous), uintptr(unsafe.Pointer(&nativeValues[0].Anonymous))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Blank", unsafe.Offsetof(nativeValues[0].Blank), uintptr(unsafe.Pointer(&nativeValues[0].Blank))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Embedded", unsafe.Offsetof(nativeValues[0].Embedded), uintptr(unsafe.Pointer(&nativeValues[0].Embedded))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.GoRecord", unsafe.Offsetof(nativeValues[0].GoRecord), uintptr(unsafe.Pointer(&nativeValues[0].GoRecord))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.NativeRecord", unsafe.Offsetof(nativeValues[0].NativeRecord), uintptr(unsafe.Pointer(&nativeValues[0].NativeRecord))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Array0", unsafe.Offsetof(nativeValues[0].Array0), uintptr(unsafe.Pointer(&nativeValues[0].Array0))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Array1", unsafe.Offsetof(nativeValues[0].Array1), uintptr(unsafe.Pointer(&nativeValues[0].Array1))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Array2", unsafe.Offsetof(nativeValues[0].Array2), uintptr(unsafe.Pointer(&nativeValues[0].Array2))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.Matrix", unsafe.Offsetof(nativeValues[0].Matrix), uintptr(unsafe.Pointer(&nativeValues[0].Matrix))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.GoArray", unsafe.Offsetof(nativeValues[0].GoArray), uintptr(unsafe.Pointer(&nativeValues[0].GoArray))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.NativeArray", unsafe.Offsetof(nativeValues[0].NativeArray), uintptr(unsafe.Pointer(&nativeValues[0].NativeArray))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.NativeArray0", unsafe.Offsetof(nativeValues[0].NativeArray0), uintptr(unsafe.Pointer(&nativeValues[0].NativeArray0))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Functions", unsafe.Offsetof(nativeValues[0].Functions), uintptr(unsafe.Pointer(&nativeValues[0].Functions))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Callbacks", unsafe.Offsetof(nativeValues[0].Callbacks), uintptr(unsafe.Pointer(&nativeValues[0].Callbacks))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.GenericScalar", unsafe.Offsetof(nativeValues[0].GenericScalar), uintptr(unsafe.Pointer(&nativeValues[0].GenericScalar))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.GenericNative", unsafe.Offsetof(nativeValues[0].GenericNative), uintptr(unsafe.Pointer(&nativeValues[0].GenericNative))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.GenericFunction", unsafe.Offsetof(nativeValues[0].GenericFunction), uintptr(unsafe.Pointer(&nativeValues[0].GenericFunction))-nativeValuesBase) && nativeOK
	// nativeOK = checkUnsafe("unsafeNativeTypes.Node", unsafe.Offsetof(nativeValues[0].Node), uintptr(unsafe.Pointer(&nativeValues[0].Node))-nativeValuesBase) && nativeOK
	nativeOK = checkUnsafe("unsafeNativeTypes.End", unsafe.Offsetof(nativeValues[0].End), uintptr(unsafe.Pointer(&nativeValues[0].End))-nativeValuesBase) && nativeOK
	var nested0 [2]unsafeNested0_2
	var nested0Alignment struct {
		Prefix byte
		Value  unsafeNested0_2
	}
	nested0Base := uintptr(unsafe.Pointer(&nested0[0]))
	nestedOK = checkUnsafe("unsafeNested0_2.Sizeof", unsafe.Sizeof(nested0[0]), uintptr(unsafe.Pointer(&nested0[1]))-nested0Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested0_2.Alignof", unsafe.Alignof(nested0[0]), uintptr(unsafe.Pointer(&nested0Alignment.Value))-uintptr(unsafe.Pointer(&nested0Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested0_2.V", unsafe.Offsetof(nested0[0].V), uintptr(unsafe.Pointer(&nested0[0].V))-nested0Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested0_2.End", unsafe.Offsetof(nested0[0].End), uintptr(unsafe.Pointer(&nested0[0].End))-nested0Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested0_2.V.V", unsafe.Offsetof(nested0[0].V.V), uintptr(unsafe.Pointer(&nested0[0].V.V))-uintptr(unsafe.Pointer(&nested0[0].V))) && nestedOK
	var nested1 [2]unsafeNested1_2
	var nested1Alignment struct {
		Prefix byte
		Value  unsafeNested1_2
	}
	nested1Base := uintptr(unsafe.Pointer(&nested1[0]))
	nestedOK = checkUnsafe("unsafeNested1_2.Sizeof", unsafe.Sizeof(nested1[0]), uintptr(unsafe.Pointer(&nested1[1]))-nested1Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested1_2.Alignof", unsafe.Alignof(nested1[0]), uintptr(unsafe.Pointer(&nested1Alignment.Value))-uintptr(unsafe.Pointer(&nested1Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested1_2.V", unsafe.Offsetof(nested1[0].V), uintptr(unsafe.Pointer(&nested1[0].V))-nested1Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested1_2.End", unsafe.Offsetof(nested1[0].End), uintptr(unsafe.Pointer(&nested1[0].End))-nested1Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested1_2.V.V", unsafe.Offsetof(nested1[0].V.V), uintptr(unsafe.Pointer(&nested1[0].V.V))-uintptr(unsafe.Pointer(&nested1[0].V))) && nestedOK
	var nested2 [2]unsafeNested2_2
	var nested2Alignment struct {
		Prefix byte
		Value  unsafeNested2_2
	}
	nested2Base := uintptr(unsafe.Pointer(&nested2[0]))
	nestedOK = checkUnsafe("unsafeNested2_2.Sizeof", unsafe.Sizeof(nested2[0]), uintptr(unsafe.Pointer(&nested2[1]))-nested2Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested2_2.Alignof", unsafe.Alignof(nested2[0]), uintptr(unsafe.Pointer(&nested2Alignment.Value))-uintptr(unsafe.Pointer(&nested2Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested2_2.V", unsafe.Offsetof(nested2[0].V), uintptr(unsafe.Pointer(&nested2[0].V))-nested2Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested2_2.End", unsafe.Offsetof(nested2[0].End), uintptr(unsafe.Pointer(&nested2[0].End))-nested2Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested2_2.V.V", unsafe.Offsetof(nested2[0].V.V), uintptr(unsafe.Pointer(&nested2[0].V.V))-uintptr(unsafe.Pointer(&nested2[0].V))) && nestedOK
	var nested3 [2]unsafeNested3_2
	var nested3Alignment struct {
		Prefix byte
		Value  unsafeNested3_2
	}
	nested3Base := uintptr(unsafe.Pointer(&nested3[0]))
	nestedOK = checkUnsafe("unsafeNested3_2.Sizeof", unsafe.Sizeof(nested3[0]), uintptr(unsafe.Pointer(&nested3[1]))-nested3Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested3_2.Alignof", unsafe.Alignof(nested3[0]), uintptr(unsafe.Pointer(&nested3Alignment.Value))-uintptr(unsafe.Pointer(&nested3Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested3_2.V", unsafe.Offsetof(nested3[0].V), uintptr(unsafe.Pointer(&nested3[0].V))-nested3Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested3_2.End", unsafe.Offsetof(nested3[0].End), uintptr(unsafe.Pointer(&nested3[0].End))-nested3Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested3_2.V.V", unsafe.Offsetof(nested3[0].V.V), uintptr(unsafe.Pointer(&nested3[0].V.V))-uintptr(unsafe.Pointer(&nested3[0].V))) && nestedOK
	var nested4 [2]unsafeNested4_2
	var nested4Alignment struct {
		Prefix byte
		Value  unsafeNested4_2
	}
	nested4Base := uintptr(unsafe.Pointer(&nested4[0]))
	nestedOK = checkUnsafe("unsafeNested4_2.Sizeof", unsafe.Sizeof(nested4[0]), uintptr(unsafe.Pointer(&nested4[1]))-nested4Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested4_2.Alignof", unsafe.Alignof(nested4[0]), uintptr(unsafe.Pointer(&nested4Alignment.Value))-uintptr(unsafe.Pointer(&nested4Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested4_2.V", unsafe.Offsetof(nested4[0].V), uintptr(unsafe.Pointer(&nested4[0].V))-nested4Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested4_2.End", unsafe.Offsetof(nested4[0].End), uintptr(unsafe.Pointer(&nested4[0].End))-nested4Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested4_2.V.V", unsafe.Offsetof(nested4[0].V.V), uintptr(unsafe.Pointer(&nested4[0].V.V))-uintptr(unsafe.Pointer(&nested4[0].V))) && nestedOK
	var nested5 [2]unsafeNested5_2
	var nested5Alignment struct {
		Prefix byte
		Value  unsafeNested5_2
	}
	nested5Base := uintptr(unsafe.Pointer(&nested5[0]))
	nestedOK = checkUnsafe("unsafeNested5_2.Sizeof", unsafe.Sizeof(nested5[0]), uintptr(unsafe.Pointer(&nested5[1]))-nested5Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested5_2.Alignof", unsafe.Alignof(nested5[0]), uintptr(unsafe.Pointer(&nested5Alignment.Value))-uintptr(unsafe.Pointer(&nested5Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested5_2.V", unsafe.Offsetof(nested5[0].V), uintptr(unsafe.Pointer(&nested5[0].V))-nested5Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested5_2.End", unsafe.Offsetof(nested5[0].End), uintptr(unsafe.Pointer(&nested5[0].End))-nested5Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested5_2.V.V", unsafe.Offsetof(nested5[0].V.V), uintptr(unsafe.Pointer(&nested5[0].V.V))-uintptr(unsafe.Pointer(&nested5[0].V))) && nestedOK
	var nested6 [2]unsafeNested6_2
	var nested6Alignment struct {
		Prefix byte
		Value  unsafeNested6_2
	}
	nested6Base := uintptr(unsafe.Pointer(&nested6[0]))
	nestedOK = checkUnsafe("unsafeNested6_2.Sizeof", unsafe.Sizeof(nested6[0]), uintptr(unsafe.Pointer(&nested6[1]))-nested6Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested6_2.Alignof", unsafe.Alignof(nested6[0]), uintptr(unsafe.Pointer(&nested6Alignment.Value))-uintptr(unsafe.Pointer(&nested6Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested6_2.V", unsafe.Offsetof(nested6[0].V), uintptr(unsafe.Pointer(&nested6[0].V))-nested6Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested6_2.End", unsafe.Offsetof(nested6[0].End), uintptr(unsafe.Pointer(&nested6[0].End))-nested6Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested6_2.V.V", unsafe.Offsetof(nested6[0].V.V), uintptr(unsafe.Pointer(&nested6[0].V.V))-uintptr(unsafe.Pointer(&nested6[0].V))) && nestedOK
	var nested7 [2]unsafeNested7_2
	var nested7Alignment struct {
		Prefix byte
		Value  unsafeNested7_2
	}
	nested7Base := uintptr(unsafe.Pointer(&nested7[0]))
	nestedOK = checkUnsafe("unsafeNested7_2.Sizeof", unsafe.Sizeof(nested7[0]), uintptr(unsafe.Pointer(&nested7[1]))-nested7Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested7_2.Alignof", unsafe.Alignof(nested7[0]), uintptr(unsafe.Pointer(&nested7Alignment.Value))-uintptr(unsafe.Pointer(&nested7Alignment))) && nestedOK
	nestedOK = checkUnsafe("unsafeNested7_2.V", unsafe.Offsetof(nested7[0].V), uintptr(unsafe.Pointer(&nested7[0].V))-nested7Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested7_2.End", unsafe.Offsetof(nested7[0].End), uintptr(unsafe.Pointer(&nested7[0].End))-nested7Base) && nestedOK
	nestedOK = checkUnsafe("unsafeNested7_2.V.V", unsafe.Offsetof(nested7[0].V.V), uintptr(unsafe.Pointer(&nested7[0].V.V))-uintptr(unsafe.Pointer(&nested7[0].V))) && nestedOK
	println("unsafe Go fields", goOK)
	println("unsafe native fields", nativeOK)
	println("unsafe nested owners", nestedOK)

	// Keep real values, including their environments, in typed storage.
	// Direct-expression unsafe constants are compared with two-element strides.
	captured := 7
	var zeroCapture struct{}
	capturedSlice := []int{1, 2}
	capturedPointer := &captured
	noCapture := func() {}
	capture := func() { println(captured) }
	captureZero := func() { println(&zeroCapture) }
	captureMany := func() { println(captured, len(capturedSlice), capturedPointer) }
	nestedClosure := func() func() { return func() { println(captured) } }()
	var loopClosures [2]func()
	for i := range loopClosures {
		loopClosures[i] = func() { println(i) }
	}
	type localFunction func()
	type localAlias = localFunction
	reader := strings.NewReader("layout")
	var readerValue strings.Reader
	reflected := reflect.ValueOf([2]int{})
	var lengthReader interface{ Len() int } = reader
	promoted := struct{ *strings.Reader }{reader}
	// TODO(#2907): cdecl function conversion.
	// goNullary := func() int32 { return 7 }
	// var implicitCallback unsafeNativeNullary = goNullary
	var recursive func(int) int
	recursive = func(n int) int {
		if n == 0 {
			return 0
		}
		return recursive(n - 1)
	}
	callableOK := true
	for _, current := range []struct {
		name            string
		value           func()
		size, alignment uintptr
	}{
		{"plain-function", main, unsafe.Sizeof(main), unsafe.Alignof(main)},
		{"no-capture", noCapture, unsafe.Sizeof(noCapture), unsafe.Alignof(noCapture)},
		{"capture-scalar", capture, unsafe.Sizeof(capture), unsafe.Alignof(capture)},
		{"capture-zero-sized", captureZero, unsafe.Sizeof(captureZero), unsafe.Alignof(captureZero)},
		{"capture-multiple", captureMany, unsafe.Sizeof(captureMany), unsafe.Alignof(captureMany)},
		{"nested-closure", nestedClosure, unsafe.Sizeof(nestedClosure), unsafe.Alignof(nestedClosure)},
		{"loop-closure-0", loopClosures[0], unsafe.Sizeof(loopClosures[0]), unsafe.Alignof(loopClosures[0])},
		{"loop-closure-1", loopClosures[1], unsafe.Sizeof(loopClosures[1]), unsafe.Alignof(loopClosures[1])},
		{"nil-function", (func())(nil), unsafe.Sizeof((func())(nil)), unsafe.Alignof((func())(nil))},
		{"boxed-closure", any(capture).(func()), unsafe.Sizeof(any(capture).(func())), unsafe.Alignof(any(capture).(func()))},
		{"local-named-function", localFunction(capture), unsafe.Sizeof(localFunction(capture)), unsafe.Alignof(localFunction(capture))},
		{"local-function-alias", localAlias(capture), unsafe.Sizeof(localAlias(capture)), unsafe.Alignof(localAlias(capture))},
		{"parenthesized-closure", (capture), unsafe.Sizeof((capture)), unsafe.Alignof((capture))},
	} {
		values := [2]func(){current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func()
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	for _, current := range []struct {
		name            string
		value           func() int
		size, alignment uintptr
	}{
		{"value-method", reflected.Len, unsafe.Sizeof(reflected.Len), unsafe.Alignof(reflected.Len)},
		{"pointer-method", reader.Len, unsafe.Sizeof(reader.Len), unsafe.Alignof(reader.Len)},
		{"addressable-pointer-method", readerValue.Len, unsafe.Sizeof(readerValue.Len), unsafe.Alignof(readerValue.Len)},
		{"interface-method", lengthReader.Len, unsafe.Sizeof(lengthReader.Len), unsafe.Alignof(lengthReader.Len)},
		{"promoted-method", promoted.Len, unsafe.Sizeof(promoted.Len), unsafe.Alignof(promoted.Len)},
		{"nil-pointer-receiver", ((*strings.Reader)(nil)).Len, unsafe.Sizeof(((*strings.Reader)(nil)).Len), unsafe.Alignof(((*strings.Reader)(nil)).Len)},
		{"parenthesized-method", (reader.Len), unsafe.Sizeof((reader.Len)), unsafe.Alignof((reader.Len))},
	} {
		values := [2]func() int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func() int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	// TODO(#2907): native callbacks and conversion to Go function values.
	// for _, current := range []struct {
	// name            string
	// value           unsafeNativeNullary
	// size, alignment uintptr
	// }{
	// {"native-symbol", unsafeNativeNullary(c.Rand), unsafe.Sizeof(unsafeNativeNullary(c.Rand)), unsafe.Alignof(unsafeNativeNullary(c.Rand))},
	// {"native-literal", unsafeNativeNullary(func() int32 { return 7 }), unsafe.Sizeof(unsafeNativeNullary(func() int32 { return 7 })), unsafe.Alignof(unsafeNativeNullary(func() int32 { return 7 }))},
	// {"native-no-env-variable", unsafeNativeNullary(goNullary), unsafe.Sizeof(unsafeNativeNullary(goNullary)), unsafe.Alignof(unsafeNativeNullary(goNullary))},
	// {"native-implicit", implicitCallback, unsafe.Sizeof(implicitCallback), unsafe.Alignof(implicitCallback)},
	// {"native-nil", unsafeNativeNullary(nil), unsafe.Sizeof(unsafeNativeNullary(nil)), unsafe.Alignof(unsafeNativeNullary(nil))},
	// {"native-alias", unsafeNativeNullaryAlias(c.Rand), unsafe.Sizeof(unsafeNativeNullaryAlias(c.Rand)), unsafe.Alignof(unsafeNativeNullaryAlias(c.Rand))},
	// {"native-alias-chain", unsafeNativeNullaryAlias2(c.Rand), unsafe.Sizeof(unsafeNativeNullaryAlias2(c.Rand)), unsafe.Alignof(unsafeNativeNullaryAlias2(c.Rand))},
	// } {
	// values := [2]unsafeNativeNullary{current.value, current.value}
	// probe := struct {
	// Prefix byte
	// Value  unsafeNativeNullary
	// }{Value: current.value}
	// stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
	// fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
	// callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
	// callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
	// callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
	// callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	// }
	// for _, current := range []struct {
	// name            string
	// value           func() int32
	// size, alignment uintptr
	// }{
	// {"C-symbol-as-Go-value", c.Rand, unsafe.Sizeof(c.Rand), unsafe.Alignof(c.Rand)},
	// {"callback-to-Go-value", (func() int32)(implicitCallback), unsafe.Sizeof((func() int32)(implicitCallback)), unsafe.Alignof((func() int32)(implicitCallback))},
	// {"nil-callback-to-Go-value", (func() int32)(unsafeNativeNullary(nil)), unsafe.Sizeof((func() int32)(unsafeNativeNullary(nil))), unsafe.Alignof((func() int32)(unsafeNativeNullary(nil)))},
	// } {
	// values := [2]func() int32{current.value, current.value}
	// probe := struct {
	// Prefix byte
	// Value  func() int32
	// }{Value: current.value}
	// stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
	// fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
	// callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
	// callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
	// callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
	// callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	// }
	for _, current := range []struct {
		name            string
		value           func(int) int
		size, alignment uintptr
	}{
		{"named-no-env", unsafeFunction(func(v int) int { return v }), unsafe.Sizeof(unsafeFunction(func(v int) int { return v })), unsafe.Alignof(unsafeFunction(func(v int) int { return v }))},
		{"named-captured", unsafeFunction(func(v int) int { return v + captured }), unsafe.Sizeof(unsafeFunction(func(v int) int { return v + captured })), unsafe.Alignof(unsafeFunction(func(v int) int { return v + captured }))},
		{"aliased-captured", unsafeFunctionAlias(func(v int) int { return v + captured }), unsafe.Sizeof(unsafeFunctionAlias(func(v int) int { return v + captured })), unsafe.Alignof(unsafeFunctionAlias(func(v int) int { return v + captured }))},
		{"recursive-function", recursive, unsafe.Sizeof(recursive), unsafe.Alignof(recursive)},
	} {
		values := [2]func(int) int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func(int) int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	for _, current := range []struct {
		name            string
		value           func(reflect.Value) int
		size, alignment uintptr
	}{
		{"value-method-expression", reflect.Value.Len, unsafe.Sizeof(reflect.Value.Len), unsafe.Alignof(reflect.Value.Len)},
	} {
		values := [2]func(reflect.Value) int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func(reflect.Value) int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	for _, current := range []struct {
		name            string
		value           func(*strings.Reader) int
		size, alignment uintptr
	}{
		{"pointer-method-expression", (*strings.Reader).Len, unsafe.Sizeof((*strings.Reader).Len), unsafe.Alignof((*strings.Reader).Len)},
	} {
		values := [2]func(*strings.Reader) int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func(*strings.Reader) int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	for _, current := range []struct {
		name            string
		value           func([]int) []int
		size, alignment uintptr
	}{
		{"instantiated-generic", slices.Clone[[]int], unsafe.Sizeof(slices.Clone[[]int]), unsafe.Alignof(slices.Clone[[]int])},
	} {
		values := [2]func([]int) []int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func([]int) []int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}
	for _, current := range []struct {
		name            string
		value           func(string, string) int
		size, alignment uintptr
	}{
		{"imported-function", strings.Compare, unsafe.Sizeof(strings.Compare), unsafe.Alignof(strings.Compare)},
	} {
		values := [2]func(string, string) int{current.value, current.value}
		probe := struct {
			Prefix byte
			Value  func(string, string) int
		}{Value: current.value}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		fieldOffset := uintptr(unsafe.Pointer(&probe.Value)) - uintptr(unsafe.Pointer(&probe))
		callableOK = checkUnsafe(current.name+".Sizeof", current.size, stride) && callableOK
		callableOK = checkUnsafe(current.name+".Alignof", current.alignment, fieldOffset) && callableOK
		callableOK = checkUnsafe(current.name+".array", unsafe.Sizeof(values), 2*stride) && callableOK
		callableOK = checkUnsafe(current.name+".field", unsafe.Offsetof(probe.Value), fieldOffset) && callableOK
	}

	// TODO(#2907): mixed native callable fields and capture conversions.
	// // Captured Go values are valid in Go containers. Native containers receive
	// // only no-env values; captured conversions are tested separately below.
	// mixed := [2]unsafeCallableMix{
	// {Go: capture, Method: reader.Len, C: implicitCallback},
	// {Go: noCapture, Method: reflected.Len, C: nil},
	// }
	// nativeMixed := [2]unsafeNativeCallableMix{
	// {Go: noCapture, Method: func() int { return 1 }, C: implicitCallback},
	// {},
	// }
	// callableOK = checkUnsafe("mixed-callables.Sizeof", unsafe.Sizeof(mixed[0]),
	// uintptr(unsafe.Pointer(&mixed[1]))-uintptr(unsafe.Pointer(&mixed[0]))) && callableOK
	// callableOK = checkUnsafe("mixed-callables.Method", unsafe.Offsetof(mixed[0].Method),
	// uintptr(unsafe.Pointer(&mixed[0].Method))-uintptr(unsafe.Pointer(&mixed[0]))) && callableOK
	// callableOK = checkUnsafe("mixed-callables.C", unsafe.Offsetof(mixed[0].C),
	// uintptr(unsafe.Pointer(&mixed[0].C))-uintptr(unsafe.Pointer(&mixed[0]))) && callableOK
	// callableOK = checkUnsafe("native-mixed-callables.Sizeof", unsafe.Sizeof(nativeMixed[0]),
	// uintptr(unsafe.Pointer(&nativeMixed[1]))-uintptr(unsafe.Pointer(&nativeMixed[0]))) && callableOK
	// callableOK = checkUnsafe("native-mixed-callables.Method", unsafe.Offsetof(nativeMixed[0].Method),
	// uintptr(unsafe.Pointer(&nativeMixed[0].Method))-uintptr(unsafe.Pointer(&nativeMixed[0]))) && callableOK
	// callableOK = checkUnsafe("native-mixed-callables.C", unsafe.Offsetof(nativeMixed[0].C),
	// uintptr(unsafe.Pointer(&nativeMixed[0].C))-uintptr(unsafe.Pointer(&nativeMixed[0]))) && callableOK

	// // Match the existing callbackmethod rejection contract. Never invoke a
	// // native callback after dropping a required Go environment.
	// rejectCapture := func(fn func() int) (rejected bool) {
	// defer func() {
	// err, ok := recover().(error)
	// rejected = ok && err.Error() == "runtime error: C callback must not capture variables"
	// }()
	// converted := unsafeNativeMethod(fn)
	// if converted == nil {
	// panic("unexpected nil callback")
	// }
	// return false
	// }
	// rejectionOK := true
	// for _, current := range []struct {
	// name  string
	// value func() int
	// }{
	// {"captured-scalar", func() int { return captured }},
	// {"captured-zero-sized", func() int { println(&zeroCapture); return 0 }},
	// {"bound-pointer-method", reader.Len},
	// {"bound-value-method", reflected.Len},
	// {"bound-interface-method", lengthReader.Len},
	// {"bound-nil-pointer-method", ((*strings.Reader)(nil)).Len},
	// } {
	// if !rejectCapture(current.value) {
	// println("unsafe callback capture accepted", current.name)
	// rejectionOK = false
	// }
	// }
	println("unsafe callable values", callableOK)
	// println("unsafe callback rejection", rejectionOK)

}
