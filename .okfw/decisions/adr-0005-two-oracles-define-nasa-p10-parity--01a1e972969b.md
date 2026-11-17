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
- On a real corpus, for example the zmake `src/` tree, each divergence needs a fix or a record in `testdata/nasa_p10/parity-divergences.json`.
- Each divergence record has a reason and a link to an issue.
- semscope does not copy a known oracle defect. An example is the merge of same-name functions into one call graph node.
- semscope keys a Zig function by its full path and its container path, for example `Mutable.toFloat`. It resolves a plain call through the lexical scopes of the caller. [ISSUE-008](/issues/issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md) records the change.
- The rules R3, R6, R8 and R10 have no oracle. Fixtures alone test them.
- For C, C++, Python and Rust, fixtures alone test all rules.

## Consequences
- The parity test runs `python3`. It fails with a clear message when `python3` is absent.
- The divergence file documents each known oracle quirk that a corpus hits.

## Results
- Fixtures: 14 trees, exact parity, including the oracle quirk that line rules apply inside Zig test helpers.
- zmake `src/`: exact parity, with 306 functions and 614 assertions in both tools.
- semscope sources: exact parity with `nasa-lint-go.py`.

## Corpora (Task 6.3)
- `scripts/parity_corpora.py` gives the default corpora of `zmake parity`.
- The Zig corpora are zmake and the standard library of the installed Zig (`zig-std-<version>`).
- The Go corpora are the fixture `scripts/testdata/nasa-lint-go/corpus-forms` and six modules of the module cache: cobra, pflag, go-sarif, CUE, go-tree-sitter and x/tools.
- A divergence record has one of three forms. A line record names one file, line and rule. A class record gives the exact count of the other differences of one rule and one side. A stats record gives the R5 counts of both tools.
- A class record is only for a defect with many instances, for example the false call edges of `nasa-lint.py` on the Zig standard library. The test fails when the count changes in either direction, and also on a stale record.
- `nasa-lint-go.py` is a semscope script. A defect that a corpus shows gets a fix and a fixture, not a record.
- ztree is not a default corpus. It is under active development, so its findings change from day to day.

Results on 2026-11-17:
- zmake, the fixture, cobra, pflag, go-sarif, CUE and go-tree-sitter: exact parity.
- x/tools: one line record. A line regex cannot see an unnamed func-typed parameter of an interface method.
- Zig 0.17.1 standard library: two class records for R1 (203 and 4), 11 line records for R4 and R7, and one stats record. [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) lists the oracle defects.
- The corpora found eight semscope defects and nine `nasa-lint-go.py` defects. Task 6.3 fixes them.

## Alternatives considered
- **Bug-for-bug parity.** Rejected: semscope then reports false recursion and misses assertions in one-line functions.
- **Parity on the semscope sources only.** Rejected: `nasa-lint.py` finds no Zig files there, so the comparison is empty.

## Related
- [ISSUE-004](/issues/issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md)
