---
id: "01a1eba3-8e60-745e-ac59-77059f1accd0"
title: "ISSUE-006: Phase 6: Exclude flag, Go type checks and parity corpora"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-6"
  - "cli"
  - "engine"
  - "parity"
  - "closed"
created: "2026-11-17T16:32:07-06:00"
timestamp: "2026-11-17T16:32:07-06:00"
description: "Add the --exclude flag, a type-aware Go check for NASA-P10-07 and more corpora for the parity test."
---

# ISSUE-006: Phase 6: Exclude flag, Go type checks and parity corpora

**Status/Phase:** closed · Phase 6

## Description
Phase 5 closed the original plan. This phase adds the next steps from the daily note of 2026-11-17. They are a command-line exclude flag, a type-aware Go check for NASA-P10-07 and more parity corpora.

## Acceptance Criteria
- [x] `--exclude PATTERN` on `scan` and `rules inspect` adds an exclude pattern. A bad pattern gives exit status 2.
- [x] The parameter `go_types` of NASA-P10-07 turns on a Go check with type information. The check finds call statements that drop an `error` result.
- [x] The default of `go_types` is `false`, so the parity with `nasa-lint-go.py` does not change.
- [x] The self-scan turns on `go_types`.
- [x] The parity test reads more than one corpus for each language. Each divergence has a fix or a record with a reason.

## Resolution
Tasks 6.1, 6.2 and 6.3 are complete.

- `--exclude` adds a path pattern on the command line. `rules inspect` shows it with the reason `--exclude flag`.
- `internal/gotypes` gives NASA-P10-07 a Go check with type information ([ADR-0009](/decisions/adr-0009-type-information-for-the-go-check-of-nasa-p10-07--01a1eba86090.md)). The self-scan turns it on. It found no defect in semscope, because golangci-lint already checks the code.
- `zmake parity` runs 9 corpora. Seven have exact parity. x/tools has one line record. The Zig 0.17.1 standard library has records that [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) explains.
- The corpora found eight semscope defects. One was an assertion failure on valid Zig input. Task 6.3 fixes all eight. [ISSUE-008](/issues/issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md) holds two Zig limits that stay open.

## Related
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)
- [ISSUE-005](/issues/issue-005-phase-5-sarif-reports-and-self-hosting-gate--01a1e9729706.md)
