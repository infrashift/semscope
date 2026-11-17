# semscope implementation plan

semscope scans source code with composable CUE rulesets. The first built-in ruleset is `nasa-p10`, the NASA Power of 10 rules. semscope writes the findings as SARIF 2.1.0, SARIF 2.2, JSON or text.

This file is the phase checklist. Each phase has a mirror issue in the OKF wiki (`.okfw/issues/`). The wiki records the decisions and the reasons for them.

## 0. Operational directives and quality gates

### 0.1 Verify-lint-repeat loop

Run the phases in sequence. Run the tasks of a phase in sequence. After each task, run the full gate:

1. Run `zmake check` for the ZMakefile.
2. Run `zmake -f Makefile check` for the Makefile.
3. Fix each failure. Then run both gates again.
4. Commit the task when both gates pass.

`zmake check` runs these checks:

- `python3 -I scripts/nasa-lint.py`: the Zig oracle. It reads `src/**/*.zig` only, so it finds no functions in this repository.
- `python3 -I scripts/nasa-lint-go.py` and its fixture tests: the Go gate for the rules R1, R2, R4, R5, R7 and R9.
- `gofmt -l`, `go vet ./...` and `golangci-lint run`.
- `go test -race ./...`.
- `okfw validate --all` and `okfw validate --okf`.
- `python3 -I scripts/ste-check.py`.
- From Task 4.1: `semscope scan --ruleset nasa-p10` on the semscope sources.

Each check must exit with status 0 and report zero warnings. Do not weaken a test or a check to make a gate pass. Each suppression comment needs a reason.

### 0.2 NASA Power of 10 rules

Every Go file in semscope obeys the Power of 10 rules of Gerard J. Holzmann. ADR-0003 in the wiki adapts the rules to Go.

- **R1 Simple control flow:** Do not use recursion, direct or indirect. Do not use `goto`. Walk trees with a Tree-sitter cursor or an explicit bounded stack.
- **R2 Fixed loop bounds:** Each loop has a static upper bound, for example `MaxFiles`, `MaxASTNodes` or `MaxTreeDepth`. A loop that reaches its bound returns an error.
- **R3 Limited allocation:** Go allocates memory at run time. ADR-0003 records this deviation. `NewEngine` sizes the worker pool, the stacks and the result buffers. Hard limits cap all other memory use.
- **R4 Short functions:** A function has 60 lines or fewer.
- **R5 Assertion density:** The average is two or more `assert.That` calls per function. Assertions check programmer errors only. Bad input gives an error, not an assertion.
- **R6 Smallest scope:** Declare each variable in the smallest scope. Do not add mutable package state.
- **R7 Check all returns:** Check each error. Do not discard a call result with `_`. Validate each parameter.
- **R8 Limited code generation:** Keep build tags and code generation to a minimum. Generated code is the same on each run.
- **R9 Restricted indirection:** Do not add function types, function tables or deep pointer chains. Use static dispatch.
- **R10 Zero warnings:** The compiler and all static analyzers report zero warnings.

### 0.3 ASD-STE100 documentation style

All Markdown files, OKF wiki pages, CLI help strings and code comments follow ASD-STE100 Simplified Technical English. `scripts/ste-check.py` checks them.

- Write with clinical precision. Do not use marketing words, for example `seamless`, `robust` or `powerful`.
- Do not use the possessive `your`. Write `the ruleset file`, not `your ruleset file`.
- Keep each procedural step to 20 words or fewer. Keep each descriptive sentence to 25 words or fewer.
- Use active voice. Name the actor: "The parser reads the source file."
- Use one term for one concept. Prefer a verb to a noun form: write "configure", not "configuration". The check does not flag noun forms, because `function` and `assertion` have the same suffixes.
- Write one instruction in each sentence.

Task order: run 2.1, then 3.1 and 3.2, then 2.2. The `scan` and `rules` commands need the ruleset engine.

## Phase 1: Repository bootstrap, gates and OKF wiki

Tracking: [ISSUE-001](.okfw/issues/issue-001-phase-1-repository-bootstrap-gates-and-okf-wiki--01a1e97296be.md)

- [x] **Task 1.0: Repository and build files.** Initialize git. Add `.gitignore`, `ZMakefile` and `Makefile`. The two build files hold the same tasks.
- [x] **Task 1.1: Go module and package layout.** Initialize `github.com/infrashift/semscope` with Go 1.25. Add these packages:
  - `cmd/semscope/`: the entry point and the `spf13/cobra` commands.
  - `internal/assert/`: the assertion primitive for rule R5.
  - `internal/version/`: the build version, commit, date and supported SARIF versions.
  - `internal/ruleset/`: the CUE ruleset loader, overlays, omissions and overrides.
  - `internal/parser/`: Tree-sitter parsers for Go, C, C++, Zig, Python and Rust.
  - `internal/engine/`: the bounded, non-recursive checks and the structural query evaluator.
  - `internal/report/`: SARIF 2.1.0 and 2.2 through `go-sarif/v3`, plus JSON and text.
  - `rules/nasa_p10/`: the embedded CUE source of the `nasa-p10` ruleset.
  - `scripts/`: the gate scripts.
  - `.golangci.yml`: a golangci-lint v2 configuration for the Power of 10 rules.
