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
- From Phase 13: `lexis scan --ruleset ste-100 --ruleset vocabulary.cue` on the Markdown files, when lexis is on the PATH.

Each check must exit with status 0 and report zero warnings. Do not weaken a test or a check to make a gate pass. Each suppression comment needs a reason.

### 0.2 NASA Power of 10 rules

Every Go file in semscope obeys the Power of 10 rules of Gerard J. Holzmann. ADR-0003 in the wiki adapts the rules to Go.

- **R1 Simple control flow:** Do not use recursion, direct or indirect. Do not use `goto`. Walk trees with a Tree-sitter cursor or an explicit bounded stack.
- **R2 Fixed loop bounds:** Each loop has a static upper bound, for example `MaxFiles`, `MaxASTNodes` or `MaxTreeDepth`. A loop that reaches its bound returns an error.
- **R3 Limited allocation:** Go allocates memory at run time. ADR-0003 records this deviation. `NewEngine` sizes the worker pool, the stacks and the result buffers. Hard limits cap all other memory use.
- **R4 Short functions:** A function has 60 lines or fewer.
- **R5 Assertion density:** The average is two or more `assert.That` calls per function. Assertions check programmer errors only. Bad input gives an error, not an assertion.
- **R6 Smallest scope:** Declare each variable in the smallest scope. Do not add mutable package state.
- **R7 Check all returns:** Check each error. Do not discard a call result with `_`. Check each parameter.
- **R8 Limited code generation:** Keep build tags and code generation to a minimum. Generated code is the same on each run.
- **R9 Restricted indirection:** Do not add function types, function tables or deep pointer chains. Use static dispatch.
- **R10 Zero warnings:** The compiler and all static analyzers report zero warnings.

### 0.3 ASD-STE100 documentation style

All Markdown files, OKF wiki pages, CLI help strings and code comments follow ASD-STE100 Simplified Technical English. `scripts/ste-check.py` checks them.

- Write precise, clinical text. Do not use marketing words, for example `seamless`, `robust` or `powerful`.
- Do not use the possessive `your`. Write `the ruleset file`, not `your ruleset file`.
- Keep each procedural step to 20 words or fewer. Keep each descriptive sentence to 25 words or fewer.
- Use active voice. Name the actor: "The parser reads the source file."
- Use one term for one concept. Prefer a verb to a noun form: write "configure", not "configuration". The check does not flag noun forms, because `function` and `assertion` have the same suffixes.
- Write one instruction in each sentence.
- From Phase 13, lexis also checks the Markdown files. It reports noun forms and -ing forms. `vocabulary.cue` lists the terms of the project that it accepts.

Task order: run 2.1, then 3.1 and 3.2, then 2.2. The `scan` and `rules` commands need the ruleset engine.

## Phase 1: Repository bootstrap, gates and OKF wiki

Issue: [ISSUE-001](.okfw/issues/issue-001-phase-1-repository-bootstrap-gates-and-okf-wiki--01a1e97296be.md)

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

Issue: [ISSUE-002](.okfw/issues/issue-002-phase-2-version-package-and-cobra-cli--01a1e97296d0.md)

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

Issue: [ISSUE-003](.okfw/issues/issue-003-phase-3-composable-cue-ruleset-engine--01a1e97296e2.md)

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

Issue: [ISSUE-004](.okfw/issues/issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md)

- [x] **Task 4.1: Non-recursive analyzer.** Write `internal/parser` and `internal/engine` under the rules R1 to R10:
  - Walk each syntax tree with a Tree-sitter cursor. Do not use recursion.
  - Build the call graph in a table with a fixed capacity. Find the cycles with an iterative depth-first search, the same search as in the oracles.
  - Write one checker for each check kind. Dispatch through a `switch`, not a function table.
  - Run user Tree-sitter queries (`ast_query`). A user then adds a code rule in CUE without a change to the Go code.
  - Return an error when the work exceeds a file, node, depth or finding limit.
  - Turn on the self-scan gate.
- [x] **Task 4.2: Differential parity tests.** Write `internal/engine/nasa_parity_test.go`. It runs the oracles and `semscope scan --ruleset nasa-p10` on these inputs:
  1. Violation fixtures in `testdata/nasa_p10/violations/`, one for each rule.
  2. Compliant fixtures in `testdata/nasa_p10/compliant/`.
  3. Custom ruleset fixtures in `testdata/custom_rulesets/` that add a rule, omit a rule and change a threshold.
  4. The semscope sources in `cmd/` and `internal/`.
  5. An optional corpus through `SEMSCOPE_PARITY_CORPUS`, for example the zmake `src/` tree.

  On the fixtures, semscope and the oracles report the same set of file, line and rule. A divergence on the corpus needs a fix, or an entry with a reason in `testdata/nasa_p10/parity-divergences.json`. semscope does not copy a known defect of an oracle.

