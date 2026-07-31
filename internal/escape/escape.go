// Package escape implements LLGo's LLVM-IR escape analysis and heap-to-stack
// transform.
package escape

import (
	"fmt"
	"strings"

	llabi "github.com/goplus/llgo/internal/abi"
	"github.com/xgo-dev/llvm"
)

const (
	runtimePrefix = "github.com/goplus/llgo/runtime/internal/runtime."
	runtimeAllocZ = runtimePrefix + "AllocZ"
	runtimeAllocU = runtimePrefix + "AllocU"

	maxDereferenceLevel = 16

	// github.com/xgo-dev/llvm does not expose these LLVM 19 C opcode names.
	opcodeAtomicCmpXchg = llvm.Opcode(56)
	opcodeAtomicRMW     = llvm.Opcode(57)
	opcodeAddrSpaceCast = llvm.Opcode(60)
	opcodeFreeze        = llvm.Opcode(68)
)

type summaryKey struct {
	fn    llvm.Value
	param int
}

type summaryState uint8

const (
	summaryUnseen summaryState = iota
	summaryEvaluating
	summaryEvaluated
)

type parameterSummary struct {
	heapLevel    int
	mutatorLevel int
	calleeLevel  int
	results      map[int]int
	alignments   map[int]int
}

func newParameterSummary() parameterSummary {
	return parameterSummary{
		heapLevel:    -1,
		mutatorLevel: -1,
		calleeLevel:  -1,
		results:      make(map[int]int),
		alignments:   make(map[int]int),
	}
}

func minLevel(dst *int, level int) bool {
	if *dst >= 0 && *dst <= level {
		return false
	}
	*dst = level
	return true
}

func (s *parameterSummary) addHeap(level int) bool {
	return minLevel(&s.heapLevel, level)
}

func (s *parameterSummary) addMutator(level int) bool {
	return minLevel(&s.mutatorLevel, level)
}

func (s *parameterSummary) addCallee(level int) bool {
	return minLevel(&s.calleeLevel, level)
}

func (s *parameterSummary) addResult(index, level int) bool {
	old, ok := s.results[index]
	if ok && old <= level {
		return false
	}
	s.results[index] = level
	return true
}

func (s *parameterSummary) addAlignment(level, align int) bool {
	if align <= s.alignments[level] {
		return false
	}
	s.alignments[level] = align
	return true
}

func (s *parameterSummary) join(src parameterSummary) bool {
	changed := false
	if src.heapLevel >= 0 {
		changed = s.addHeap(src.heapLevel) || changed
	}
	if src.mutatorLevel >= 0 {
		changed = s.addMutator(src.mutatorLevel) || changed
	}
	if src.calleeLevel >= 0 {
		changed = s.addCallee(src.calleeLevel) || changed
	}
	for index, level := range src.results {
		changed = s.addResult(index, level) || changed
	}
	for level, align := range src.alignments {
		changed = s.addAlignment(level, align) || changed
	}
	return changed
}

type summaryEntry struct {
	state      summaryState
	summary    parameterSummary
	dependents map[summaryKey]struct{}
}

type analyzer struct {
	mod llvm.Module

	summaries   map[summaryKey]*summaryEntry
	summaryKeys []summaryKey
	pending     []summaryKey
	pendingSet  map[summaryKey]bool
}

func newAnalyzer(mod llvm.Module) *analyzer {
	a := &analyzer{
		mod:        mod,
		summaries:  make(map[summaryKey]*summaryEntry),
		pendingSet: make(map[summaryKey]bool),
	}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() || isRuntimeFunction(fn) {
			continue
		}
		for index, param := range fn.Params() {
			if !isPointer(param.Type()) {
				continue
			}
			key := summaryKey{fn: fn, param: index}
			a.summaries[key] = &summaryEntry{
				summary:    newParameterSummary(),
				dependents: make(map[summaryKey]struct{}),
			}
			a.summaryKeys = append(a.summaryKeys, key)
		}
	}
	return a
}

func isRuntimeFunction(fn llvm.Value) bool {
	return strings.HasPrefix(fn.Name(), runtimePrefix)
}

func isPointer(typ llvm.Type) bool {
	return typ.TypeKind() == llvm.PointerTypeKind
}

func (a *analyzer) solveSummaries() {
	for _, key := range a.summaryKeys {
		a.evaluateSummary(key)
	}
	for len(a.pending) != 0 {
		key := a.pending[0]
		a.pending = a.pending[1:]
		delete(a.pendingSet, key)
		a.evaluateSummary(key)
	}
}

func (a *analyzer) enqueueSummary(key summaryKey) {
	if a.pendingSet[key] {
		return
	}
	a.pendingSet[key] = true
	a.pending = append(a.pending, key)
}