- [x] **Task 1.2: Baseline linter scripts.** Copy `nasa-lint.py` and `ste-check.py` from zmake into `scripts/`. `nasa-lint.py` reads Zig only. It becomes the Zig oracle for the parity tests. Write `scripts/nasa-lint-go.py`, a Go port of its six checks, with golden fixture tests.
- [x] **Task 1.2b: Documentation style gate.** Remove the open code fence that hid this file from `ste-check.py`. Rewrite this file in STE. Add the marketing words and `your` to the banned list. Extend the check to Go comments and to cobra `Short` and `Long` strings.
- [x] **Task 1.3: OKF wiki.** Copy the wiki layout of zmake: `architecture/`, `specs/`, `decisions/`, `issues/`, `daily/` and two root guides. Write the architecture overview, the specs, ADR-0001 to ADR-0006 and one issue for each phase. Record the check inventory of `nasa-lint.py` as a spec. Use the okfw MCP server or the `okfw` CLI.

## Phase 2: Version package and Cobra CLI

Tracking: [ISSUE-002](.okfw/issues/issue-002-phase-2-version-package-and-cobra-cli--01a1e97296d0.md)

- [x] **Task 2.1: Version package.** Add `internal/version` with `Version`, `Commit` and `BuildDate`. The `build` task sets them through `-ldflags`. Without `-ldflags`, the package reads the VCS fields of `debug.ReadBuildInfo`. The package lists the supported SARIF versions `2.1.0` and `2.2`.
- [x] **Task 2.2: Cobra CLI.** Build the CLI with `github.com/spf13/cobra`:
  - `semscope --version`, `semscope -v` and `semscope version` print the version metadata. The output is the same on each run.
  - `semscope scan [paths...]` takes the flags in the next items.
  - `--ruleset, -r <name-or-path>`: a built-in ruleset (default `nasa-p10`), or one or more `.cue` files or directories.
  - `--enable-rule <id[,id...]>`: enable the given rules.
  - `--omit-rule, --disable-rule <id[,id...]>`: remove the given rules from the active ruleset.
  - `--set <rule_id.param=value>`: override a rule parameter, for example `--set NASA-P10-04.max_lines=50`.
  - `--format <sarif|text|json>`: the output format.
  - `--sarif-version <2.1.0|2.2>`: the SARIF version. The default is `2.1.0`. SARIF 2.2 is an OASIS draft.
  - `--output, -o <file>`: the output file. The default is stdout.
  - `semscope rules list` and `semscope rules inspect <ruleset>` show the active rules, the omitted rules and the effective parameters.
  - An unknown rule ID or parameter in a flag gives an error.

## Phase 3: Composable CUE ruleset engine

Tracking: [ISSUE-003](.okfw/issues/issue-003-phase-3-composable-cue-ruleset-engine--01a1e97296e2.md)

- [x] **Task 3.1: CUE ruleset schema.** Write `internal/ruleset/schema.cue`. Load it with `cuelang.org/go/cue/cuecontext`. The schema has these parts:
  - The `rules` map uses the stable rule ID as the key: `rules: [ID=string]: #Rule & {id: ID}`.
  - `id`: the stable identifier, for example `"NASA-P10-01"`.
  - `enabled`: `bool | *true`. An overlay sets `enabled: false` to omit a rule.
  - `severity`: `*"error" | "warning" | "note"`.
  - `languages`: a list of `go`, `c`, `cpp`, `zig`, `python` and `rust`.
  - `title` and `message`: a clinical description and the finding template.
  - `params`: an open struct of tunable values, for example `max_lines: *60 | int`.
  - `check`: a union with the key `kind`. The kinds are `no_goto_or_recursion`, `bounded_loops`, `no_dynamic_alloc`, `max_function_lines`, `min_assertions`, `narrowest_scope`, `checked_returns`, `restricted_preprocessor`, `max_pointer_indirection`, `ast_query` and `token_ban`.
  - The loader reads a base ruleset. It unifies the overlay files in sequence. It applies `--set`, then `--enable-rule` and `--omit-rule`.
  - The base ruleset declares each tunable value as a CUE default. An overlay value then replaces the default and does not conflict with it.