## Phase 5: SARIF reports and self-hosting gate

Issue: [ISSUE-005](.okfw/issues/issue-005-phase-5-sarif-reports-and-self-hosting-gate--01a1e9729706.md)

- [x] **Task 5.1: SARIF producer.** Write the SARIF writer with `github.com/owenrumney/go-sarif/v3`:
  - Use `pkg/report/v210/sarif` for SARIF 2.1.0 and `pkg/report/v22/sarif` for SARIF 2.2. The `--sarif-version` flag selects one.
  - Set `tool.driver.name` to `"semscope"`. Set `tool.driver.version`, `tool.driver.informationUri` and `tool.driver.rules`. The rules list holds only the active rules.
  - Give each finding a repository-relative `artifactLocation.uri`, a `ruleId` and a `level`. Add the 1-based `startLine`, `startColumn`, `endLine` and `endColumn`.
  - Call `rep.Validate()` before the writer writes the report.
- [x] **Task 5.2: Self-scan gate and wiki sync.** Run the full gate:
  1. Run `zmake check` and `zmake -f Makefile check`.
  2. Run `zmake build`.
  3. Run `./bin/semscope --version`.
  4. Run `./bin/semscope scan --ruleset nasa-p10 -r .semscope.cue --format sarif --sarif-version 2.1.0 --output semscope-self.sarif .`.
  5. Run `./bin/semscope scan --ruleset nasa-p10 -r .semscope.cue --format sarif --sarif-version 2.2 --output semscope-self-v22.sarif .`.
  6. Run `zmake sarif`. It runs steps 2, 4 and 5.

  Write `README.md`. Sync the architecture records and the notes on how to use semscope to the OKF wiki. Verify zero violations in the repository.

## Phase 6: Exclude flag, Go type checks and parity corpora

Issue: [ISSUE-006](.okfw/issues/issue-006-phase-6-exclude-flag-go-type-checks-and-parity-corpora--01a1eba38e60.md)

- [x] **Task 6.1: Exclude flag.** Add `--exclude PATTERN` to `scan` and `rules inspect`. Repeat the flag to add patterns.
  - A flag pattern has the same syntax as a pattern in the `exclude` map of a ruleset.
  - The loader adds each flag pattern to the exclude map with the reason `--exclude flag`.
  - A bad pattern gives an error and exit status 2.
- [x] **Task 6.2: Type-aware Go check for NASA-P10-07.** Add the parameter `go_types: *false | bool` to `checked_returns`.
  - When `go_types` is `true`, semscope loads the Go packages with type information (`golang.org/x/tools/go/packages`).
  - The check finds each call statement, `defer` statement and `go` statement that drops an `error` result.
  - The parameter `go_ignore_calls` names calls that never return an error, for example `(*strings.Builder).WriteString`.
  - The default is `false`, so the parity with `nasa-lint-go.py` does not change. The self-scan overlay `.semscope.cue` sets it to `true`.
- [x] **Task 6.3: More parity corpora.** The corpus parity test reads a list of Zig trees and a list of Go trees. `scripts/parity_corpora.py` gives the default lists.
  - Zig: zmake and the Zig standard library when they are on the disk. ztree is not a default corpus, because it changes each day.
  - Go: the fixture `corpus-forms` and six source modules of the semscope dependencies in the module cache.
  - Each divergence gets a fix or a record with a reason in `testdata/nasa_p10/parity-divergences.json`.
  - A record is a line record, a class record with an exact count, or a stats record.

## Phase 7: Zig limits of ISSUE-008 and the oracle defects of ISSUE-007

Issue: [ISSUE-008](.okfw/issues/issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md)

- [x] **Task 7.1: Zig container keys.** Key each Zig function by its container path, for example `Mutable.toFloat`.
  - A plain call looks in the container of the caller, then in each outer container, then in the file.
  - A call through `self`, `Self` or `@This()` looks in the container of the caller only.
- [x] **Task 7.2: Zig 0.17 grammar.** Zig 0.17 has no keywords `async`, `await` and `usingnamespace`. No release of `tree-sitter-zig` removes them.
  - Keep a patched copy of the v1.1.2 grammar in `third_party/tree-sitter-zig`. A `replace` directive in `go.mod` selects it.
  - Generate `src/parser.c` with `tree-sitter-cli` 0.25.10.
- [x] **Task 7.3: Oracle defects.** Fix the defects of `nasa-lint.py` that ISSUE-007 lists.
  - Use the semscope node keys in R1, count each line for the innermost function, and read R7 across lines.
  - Add the fixture `scripts/testdata/nasa-lint/corpus-forms` and fix the semscope defects that the fix shows.
- [x] **Task 7.4: Zig 0.15 asm clobbers.** Let `asm_clobbers` of the patched grammar read an anonymous struct literal (ISSUE-009).
  - Put the SHA-256 of `parser.c` in `binding.go`, so the Go build cache sees a new parser.

