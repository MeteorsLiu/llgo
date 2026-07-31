//go:build !llgo

package escape

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	llabi "github.com/goplus/llgo/internal/abi"
	"github.com/goplus/llgo/internal/cabi"
	"github.com/goplus/llgo/ssa"
	"github.com/xgo-dev/llvm"
)

const allocationDeclarations = `
declare ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocZ"(i64)
declare ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64)
`

func parseModule(t *testing.T, ir string) (llvm.Context, llvm.Module) {
	t.Helper()
	ctx := llvm.NewContext()
	path := filepath.Join(t.TempDir(), "input.ll")
	if err := os.WriteFile(path, []byte(ir), 0o644); err != nil {
		ctx.Dispose()
		t.Fatal(err)
	}
	buffer, err := llvm.NewMemoryBufferFromFile(path)
	if err != nil {
		ctx.Dispose()
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buffer)
	if err != nil {
		ctx.Dispose()
		t.Fatalf("parse LLVM IR: %v\n%s", err, ir)
	}
	return ctx, mod
}

func transformModule(t *testing.T, ir string) (llvm.Context, llvm.Module) {
	t.Helper()
	ctx, mod := parseModule(t, ir)
	if err := TransformModule(mod); err != nil {
		mod.Dispose()
		ctx.Dispose()
		t.Fatalf("TransformModule: %v\n%s", err, mod.String())
	}
	return ctx, mod
}

func requireStackAllocation(t *testing.T, mod llvm.Module, name string) {
	t.Helper()
	ir := mod.NamedFunction(name).String()
	if strings.Contains(ir, runtimeAllocZ) || strings.Contains(ir, runtimeAllocU) {
		t.Fatalf("%s allocation was not moved to the stack:\n%s", name, ir)
	}
	if !strings.Contains(ir, "alloca i8") {
		t.Fatalf("%s has no stack allocation:\n%s", name, ir)
	}
}

func requireHeapAllocation(t *testing.T, mod llvm.Module, name string) {
	t.Helper()
	ir := mod.NamedFunction(name).String()
	if !strings.Contains(ir, runtimeAllocZ) && !strings.Contains(ir, runtimeAllocU) {
		t.Fatalf("%s allocation unexpectedly moved to the stack:\n%s", name, ir)
	}
}

func TestSafeUsesAndTransparentPropagation(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
` + allocationDeclarations + `
declare void @llvm.lifetime.start.p0(i64 immarg, ptr nocapture)
declare void @llvm.lifetime.end.p0(i64 immarg, ptr nocapture)
declare void @llvm.memset.p0.i64(ptr writeonly, i8, i64, i1 immarg)
declare void @llvm.memcpy.p0.p0.i64(ptr noalias writeonly, ptr noalias readonly, i64, i1 immarg)
@bytes = global [8 x i8] zeroinitializer

define i64 @load_store() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i64 7, ptr %p, align 8
  %v = load i64, ptr %p, align 8
  ret i64 %v
}

define void @gep() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = getelementptr i8, ptr %p, i64 1
  store i8 7, ptr %q, align 1
  ret void
}

define void @bitcast() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = bitcast ptr %p to ptr
  store i8 7, ptr %q, align 1
  ret void
}

define void @select_value(i1 %cond) {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = select i1 %cond, ptr %p, ptr %p
  store i8 7, ptr %q, align 1
  ret void
}

define void @phi_cycle(i1 %again) {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  br label %loop
loop:
  %q = phi ptr [ %p, %entry ], [ %q, %loop ]
  store i8 7, ptr %q, align 1
  br i1 %again, label %loop, label %exit
exit:
  ret void
}

define i1 @compare() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %same = icmp eq ptr %p, null
  ret i1 %same
}

define void @atomic_addresses() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store atomic i32 1, ptr %p monotonic, align 4
  %v = load atomic i32, ptr %p monotonic, align 4
  %cx = cmpxchg ptr %p, i32 %v, i32 2 monotonic monotonic, align 4
  %rmw = atomicrmw add ptr %p, i32 1 monotonic, align 4
  ret void
}

define void @freeze_value() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = freeze ptr %p
  store i8 7, ptr %q
  ret void
}

define void @volatile_addresses() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store volatile i8 7, ptr %p
  %v = load volatile i8, ptr %p
  ret void
}

