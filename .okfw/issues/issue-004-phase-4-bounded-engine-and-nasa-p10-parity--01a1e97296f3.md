---
x-okfw-id: "01a1e972-96f3-7e23-b94a-ba41c455c04d"
title: "ISSUE-004: Phase 4: Bounded engine and nasa-p10 parity"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-4"
  - "engine"
  - "closed"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T07:01:10-06:00"
description: "Add the Tree-sitter parsers, the bounded engine and the parity tests against both oracles."
---

# ISSUE-004: Phase 4: Bounded engine and nasa-p10 parity

**Status/Phase:** closed · Phase 4

## Description
Add the Tree-sitter parsers for six languages, the bounded engine and the parity tests against both oracles.

## Acceptance Criteria
- [x] One adapter for each of Go, C, C++, Zig, Python and Rust.
- [x] Tree walks use a cursor. The cycle search is iterative and its stacks have a fixed capacity. It is a three-color depth-first search, the same as in the oracles.
- [x] Each check kind has a checker. A `switch` dispatches them.
- [x] Every bound gives an error when the work exceeds it.
- [x] The self-scan gate runs in `zmake check`.
- [x] Fixture parity with `nasa-lint.py` and `nasa-lint-go.py` is exact.
- [x] The corpus run on zmake `src/` shows no divergence without a recorded reason.

## Notes
- Parity follows [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md). The Zig grammar comes from `tree-sitter-grammars/tree-sitter-zig`.

## Resolution
Tasks 4.1 and 4.2 are complete.

- `TestNasaParityFixtures` compares 14 trees: six oracle rules and one compliant tree, for Zig and for Go. The sets of file, line and rule are equal. The function and assertion counts of R5 are equal.
- `TestNasaParitySemscopeSources` compares `nasa-lint-go.py` and semscope on the semscope sources.
- `TestSingleRuleExpected` checks each of the ten rules in all six languages against `expected.txt`.
- `zmake parity` runs `TestNasaParityCorpus` on zmake `src/`: 306 functions and 614 assertions in both tools, no findings, and no divergence. `parity-divergences.json` is empty.
- The full `nasa-p10` ruleset on zmake found one false positive of the Zig adapter. A `var` in a comptime block counted as container-level state. The adapter now needs a container parent. The sample `sample_b.zig` holds the case.

## Related
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md)
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)
