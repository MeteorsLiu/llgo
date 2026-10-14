// Command layoutlint compares Go source packages with LLGo through native LLDB.
package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"go/types"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type runConfig struct {
	llgo, llgen, adapter, toolDir, root, output string
	goos, goarch, goversion                     string
	levels, arguments                           []string
	test                                        bool
}
type caseResult struct {
	Case        string       `json:"case"`
	Level       string       `json:"level,omitempty"`
	Status      string       `json:"status"`
	Reason      string       `json:"reason,omitempty"`
	Directory   string       `json:"directory,omitempty"`
	Excluded    []string     `json:"excluded,omitempty"`
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
	flag.StringVar(&config.llgen, "llgen", "llgen", "llgen executable for pre-ABI IR checks")
	flag.StringVar(&config.adapter, "lldb-dap", "lldb-dap", "native LLDB debug adapter executable")
	flag.StringVar(&config.root, "root", ".", "working directory for input package patterns or Go files")
	flag.StringVar(&config.output, "output", "", "new artifact directory (default: temporary directory)")
	flag.StringVar(&levels, "opt", "0,2", "comma-separated LLGo optimization levels: 0,2")
	flag.BoolVar(&config.test, "test", false, "build and observe the input packages' existing Go tests")
	flag.Parse()
	patterns := flag.Args()
	for i, arg := range patterns {
		if arg == "--" {
			config.arguments = patterns[i+1:]
			patterns = patterns[:i]
			break
		}
	}
	if len(patterns) == 0 {
		patterns = []string{"."}
	}
	if err := run(config, levels, patterns); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(config runConfig, levels string, patterns []string) error {
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
		config.output, err = os.MkdirTemp("", "llgo-layoutlint-*")
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
	for _, tool := range []*string{&config.llgo, &config.llgen, &config.adapter} {
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
		return fmt.Errorf("native LLDB does not support GOOS=%s", config.goos)
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
	fixtures, err := loadFixtures(config, patterns)
	if err != nil {
		return err
	}
	report := runReport{GoVersion: config.goversion, GOOS: config.goos, GOARCH: config.goarch, Compiler: strings.TrimSpace(string(compiler))}
	failed := false
	for caseIndex, fixture := range fixtures {
		for _, level := range config.levels {
			fmt.Printf("RUN %s O%s: %d types, %d observation sites\n", fixture.name, level, len(fixture.cases), len(fixture.sites))
			directory := filepath.Join(config.output, fmt.Sprintf("case-%d-O%s", caseIndex, level))
			result := runExisting(config, fixture, level, directory)
			report.Results = append(report.Results, result)
			failed = failed || result.Status != "pass" && result.Status != "warning"
			warnings := 0
			for _, diff := range result.Differences {
				if diff.Warning {
					warnings++
				}
			}
			fmt.Printf("%s %s O%s: %d differences, %d warnings %s\n", strings.ToUpper(result.Status), fixture.name, level, len(result.Differences)-warnings, warnings, result.Reason)
			if err := writeJSON(filepath.Join(config.output, "report.json"), report); err != nil {
				return err
			}
			if err := writeAnnotations(filepath.Join(config.output, fmt.Sprintf("case-%d", caseIndex), "annotated"), fixture.files, report, fixture.name); err != nil {
				return err
			}
		}
	}
	if failed {
		return fmt.Errorf("layout checks failed; see %s", filepath.Join(config.output, "report.json"))
	}
	return nil
}

func runExisting(config runConfig, fixture *existingFixture, level, directory string) (result caseResult) {
	result = caseResult{Case: fixture.name, Level: level, Status: "error", Directory: directory, Excluded: fixture.excluded}
	fail := func(err error) caseResult { result.Reason = err.Error(); return result }
	if err := os.MkdirAll(directory, 0755); err != nil {
		return fail(err)
	}
	if fixture.loadError != nil {
		return fail(fixture.loadError)
	}
	scratch, err := os.MkdirTemp("", "llgo-layoutlint-work-")
	if err != nil {
		return fail(err)
	}
	defer func() {
		if err := os.RemoveAll(scratch); err != nil {
			result.Status = "error"
			result.Reason += fmt.Sprintf("; remove observation workspace: %v", err)
		}
	}()
	workRoot := filepath.Join(scratch, "work")
	// Keep the module boundary and package paths intact. A package such as
	// example.org/project/a must not become a submodule requiring its own parent.
	if err := copyModule(fixture.moduleRoot, workRoot, config.output); err != nil {
		return fail(err)
	}
	if err := os.WriteFile(filepath.Join(workRoot, "go.mod"), []byte(fixture.module), 0644); err != nil {
		return fail(err)
	}
	relative, err := filepath.Rel(fixture.moduleRoot, fixture.directory)
	if err != nil {
		return fail(err)
	}
	work := filepath.Join(workRoot, relative)
	sourceRoot := filepath.Join(directory, "source")
	sourceDirectory := filepath.Join(sourceRoot, relative)
	if err := os.MkdirAll(sourceDirectory, 0755); err != nil {
		return fail(err)
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "go.mod"), []byte(fixture.module), 0644); err != nil {
		return fail(err)
	}
	for _, file := range fixture.files {
		if err := os.WriteFile(filepath.Join(work, filepath.Base(file.path)), []byte(file.instrumented), 0644); err != nil {
			return fail(err)
		}
		if err := os.WriteFile(filepath.Join(sourceDirectory, filepath.Base(file.path)), []byte(file.instrumented), 0644); err != nil {
			return fail(err)
		}
	}
	command := "build"
	if fixture.test {
		command = "test"
	}
	buildArgs := []string{command}
	if fixture.test {
		buildArgs = append(buildArgs, "-c")
	}
	baseline := filepath.Join(directory, "reference"+executableSuffix(config.goos))
	args := append(append(append([]string(nil), buildArgs...), "-o", baseline), fixture.arguments...)
	fmt.Println("  Go reference build")
	if _, err := runCommand(work, filepath.Join(directory, "go-build.log"), "go", args...); err != nil {
		return fail(err)
	}
	output, err := runCommand(work, filepath.Join(directory, "go-run.log"), baseline, config.arguments...)
	if err != nil {
		return fail(err)
	}
	reference, err := parseReference(string(output), fixture)
	if err != nil {
		return fail(err)
	}
	if err := writeJSON(filepath.Join(directory, "baseline.json"), reference); err != nil {
		return fail(err)
	}
	expected := make([]sample, len(reference))
	sizes := types.SizesFor("gc", config.goarch)
	for i, current := range reference {
		id := int(current.Words[0])
		expected[i] = current
		if id%2 == 0 {
			adjusted, err := expectedReference(fixture.cases[id/2], config.goarch, []sample{current})
			if err != nil {
				return fail(err)
			}
			expected[i] = adjusted[0]
		} else {
			words := append([]uint64(nil), current.Words...)
			expected[i].Words = words
			site := fixture.sites[id/2]
			for j := 1; j < len(words); {
				slot := int(words[j])
				words[j+2] = uint64(expectedLayout(site.types[slot], sizes).size)
				j += 4 + int(words[j+3])
			}
		}
	}
	if err := writeJSON(filepath.Join(directory, "expected.json"), expected); err != nil {
		return fail(err)
	}
	executable := filepath.Join(directory, "observed"+executableSuffix(config.goos))
	args = append(append(append([]string(nil), buildArgs...), "-O"+level, "-ldflags=-w=false", "-o", executable), fixture.arguments...)
	fmt.Println("  LLGo observation build")
	if _, err := runCommand(work, filepath.Join(directory, "llgo-build.log"), config.llgo, args...); err != nil {
		return fail(err)
	}
	layoutWords := make([]uint64, headerWords)
	layoutWords[17] = uint64(sizes.Sizeof(types.Typ[types.Uintptr]))
	if binary.NativeEndian.Uint16([]byte{1, 0}) == 1 {
		layoutWords[18] = 1
	}
	fmt.Printf("  LLDB: %d snapshots\n", len(reference))
	observed, err := inspect(config.adapter, executable, work, fixture, filepath.Join(directory, "lldb.log"), reference, layoutWords, config.arguments)
	if saveErr := writeJSON(filepath.Join(directory, "observed.json"), observed); saveErr != nil {
		return fail(saveErr)
	}
	if err != nil {
		return fail(err)
	}
	typeReferences := make(map[int]sample)
	for i, want := range expected {
		id := int(want.Words[0])
		if id%2 == 0 {
			typeReferences[id/2] = want
			result.Differences = append(result.Differences, compare(fixture.cases[id/2], []sample{want}, []sample{observed[i]})...)
		} else {
			result.Differences = append(result.Differences, compareValues(fixture.sites[id/2], want, observed[i], layoutWords)...)
		}
	}
	data, err := dwarfData(executable, config.goos, config.toolDir)
	if err != nil {
		return fail(err)
	}
	for i, c := range fixture.cases {
		want, ok := typeReferences[i]
		if !ok {
			return fail(fmt.Errorf("type %s was not observed", c.displayName))
		}
		differences, err := checkDWARF(c, data, want)
		if err != nil {
			return fail(err)
		}
		result.Differences = append(result.Differences, differences...)
	}
	// Give llgen a source view where existing test functions are ordinary Go
	// files. This retains their types without requiring a synthetic test main.
	type irInput struct{ directory, artifacts string }
	irDirectories := make(map[string]irInput)
	for _, file := range fixture.files {
		input, ok := irDirectories[file.packagePath]
		if !ok {
			name := fmt.Sprintf("ir-%d", len(irDirectories))
			irRoot := filepath.Join(scratch, name)
			if err := copyModule(fixture.moduleRoot, irRoot, config.output); err != nil {
				return fail(err)
			}
			if err := os.WriteFile(filepath.Join(irRoot, "go.mod"), []byte(fixture.module), 0644); err != nil {
				return fail(err)
			}
			irDir := filepath.Join(irRoot, relative)
			if file.packageName != fixture.files[0].packageName {
				irDir += "_test"
			}
			input = irInput{directory: irDir, artifacts: filepath.Join(directory, name)}
			irDirectories[file.packagePath] = input
			if err := os.MkdirAll(irDir, 0755); err != nil {
				return fail(err)
			}
			if err := os.Mkdir(input.artifacts, 0755); err != nil {
				return fail(err)
			}
		}
		name := strings.TrimSuffix(filepath.Base(file.path), "_test.go")
		if name != filepath.Base(file.path) {
			name += "_layout.go"
		}
		if err := os.WriteFile(filepath.Join(input.directory, name), []byte(file.instrumented), 0644); err != nil {
			return fail(err)
		}
	}
	for prefix, input := range irDirectories {
		irDir := input.directory
		fmt.Println("  DWARF and IR:", prefix)
		packagePath, err := runCommand(irDir, filepath.Join(input.artifacts, "go-list.log"), "go", "list", "-f", "{{.ImportPath}}", ".")
		if err != nil {
			return fail(err)
		}
		if _, err := runCommand(irDir, filepath.Join(input.artifacts, "llgen.log"), config.llgen, "-phase=pre-abi", "."); err != nil {
			return fail(err)
		}
		ir, err := os.ReadFile(filepath.Join(irDir, "llgo_autogen.ll"))
		if err != nil {
			return fail(err)
		}
		irPath := filepath.Join(input.artifacts, "llgo_autogen.ll")
		if err := os.WriteFile(irPath, ir, 0644); err != nil {
			return fail(err)
		}
		if _, err := runCommand(irDir, filepath.Join(input.artifacts, "verify-ir.log"), filepath.Join(config.toolDir, "opt"+executableSuffix(config.goos)), "-passes=verify", "-disable-output", irPath); err != nil {
			return fail(err)
		}
		for i, c := range fixture.cases {
			if c.irPrefix != prefix && c.irPrefix != "main" {
				continue
			}
			if c.irPrefix != "main" {
				c.irPrefix = strings.TrimSpace(string(packagePath))
			}
			differences, err := checkIR(c, irPath, typeReferences[i])
			if err != nil {
				return fail(err)
			}
			result.Differences = append(result.Differences, differences...)
		}
	}
	result.Status = "pass"
	for _, diff := range result.Differences {
		if !diff.Warning {
			result.Status = "fail"
			break
		}
		result.Status = "warning"
	}
	return result
}

func copyModule(source, destination, output string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Name() == ".git" || path == output {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			if !filepath.IsAbs(link) {
				link = filepath.Join(filepath.Dir(path), link)
			}
			return os.Symlink(link, target)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
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
		if len(words) == 0 {
			return nil, fmt.Errorf("empty observation")
		}
		id := int(words[0])
		index := id / 2
		if id%2 == 0 {
			if index >= len(fixture.cases) || len(words) != headerWords+len(fixture.cases[index].fields)*fieldWords {
				return nil, fmt.Errorf("invalid type observation %d", id)
			}
			if words[3] != words[6] || words[6] != words[11] || words[8] != words[12] || words[9] != words[13] {
				return nil, fmt.Errorf("Go ABI prefix disagrees with reflect: %v", words[:headerWords])
			}
		} else {
			if index >= len(fixture.sites) {
				return nil, fmt.Errorf("invalid value observation %d", id)
			}
			for j := 1; j < len(words); {
				if j+4 > len(words) || words[j] >= uint64(len(fixture.sites[index].names)) || words[j+3] > uint64(len(words)-j-4) {
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
	if len(result) == 0 {
		return nil, fmt.Errorf("input produced no layout observations")
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
		return output, writeErr
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
