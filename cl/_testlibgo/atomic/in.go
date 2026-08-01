// LITTEST
package main

import (
	"sync/atomic"
)

// CHECK-LABEL: define void @main.main(){{.*}} {
func main() {
	var v int64

	// CHECK: store atomic i64 100, ptr %.stack seq_cst, align 8
	// CHECK: %0 = load atomic i64, ptr %.stack seq_cst, align 8
	atomic.StoreInt64(&v, 100)
	println("store:", atomic.LoadInt64(&v))

	// CHECK: %1 = atomicrmw add ptr %.stack, i64 1 seq_cst, align 8
	// CHECK: %2 = add i64 %1, 1
	// CHECK: %3 = load i64, ptr %.stack, align 8
	ret := atomic.AddInt64(&v, 1)
	println("ret:", ret, "v:", v)

	// CHECK: %4 = cmpxchg ptr %.stack, i64 100, i64 102 seq_cst seq_cst, align 8
	// CHECK: %5 = extractvalue { i64, i1 } %4, 1
	// CHECK: %6 = load i64, ptr %.stack, align 8
	swp := atomic.CompareAndSwapInt64(&v, 100, 102)
	println("swp:", swp, "v:", v)

	// CHECK: %7 = cmpxchg ptr %.stack, i64 101, i64 102 seq_cst seq_cst, align 8
	// CHECK: %8 = extractvalue { i64, i1 } %7, 1
	// CHECK: %9 = load i64, ptr %.stack, align 8
	swp = atomic.CompareAndSwapInt64(&v, 101, 102)
	println("swp:", swp, "v:", v)

	// CHECK: %10 = atomicrmw add ptr %.stack, i64 -1 seq_cst, align 8
	// CHECK: %11 = add i64 %10, -1
	// CHECK: %12 = load i64, ptr %.stack, align 8
	ret = atomic.AddInt64(&v, -1)
	println("ret:", ret, "v:", v)
}
