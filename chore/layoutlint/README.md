# layoutlint

This standalone Go tool takes Go package patterns or source files, builds observation copies with Go and LLGo, and uses native `lldb-dap` to check the LLGo executable. Input programs need no layout-specific comments or breakpoint markers. There is no built-in fixture or case list.

This developer command lives beside `chore/llgen` and `chore/litgen`; the littest checking library itself lives in `internal/littest`. Build it from the repository root with `go build -o /path/to/layoutlint ./chore/layoutlint`. It is part of the root module and is also installed by the existing `go install ./...` tool build. Supply the existing LLGo compiler, `llgen`, and native debug adapter:

```sh
layoutlint -root /path/to/project \
  -llgo /path/to/llgo -llgen /path/to/llgen \
  -lldb-dap /path/to/lldb-dap ./some/package

layoutlint -root /path/to/project \
  -llgo /path/to/llgo -llgen /path/to/llgen \
  -lldb-dap /path/to/lldb-dap first.go second.go

layoutlint -root /path/to/project -test \
  -llgo /path/to/llgo -llgen /path/to/llgen \
  -lldb-dap /path/to/lldb-dap ./some/package -- -test.run=TestExample
```

`-root` is the input's working directory and defaults to the current directory. With no positional input, the pattern is `.`. `-test` builds and runs existing tests with `go test -c` and `llgo test -c`, including packages containing only external `package name_test` files; it also accepts selected `_test.go` files. Arguments following `--` after the input are passed to both executables. `-opt 0,2` is the default. `-output` must name a new directory; otherwise the tool creates and prints a temporary artifact directory. Invoke separate modules with their own `-root` rather than combining their package patterns in one call.

Configure LLVM and `LLGO_ROOT` for the compiler being tested. `LLVM_CONFIG` locates LLVM's tools. The LLVM linker must be available to the compiler, and the native debug adapter must be able to launch local processes. No Python script or LLGo debugger presentation plugin is used.

## Processing an input

The Go package loader selects the input files and resolves their types and imports. The tool discovers concrete named and anonymous structs, including concrete generic instances available in the loaded syntax. It records source locations directly from the AST.

A temporary copy retains the module boundary and the package's relative path, including other source files and test resources. Only the selected input files receive observation code. Existing source and littest assertions remain untouched. This extra observation build is separate from the original test's ordinary build.

Struct observers measure adjacent array elements, typed field addresses, `unsafe` sizes and alignment, reflection, and the ABI type prefix. Value observers are inserted around existing assignments, declarations, expression statements and returns. They capture visible variables and accessible fields; they do not invent new application writes or test scenarios. Both compilers run the same observation source.

The Go run supplies the baseline. LLDB stops inside the generated observer, reads the observation words, and independently reads the descriptor prefix and scalar/string/slice storage. DWARF checks follow the observer's actual `items` variable to its array element type, so aliases and anonymous structs do not require a guessed type name. A separate IR source view includes existing test functions as ordinary source files for `llgen`.

## Expected representation and coverage

`baseline.json` preserves Go's measurements. `expected.json` records the LLGo expectation. Ordinary function values intentionally occupy two words in LLGo: code and environment. When they occur inside structs or arrays, the tool recomputes enclosing offsets, sizes, trailing padding, array stride and `PtrBytes` using that representation and Go alignment rules. Pointer referents and function signatures do not expand their containing value.

| Check | Measurements |
| --- | --- |
| ABI metadata | Size, alignment, field alignment, kind, `PtrBytes`, field offsets and field type sizes; independent descriptor-prefix reads. |
| Physical layout | `unsafe` constants, typed field-address differences, object alignment and array stride. |
| Access and values | Values surrounding existing statements, accessible aggregate fields, and LLDB reads of scalar storage, strings and slice headers. |
| Calls and returns | Variables visible at existing call/return sites and caller assignments, as exercised by the input program. |
| DWARF | Struct size and nested field offsets/sizes through the actual observer variable's type reference. |
| LLVM IR | Module verification and load/store types and alignment promises in generated private accessors for eligible scalar/pointer fields. |

`PtrBytes` is a metadata check, not a GC-liveness test. Pointer values compare nilness rather than cross-process addresses; root struct pointers are dereferenced once. Map values compare length, not iteration order. Struct value comparisons recurse through fields accessible to the input package, avoiding foreign private state such as a test framework's timestamps and stack PCs. The input must produce a corresponding observation sequence in both runs; different sequences or payload lengths are reported as errors, not accepted as layout passes.

Slice capacity differences between Go and LLGo are warnings: append can expose different allocation size classes. Both capacities remain in the observations and diagnostics, including slices nested inside other values. Types, sizes, lengths and element values still require equality. LLDB's raw slice capacity must match LLGo's own reflected capacity, and a capacity below the slice length is an error. Existing input assertions about capacity are unchanged.

`llgo:cdecl`, `llgo:stdcall` and legacy `llgo:type C` types are excluded from ordinary-layout comparisons, as are containing types that depend on their representation. Exclusions are named in the report. A program that cannot build with the Go reference compiler is reported as an error; the tool has no fixture-specific source removal rules.

## Reports and source diagnostics

The output contains a top-level `report.json`, separate build/debug artifacts for each input and optimization level, and an `annotated` directory for each input. Each run retains its selected observation source, executables, logs, baseline, expectation, LLDB observations, DWARF artifacts and LLVM IR. Complete module workspaces are temporary and are removed after each check. A failed package load or comparison is recorded without preventing other inputs from running; any failure makes the final exit status nonzero. Progress identifies the current package and build/debug phase.

Annotated source appends diagnostics to the affected declaration or statement:

```go
fn func(string) (int, error) // layoutlint: reflected size = 8 bytes; expected 16 (O0, O2)
```

Comments name the mismatched property and show its observed and expected values. Size errors point to the variable declaration or struct field, while value errors point to the observed statement. Repeated size errors and identical optimization results share one annotation; different values remain separate. Full evidence and original file/line locations remain in JSON. Missing symbols, debugger errors, or missing DWARF/IR fail the run. Known compiler discrepancies remain failures.

Capacity diagnostics carry `"warning": true` in `report.json` and appear as `// layoutlint: warning: app capacity = 2; Go reports 8 (O0, O2)` in annotated source. A result containing only warnings has status `warning` and does not make the command exit nonzero; any error still fails the run.

The LLGo workflow invokes the command separately from existing test runners, initially only in `MeteorsLiu/llgo`. It enumerates compiler fixture directories, checks runtime fixtures from the runtime module, and checks existing test packages. The core job reuses the package list printed by its existing runner and prepares `.github/test-go.mod` in a temporary source copy; the focused Linux ARM64 and macOS Intel jobs retain the root module. Reports include the selected input lists, observations and annotated source. Any batch failure fails the final layout check. Fixture discovery includes inputs with specialized build requirements; unsupported inputs remain visible errors rather than being silently omitted. The command itself contains no repository-specific case list or CI hooks.