## Phase 8: Documentation site on GitHub Pages

Issue: [ISSUE-010](.okfw/issues/issue-010-phase-8-documentation-site-on-github-pages--01a1ed3cf4ce.md)

- [x] **Task 8.1: Site scaffold and build tasks.** Add an Astro Starlight site in `docs/`, the same setup as the zmake site ([ADR-0011](.okfw/decisions/adr-0011-document-semscope-with-astro-starlight-on-github-pages--01a1ed3cd653.md)).
  - Use Astro 7, Starlight 0.41 and bun 1.3.9. Commit `docs/bun.lock`.
  - Set the Astro `base` to `/semscope`. Use the Starlight `splash` template for the landing page.
  - Add the tasks `docs-dev` and `docs-build` to the `ZMakefile` and the `Makefile`. `check` does not build the site.
- [x] **Task 8.2: Pages.** Write the guides, the reference and the project pages in ASD-STE100.
  - Guides: scanning, rulesets and overlays, output formats, suppressions and excludes, CI integration.
  - Reference: the CLI, the `nasa-p10` rules, the ruleset schema and the make targets.
  - Take each example output from a real run of semscope.
- [x] **Task 8.3: Pages workflow.** Add `.github/workflows/docs-release.yml`.
  - A pull request builds the site. A push to `main` builds the site and deploys it to GitHub Pages.
  - The repository Pages source must be "GitHub Actions".

## Phase 9: The Zig grammar of ztree

Issue: [ISSUE-011](.okfw/issues/issue-011-phase-9-the-zig-grammar-of-ztree--01a1f3fdc1c1.md)

- [x] **Task 9.1: Zig grammar.** Take the `grammar.js` of ztree in `third_party/tree-sitter-zig` ([ADR-0012](.okfw/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md)).
  - The grammar adds empty containers and three rare forms to the patch of ADR-0010.
  - Generate `src/parser.c` with `tree-sitter-cli` 0.25.10 and put its SHA-256 in `binding.go`.

## Phase 10: The ztree runtime

Issue: [ISSUE-012](.okfw/issues/issue-012-phase-10-the-ztree-runtime--01a1f4bf5f39.md)

- [x] **Task 10.1: Merge.** Merge the branch `ztree-runtime` into `main` ([ADR-0013](.okfw/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md)). The parser, the walker and the `ast_query` checker use the ztree Go binding.
- [x] **Task 10.2: Build task and documents.** Add the task `ztree` to the `ZMakefile` and the `Makefile`. It builds `libztree.a` before the Go tasks.
  - Update the README, the architecture overview and the site pages.
  - File [ISSUE-013](.okfw/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md) for the Zig grammar copy that has no user now.

## Phase 11: The Zig grammar lives in ztree only

Issue: [ISSUE-013](.okfw/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md)

- [x] **Task 11.1: Remove the grammar copy.** Remove `third_party/tree-sitter-zig`, its `replace` directive and `TestZigParserHash` ([ADR-0014](.okfw/decisions/adr-0014-the-zig-grammar-lives-in-ztree-only--01a1f4d211fc.md)).
  - Remove the `third_party` entry of `.semscope.cue`.
  - Mark ADR-0010 and ADR-0012 as deprecated. Update the site pages.

## Phase 12: The Go parity corpus

Issue: [ISSUE-014](.okfw/issues/issue-014-the-go-parity-corpus-lost-five-modules--01a1f5102109.md)

- [x] **Task 12.1: Corpus list.** Remove `go-tree-sitter` from `GO_MODULES` in `scripts/parity_corpora.py`. Run one `go list -m` for each module, so the script skips only a module that it cannot find.
  - `zmake parity` passes on the three Zig corpora and the six Go corpora with the ztree runtime.
  - Update the corpus list of ADR-0005.

## Phase 13: lexis checks the documentation

Issue: [ISSUE-015](.okfw/issues/issue-015-phase-13-lexis-checks-the-documentation--01a1f5265a5d.md)

- [x] **Task 13.1: Vocabulary and text.** Bring the lexis findings of the Markdown documentation to 0 ([ADR-0015](.okfw/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md)).
  - Put the technical terms of semscope in `vocabulary.cue`, each with a reason.
  - Rewrite the other prose. Keep the decisions and the facts of each ADR.
- [x] **Task 13.2: Gate.** Add the gate `docs-lexis` to the `ZMakefile` and the `Makefile`. `check` runs it after `docs-style`.
  - The gate runs lexis only when lexis is on the PATH. Each finding fails it.
  - Update `CLAUDE.md`, the wiki and the site pages.

Acceptance criteria:

- `zmake check` and `zmake -f Makefile check` pass with lexis on the PATH.
- The lexis scan gives 0 errors, 0 warnings and 0 notes.
- `zmake -n check` and `make -n check` print the same commands.
