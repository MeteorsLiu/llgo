//go:build !llgo

package build

import (
	"strings"
	"testing"

	"github.com/xgo-dev/llvm"
)

func TestEscapePhase1RunsAfterModuleHook(t *testing.T) {
	conf := NewDefaultConf(ModeGen)
	var before string
	conf.ModuleHook = func(pkg Package) {
		if fn := findLLVMFunction(pkg.LPkg.Module(), ".Local"); !fn.IsNil() {
			before = fn.String()
		}
	}
	pkgs, err := Do([]string{"./testdata/escape_phase1"}, conf)
	if err != nil {
		t.Fatalf("generate escape Phase 1 fixture: %v", err)
	}
	if len(pkgs) != 1 || pkgs[0].LPkg == nil {
		t.Fatalf("generated packages = %#v", pkgs)
	}
	defer pkgs[0].LPkg.Prog.Dispose()

	if !strings.Contains(before, "runtime.AllocZ") {
		t.Fatalf("pre-transform Local has no heap allocation:\n%s", before)
	}
	after := findLLVMFunction(pkgs[0].LPkg.Module(), ".Local").String()
	if strings.Contains(after, "runtime.AllocZ") || !strings.Contains(after, "alloca i8") {
		t.Fatalf("post-transform Local was not moved to the stack:\n%s", after)
	}
}

func findLLVMFunction(mod llvm.Module, suffix string) llvm.Value {
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if strings.HasSuffix(fn.Name(), suffix) {
			return fn
		}
	}
	return llvm.Value{}
}
