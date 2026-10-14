// Command layout compares the existing runtime debugger fixture with Go.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type runConfig struct {
	llgo, llgen, adapter, toolDir, root, output string
	goos, goarch, goversion                     string
	levels                                      []string
}
type caseResult struct {
	Case        string       `json:"case"`
	Level       string       `json:"level,omitempty"`
	Status      string       `json:"status"`
	Reason      string       `json:"reason,omitempty"`
	Directory   string       `json:"directory,omitempty"`
	Differences []difference `json:"differences,omitempty"`
}
type runReport struct {
	GoVersion string       `json:"go_version"`
	GOOS      string       `json:"goos"`
	GOARCH    string       `json:"goarch"`
	Compiler  string       `json:"llgo_version"`
	Results   []caseResult `json:"results"`
}

func main() {
	var config runConfig
	var levels string
	flag.StringVar(&config.llgo, "llgo", "llgo", "LLGo compiler executable")
	flag.StringVar(&config.llgen, "llgen", "", "llgen executable (default: build chore/llgen)")
	flag.StringVar(&config.adapter, "lldb-dap", "lldb-dap", "native LLDB debug adapter executable")
	flag.StringVar(&config.root, "root", os.Getenv("LLGO_ROOT"), "repository containing the existing debugger fixtures")
	flag.StringVar(&config.output, "output", "", "new artifact directory (default: temporary directory)")
	flag.StringVar(&levels, "opt", "0,2", "comma-separated LLGo optimization levels: 0,2")
	flag.Parse()
	if err := run(config, levels); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(config runConfig, levels string) error {
	if config.root == "" {
		_, source, _, ok := runtime.Caller(0)
		if !ok {
			return fmt.Errorf("cannot locate repository")
		}
		config.root = filepath.Clean(filepath.Join(filepath.Dir(source), "..", "..", ".."))
	}
	var err error
	config.root, err = filepath.Abs(config.root)
	if err != nil {
		return err
	}
	config.levels = strings.Split(levels, ",")
	for _, level := range config.levels {
		if level != "0" && level != "2" {
			return fmt.Errorf("unsupported optimization level %q", level)
		}
	}
	if config.output == "" {
		config.output, err = os.MkdirTemp("", "llgo-layout-*")
	} else {
		config.output, err = filepath.Abs(config.output)
		if err == nil {
			err = os.Mkdir(config.output, 0755)
		}
	}
	if err != nil {
		return err
	}
	fmt.Println("Artifacts:", config.output)
	for _, tool := range []*string{&config.llgo, &config.adapter} {
		path, err := exec.LookPath(*tool)
		if err != nil {
			return err
		}
		*tool, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}
	envData, err := runCommand(config.root, filepath.Join(config.output, "go-env.log"), "go", "env", "-json", "GOVERSION", "GOOS", "GOARCH")
	if err != nil {
		return err
	}
	var target struct{ GOVERSION, GOOS, GOARCH string }
	if err := json.Unmarshal(envData, &target); err != nil {
		return err
	}
	config.goversion, config.goos, config.goarch = target.GOVERSION, target.GOOS, target.GOARCH
	if config.goos != "darwin" && config.goos != "linux" && config.goos != "windows" {
		return fmt.Errorf("native runner does not support GOOS=%s", config.goos)
	}
	llvmConfig := os.Getenv("LLVM_CONFIG")
	if llvmConfig == "" {
		llvmConfig = "llvm-config"
	}
	bindir, err := runCommand(config.root, filepath.Join(config.output, "llvm-tools.log"), llvmConfig, "--bindir")
	if err != nil {
		return err
	}
	config.toolDir = strings.TrimSpace(string(bindir))
	compiler, err := runCommand(config.root, filepath.Join(config.output, "llgo-version.log"), config.llgo, "version")
	if err != nil {
		return err
	}
	if config.llgen == "" {
		config.llgen = filepath.Join(config.output, "llgen"+executableSuffix(config.goos))
		if _, err := runCommand(config.root, filepath.Join(config.output, "build-llgen.log"), "go", "build", "-o", config.llgen, "./chore/llgen"); err != nil {
			return err
		}
	}
	fixture, err := loadExisting(config.root, config.goarch)
	if err != nil {
		return err
	}
	report := runReport{GoVersion: config.goversion, GOOS: config.goos, GOARCH: config.goarch, Compiler: strings.TrimSpace(string(compiler))}
	failed := false
	for _, level := range config.levels {
		fmt.Println("RUN runtime O" + level)
		result := runExisting(config, fixture, level)
		report.Results = append(report.Results, result)
		failed = failed || result.Status != "pass"
		fmt.Printf("%s runtime O%s: %d differences %s\n", strings.ToUpper(result.Status), level, len(result.Differences), result.Reason)
		if err := writeJSON(filepath.Join(config.output, "report.json"), report); err != nil {
			return err
		}
	}
	if failed {
		return fmt.Errorf("layout checks failed; see %s", filepath.Join(config.output, "report.json"))
	}
	return nil
}

func runExisting(config runConfig, fixture *existingFixture, level string) (result caseResult) {
	result = caseResult{Case: "runtime", Level: level, Status: "error", Directory: filepath.Join(config.output, "runtime-O"+level)}
	fail := func(err error) caseResult { result.Reason = err.Error(); return result }
	if err := os.Mkdir(result.Directory, 0755); err != nil {
		return fail(err)
	}
	for _, file := range []struct{ name, source string }{{"main.go", fixture.mainSource}, {"layout_observe.go", fixture.observerSource}, {"go.mod", fixture.module}} {
		if err := os.WriteFile(filepath.Join(result.Directory, file.name), []byte(file.source), 0644); err != nil {
			return fail(err)
		}
	}
	baseline := filepath.Join(result.Directory, "reference"+executableSuffix(config.goos))
	if _, err := runCommand(result.Directory, filepath.Join(result.Directory, "go-build.log"), "go", "build", "-o", baseline, "."); err != nil {
		return fail(err)
	}
	output, err := runCommand(result.Directory, filepath.Join(result.Directory, "go-run.log"), baseline)
	if err != nil {
		return fail(err)
	}
	reference, err := parseReference(string(output), fixture)
	if err != nil {
		return fail(err)
	}
	if err := writeJSON(filepath.Join(result.Directory, "baseline.json"), reference); err != nil {
		return fail(err)
	}
	expected := make([]sample, len(reference))
	sizes := types.SizesFor("gc", config.goarch)
	for i, current := range reference {
		id := int(current.Words[0])
		expected[i] = current
		if id < len(fixture.cases) {
			adjusted, err := expectedReference(fixture.cases[id], config.goarch, []sample{current})
			if err != nil {
				return fail(err)
			}
			expected[i] = adjusted[0]
		} else {
			words := append([]uint64(nil), current.Words...)
			expected[i].Words = words
			site := fixture.sites[id-len(fixture.cases)]
			for j := 1; j < len(words); {
				slot := int(words[j])
				words[j+2] = uint64(expectedLayout(site.types[slot], sizes).size)
				j += 4 + int(words[j+3])
			}
		}
	}
	if err := writeJSON(filepath.Join(result.Directory, "expected.json"), expected); err != nil {
		return fail(err)
	}
	executable := filepath.Join(result.Directory, "observed"+executableSuffix(config.goos))
	if _, err := runCommand(result.Directory, filepath.Join(result.Directory, "llgo-build.log"), config.llgo, "build", "-O"+level, "-ldflags=-w=false", "-o", executable, "."); err != nil {
		return fail(err)
	}
	observed, err := inspect(config.adapter, executable, fixture, filepath.Join(result.Directory, "lldb.log"), reference)
	if saveErr := writeJSON(filepath.Join(result.Directory, "observed.json"), observed); saveErr != nil {
		return fail(saveErr)
	}
	if err != nil {
		return fail(err)
	}
	for i, want := range expected {
		id := int(want.Words[0])
		if id < len(fixture.cases) {
			result.Differences = append(result.Differences, compare(fixture.cases[id], []sample{want}, []sample{observed[i]})...)
		} else {
			result.Differences = append(result.Differences, compareValues(fixture.sites[id-len(fixture.cases)], want, observed[i], reference[0].Words)...)
		}
	}
	data, err := dwarfData(executable, config.goos, config.toolDir)
	if err != nil {
		return fail(err)
	}
	for i, c := range fixture.cases {
		differences, err := checkDWARF(c, data, expected[i])
		if err != nil {
			return fail(err)
		}
		result.Differences = append(result.Differences, differences...)
	}
	if _, err := runCommand(result.Directory, filepath.Join(result.Directory, "llgen.log"), config.llgen, "-phase=pre-abi", "."); err != nil {
		return fail(err)
	}
	irPath := filepath.Join(result.Directory, "llgo_autogen.ll")
	if _, err := runCommand(result.Directory, filepath.Join(result.Directory, "verify-ir.log"), filepath.Join(config.toolDir, "opt"+executableSuffix(config.goos)), "-passes=verify", "-disable-output", irPath); err != nil {
		return fail(err)
	}
	for i, c := range fixture.cases {
		differences, err := checkIR(c, irPath, expected[i])
		if err != nil {
			return fail(err)
		}
		result.Differences = append(result.Differences, differences...)
	}
	result.Status = "pass"
	if len(result.Differences) != 0 {
		result.Status = "fail"
	}
	return result
}

func parseReference(output string, fixture *existingFixture) ([]sample, error) {
	var result []sample
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line, ok := strings.CutPrefix(scanner.Text(), "LLGO_LAYOUT ")
		if !ok {
			continue
		}
		if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
			return nil, fmt.Errorf("invalid observation: %s", line)
		}
		fields := strings.Fields(line[1 : len(line)-1])
		words := make([]uint64, len(fields))
		for i, field := range fields {
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return nil, err
			}
			words[i] = value
		}
		if len(words) == 0 || words[0] >= uint64(len(fixture.cases)+len(fixture.sites)) {
			return nil, fmt.Errorf("invalid observation identifier: %v", words)
		}
		id := int(words[0])
		if id < len(fixture.cases) {
			if id != len(result) || len(words) != headerWords+len(fixture.cases[id].fields)*fieldWords {
				return nil, fmt.Errorf("invalid type observation %d", id)
			}
			if words[3] != words[6] || words[6] != words[11] || words[8] != words[12] || words[9] != words[13] {
				return nil, fmt.Errorf("Go ABI prefix disagrees with reflect: %v", words[:headerWords])
			}
		} else {
			for j := 1; j < len(words); {
				if j+4 > len(words) || words[j] >= uint64(len(fixture.sites[id-len(fixture.cases)].names)) || words[j+3] > uint64(len(words)-j-4) {
					return nil, fmt.Errorf("invalid value observation %d", id)
				}
				j += 4 + int(words[j+3])
			}
		}
		result = append(result, sample{Words: words})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(result) <= len(fixture.cases) {
		return nil, fmt.Errorf("existing checkpoints produced no value observations")
	}
	return result, nil
}

func executableSuffix(goos string) string {
	if goos == "windows" {
		return ".exe"
	}
	return ""
}
func runCommand(directory, logPath, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	log := fmt.Sprintf("%s %q\n%s", name, args, output)
	if writeErr := os.WriteFile(logPath, []byte(log), 0644); writeErr != nil {
		return nil, writeErr
	}
	if err != nil {
		return output, fmt.Errorf("%s: %w; see %s", name, err, logPath)
	}
	return output, nil
}
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
