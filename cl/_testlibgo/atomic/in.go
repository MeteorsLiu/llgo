// LITTEST
package main

import (
	"sync/atomic"
)

// ESCAPE-LABEL: define void @main.main(){{.*}} {
// ESCAPE: %.stack = alloca i8, i64 8, align 8
// ESCAPE: call void @llvm.memset.p0.i64(ptr %.stack, i8 0, i64 8, i1 false)
// ESCAPE: store atomic i64 100, ptr %.stack seq_cst, align 8
// ESCAPE: load atomic i64, ptr %.stack seq_cst, align 8
// ESCAPE: atomicrmw add ptr %.stack, i64 1 seq_cst, align 8
// ESCAPE: cmpxchg ptr %.stack, i64 100, i64 102 seq_cst seq_cst, align 8
// ESCAPE: cmpxchg ptr %.stack, i64 101, i64 102 seq_cst seq_cst, align 8
// ESCAPE: atomicrmw add ptr %.stack, i64 -1 seq_cst, align 8
// ESCAPE: ret void

// CHECK-LABEL: define void @main.main(){{.*}} {
func main() {
	var v int64

	// CHECK: store atomic i64 100, ptr %0 seq_cst, align 8
	// CHECK: %1 = load atomic i64, ptr %0 seq_cst, align 8
	atomic.StoreInt64(&v, 100)
	println("store:", atomic.LoadInt64(&v))

	// CHECK: %2 = atomicrmw add ptr %0, i64 1 seq_cst, align 8
	// CHECK: %3 = add i64 %2, 1
	// CHECK: %4 = load i64, ptr %0, align 8
	ret := atomic.AddInt64(&v, 1)
	println("ret:", ret, "v:", v)

	// CHECK: %5 = cmpxchg ptr %0, i64 100, i64 102 seq_cst seq_cst, align 8
	// CHECK: %6 = extractvalue { i64, i1 } %5, 1
	// CHECK: %7 = load i64, ptr %0, align 8
	swp := atomic.CompareAndSwapInt64(&v, 100, 102)
	println("swp:", swp, "v:", v)

	// CHECK: %8 = cmpxchg ptr %0, i64 101, i64 102 seq_cst seq_cst, align 8
	// CHECK: %9 = extractvalue { i64, i1 } %8, 1
	// CHECK: %10 = load i64, ptr %0, align 8
	swp = atomic.CompareAndSwapInt64(&v, 101, 102)
	println("swp:", swp, "v:", v)

	// CHECK: %11 = atomicrmw add ptr %0, i64 -1 seq_cst, align 8
	// CHECK: %12 = add i64 %11, -1
	// CHECK: %13 = load i64, ptr %0, align 8
	ret = atomic.AddInt64(&v, -1)
	println("ret:", ret, "v:", v)
}
