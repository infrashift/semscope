---
x-okfw-id: "01a1e972-96d0-776e-af09-67e792f2839a"
title: "ISSUE-002: Phase 2: Version package and Cobra CLI"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-2"
  - "cli"
  - "closed"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T06:35:40-06:00"
description: "Add internal/version and the cobra commands version, scan and rules with every ruleset flag."
---

# ISSUE-002: Phase 2: Version package and Cobra CLI

**Status/Phase:** closed · Phase 2

## Description
Add `internal/version` and the cobra command tree. Run Task 2.2 after Phase 3, because the `scan` and `rules` commands need the ruleset engine.

## Acceptance Criteria
- [x] `internal/version` holds `Version`, `Commit` and `BuildDate` from `-ldflags`, with a fallback to `debug.ReadBuildInfo`.
- [x] `internal/version` lists the SARIF versions `2.1.0` and `2.2`.
- [x] `semscope --version`, `semscope -v` and `semscope version` print the same metadata.
- [x] `semscope scan` accepts every flag of the [CLI contract](/specs/cli-contract--01a1e9729633.md).
- [x] `semscope rules list` and `semscope rules inspect` work.
- [x] Golden tests cover the help text and the version output.

## Notes
- Cobra callbacks are function literals. ADR-0003 allows them.

## Resolution
Tasks 2.1 and 2.2 are complete. `semscope scan` validates the flags and loads the ruleset. Until Task 4.1, it then exits with status 2 and the message that the engine is not available. [ADR-0007](/decisions/adr-0007-cobra-parses-the-command-line--01a1e97fa733.md) records the cobra choice.

## Related
- [CLI contract](/specs/cli-contract--01a1e9729633.md)
