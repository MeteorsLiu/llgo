// Package escape implements LLGo's LLVM-IR escape analysis and heap-to-stack
// transform.
package escape

import (
	"strings"

	llabi "github.com/goplus/llgo/internal/abi"
	"github.com/xgo-dev/llvm"
)

const (
	runtimePrefix = "github.com/goplus/llgo/runtime/internal/runtime."
	runtimeAllocZ = runtimePrefix + "AllocZ"
	runtimeAllocU = runtimePrefix + "AllocU"

	// github.com/xgo-dev/llvm does not expose these LLVM 19 C opcode names.
	opcodeAtomicCmpXchg = llvm.Opcode(56)
	opcodeAtomicRMW     = llvm.Opcode(57)
	opcodeAddrSpaceCast = llvm.Opcode(60)
	opcodeFreeze        = llvm.Opcode(68)
)

type paramKey struct {
	fn    llvm.Value
	param int
}

type locationEdge struct {
	from    llvm.Value
	to      llvm.Value
	operand int
}

// locationGraph records diagnostic pointer flows but is never queried by the
// optimization analyses.
type locationGraph struct {
	edges map[locationEdge]struct{}
}

func newLocationGraph() locationGraph {
	return locationGraph{edges: make(map[locationEdge]struct{})}
}

func (g *locationGraph) addUse(from llvm.Value, state useState) {
	g.edges[locationEdge{from: from, to: state.user, operand: state.operand}] = struct{}{}
}

func (g *locationGraph) addFlow(from, to llvm.Value) {
	g.edges[locationEdge{from: from, to: to, operand: -1}] = struct{}{}
}

type analyzer struct {
	noCapture     map[paramKey]bool
	noCaptureKeys []paramKey
	locations     locationGraph
}

