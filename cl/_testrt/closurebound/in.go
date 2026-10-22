// LITTEST darwin/arm64 linux/amd64
package main

import "unsafe"

var my = demo2{}.encode

// TODO(#2907): defer cdecl function conversion; retain Go method layouts.
// //llgo:cdecl
// type nativeBound func() int

type demo1 struct {
}

func (se demo1) encode() int {
	return 1
}

type demo2 struct {
}

func (se demo2) encode() int {
	return 2
}

func main() {
	se := demo1{}
	f := se.encode
	if f() != 1 {
		panic("error")
	}
	// Empty receivers still require a Go method environment. Layout must not
	// shrink because the receiver occupies no bytes or the value is global.
	promoted := struct{ demo1 }{se}
	var dynamic interface{ encode() int } = se
	layoutOK := true
	// rejectionOK := true
	// reject := func(fn func() int) (rejected bool) {
	// defer func() {
	// err, ok := recover().(error)
	// rejected = ok && err.Error() == "runtime error: C callback must not capture variables"
	// }()
	// converted := nativeBound(fn)
	// if converted == nil {
	// panic("unexpected nil callback")
	// }
	// return false
	// }
	for _, fn := range []func() int{f, my, promoted.encode, dynamic.encode} {
		values := [2]func() int{fn, fn}
		probe := struct {
			B byte
			F func() int
		}{F: fn}
		stride := uintptr(unsafe.Pointer(&values[1])) - uintptr(unsafe.Pointer(&values[0]))
		field := uintptr(unsafe.Pointer(&probe.F)) - uintptr(unsafe.Pointer(&probe))
		layoutOK = layoutOK && unsafe.Sizeof(fn) == stride && unsafe.Alignof(fn) == field && unsafe.Offsetof(probe.F) == field
		// rejectionOK = reject(fn) && rejectionOK
	}
	methods := [2]func(demo1) int{demo1.encode, demo1.encode}
	pointerMethods := [2]func(*demo1) int{(*demo1).encode, (*demo1).encode}
	layoutOK = layoutOK && unsafe.Sizeof(demo1.encode) == uintptr(unsafe.Pointer(&methods[1]))-uintptr(unsafe.Pointer(&methods[0]))
	layoutOK = layoutOK && unsafe.Sizeof((*demo1).encode) == uintptr(unsafe.Pointer(&pointerMethods[1]))-uintptr(unsafe.Pointer(&pointerMethods[0]))
	println("zero-sized receiver layout", layoutOK)
	// println("zero-sized receiver callback rejection", rejectionOK)
}

// Pointer receiver wrappers must guard nil before calling the value receiver.
// CHECK-LABEL: define i64 @"main.(*demo1).encode"(ptr %0){{.*}} {
// CHECK: %[[D1_NIL:[0-9]+]] = icmp eq ptr %0, null
// CHECK: call void @"{{.*}}/runtime/internal/runtime.PanicWrapNilPointer"(i1 %[[D1_NIL]],{{.*}})
// CHECK: call i64 @main.demo1.encode(%main.demo1 zeroinitializer)

// CHECK-LABEL: define i64 @"main.(*demo2).encode"(ptr %0){{.*}} {
// CHECK: %[[D2_NIL:[0-9]+]] = icmp eq ptr %0, null
// CHECK: call void @"{{.*}}/runtime/internal/runtime.PanicWrapNilPointer"(i1 %[[D2_NIL]],{{.*}})
// CHECK: call i64 @main.demo2.encode(%main.demo2 zeroinitializer)

// The package-global method value stores both the selected bound wrapper and
// the zero-sized receiver in init.
// CHECK-LABEL: define void @main.init(){{.*}} {
// CHECK: store { ptr, ptr } { ptr @"main.demo2.encode$bound", ptr @"__llgo.moduleZeroSizedAlloc$" }, ptr @main.my

// CHECK-LABEL: define void @main.main(){{.*}} {
// ARM64: %[[RESULT:[0-9]+]] = call i64 @"main.demo1.encode$bound"(ptr swiftself @"__llgo.moduleZeroSizedAlloc$")
// AMD64: %[[RESULT:[0-9]+]] = call i64 @"main.demo1.encode$bound"(ptr nest @"__llgo.moduleZeroSizedAlloc$")
// CHECK: %[[BAD:[0-9]+]] = icmp ne i64 %[[RESULT]], 1
// CHECK: br i1 %[[BAD]]

// ARM64-LABEL: define i64 @"main.demo2.encode$bound"(ptr swiftself %0){{.*}} {
// AMD64-LABEL: define i64 @"main.demo2.encode$bound"(ptr nest %0){{.*}} {
// CHECK: br i1 %{{[0-9]+}}, label %{{[^,]+}}, label %{{[^ ]+}}
// CHECK: call void @"{{.*}}/runtime/internal/runtime.AssertNilDeref"(i1 true)
// CHECK: %[[D2_RESULT:[0-9]+]] = call i64 @main.demo2.encode(%main.demo2 zeroinitializer)
// CHECK: ret i64 %[[D2_RESULT]]

// ARM64-LABEL: define i64 @"main.demo1.encode$bound"(ptr swiftself %0){{.*}} {
// AMD64-LABEL: define i64 @"main.demo1.encode$bound"(ptr nest %0){{.*}} {
// CHECK: br i1 %{{[0-9]+}}, label %{{[^,]+}}, label %{{[^ ]+}}
// CHECK: call void @"{{.*}}/runtime/internal/runtime.AssertNilDeref"(i1 true)
// CHECK: %[[D1_RESULT:[0-9]+]] = call i64 @main.demo1.encode(%main.demo1 zeroinitializer)
// CHECK: ret i64 %[[D1_RESULT]]
