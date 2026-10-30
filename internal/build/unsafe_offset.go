package build

import (
	"go/ast"
	"go/types"

	"github.com/xgo-dev/llgo/internal/packages"
)

// unsafeOffsetOwner bridges the context omitted by types.Sizes.Offsetsof.
//
// Identity: for `type G struct{F func(); Tail uintptr}; type C G`, G and C may
// share every field object, while cdecl C stores F in one word instead of two.
// Selection.Recv retains the actual owner of unsafe.Offsetof(c.Tail).
//
// Timing: go/types records a fresh Selection before each Offsetof evaluation.
// A repeated const initializer reuses its AST, so track Selection identity,
// not whether the expression already has a constant in Info.Types. Promoted
// fields issue one callback for each struct along Selection.Index.
//
// Lifetime: this state belongs to one checker and is rebuilt on typed rewrite.
// It does not change the AST, field objects, or shared Program type metadata.
func unsafeOffsetOwner(pkg *packages.Package) func([]*types.Var) types.Type {
	type query struct {
		selector  *ast.SelectorExpr
		callee    *ast.Ident
		selection *types.Selection
		next      int
	}
	var queries []query
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			selector, ok := ast.Unparen(call.Args[0]).(*ast.SelectorExpr)
			if !ok {
				return true
			}
			var callee *ast.Ident
			switch fun := ast.Unparen(call.Fun).(type) {
			case *ast.SelectorExpr:
				callee = fun.Sel
			case *ast.Ident:
				callee = fun
			}
			if callee != nil && callee.Name == "Offsetof" {
				queries = append(queries, query{selector: selector, callee: callee})
			}
			return true
		})
	}
	if len(queries) == 0 {
		return nil
	}
	info, owner := pkg.TypesInfo, pkg.Types
	builtin := types.Unsafe.Scope().Lookup("Offsetof")
	return func(fields []*types.Var) types.Type {
		if owner.Complete() {
			return nil
		}
		match := -1
		var result types.Type
		for i := range queries {
			q := &queries[i]
			selection := info.Selections[q.selector]
			if selection == nil || info.ObjectOf(q.callee) != builtin {
				continue
			}
			if selection != q.selection {
				q.selection, q.next = selection, 0
			}
			index := selection.Index()
			if q.next == len(index) {
				continue
			}
			current := selection.Recv()
			for _, field := range index[:q.next] {
				current = current.Underlying().(*types.Struct).Field(field).Type()
			}
			st := current.Underlying().(*types.Struct)
			if st.NumFields() != len(fields) {
				continue
			}
			same := true
			for j, field := range fields {
				if st.Field(j) != field {
					same = false
					break
				}
			}
			if !same {
				continue
			}
			if match != -1 {
				panic("unsafe.Offsetof: multiple active selections for the same fields")
			}
			match, result = i, current
		}
		if match != -1 {
			queries[match].next++
		}
		return result
	}
}