func newAnalyzer(mod llvm.Module) *analyzer {
	a := &analyzer{
		noCapture: make(map[paramKey]bool),
		locations: newLocationGraph(),
	}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() || isRuntimeFunction(fn) {
			continue
		}
		for index, param := range fn.Params() {
			if !isPointer(param.Type()) {
				continue
			}
			key := paramKey{fn: fn, param: index}
			a.noCapture[key] = true
			a.noCaptureKeys = append(a.noCaptureKeys, key)
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

func (a *analyzer) solveNoCapture() {
	for {
		changed := false
		for _, key := range a.noCaptureKeys {
			if a.noCapture[key] && !a.parameterNoCapture(key) {
				a.noCapture[key] = false
				changed = true
			}
		}
		if !changed {
			return
		}
	}
}

type useState struct {
	user    llvm.Value
	operand int
}

type walkResult struct {
	escaped   bool
	alignment int
}

func (a *analyzer) addUses(worklist *[]useState, value llvm.Value) {
	for use := value.FirstUse(); !use.IsNil(); use = use.NextUse() {
		user := use.User()
		for operand := 0; operand < user.OperandsCount(); operand++ {
			if user.Operand(operand) == value {
				state := useState{user: user, operand: operand}
				*worklist = append(*worklist, state)
				a.locations.addUse(value, state)
			}
		}
	}
}

type useAction uint8

const (
	useSafe useAction = iota
	useFollow
	useCapture
)

func (a *analyzer) checkForAllUses(root llvm.Value, classify func(useState) useAction) bool {
	worklist := make([]useState, 0, 8)
	a.addUses(&worklist, root)
	visited := make(map[useState]struct{})
	valid := true

	for len(worklist) != 0 {
		state := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if _, ok := visited[state]; ok {
			continue
		}
		visited[state] = struct{}{}

		if state.user.InstructionOpcode() == llvm.Store && state.operand == 0 {
			copies, ok := exactLocalCopies(state.user)
			if ok {
				for _, copy := range copies {
					a.locations.addFlow(state.user.Operand(0), copy)
					a.addUses(&worklist, copy)
				}
				continue
			}
		}

		switch classify(state) {
		case useFollow:
			a.addUses(&worklist, state.user)
		case useCapture:
			valid = false
		}
	}
	return valid
}

func (a *analyzer) parameterNoCapture(key paramKey) bool {
	return a.checkForAllUses(key.fn.Param(key.param), a.classifyNoCaptureUse)
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

func definedCallParameter(call llvm.Value, argument int) (paramKey, bool) {
	fn := call.CalledValue().IsAFunction()
	if fn.IsNil() || fn.IntrinsicID() != 0 || fn.IsDeclaration() || isRuntimeFunction(fn) || argument >= fn.ParamsCount() {
		return paramKey{}, false
	}
	return paramKey{fn: fn, param: argument}, true
}

func (a *analyzer) classifyNoCaptureUse(state useState) useAction {
	user := state.user
	switch user.InstructionOpcode() {
	case llvm.Call, llvm.Invoke:
		if a.callArgumentNoCapture(state) {
			return useSafe
		}
	case llvm.Load:
		return useSafe
	case llvm.Store:
		if state.operand == 1 {
			return useSafe
		}
	case opcodeAtomicRMW:
		if state.operand == 0 {
			return useSafe
		}
	case opcodeAtomicCmpXchg:
		if state.operand == 0 {
			return useSafe
		}
	case llvm.VAArg:
		return useSafe
	case llvm.GetElementPtr:
		if state.operand == 0 && user.Type().TypeKind() != llvm.VectorTypeKind {
			return useFollow
		}
	case llvm.BitCast, llvm.PHI, opcodeAddrSpaceCast:
		return useFollow
	case llvm.Select:
		if state.operand == 1 || state.operand == 2 {
			return useFollow
		}
	case opcodeFreeze:
		if state.operand == 0 {
			return useFollow
		}
	}
	return useCapture
}

func (a *analyzer) requiredAlignment(root paramKey) int {
	alignment := 0
	worklist := []paramKey{root}
	visited := make(map[paramKey]struct{})

	for len(worklist) != 0 {
		key := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		if _, ok := visited[key]; ok {
			continue
		}
		visited[key] = struct{}{}

		a.checkForAllUses(key.fn.Param(key.param), func(state useState) useAction {
			user := state.user
			switch user.InstructionOpcode() {
			case llvm.Load:
				if user.Alignment() > alignment {
					alignment = user.Alignment()
				}
			case llvm.Store:
				if state.operand == 1 && user.Alignment() > alignment {
					alignment = user.Alignment()
				}
			case llvm.Call, llvm.Invoke:
				if callee, ok := definedCallParameter(user, state.operand); ok && a.noCapture[callee] {
					worklist = append(worklist, callee)
				}
			case llvm.GetElementPtr:
				if state.operand == 0 {
					return useFollow
				}
			case llvm.BitCast, llvm.PHI, opcodeAddrSpaceCast:
				return useFollow
			case llvm.Select:
				if state.operand == 1 || state.operand == 2 {
					return useFollow
				}
			case opcodeFreeze:
				if state.operand == 0 {
					return useFollow
				}
			case opcodeAtomicCmpXchg, opcodeAtomicRMW:
				if state.operand == 0 && user.Alignment() > alignment {
					alignment = user.Alignment()
				}
			}
			return useSafe
		})
	}
	return alignment
}

func (a *analyzer) allocationUses(root llvm.Value) walkResult {
	result := walkResult{}
	result.escaped = !a.checkForAllUses(root, func(state useState) useAction {
		user := state.user
		switch user.InstructionOpcode() {
		case llvm.Load:
			if user.Alignment() > result.alignment {
				result.alignment = user.Alignment()
			}
			return useSafe
		case llvm.Store:
			if state.operand == 1 {
				if user.Alignment() > result.alignment {
					result.alignment = user.Alignment()
				}
				return useSafe
			}
		case llvm.Call, llvm.Invoke:
			if a.callArgumentNoCapture(state) {
				if key, ok := definedCallParameter(user, state.operand); ok {
					if alignment := a.requiredAlignment(key); alignment > result.alignment {
						result.alignment = alignment
					}
				}
				return useSafe
			}
		case llvm.GetElementPtr:
			if state.operand == 0 {
				return useFollow
			}
		case llvm.BitCast, llvm.PHI:
			return useFollow
		case llvm.Select:
			if state.operand == 1 || state.operand == 2 {
				return useFollow
			}
		case opcodeFreeze:
			if state.operand == 0 {
				return useFollow
			}
		case opcodeAtomicCmpXchg:
			if state.operand == 0 {
				if user.Alignment() > result.alignment {
					result.alignment = user.Alignment()
				}
				return useSafe
			}
		case opcodeAtomicRMW:
			if state.operand == 0 {
				if user.Alignment() > result.alignment {
					result.alignment = user.Alignment()
				}
				return useSafe
			}
		}
		return useCapture
	})
	return result
}

func (a *analyzer) callArgumentNoCapture(state useState) bool {
	call := state.user
	calleeOperand := callCalleeOperand(call)
	if state.operand == calleeOperand {
		return true
	}

	callee := call.CalledValue()
	if !callee.IsAInlineAsm().IsNil() {
		return false
	}
	fn := callee.IsAFunction()
	if fn.IsNil() {
		return false
	}
	if fn.IntrinsicID() != 0 {
		return intrinsicArgumentNoCapture(call, fn, state.operand)
	}
	key, ok := definedCallParameter(call, state.operand)
	if !ok {
		return false
	}

	a.locations.addFlow(call.Operand(state.operand), key.fn.Param(key.param))
	return a.noCapture[key]
}

func intrinsicArgumentNoCapture(call, fn llvm.Value, argument int) bool {
	name := fn.Name()
	if argument < 0 {
		return false
	}
	if strings.HasPrefix(name, "llvm.dbg.") ||
		strings.HasPrefix(name, "llvm.lifetime.start.") ||
		strings.HasPrefix(name, "llvm.lifetime.end.") ||
		strings.HasPrefix(name, "llvm.objectsize.") {
		return true
	}
	kind := llvm.AttributeKindID("nocapture")
	attributeIndex := argument + 1
	if attr := call.GetCallSiteEnumAttribute(attributeIndex, kind); !attr.IsNil() {
		return true
	}
	return !fn.GetEnumAttributeAtIndex(attributeIndex, kind).IsNil()
}

type allocationPlan struct {
	call      llvm.Value
	size      llvm.Value
	zero      bool
	alignment int
}

// TransformModule analyzes eligible LLGo allocations and rewrites proven-local
// AllocZ and AllocU calls.
func TransformModule(mod llvm.Module) {
	a := newAnalyzer(mod)
	a.solveNoCapture()

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

	result := a.allocationUses(call)
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
