//go:build !llgo
// +build !llgo

package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModuleHookReceivesMainPackageModule(t *testing.T) {
	conf := NewDefaultConf(ModeGen)

	counts := make(map[string]int)
	snapshots := make(map[string]string)
	conf.ModuleHook = func(pkg Package) {
		counts[pkg.PkgPath]++
		if _, ok := snapshots[pkg.PkgPath]; !ok {
			snapshots[pkg.PkgPath] = pkg.LPkg.String()
		}
	}

	pkgs, err := Do([]string{"../../cl/_testgo/print"}, conf)
	if err != nil {
		t.Fatalf("Do failed: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected 1 initial package, got %d", len(pkgs))
	}

	mainPkg := pkgs[0].PkgPath
	if counts[mainPkg] != 1 {
		t.Fatalf("expected hook to fire once for %s, got %d", mainPkg, counts[mainPkg])
	}
	if snapshots[mainPkg] == "" {
		t.Fatalf("expected non-empty module snapshot for %s", mainPkg)
	}
}

func TestNoBoundsConfigControlsGeneratedIR(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "case.go")
	source := "package p\ntype A [0]byte\nfunc Get(a *A, i int) byte { return a[i] }\n"
	if err := os.WriteFile(goFile, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	buildIR := func(noBounds bool) string {
		t.Helper()
		conf := NewDefaultConf(ModeGen)
		conf.NoBounds = noBounds
		var ir string
		conf.ModuleHook = func(pkg Package) {
			if pkg.PkgPath == "command-line-arguments" {
				ir = pkg.LPkg.String()
			}
		}
		pkgs, err := Do([]string{goFile}, conf)
		if len(pkgs) != 0 && pkgs[0].LPkg != nil {
			defer pkgs[0].LPkg.Prog.Dispose()
		}
		if err != nil {
			t.Fatalf("Do(NoBounds=%v): %v", noBounds, err)
		}
		if ir == "" {
			t.Fatalf("Do(NoBounds=%v) produced no module snapshot", noBounds)
		}
		return ir
	}

	checked := buildIR(false)
	if !strings.Contains(checked, "CheckIndexRange") {
		t.Fatalf("default build did not emit checked index IR:\n%s", checked)
	}
	unchecked := buildIR(true)
	if strings.Contains(unchecked, "CheckIndexRange") {
		t.Fatalf("NoBounds build did not emit unchecked index IR:\n%s", unchecked)
	}
}