func (a *analyzer) evaluateSummary(key summaryKey) {
	entry := a.summaries[key]
	if entry == nil || entry.state == summaryEvaluating {
		return
	}
	entry.state = summaryEvaluating
	result := a.walk(key.fn.Param(key.param), true, key)
	entry.state = summaryEvaluated
	if entry.summary.join(result.summary) {
		for dependent := range entry.dependents {
			a.enqueueSummary(dependent)
		}
	}
}

func (a *analyzer) summaryFor(key, consumer summaryKey) (parameterSummary, bool) {
	entry := a.summaries[key]
	if entry == nil {
		return parameterSummary{}, false
	}
	entry.dependents[consumer] = struct{}{}
	if entry.state == summaryUnseen {
		a.evaluateSummary(key)
	}
	return entry.summary, true
}

type useState struct {
	user    llvm.Value
	operand int
	derefs  int
}

type walkResult struct {
	escaped   bool
	alignment int
	summary   parameterSummary
}

func newWalkResult(parameter bool) walkResult {
	result := walkResult{}
	if parameter {
		result.summary = newParameterSummary()
	}
	return result
}

func addUses(worklist *[]useState, value llvm.Value, derefs int) {
	for use := value.FirstUse(); !use.IsNil(); use = use.NextUse() {
		user := use.User()
		for operand := 0; operand < user.OperandsCount(); operand++ {
			if user.Operand(operand) == value {
				*worklist = append(*worklist, useState{user: user, operand: operand, derefs: derefs})
			}
		}
	}
}

func (a *analyzer) walk(root llvm.Value, parameter bool, consumer summaryKey) walkResult {
	result := newWalkResult(parameter)
	worklist := make([]useState, 0, 8)
	addUses(&worklist, root, 0)
	visited := make(map[useState]struct{})

	for len(worklist) != 0 {
		state := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if _, ok := visited[state]; ok {
			continue
		}
		visited[state] = struct{}{}
		if state.derefs > maxDereferenceLevel {
			a.recordEscape(&result, parameter, maxDereferenceLevel)
			continue
		}

		user := state.user
		opcode := user.InstructionOpcode()
		switch opcode {
		case llvm.Load:
			a.recordAddressUse(&result, parameter, state.derefs, user.Alignment(), false)
			if parameter && isPointer(user.Type()) {
				addUses(&worklist, user, state.derefs+1)
			}
		case llvm.Store:
			switch state.operand {
			case 0:
				copies, ok := exactLocalCopies(user)
				if !ok {
					a.recordEscape(&result, parameter, state.derefs)
					continue
				}
				for _, copy := range copies {
					addUses(&worklist, copy, state.derefs)
				}
			case 1:
				a.recordAddressUse(&result, parameter, state.derefs, user.Alignment(), true)
			default:
				a.recordEscape(&result, parameter, state.derefs)
			}
		case llvm.GetElementPtr:
			if state.operand == 0 {
				addUses(&worklist, user, state.derefs)
			} else {
				a.recordEscape(&result, parameter, state.derefs)
			}
		case llvm.BitCast, llvm.PHI:
			addUses(&worklist, user, state.derefs)
		case llvm.Select:
			if state.operand == 1 || state.operand == 2 {
				addUses(&worklist, user, state.derefs)
			} else {
				a.recordEscape(&result, parameter, state.derefs)
			}
		case opcodeFreeze:
			if state.operand == 0 {
				addUses(&worklist, user, state.derefs)
			} else {
				a.recordEscape(&result, parameter, state.derefs)
			}
		case llvm.ICmp:
			// A comparison observes pointer identity but does not retain it.
		case llvm.Ret:
			if parameter && state.operand == 0 {
				result.summary.addResult(0, state.derefs)
			} else {
				a.recordEscape(&result, parameter, state.derefs)
			}
		case llvm.Call, llvm.Invoke:
			a.handleCall(&result, &worklist, state, parameter, consumer)
		case opcodeAtomicCmpXchg:
			switch state.operand {
			case 0:
				a.recordAddressUse(&result, parameter, state.derefs, user.Alignment(), true)
			case 1:
				// Comparing the expected value does not publish it.
			default:
				a.recordEscape(&result, parameter, state.derefs)
			}
		case opcodeAtomicRMW:
			if state.operand == 0 {
				a.recordAddressUse(&result, parameter, state.derefs, user.Alignment(), true)
			} else {
				a.recordEscape(&result, parameter, state.derefs)
			}
		case llvm.PtrToInt, llvm.IntToPtr, opcodeAddrSpaceCast:
			a.recordEscape(&result, parameter, state.derefs)
		default:
			a.recordEscape(&result, parameter, state.derefs)
		}
	}
	return result
}

func (a *analyzer) recordEscape(result *walkResult, parameter bool, level int) {
	if parameter {
		result.summary.addHeap(level)
		return
	}
	result.escaped = true
}

