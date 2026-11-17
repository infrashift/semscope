---
id: "01a1e972-95ff-7186-893d-12deaeef1747"
title: "semscope architecture overview"
type: "Architecture"
tags:
  - "architecture"
  - "overview"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:21:28-06:00"
description: "The packages of semscope and the data flow from the command line to a SARIF report."
status: "draft"
---

# semscope architecture overview

semscope is a Go command-line tool. It reads a CUE ruleset, parses source files with Tree-sitter and writes the findings as SARIF, JSON or text.

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
| `internal/report` | SARIF 2.1.0, SARIF 2.2, JSON and text writers. |
| `rules/nasa_p10` | The CUE source of the `nasa-p10` ruleset, embedded in the binary. |

## Bounds
The engine has a fixed upper bound for each kind of work: files, file size, syntax nodes, tree depth, call graph nodes and findings. When the work exceeds a bound, the engine returns an error. It does not truncate the result.

## Related
- [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md): composable CUE rulesets.
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md): the Power of 10 rules in Go.
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md): Tree-sitter for all six languages.
- [CUE ruleset schema](/specs/cue-ruleset-schema--01a1e9729645.md) and [CLI contract](/specs/cli-contract--01a1e9729633.md).