define void @safe_intrinsics() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @llvm.lifetime.start.p0(i64 8, ptr %p)
  call void @llvm.memset.p0.i64(ptr %p, i8 0, i64 8, i1 false)
  call void @llvm.memcpy.p0.p0.i64(ptr @bytes, ptr %p, i64 8, i1 false)
  call void @llvm.lifetime.end.p0(i64 8, ptr %p)
  ret void
}

define void @cmpxchg_expected() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %old = cmpxchg ptr @escaped_pointer, ptr %p, ptr null monotonic monotonic, align 8
  ret void
}

@escaped_pointer = global ptr null
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	for _, name := range []string{"load_store", "gep", "bitcast", "select_value", "phi_cycle", "compare", "atomic_addresses", "freeze_value", "volatile_addresses", "safe_intrinsics", "cmpxchg_expected"} {
		requireStackAllocation(t, mod, name)
	}
	if got := mod.NamedFunction("load_store").String(); !strings.Contains(got, "alloca i8, i64 8, align 8") {
		t.Fatalf("load alignment was not preserved:\n%s", got)
	}
}

func TestExactLocalCopy(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
@escaped = global ptr null
` + allocationDeclarations + `
declare void @clobber(ptr)

define void @exact_copy() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  store ptr %p, ptr %slot
  %q = load ptr, ptr %slot
  store i8 1, ptr %q
  ret void
}

define void @alias() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  %alias = getelementptr ptr, ptr %slot, i64 0
  store ptr %p, ptr %slot
  %q = load ptr, ptr %alias
  store i8 1, ptr %q
  ret void
}

define void @slot_address_escape() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  store ptr %p, ptr %slot
  store ptr %slot, ptr @escaped
  ret void
}

define void @unknown_clobber() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  store ptr %p, ptr %slot
  call void @clobber(ptr %slot)
  ret void
}

define void @unsupported_reader() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  store ptr %p, ptr %slot
  %q = load i64, ptr %slot
  ret void
}

define void @other_store() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %slot = alloca ptr
  store ptr %p, ptr %slot
  store ptr null, ptr %slot
  ret void
}

define void @global_destination() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store ptr %p, ptr @escaped
  ret void
}

define void @nonlocal_destination(ptr %dst) {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store ptr %p, ptr %dst
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	requireStackAllocation(t, mod, "exact_copy")
	for _, name := range []string{"alias", "slot_address_escape", "unknown_clobber", "unsupported_reader", "other_store", "global_destination", "nonlocal_destination"} {
		requireHeapAllocation(t, mod, name)
	}
}

func TestDirectCallSummariesAndRecursion(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
@escaped = global ptr null
` + allocationDeclarations + `

define void @read(ptr %p) {
entry:
  %v = load i8, ptr %p
  ret void
}

define void @save(ptr %p) {
entry:
  store ptr %p, ptr @escaped
  ret void
}

define ptr @identity(ptr %p) {
entry:
  ret ptr %p
}

define void @self_safe(ptr %p) {
entry:
  call void @self_safe(ptr %p)
  ret void
}

define void @self_sink(ptr %p) {
entry:
  call void @self_sink(ptr %p)
  store ptr %p, ptr @escaped
  ret void
}

define void @mutual_safe_a(ptr %p) {
entry:
  call void @mutual_safe_b(ptr %p)
  ret void
}

define void @mutual_safe_b(ptr %p) {
entry:
  call void @mutual_safe_a(ptr %p)
  ret void
}

define void @mutual_sink_a(ptr %p) {
entry:
  call void @mutual_sink_b(ptr %p)
  ret void
}

define void @mutual_sink_b(ptr %p) {
entry:
  call void @mutual_sink_a(ptr %p)
  store ptr %p, ptr @escaped
  ret void
}

define void @known_no_leak() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @read(ptr %p)
  ret void
}

define void @known_heap_leak() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @save(ptr %p)
  ret void
}

define void @result_local() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = call ptr @identity(ptr %p)
  store i8 1, ptr %q
  ret void
}

define void @result_escape() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = call ptr @identity(ptr %p)
  store ptr %q, ptr @escaped
  ret void
}

define void @direct_recursion_safe() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @self_safe(ptr %p)
  ret void
}

define void @direct_recursion_sink() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @self_sink(ptr %p)
  ret void
}

