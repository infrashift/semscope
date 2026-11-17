# `PLAN.md` — Semscope Implementation Plan for Claude Code

Save the following specification as `PLAN.md` in the root of `[github.com/infrashift/semscope](https://github.com/infrashift/semscope)`:

```markdown
# SEMSCOPE Implementation Plan

## 0. Operational Directives & Quality Gates (Mandatory)

1. **Iterative Verify-Lint-Repeat Loop:**
   Execute this plan strictly phase-by-phase and task-by-task. After completing any code or documentation task, run the full verification suite:
   - `python3 scripts/nasa-lint.py`
   - `python3 scripts/ste-check.py`
   - `go test -v -race ./...`
   - `golangci-lint run`
   Do not proceed to the next task until all checks exit with status `0` and zero warnings.

2. **NASA Power of 10 Code Compliance:**
   Every Go file in `semscope` must obey Gerard J. Holzmann's Power of 10 rules (https://en.wikipedia.org/wiki/The_Power_of_10:_Rules_for_Developing_Safety-Critical_Code):
   - **Rule 1 (Simple Control Flow):** Zero recursion (direct or indirect), zero `goto`. Use explicit bounded stacks for AST/CST traversal and call-graph cycle checks.
   - **Rule 2 (Fixed Loop Bounds):** Every loop must have a statically verifiable upper bound (`MaxFiles`, `MaxASTNodes`, `MaxStackDepth`) with an explicit guard error if breached.
   - **Rule 3 (No Dynamic Allocation After Init):** Pre-allocate worker pools, AST node traversal stacks, and result buffers during engine initialization (`NewEngine`).
   - **Rule 4 (Short Functions):** Maximum 60 lines per function.
   - **Rule 5 (Assertion Density):** Enforce minimum two precondition/invariant checks per function where applicable.
   - **Rule 6 (Smallest Scope):** Declare variables at the narrowest lexical scope.
   - **Rule 7 (Check All Returns):** Check every error return value and parameter; never discard errors with `_`.
   - **Rule 8 (Restricted Preprocessor / Codegen):** Keep build tags and code generation minimal and deterministic.
   - **Rule 9 (Restricted Indirection):** Avoid deep pointer chains and dynamic function pointer tables where static dispatch suffices.
   - **Rule 10 (Zero Warnings):** Compile and pass all static analyzers with zero warnings.

3. **ASD-STE100 & Clinical Documentation Style (Vizzini Standard):**
   All Markdown files, OKF wiki pages, CLI help strings, and code comments must pass `scripts/ste-check.py` and adhere to these constraints:
   - **Clinical Precision & Zero Marketing Fluff:** Write with dry, clinical objectivity. Never use marketing adjectives or intensifiers ("seamless", "blazingly fast", "robust", "effortless", "powerful", "modern", "out-of-the-box", "simply", "easily").
   - **Non-Possessive Pronouns:** Never use possessive second-person pronouns ("your"). Replace "your" with deterministic articles such as "the" (e.g., write "the ruleset file" instead of "your ruleset file").
   - **Sentence Length Limits:** Maximum **20 words** for procedural steps (imperatives, list items). Maximum **25 words** for descriptive prose sentences.
   - **Active Voice Only:** Prohibit passive constructions ("is parsed by", "was created", "will be evaluated"). State the actor explicitly ("The parser reads the source file").
   - **Lexical Economy (One Term Per Concept):** Maintain a 1:1 mapping between technical terms and concepts. Avoid nominalizations (`-tion`, `-ment`, `-ance`) and ambiguous `-ing` gerunds.
   - **One Instruction Per Sentence:** Do not combine sequential actions into a single sentence.

---

## Phase 1: Repository Bootstrap, Baseline Linters & OKF Wiki

- [ ] **Task 1.1: Go Module & Directory Layout**
  Initialize `github.com/infrashift/semscope` (Go 1.23+). Create the package structure:
  - `cmd/semscope/` (Entrypoint and `spf13/cobra` CLI commands)
  - `internal/version/` (Build-time version, commit, date, and supported SARIF versions)
  - `internal/ruleset/` (CUE ruleset loader, rule composition/omission/override evaluator, and embedded `nasa-p10` ruleset)
  - `internal/parser/` (Go `go/ast` + Tree-sitter CST parsers for multi-language code analysis)
  - `internal/engine/` (Non-recursive bounded AST matcher, structural query evaluator, and metric analyzer)
  - `internal/report/` (SARIF v2.1.0 and v2.2+ generator via `go-sarif/v3`, plus text/JSON output)
  - `rules/nasa_p10/` (Embedded CUE definition of the NASA Power of 10 ruleset)
  - `scripts/` (Baseline linter scripts copied from `zmake`)

- [ ] **Task 1.2: Copy Baseline Linter Scripts from `zmake`**
  Copy the two reference linter scripts from `_my_workspace_/github.com/infrashift/zmake/scripts/` into `scripts/`:
  - `scripts/nasa-lint.py` (Serves as both the mandatory code gate and the **baseline test oracle** for `semscope`'s NASA Power of 10 ruleset).
  - `scripts/ste-check.py` (Enforces ASD-STE100 compliance on all documentation and comments).
  Inspect `scripts/nasa-lint.py` thoroughly and document every check, threshold, regex/AST rule, and file-type target it implements.

- [ ] **Task 1.3: OKF Wiki Initialization & MCP Connection**
  - Review `_my_workspace_/github.com/infrashift/zmake/.okfw` to identify the common OKF wiki bundles, frontmatter schemas, and MADR templates.
  - Initialize the repository OKF wiki using `~/.local/bin/okfw` configured with the standard bundles from `zmake/.okfw`.
  - Connect to the OKF wiki MCP server and author the foundational architecture docs and ADR-0001 (Composable CUE Code Rulesets and NASA Power of 10 Baseline).
  - Run `python3 scripts/ste-check.py` on all wiki and Markdown files until green.

---

## Phase 2: Cobra CLI, `--version`, and Ruleset Overrides

- [ ] **Task 2.1: Version Package (`internal/version`)**
  Implement `internal/version` with build-time `-ldflags` variables (`Version`, `Commit`, `BuildDate`) and supported SARIF schema versions (`2.1.0`, `2.2`).

- [ ] **Task 2.2: Cobra CLI & Flags (`cmd/semscope`)**
  Build the CLI using `github.com/spf13/cobra`:
  - Support `semscope --version` (and `-v` / `semscope version`) outputting deterministic version metadata.
  - Implement `semscope scan [paths...]` with flags for full ruleset customization:
    - `--ruleset, -r <name-or-path>`: Select a built-in ruleset (default: `nasa-p10`) or one or more user-defined `.cue` files/directories.
    - `--enable-rule <id[,id...]>`: Explicitly enable specific rules.
    - `--omit-rule, --disable-rule <id[,id...]>`: Omit specific rules from the active ruleset.
    - `--set <rule_id.param=value>`: Override rule parameters or thresholds from the CLI (e.g., `--set NASA-P10-04.max_lines=50`).
    - `--format <sarif|text|json>`: Output format.
    - `--sarif-version <2.1.0|2.2>`: Target SARIF specification version (default `2.1.0`, supporting `2.2` and higher).
    - `--output, -o <file>`: Output file path (defaults to stdout).
  - Implement `semscope rules list` and `semscope rules inspect <ruleset>` to display active rules, omitted rules, and effective parameter bounds.
  - Verify with `nasa-lint.py`, `ste-check.py`, and `go test ./...`. Repeat until green.

---

## Phase 3: Composable CUE Ruleset Engine & NASA Power of 10 (`internal/ruleset`)

- [ ] **Task 3.1: Extensible CUE Ruleset Schema (`internal/ruleset/schema.cue`)**
  Using `cuelang.org/go/cue/cuecontext`, define a strict CUE schema supporting user-defined rulesets and composition:
  - Rules are keyed by stable ID (`rules: [ID=string]: #Rule & { id: ID }`).
  - Every `#Rule` includes:
    - `id`: Stable identifier (e.g., `"NASA-P10-01"`).
    - `enabled`: `bool | *true` (allows users to omit a rule in a custom overlay via `rules: "NASA-P10-03": enabled: false`).
    - `severity`: `*"error" | "warning" | "note"`.
    - `languages`: `[...("go" | "c" | "cpp" | "zig" | "python" | "rust")]`.
    - `title` and `message`: Clinical description and violation template.
    - `params`: Open struct `{ [string]: int | string | bool | [...string] }` for user-tunable thresholds (e.g., `max_lines: *60 | int`, `min_assertions: *2 | int`).
    - `check`: Discriminated union defining the analyzer check type (`"no_goto_or_recursion"`, `"bounded_loops"`, `"no_dynamic_alloc"`, `"max_function_lines"`, `"min_assertions"`, `"narrowest_scope"`, `"checked_returns"`, `"restricted_preprocessor"`, `"max_pointer_indirection"`, `"ast_query"`, `"token_ban"`).
  - Support loading a base ruleset (like `nasa-p10`), unifying user `.cue` overlay files to add new custom rules, omitting disabled rules (`enabled: false` or `--omit-rule`), and applying parameter overrides.

