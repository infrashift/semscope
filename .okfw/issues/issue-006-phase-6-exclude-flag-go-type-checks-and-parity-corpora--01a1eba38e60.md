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
  - "in-progress"
created: "2026-11-17T16:32:07-06:00"
timestamp: "2026-11-17T16:32:07-06:00"
description: "Add the --exclude flag, a type-aware Go check for NASA-P10-07 and more corpora for the parity test."
---

# ISSUE-006: Phase 6: Exclude flag, Go type checks and parity corpora

**Status/Phase:** in-progress · Phase 6

## Description
Phase 5 closed the original plan. This phase adds the next steps from the daily note of 2026-11-17. They are a command-line exclude flag, a type-aware Go check for NASA-P10-07 and more parity corpora.

## Acceptance Criteria
- [x] `--exclude PATTERN` on `scan` and `rules inspect` adds an exclude pattern. A bad pattern gives exit status 2.
- [x] The parameter `go_types` of NASA-P10-07 turns on a Go check with type information. The check finds call statements that drop an `error` result.
- [x] The default of `go_types` is `false`, so the parity with `nasa-lint-go.py` does not change.
- [x] The self-scan turns on `go_types`.
- [ ] The parity test reads more than one corpus for each language. Each divergence has a fix or a record with a reason.

## Related
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)
- [ISSUE-005](/issues/issue-005-phase-5-sarif-reports-and-self-hosting-gate--01a1e9729706.md)