func (a *analyzer) recordAddressUse(result *walkResult, parameter bool, level, align int, mutator bool) {
	if parameter {
		result.summary.addAlignment(level, align)
		if mutator {
			result.summary.addMutator(level)
		}
		return
	}
	if align > result.alignment {
		result.alignment = align
	}
}

func exactLocalCopies(store llvm.Value) ([]llvm.Value, bool) {
	if store.IsVolatile() || store.Ordering() != llvm.AtomicOrderingNotAtomic {
		return nil, false
	}
	slot := store.Operand(1).IsAAllocaInst()
	if slot.IsNil() {
		return nil, false
	}

	stores := 0
	var copies []llvm.Value
	for use := slot.FirstUse(); !use.IsNil(); use = use.NextUse() {
		user := use.User()
		switch {
		case !user.IsAStoreInst().IsNil() && user.Operand(1) == slot:
			stores++
			if user != store || user.IsVolatile() || user.Ordering() != llvm.AtomicOrderingNotAtomic {
				return nil, false
			}
		case !user.IsALoadInst().IsNil() && user.Operand(0) == slot:
			if user.IsVolatile() || user.Ordering() != llvm.AtomicOrderingNotAtomic || !isPointer(user.Type()) {
				return nil, false
			}
			copies = append(copies, user)
		default:
			return nil, false
		}
	}
	return copies, stores == 1
}

func callCalleeOperand(call llvm.Value) int {
	called := call.CalledValue()
	index := -1
	for operand := 0; operand < call.OperandsCount(); operand++ {
		if call.Operand(operand) == called {
			index = operand
		}
	}
	return index
}

func (a *analyzer) handleCall(result *walkResult, worklist *[]useState, state useState, parameter bool, consumer summaryKey) {
	call := state.user
	calleeOperand := callCalleeOperand(call)
	if state.operand == calleeOperand {
		if parameter {
			result.summary.addCallee(state.derefs)
		} else {
			result.escaped = true
		}
		return
	}

	callee := call.CalledValue()
	if !callee.IsAInlineAsm().IsNil() {
		a.recordEscape(result, parameter, state.derefs)
		return
	}
	fn := callee.IsAFunction()
	if fn.IsNil() {
		a.recordEscape(result, parameter, state.derefs)
		return
	}
	if fn.IntrinsicID() != 0 {
		if !a.handleIntrinsic(result, state, parameter) {
			a.recordEscape(result, parameter, state.derefs)
		}
		return
	}
	if fn.IsDeclaration() || isRuntimeFunction(fn) || state.operand >= fn.ParamsCount() {
		a.recordEscape(result, parameter, state.derefs)
		return
	}

	calleeSummary, ok := a.summaryFor(summaryKey{fn: fn, param: state.operand}, consumer)
	if !ok {
		a.recordEscape(result, parameter, state.derefs)
		return
	}
	a.composeSummary(result, worklist, call, state.derefs, parameter, calleeSummary)
}

func (a *analyzer) handleIntrinsic(result *walkResult, state useState, parameter bool) bool {
	name := state.user.CalledValue().Name()
	switch {
	case strings.HasPrefix(name, "llvm.dbg."),
		strings.HasPrefix(name, "llvm.lifetime.start."),
		strings.HasPrefix(name, "llvm.lifetime.end."),
		strings.HasPrefix(name, "llvm.objectsize."):
		return true
	case strings.HasPrefix(name, "llvm.memset."):
		if state.operand == 0 {
			a.recordAddressUse(result, parameter, state.derefs, 0, true)
			return true
		}
	case strings.HasPrefix(name, "llvm.memcpy."), strings.HasPrefix(name, "llvm.memmove."):
		if state.operand == 0 {
			a.recordAddressUse(result, parameter, state.derefs, 0, true)
			return true
		}
		if state.operand == 1 {
			return true
		}
	}
	return false
}

func (a *analyzer) composeSummary(result *walkResult, worklist *[]useState, call llvm.Value, level int, parameter bool, summary parameterSummary) {
	if summary.heapLevel >= 0 {
		if parameter {
			result.summary.addHeap(level + summary.heapLevel)
		} else if summary.heapLevel == 0 {
			result.escaped = true
		}
	}
	if parameter {
		if summary.mutatorLevel >= 0 {
			result.summary.addMutator(level + summary.mutatorLevel)
		}
		if summary.calleeLevel >= 0 {
			result.summary.addCallee(level + summary.calleeLevel)
		}
		for derefs, align := range summary.alignments {
			result.summary.addAlignment(level+derefs, align)
		}
	} else if align := summary.alignments[0]; align > result.alignment {
		result.alignment = align
	}

	for index, derefs := range summary.results {
		if index != 0 || !isPointer(call.Type()) {
			a.recordEscape(result, parameter, level)
			continue
		}
		if !parameter && derefs != 0 {
			continue
		}
		addUses(worklist, call, level+derefs)
	}
}

