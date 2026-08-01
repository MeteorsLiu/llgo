// LITTEST
package main

import (
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/c/sync/atomic"
)

// CHECK-LABEL: define void @main.main(){{.*}} {
func main() {
	var v int64

	// CHECK: store atomic i64 100, ptr %.stack seq_cst, align 8
	atomic.Store(&v, 100)
	// CHECK: %0 = load atomic i64, ptr %.stack seq_cst, align 8
	c.Printf(c.Str("store: %ld\n"), atomic.Load(&v))
	// CHECK: %2 = atomicrmw add ptr %.stack, i64 1 seq_cst, align 8
	ret := atomic.Add(&v, 1)
	c.Printf(c.Str("ret: %ld, v: %ld\n"), ret, v)

	// CHECK: %5 = cmpxchg ptr %.stack, i64 100, i64 102 seq_cst seq_cst, align 8
	ret, _ = atomic.CompareAndExchange(&v, 100, 102)
	c.Printf(c.Str("ret: %ld vs 100, v: %ld\n"), ret, v)

	// CHECK: %10 = cmpxchg ptr %.stack, i64 101, i64 102 seq_cst seq_cst, align 8
	ret, _ = atomic.CompareAndExchange(&v, 101, 102)
	c.Printf(c.Str("ret: %ld vs 101, v: %ld\n"), ret, v)

	// CHECK: %15 = atomicrmw sub ptr %.stack, i64 1 seq_cst, align 8
	ret = atomic.Sub(&v, 1)
	c.Printf(c.Str("ret: %ld, v: %ld\n"), ret, v)
}
