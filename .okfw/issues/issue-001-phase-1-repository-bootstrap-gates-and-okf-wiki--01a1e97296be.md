---
x-okfw-id: "01a1e972-96be-77de-bcb0-886662cd4f77"
title: "ISSUE-001: Phase 1: Repository bootstrap, gates and OKF wiki"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-1"
  - "bootstrap"
  - "closed"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T06:21:28-06:00"
description: "Bootstrap git, the zmake gates, the Go module, the oracle scripts, the STE gate and the wiki."
---

# ISSUE-001: Phase 1: Repository bootstrap, gates and OKF wiki

**Status/Phase:** closed · Phase 1

## Description
Bootstrap the repository: git, the zmake gates, the Go module, the oracle scripts, the STE gate and this wiki.

## Acceptance Criteria
- [x] `zmake check` and `zmake -f Makefile check` pass.
- [x] The Go module has the package layout of `PLAN.md` and the package `internal/assert`.
- [x] `scripts/nasa-lint-go.py` passes its golden fixture tests.
- [x] `scripts/ste-check.py` reads `PLAN.md`, the Go comments and the cobra help strings.
- [x] The wiki holds the guides, the architecture overview, four specs, ADR-0001 to ADR-0006 and the phase issues.

## Notes
- `nasa-lint.py` finds no functions in this repository, because it reads `src/**/*.zig` only. See [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md).
- Before Task 1.2b, `ste-check.py` skipped `PLAN.md`. A `markdown` code fence at line 5 never closed.
- The session that wrote the wiki had no okfw MCP tools. It used the `okfw` CLI.

## Resolution
Tasks 1.0 to 1.3 are complete. Each task has one commit. Both gates pass.

## Related
- [Development workflow](/development-workflow--01a1e97295ed.md)
- [ADR-0006](/decisions/adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md)