- [ ] **Task 3.2: First Supported Ruleset — NASA Power of 10 (`rules/nasa_p10/nasa_p10.cue`)**
  Implement the complete NASA Power of 10 ruleset (`nasa-p10`) in embedded CUE, matching and generalizing every check in `scripts/nasa-lint.py`:
  - `NASA-P10-01`: Simple control flow (prohibit `goto`, `setjmp`/`longjmp`, and direct/indirect recursion).
  - `NASA-P10-02`: Fixed upper bounds on all loops (reject unbounded `for {}` / `while(true)` or loops lacking an explicit upper-bound comparison or range bound).
  - `NASA-P10-03`: No dynamic heap allocation after initialization (flag runtime heap allocations/escapes or banned allocator calls outside init phases).
  - `NASA-P10-04`: Function length limit (`params.max_lines: *60 | int`).
  - `NASA-P10-05`: Assertion / precondition density (`params.min_assertions: *2 | int` per non-trivial function).
  - `NASA-P10-06`: Smallest possible lexical scope for data declarations (flag mutable package/global state or wide-scope variables).
  - `NASA-P10-07`: Check return values of all non-void functions and validate parameters (flag blank-identifier `_ = fn()` error drops and unhandled call expressions).
  - `NASA-P10-08`: Restricted preprocessor / macro / conditional build complexity.
  - `NASA-P10-09`: Restricted pointer indirection (`params.max_pointer_depth: *1 | int`) and function pointer restrictions.
  - `NASA-P10-10`: Zero compiler/analyzer warnings and banned unsafe constructs.
  - Write unit tests verifying that user CUE files can: (a) add a custom rule `CUSTOM-01`, (b) omit `NASA-P10-05` via `enabled: false` or `--omit-rule`, and (c) modify `NASA-P10-04.params.max_lines` to `40`.

