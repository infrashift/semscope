---
id: "01a1e972-969b-7d42-af22-04302b03a42b"
title: "ADR-0005: Two oracles define nasa-p10 parity"
type: "ADR"
tags:
  - "adr"
  - "parity"
  - "testing"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "nasa-lint.py (Zig) and nasa-lint-go.py (Go) are the parity oracles; fixtures match exactly and corpus divergences need a recorded reason."
---

# ADR-0005: Two oracles define nasa-p10 parity

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request (`PLAN.md` Task 4.2). User decision on 2026-11-17 to add a Go oracle.

## Context
`PLAN.md` asks for full detection parity with `nasa-lint.py`. That script reads Zig only, has heuristic quirks and covers six of the ten rules. The [check inventory](/specs/nasa-lint-py-check-inventory--01a1e9729610.md) lists the quirks.

## Decision
- `nasa-lint.py` is the Zig oracle. It stays byte-identical with the zmake copy.
- `nasa-lint-go.py` is the Go oracle. Its [check inventory](/specs/nasa-lint-go-py-check-inventory--01a1e9729621.md) lists its rules.
- On the fixtures in `testdata/nasa_p10/`, semscope and the oracle report the same set of file, line and rule. The R5 averages match to two decimals.
- On a real corpus, for example the zmake `src/` tree, each divergence needs a fix or an entry in `testdata/nasa_p10/parity-divergences.json`.
- Each divergence entry has a reason and a link to an issue.
- semscope does not copy a known oracle defect, for example the merge of same-name functions into one call graph node.
- The rules R3, R6, R8 and R10 have no oracle. Fixtures alone test them.
- For C, C++, Python and Rust, fixtures alone test all rules.

## Consequences
- The parity test runs `python3`. It fails with a clear message when `python3` is absent.
- The divergence file documents each known oracle quirk that a corpus hits.

## Results
- Fixtures: 14 trees, exact parity, including the oracle quirk that line rules apply inside Zig test helpers.
- zmake `src/`: exact parity, with 306 functions and 614 assertions in both tools.
- semscope sources: exact parity with `nasa-lint-go.py`.

## Alternatives considered
- **Bug-for-bug parity.** Rejected: semscope then reports false recursion and misses assertions in one-line functions.
- **Parity on the semscope sources only.** Rejected: `nasa-lint.py` finds no Zig files there, so the comparison is empty.

## Related
- [ISSUE-004](/issues/issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md)