type allocationPlan struct {
	call      llvm.Value
	size      llvm.Value
	zero      bool
	alignment int
}

// TransformModule analyzes eligible LLGo allocations, rewrites proven-local
// AllocZ and AllocU calls, and verifies the resulting LLVM module.
func TransformModule(mod llvm.Module) error {
	a := newAnalyzer(mod)
	a.solveSummaries()

	var plans []allocationPlan
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() {
			continue
		}
		for block := fn.FirstBasicBlock(); !block.IsNil(); block = llvm.NextBasicBlock(block) {
			for instr := block.FirstInstruction(); !instr.IsNil(); instr = llvm.NextInstruction(instr) {
				plan, ok := a.planAllocation(instr)
				if ok {
					plans = append(plans, plan)
				}
			}
		}
	}

	for _, plan := range plans {
		rewriteAllocation(mod.Context(), plan)
	}
	if len(plans) != 0 {
		if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
			return fmt.Errorf("verify heap-to-stack transform: %w", err)
		}
	}
	return nil
}

func (a *analyzer) planAllocation(instr llvm.Value) (allocationPlan, bool) {
	call := instr.IsACallInst()
	if call.IsNil() || call.OperandsCount() < 2 || !isPointer(call.Type()) {
		return allocationPlan{}, false
	}
	callee := call.CalledValue().IsAFunction()
	if callee.IsNil() {
		return allocationPlan{}, false
	}
	name := callee.Name()
	if name != runtimeAllocZ && name != runtimeAllocU {
		return allocationPlan{}, false
	}
	size := call.Operand(0).IsAConstantInt()
	if size.IsNil() {
		return allocationPlan{}, false
	}
	sizeBytes := size.ZExtValue()
	if sizeBytes == 0 || sizeBytes > llabi.MaxImplicitStackVarSize || blockInCycle(call.InstructionParent()) {
		return allocationPlan{}, false
	}

	result := a.walk(call, false, summaryKey{})
	if result.escaped {
		return allocationPlan{}, false
	}
	alignment := callResultAlignment(call, callee)
	if result.alignment > alignment {
		alignment = result.alignment
	}
	if alignment == 0 {
		alignment = 1
	}
	if alignment&(alignment-1) != 0 || alignment > int(llabi.MaxImplicitStackVarSize) {
		return allocationPlan{}, false
	}
	return allocationPlan{call: call, size: size, zero: name == runtimeAllocZ, alignment: alignment}, true
}

func callResultAlignment(call, callee llvm.Value) int {
	kind := llvm.AttributeKindID("align")
	align := 0
	if attr := call.GetCallSiteEnumAttribute(0, kind); !attr.IsNil() {
		align = int(attr.GetEnumValue())
	}
	if attr := callee.GetEnumAttributeAtIndex(0, kind); !attr.IsNil() && int(attr.GetEnumValue()) > align {
		align = int(attr.GetEnumValue())
	}
	return align
}

func blockInCycle(start llvm.BasicBlock) bool {
	visited := make(map[llvm.BasicBlock]bool)
	var visit func(llvm.BasicBlock) bool
	visit = func(block llvm.BasicBlock) bool {
		if block == start {
			return true
		}
		if visited[block] {
			return false
		}
		visited[block] = true
		terminator := block.LastInstruction()
		for index := 0; index < terminator.SuccessorsCount(); index++ {
			if visit(terminator.Successor(index)) {
				return true
			}
		}
		return false
	}

	terminator := start.LastInstruction()
	for index := 0; index < terminator.SuccessorsCount(); index++ {
		if visit(terminator.Successor(index)) {
			return true
		}
	}
	return false
}

func rewriteAllocation(ctx llvm.Context, plan allocationPlan) bool {
	builder := ctx.NewBuilder()
	defer builder.Dispose()
	builder.SetInsertPointBefore(plan.call)
	stack := llvm.CreateArrayAlloca(builder, ctx.Int8Type(), plan.size)
	if stack.Type().PointerAddressSpace() != plan.call.Type().PointerAddressSpace() {
		stack.EraseFromParentAsInstruction()
		return false
	}
	stack.SetName(plan.call.Name() + ".stack")
	stack.SetAlignment(plan.alignment)
	if plan.zero {
		builder.CreateIntrinsic(ctx.VoidType(), llvm.LookupIntrinsicID("llvm.memset"), []llvm.Value{
			stack,
			llvm.ConstInt(ctx.Int8Type(), 0, false),
			plan.size,
			llvm.ConstInt(ctx.Int1Type(), 0, false),
		}, "")
	}
	plan.call.ReplaceAllUsesWith(stack)
	plan.call.EraseFromParentAsInstruction()
	return true
}
