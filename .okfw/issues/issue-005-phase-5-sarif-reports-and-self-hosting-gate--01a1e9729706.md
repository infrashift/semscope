---
id: "01a1e972-9706-7158-a6fa-fe4ee1bdbe8f"
title: "ISSUE-005: Phase 5: SARIF reports and self-hosting gate"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-5"
  - "report"
  - "closed"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T07:07:37-06:00"
description: "Add the SARIF 2.1.0 and 2.2 writers and run the final self-hosting gate."
---

# ISSUE-005: Phase 5: SARIF reports and self-hosting gate

**Status/Phase:** closed · Phase 5

## Description
Add the SARIF 2.1.0 and SARIF 2.2 writers through `go-sarif/v3`, then run the final self-scan gate.

## Acceptance Criteria
- [x] `--sarif-version` selects `pkg/report/v210/sarif` or `pkg/report/v22/sarif`.
- [x] `tool.driver.rules` holds only the active rules.
- [x] Each result has a repository-relative URI, 1-based lines and columns, a `ruleId` and a `level`.
- [x] `rep.Validate()` runs before each write.
- [x] Both self-scan SARIF reports exit 0.
- [x] `README.md` passes `ste-check.py`.

## Notes
- go-sarif gives each SARIF 2.2 report a random `guid`. semscope sets a version 5 UUID of the report content instead, so the same scan gives the same report.
- The whole-repository scan found a defect: the Python dynamic-code check matched `re.compile`. It now matches plain builtin calls only.
- SARIF 2.2 is an OASIS draft. The help text says so.

## Resolution
Tasks 5.1 and 5.2 are complete. The final gate on 2026-11-17 gave these results:

- `zmake check` and `zmake -f Makefile check`: exit 0.
- `nasa-lint-go.py`: 157 functions, 314 assertions, no errors. `ste-check.py --strict`: no errors and no warnings.
- `./bin/semscope --version`: the version, the commit and the build date from `-ldflags`.
- The SARIF 2.1.0 and SARIF 2.2 self-scans: 9 active rules, no results, exit 0.
- `go test -v -race ./...`: 123 tests pass. `zmake parity` passes on zmake `src/`.
- The okfw MCP tools were not loaded in the build session. The wiki updates used the `okfw` CLI.

## Related
- [Architecture overview](/architecture/semscope-architecture-overview--01a1e97295ff.md)