define void @mutual_recursion_safe() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @mutual_safe_a(ptr %p)
  ret void
}

define void @mutual_recursion_sink() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @mutual_sink_a(ptr %p)
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	for _, name := range []string{"known_no_leak", "result_local", "direct_recursion_safe", "mutual_recursion_safe"} {
		requireStackAllocation(t, mod, name)
	}
	for _, name := range []string{"known_heap_leak", "result_escape", "direct_recursion_sink", "mutual_recursion_sink"} {
		requireHeapAllocation(t, mod, name)
	}
}

func TestUnknownCallBoundaries(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
` + allocationDeclarations + `
declare void @c_function(ptr)
declare void @"github.com/goplus/llgo/runtime/internal/runtime.ChanSend"(ptr)
declare ptr @llvm.launder.invariant.group.p0(ptr)

define void @bodyless_c() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @c_function(ptr %p)
  ret void
}

define void @runtime_call() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void @"github.com/goplus/llgo/runtime/internal/runtime.ChanSend"(ptr %p)
  ret void
}

define void @indirect_call(ptr %fn) {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void %fn(ptr %p)
  ret void
}

define void @inline_asm() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  call void asm sideeffect "", "r"(ptr %p)
  ret void
}

define void @unknown_intrinsic() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = call ptr @llvm.launder.invariant.group.p0(ptr %p)
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	for _, name := range []string{"bodyless_c", "runtime_call", "indirect_call", "inline_asm", "unknown_intrinsic"} {
		requireHeapAllocation(t, mod, name)
	}
}

func TestInvokeUsesDirectSummaryAndRejectsUnknownCallee(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
` + allocationDeclarations + `
declare i32 @personality(...)
declare void @external(ptr)

define void @read_invoke(ptr %p) {
entry:
  %v = load i8, ptr %p
  ret void
}

define void @known_invoke() personality ptr @personality {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  invoke void @read_invoke(ptr %p) to label %done unwind label %failed
done:
  ret void
failed:
  %landing = landingpad { ptr, i32 } cleanup
  resume { ptr, i32 } %landing
}

define void @unknown_invoke() personality ptr @personality {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  invoke void @external(ptr %p) to label %done unwind label %failed
done:
  ret void
failed:
  %landing = landingpad { ptr, i32 } cleanup
  resume { ptr, i32 } %landing
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	requireStackAllocation(t, mod, "known_invoke")
	requireHeapAllocation(t, mod, "unknown_invoke")
}

func TestConversionsReturnsAndAtomicPublicationEscape(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-p1:64:64-i64:64-n32:64-S128"
@escaped = global ptr null
` + allocationDeclarations + `

define i64 @ptr_to_int() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %n = ptrtoint ptr %p to i64
  ret i64 %n
}

define void @address_space_cast() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %q = addrspacecast ptr %p to ptr addrspace(1)
  ret void
}

define ptr @returned() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  ret ptr %p
}

define void @atomic_store_value() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store atomic ptr %p, ptr @escaped release, align 8
  ret void
}

define void @cmpxchg_replacement() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %old = cmpxchg ptr @escaped, ptr null, ptr %p monotonic monotonic, align 8
  ret void
}

define void @atomicrmw_value() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  %old = atomicrmw xchg ptr @escaped, ptr %p monotonic, align 8
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	for _, name := range []string{"ptr_to_int", "address_space_cast", "returned", "atomic_store_value", "cmpxchg_replacement", "atomicrmw_value"} {
		requireHeapAllocation(t, mod, name)
	}
}

