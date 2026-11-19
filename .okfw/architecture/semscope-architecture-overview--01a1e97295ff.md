---
id: "01a1e972-95ff-7186-893d-12deaeef1747"
title: "semscope architecture overview"
type: "Architecture"
tags:
  - "architecture"
  - "overview"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T07:07:37-06:00"
description: "The packages of semscope and the data flow from the command line to a SARIF report."
status: "stable"
---

# semscope architecture overview

semscope is a Go command-line tool. It reads a CUE ruleset, parses source files with ztree, a Tree-sitter runtime in Zig, and writes the findings as SARIF, JSON or text.

## Data flow
1. `cmd/semscope` parses the flags with cobra.
2. `internal/ruleset` loads the base ruleset and unifies the overlays. It applies `--set`, `--enable-rule` and `--omit-rule`.
3. `internal/parser` parses each source file with the Tree-sitter grammar of its language. It gives a flat record of the file to the engine.
4. `internal/engine` runs the check of each active rule on each record. It builds the call graph for the recursion check.
5. `internal/report` writes the findings in the format that the flags select.

## Packages
| Package | Responsibility |
|---|---|
| `cmd/semscope` | Entry point and cobra commands: `scan`, `rules`, `version`. |
| `internal/assert` | `assert.That`, the assertion primitive for rule R5. |
| `internal/version` | Build metadata from `-ldflags` and the supported SARIF versions. |
| `internal/ruleset` | CUE schema, loader, overlays, omissions and overrides. Gives the resolved rules. |
| `internal/parser` | One Tree-sitter adapter for each language. Gives a flat record of each file. |
| `internal/engine` | Bounded worker pool, checkers, call graph, structural queries. |
| `internal/gotypes` | Loads Go packages with type information. Finds the calls that drop an error result. |
| `internal/report` | SARIF 2.1.0, SARIF 2.2, JSON and text writers. |
| `rules/nasa_p10` | The CUE source of the `nasa-p10` ruleset, embedded in the binary. |

## Bounds
The engine has a fixed upper bound for each kind of work: files, file size, syntax nodes, tree depth, call graph nodes and findings. When the work exceeds a bound, the engine returns an error. It does not truncate the result.

## Parser
- One `parser.Parser` holds a Tree-sitter parser and the six grammars. A worker owns one parser.
- The parser, the trees and the queries come from the ztree Go binding. cgo links `libztree.a` of the ztree checkout ([ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md)).
- The walker visits the tree in pre-order with a `TreeCursor`. It does not use recursion. `MaxNodes` and `MaxDepth` bound the walk.
- A stack with a fixed capacity (`MaxNesting`) tracks the enclosing function. The walker pops a function when a node starts after its last byte.
- One adapter for each language turns nodes into a flat `parser.File`: functions, calls and sites. A site is a construct that a check reports, for example an unbounded loop or a global variable.
- The adapter skips a Zig `test` block and a Rust `#[cfg(test)]` module. A Go `_test.go` file and a Python test file are exempt as a whole.
- `File.Stripped` holds the source lines with comments and strings blanked. `token_ban` rules read it.

## Engine
- `engine.New` compiles the patterns and queries once and allocates the worker pool: one parser and one query cursor for each worker.
- `Engine.Scan` collects the files in sorted sequence. It skips hidden, `vendor`, `testdata`, `node_modules`, `zig-cache`, `zig-out`, `target` and `__pycache__` directories.
- Each worker parses a file, runs the per-file checks and closes the tree. A `switch` on the check kind selects the checker.
- After the pool, the engine runs the checks over all files: the call graph for recursion and the assertion average for each language.
- The cycle search is an iterative three-color depth-first search over sorted nodes, with one finding for each back edge. The oracles use the same search, so the findings match. A Tarjan search gives components, not the cycle path of each back edge.
- Each finding goes through the suppression check: `semscope: allow(<ID>)`, or `nasa-lint: allow(R<n>)` for a NASA-P10 rule, on the line or the line above.

## Reports
- `internal/gotypes` runs only when a `checked_returns` rule sets `go_types`. [ADR-0009](/decisions/adr-0009-type-information-for-the-go-check-of-nasa-p10-07--01a1eba86090.md) records the design.
- `internal/report` writes text, JSON and SARIF. [ADR-0008](/decisions/adr-0008-go-sarif-writes-the-sarif-reports--01a1e9998cbd.md) records the SARIF design.
- The JSON report also gives the function and assertion counts of each language. The parity tests read them.
- Each report is the same on each run. The engine sorts the findings, and the SARIF 2.2 `guid` comes from the report content.

## Self-scan
`zmake selfscan` runs `semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning .` on the whole repository. The overlay `.semscope.cue` turns off NASA-P10-03 for the R3 deviation of ADR-0003. It also excludes the Python gate scripts. `zmake sarif` writes the same scan as SARIF 2.1.0 and SARIF 2.2.

## Documentation checks
`zmake docs-style` checks the documentation with `scripts/ste-check.py`. `zmake docs-lexis` scans the Markdown files with lexis and the repository vocabulary `vocabulary.cue`, when lexis is on the PATH. [ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md) records the gate.

## Related
- [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md): composable CUE rulesets.
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md): the Power of 10 rules in Go.
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md): Tree-sitter for all six languages.
- [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md): the ztree runtime.
- [CUE ruleset schema](/specs/cue-ruleset-schema--01a1e9729645.md) and [CLI contract](/specs/cli-contract--01a1e9729633.md).
