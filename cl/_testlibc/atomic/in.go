// LITTEST
package main

import (
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/c/sync/atomic"
)

// ESCAPE-LABEL: define void @main.main(){{.*}} {
// ESCAPE: %.stack = alloca i8, i64 8, align 8
// ESCAPE: call void @llvm.memset.p0.i64(ptr %.stack, i8 0, i64 8, i1 false)
// ESCAPE: store atomic i64 100, ptr %.stack seq_cst, align 8
// ESCAPE: load atomic i64, ptr %.stack seq_cst, align 8
// ESCAPE: atomicrmw add ptr %.stack, i64 1 seq_cst, align 8
// ESCAPE: cmpxchg ptr %.stack, i64 100, i64 102 seq_cst seq_cst, align 8
// ESCAPE: cmpxchg ptr %.stack, i64 101, i64 102 seq_cst seq_cst, align 8
// ESCAPE: atomicrmw sub ptr %.stack, i64 1 seq_cst, align 8
// ESCAPE: ret void

// CHECK-LABEL: define void @main.main(){{.*}} {
func main() {
	var v int64

	// CHECK: store atomic i64 100, ptr %0 seq_cst, align 8
	atomic.Store(&v, 100)
	// CHECK: %1 = load atomic i64, ptr %0 seq_cst, align 8
	c.Printf(c.Str("store: %ld\n"), atomic.Load(&v))
	// CHECK: %3 = atomicrmw add ptr %0, i64 1 seq_cst, align 8
	ret := atomic.Add(&v, 1)
	c.Printf(c.Str("ret: %ld, v: %ld\n"), ret, v)

	// CHECK: %6 = cmpxchg ptr %0, i64 100, i64 102 seq_cst seq_cst, align 8
	ret, _ = atomic.CompareAndExchange(&v, 100, 102)
	c.Printf(c.Str("ret: %ld vs 100, v: %ld\n"), ret, v)

	// CHECK: %11 = cmpxchg ptr %0, i64 101, i64 102 seq_cst seq_cst, align 8
	ret, _ = atomic.CompareAndExchange(&v, 101, 102)
	c.Printf(c.Str("ret: %ld vs 101, v: %ld\n"), ret, v)

	// CHECK: %16 = atomicrmw sub ptr %0, i64 1 seq_cst, align 8
	ret = atomic.Sub(&v, 1)
	c.Printf(c.Str("ret: %ld, v: %ld\n"), ret, v)
}
