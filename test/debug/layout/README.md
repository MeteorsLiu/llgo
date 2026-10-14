# Layout checks for the existing debugger fixture

This Go tool builds `test/debug/runtime` with Go and LLGo, then reads the LLGo program through native `lldb-dap`. It discovers the fixture's existing named structs and `LLDB_BREAK` checkpoints. There is no separate list of handwritten test cases, and the runner uses no Python scripts or LLGo debugger presentation plugin.

```sh
go run . -root /path/to/llgo -llgo /path/to/llgo-binary -llgen /path/to/llgen -lldb-dap /path/to/lldb-dap
```

Configure the same LLVM 22 build environment and `LLGO_ROOT` used by the existing debugger suite. `LLVM_CONFIG` locates LLVM's tools. On macOS, `xcrun --find lldb-dap` locates Apple's adapter. `-llgen` can select the compiler lane's already-built tool; without it the runner builds `chore/llgen` privately. `-opt 0,2` is the default. `-output` must name a new directory; otherwise a temporary directory is created and printed.

## Go reference and LLGo expectation

`baseline.json` preserves the Go program's observations. `expected.json` contains the expectation used for LLGo. Ordinary function values occupy two words in that expectation: the code pointer and the environment pointer. A function value therefore occupies 8 bytes on 32-bit targets and 16 bytes on 64-bit targets. This representation difference is intentional.

For a struct or array containing function values, the runner recomputes field offsets, element sizes, trailing padding and `PtrBytes` from the adjusted element sizes and Go alignment rules. It does not add a word to an already-padded Go offset. For example, expanding a 4-byte function value into the 4-byte gap before an 8-byte-aligned field leaves that next field's offset unchanged. Pointer referents and the parameter/result types of a function do not expand the containing value.

The current `StructWithAllTypeFields` fixture has Go/LLGo expected sizes of 280/288 bytes on amd64 and arm64, and 172/176 bytes on 386 and ARM32. Its fields following `fn` use the recomputed offsets. Layout changes are not applied to the semantic values captured at checkpoints; those still compare directly with Go.

## Observations

The runner writes temporary instrumented copies of the existing source. The original fixture is never edited. At program entry, private observers measure each existing struct using two adjacent elements, typed field addresses, `unsafe` constants and reflection. At existing breakpoints, observers capture visible variables, aggregate values and individual fields. LLDB reads the observation buffers, actual scalar storage, strings, slice headers and struct type-descriptor prefixes while those objects are live.

| Report category | Coverage from the existing fixture |
| --- | --- |
| `metadata` | Struct size, alignment, field alignment, kind, `PtrBytes`, field offsets and field type sizes. LLDB reads the descriptor prefix independently. |
| `layout` | Compiled field-address differences, array stride, object alignment and `unsafe` constants. |
| `access` | Values before/after the fixture's existing writes, typed field reads, reflected aggregate reads, and LLDB raw-memory reads. |
| `call_return` | Existing by-value struct arguments and their mutations. The fixture's scalar/error return results are also observed in its caller; it does not currently exercise a by-value struct return. |
| `value_semantics` | Existing interface operations, closure calls, method values, map lookup results and channel results. Map iteration order and absolute pointer addresses are not compared. |
| `dwarf` | Existing struct sizes and nested field offsets/sizes, following typedefs. |
| `ir` | LLVM verification plus load/store widths and alignment promises in private accessors generated for the existing scalar/pointer fields. The accessors do not add runtime writes. |

`PtrBytes` is checked only as metadata. There is no GC-liveness test. Arrays and structs are traversed by value; pointer fields compare nil/non-nil, with one level of dereferencing for existing struct-pointer variables. No additional copy, equality or return scenario is invented for coverage absent from the existing program.

The approved temporary overlay removes the `RuntimeGoroutineValues` call and implementation and omits `goroutines.go`, whose native traceback symbols cannot link with Go. Both compilers observe the same overlay. The full existing LLDB suite still runs its original source, including that scenario. C/native callback fixtures are not selected; ordinary-layout preflight rejects `llgo:cdecl`, `llgo:stdcall` and legacy `llgo:type C` declarations.

## CI and results

The shared `.github/actions/check-layout` action receives the already-built `llgo` and `llgen` paths and a unique `artifact-name`. `source-root` defaults to the checkout; `runtime-root` defaults to that source root. Release checks pass the installed compiler and its installed runtime explicitly, while reading the fixtures from the checkout. The action prepares the native debug adapter, runs O0/O2 checks and uploads observations even on failure. It does not build or replace the selected compiler.

The full check belongs to the external [LLGo compatibility workflows](https://github.com/llgo-compat/ci). Its existing `GOROOT` scheduled/manual entry runs an independent native-layout matrix once per platform/ABI, without multiplying the work across corpus shards. The selected `llgo_repository` and `llgo_ref` supply the compiler, fixture and shared action. Ordinary LLGo pull-request compiler jobs do not invoke this lengthy check.

The native matrix covers Linux amd64/arm64, macOS amd64/arm64, and Windows MSVC/MinGW on amd64/arm64/386. Windows debugger environment values are scoped to the checking process; the action does not export them into subsequent job steps. Layout failures retain a nonzero exit status. Wasm, embedded and cross-target layout execution remain outside this native action.

Each optimization level retains the instrumented source, Go/LLGo executables, build logs, `baseline.json`, `expected.json`, `observed.json`, a DAP transcript, DWARF artifacts and pre-ABI LLVM IR. `report.json` lists mismatches with their category, checkpoint and field or variable. Missing symbols, failed debugger operations and missing DWARF/IR are errors rather than passes.

`annotated/main.go` copies the original fixture and adds `LAYOUT-FAIL` comments before the relevant type, field or `LLDB_BREAK` statement. Each comment includes the original source line, category, checkpoint when applicable, expected value and observed value. For example, `StructWithAllTypeFields.fn size; want=16; got=8` appears beside the `fn` field, while `(s).fn size` appears at the checkpoint where `s` was observed. Nested field offsets are annotated at the outer field whose storage contains them. Identical O0/O2 differences share one comment; different values remain separate. The header retains each run's status and any execution error. This is a generated diagnostic view, not a new source of assertions: the original fixture, full JSON differences and failure status are unchanged.

Native macOS/arm64 has been exercised at O0 and O2, including the shared action's Unix command with a separate runtime directory containing spaces. The runner also cross-compiles for Linux/amd64 and Windows/386, amd64 and arm64; Linux/ARM64, Intel macOS, hosted/self-hosted Linux permissions, Windows adapter distributions and the actual release-archive combinations still require their CI runs.

The audited main revision currently fails the metadata assertions: `PtrBytes` loses the final pointer-bearing field's extent when scalar fields follow it, and reflected function types report one word despite two-word storage. The runner keeps the two-word expectation and reports those failures. It does not whitelist them or change the compiler.