---

## Phase 4: Bounded Static Analysis Engine & `nasa-lint.py` Parity (`internal/engine`)

- [ ] **Task 4.1: Non-Recursive AST Analyzer (`internal/parser` & `internal/engine`)**
  Implement the code analysis engine adhering strictly to NASA Power of 10 Rules 1–10:
  - Use explicit pre-allocated node stacks (`[]astNodeFrame`) to walk syntax trees with zero recursion.
  - Implement call-graph construction using iterative Tarjan/BFS cycle detection on a fixed-capacity adjacency table to detect direct and indirect recursion (`NASA-P10-01`).
  - Implement AST structural checkers for Rules 1 through 10 plus custom user-defined Tree-sitter S-expression / AST pattern queries (`ast_query`) so users can define arbitrary new code rules in CUE without modifying Go code.

- [ ] **Task 4.2: Differential Parity Testing Against `scripts/nasa-lint.py`**
  Create a comprehensive test harness (`internal/engine/nasa_parity_test.go`) that runs both `python3 scripts/nasa-lint.py` and `semscope scan --ruleset nasa-p10` against:
  1. Dedicated violation fixtures (`testdata/nasa_p10/violations/`) triggering each of the 10 rules individually.
  2. Clean compliant fixtures (`testdata/nasa_p10/compliant/`).
  3. Custom user ruleset fixtures (`testdata/custom_rulesets/`) testing rule addition, rule omission, and threshold changes.
  4. The `semscope` source tree itself (`cmd/`, `internal/`).
  Iterate until `semscope` achieves 100% detection parity with `nasa-lint.py` and passes all checks.

---

## Phase 5: SARIF v2.1.0+ Reporting & Final Self-Hosting Gate (`internal/report`)

- [ ] **Task 5.1: SARIF v2.1.0 and v2.2+ Producer (`github.com/owenrumney/go-sarif/v3`)**
  Implement the SARIF reporter using `github.com/owenrumney/go-sarif/v3`:
  - Support both `pkg/report/v210/sarif` (SARIF v2.1.0) and `pkg/report/v22/sarif` (SARIF v2.2) selectable via `--sarif-version`.
  - Populate `tool.driver.name` (`"semscope"`), `tool.driver.version`, `tool.driver.informationUri`, and `tool.driver.rules` (only including active, non-omitted rules from the resolved CUE ruleset).
  - Record each violation with repository-relative `artifactLocation.uri`, 1-based `startLine`, `startColumn`, `endLine`, `endColumn`, `ruleId`, and `level`.
  - Validate the generated SARIF report via `rep.Validate()` before writing to output.

- [ ] **Task 5.2: Final Self-Hosting Verification & OKF Wiki Sync**
  Execute the full self-hosting gate:
  1. `python3 scripts/nasa-lint.py`
  2. `python3 scripts/ste-check.py`
  3. `./bin/semscope --version`
  4. `./bin/semscope scan --ruleset nasa-p10 --format sarif --sarif-version 2.1.0 --output semscope-self.sarif .`
  5. `./bin/semscope scan --ruleset nasa-p10 --format sarif --sarif-version 2.2 --output semscope-self-v22.sarif .`
  6. `go test -v -race ./...`
  Sync final architecture records and usage documentation to the OKF wiki via MCP and verify zero violations across the entire repository.