- [x] **Task 3.2: NASA Power of 10 ruleset.** Write `rules/nasa_p10/nasa_p10.cue` with parameters for each language. It matches every check of `nasa-lint.py` and `nasa-lint-go.py`. It adds checks for the other rules.
  - `NASA-P10-01`: no `goto`, no `setjmp` or `longjmp`, and no direct or indirect recursion.
  - `NASA-P10-02`: no loop without a static bound, for example `for {}`, `while (true)` or `loop {}`.
  - `NASA-P10-03`: no allocator calls outside the init functions.
  - `NASA-P10-04`: function length, `params.max_lines: *60 | int`.
  - `NASA-P10-05`: assertion density, `params.min_assertions: *2 | int` and `params.mode: *"average" | "per_function"`.
  - `NASA-P10-06`: no mutable global state.
  - `NASA-P10-07`: no discarded call results and no swallowed errors.
  - `NASA-P10-08`: limits on macros, conditional compilation and build directives.
  - `NASA-P10-09`: pointer depth, `params.max_pointer_depth: *1 | int`, and no function pointer types.
  - `NASA-P10-10`: no unsafe constructs and no lint suppressions without a reason.
  - Unit tests: a user CUE file adds the custom rule `CUSTOM-01`. A user CUE file or `--omit-rule` omits `NASA-P10-05`. A user CUE file or `--set` sets the `max_lines` of `NASA-P10-04` to 40.

## Phase 4: Bounded static analysis engine and parity

Tracking: [ISSUE-004](.okfw/issues/issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md)

- [x] **Task 4.1: Non-recursive analyzer.** Implement `internal/parser` and `internal/engine` under the rules R1 to R10:
  - Walk each syntax tree with a Tree-sitter cursor. Do not use recursion.
  - Build the call graph in a table with a fixed capacity. Find the cycles with an iterative depth-first search, the same search as in the oracles.
  - Implement one checker for each check kind. Dispatch through a `switch`, not a function table.
  - Run user Tree-sitter queries (`ast_query`). A user then adds a code rule in CUE without a change to the Go code.
  - Return an error when the work exceeds a file, node, depth or finding limit.
  - Turn on the self-scan gate.
- [ ] **Task 4.2: Differential parity tests.** Write `internal/engine/nasa_parity_test.go`. It runs the oracles and `semscope scan --ruleset nasa-p10` on these inputs:
  1. Violation fixtures in `testdata/nasa_p10/violations/`, one for each rule.
  2. Compliant fixtures in `testdata/nasa_p10/compliant/`.
  3. Custom ruleset fixtures in `testdata/custom_rulesets/` for rule addition, rule omission and threshold changes.
  4. The semscope sources in `cmd/` and `internal/`.
  5. An optional corpus through `SEMSCOPE_PARITY_CORPUS`, for example the zmake `src/` tree.

  On the fixtures, semscope and the oracles report the same set of file, line and rule. A divergence on the corpus needs a fix, or an entry with a reason in `testdata/nasa_p10/parity-divergences.json`. semscope does not copy a known defect of an oracle.

## Phase 5: SARIF reports and self-hosting gate

Tracking: [ISSUE-005](.okfw/issues/issue-005-phase-5-sarif-reports-and-self-hosting-gate--01a1e9729706.md)

- [ ] **Task 5.1: SARIF producer.** Implement the SARIF writer with `github.com/owenrumney/go-sarif/v3`:
  - Use `pkg/report/v210/sarif` for SARIF 2.1.0 and `pkg/report/v22/sarif` for SARIF 2.2. The `--sarif-version` flag selects one.
  - Set `tool.driver.name` to `"semscope"`. Set `tool.driver.version`, `tool.driver.informationUri` and `tool.driver.rules`. The rules list holds only the active rules.
  - Give each finding a repository-relative `artifactLocation.uri`, a `ruleId` and a `level`. Add the 1-based `startLine`, `startColumn`, `endLine` and `endColumn`.
  - Call `rep.Validate()` before the writer writes the report.
- [ ] **Task 5.2: Self-hosting gate and wiki sync.** Run the full gate:
  1. Run `zmake check` and `zmake -f Makefile check`.
  2. Run `zmake build`.
  3. Run `./bin/semscope --version`.
  4. Run `./bin/semscope scan --ruleset nasa-p10 --format sarif --sarif-version 2.1.0 --output semscope-self.sarif .`.
  5. Run `./bin/semscope scan --ruleset nasa-p10 --format sarif --sarif-version 2.2 --output semscope-self-v22.sarif .`.

  Write `README.md`. Sync the architecture records and the usage notes to the OKF wiki. Verify zero violations in the repository.