func TestParameterSummaryFacts(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
@escaped = global ptr null

define ptr @facts(ptr %p) {
entry:
  %q = load ptr, ptr %p, align 8
  store ptr %q, ptr @escaped
  store i8 1, ptr %p, align 4
  ret ptr %p
}

define ptr @forward(ptr %p) {
entry:
  %q = call ptr @facts(ptr %p)
  ret ptr %q
}

define void @call_parameter(ptr %fn) {
entry:
  call void %fn()
  ret void
}
`
	ctx, mod := parseModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	a := newAnalyzer(mod)
	a.solveSummaries()

	facts := a.summaries[summaryKey{fn: mod.NamedFunction("facts"), param: 0}].summary
	if facts.heapLevel != 1 || facts.mutatorLevel != 0 || facts.results[0] != 0 || facts.alignments[0] != 8 {
		t.Fatalf("facts summary = %#v", facts)
	}
	forward := a.summaries[summaryKey{fn: mod.NamedFunction("forward"), param: 0}].summary
	if forward.heapLevel != 1 || forward.mutatorLevel != 0 || forward.results[0] != 0 || forward.alignments[0] != 8 {
		t.Fatalf("forward summary = %#v", forward)
	}
	callee := a.summaries[summaryKey{fn: mod.NamedFunction("call_parameter"), param: 0}].summary
	if callee.calleeLevel != 0 || callee.heapLevel != -1 {
		t.Fatalf("callee summary = %#v", callee)
	}
}

func TestAllocationRewriteAndPolicyRejections(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-i64:64-n32:64-S128"
@escaped = global ptr null
` + allocationDeclarations + `

define void @alloc_z() {
entry:
  %p = call align 16 ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocZ"(i64 8)
  store i8 1, ptr %p
  ret void
}

define void @alloc_u() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i8 1, ptr %p
  ret void
}

define void @escaping() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store ptr %p, ptr @escaped
  ret void
}

define void @loop(i1 %again) {
entry:
  br label %body
body:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i8 1, ptr %p
  br i1 %again, label %body, label %exit
exit:
  ret void
}

define void @dynamic_size(i64 %size) {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 %size)
  store i8 1, ptr %p
  ret void
}

define void @too_large() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 65537)
  store i8 1, ptr %p
  ret void
}

define void @at_limit() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 65536)
  store i8 1, ptr %p
  ret void
}

define void @zero_size() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 0)
  %same = icmp eq ptr %p, null
  ret void
}

define void @excessive_alignment() {
entry:
  %p = call align 131072 ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i8 1, ptr %p
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	for _, name := range []string{"alloc_z", "alloc_u", "at_limit"} {
		requireStackAllocation(t, mod, name)
	}
	for _, name := range []string{"escaping", "loop", "dynamic_size", "too_large", "zero_size", "excessive_alignment"} {
		requireHeapAllocation(t, mod, name)
	}
	zeroed := mod.NamedFunction("alloc_z").String()
	if !strings.Contains(zeroed, "alloca i8, i64 8, align 16") || !strings.Contains(zeroed, "call void @llvm.memset") {
		t.Fatalf("AllocZ initialization/alignment was not preserved:\n%s", zeroed)
	}
	uninitialized := mod.NamedFunction("alloc_u").String()
	if strings.Contains(uninitialized, "llvm.memset") {
		t.Fatalf("AllocU was unexpectedly initialized:\n%s", uninitialized)
	}
}

func TestAllocaAddressSpaceMismatchRejectsRewrite(t *testing.T) {
	ir := `
target datalayout = "e-p:64:64-A5"
` + allocationDeclarations + `

define void @wrong_address_space() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i8 1, ptr %p
  ret void
}
`
	ctx, mod := transformModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	requireHeapAllocation(t, mod, "wrong_address_space")
}

func TestTransformPrecedesLargeAggregateAndCABILowering(t *testing.T) {
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	ir := `
%Large = type [65537 x i8]
` + allocationDeclarations + `

define void @"pkg.local"() {
entry:
  %p = call ptr @"github.com/goplus/llgo/runtime/internal/runtime.AllocU"(i64 8)
  store i8 1, ptr %p
  ret void
}

define %Large @"pkg.large"(ptr %src) {
entry:
  %value = load %Large, ptr %src, align 1
  ret %Large %value
}
`
	ctx, mod := parseModule(t, ir)
	defer ctx.Dispose()
	defer mod.Dispose()
	prog := ssa.NewProgram(&ssa.Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH})
	defer prog.Dispose()
	mod.SetDataLayout(prog.DataLayout())
	mod.SetTarget(prog.Target().Spec().Triple)

	if err := TransformModule(mod); err != nil {
		t.Fatalf("escape transform: %v", err)
	}
	requireStackAllocation(t, mod, "pkg.local")
	llabi.LowerLargeAggregates(prog.TargetData(), mod)
	cabi.NewTransformer(prog, "", "", cabi.ModeAllFunc, false).TransformModule("pkg", mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatalf("module invalid after escape, large aggregate, and C ABI transforms: %v\n%s", err, mod.String())
	}
}
