// LITTEST linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/386 windows/386
package main

import (
	"reflect"
	"unsafe"

	"github.com/xgo-dev/llgo/runtime/abi"
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
}
