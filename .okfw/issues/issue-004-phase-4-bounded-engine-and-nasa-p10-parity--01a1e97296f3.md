---
id: "01a1e972-96f3-7e23-b94a-ba41c455c04d"
title: "ISSUE-004: Phase 4: Bounded engine and nasa-p10 parity"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-4"
  - "engine"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "Add the Tree-sitter parsers, the bounded engine and the parity tests against both oracles."
---

# ISSUE-004: Phase 4: Bounded engine and nasa-p10 parity

**Status/Phase:** open · Phase 4

## Description
Add the Tree-sitter parsers for six languages, the bounded engine and the parity tests against both oracles.

## Acceptance Criteria
- [ ] One adapter for each of Go, C, C++, Zig, Python and Rust.
- [ ] Tree walks use a cursor. The call graph uses an iterative Tarjan search on a table with a fixed capacity.
- [ ] Each check kind has a checker. A `switch` dispatches them.
- [ ] Every bound gives an error when the work exceeds it.
- [ ] The self-scan gate runs in `zmake check`.
- [ ] Fixture parity with `nasa-lint.py` and `nasa-lint-go.py` is exact.
- [ ] The corpus run on zmake `src/` shows no divergence without a recorded reason.

## Notes
- Parity follows [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md). The Zig grammar comes from `tree-sitter-grammars/tree-sitter-zig`.

## Related
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md)
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)
