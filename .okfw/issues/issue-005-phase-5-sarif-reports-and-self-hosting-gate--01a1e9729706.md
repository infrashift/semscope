---
id: "01a1e972-9706-7158-a6fa-fe4ee1bdbe8f"
title: "ISSUE-005: Phase 5: SARIF reports and self-hosting gate"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-5"
  - "report"
  - "in-progress"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T07:06:18-06:00"
description: "Add the SARIF 2.1.0 and 2.2 writers and run the final self-hosting gate."
---

# ISSUE-005: Phase 5: SARIF reports and self-hosting gate

**Status/Phase:** in-progress · Phase 5

## Description
Add the SARIF 2.1.0 and SARIF 2.2 writers through `go-sarif/v3`, then run the final self-hosting gate.

## Acceptance Criteria
- [x] `--sarif-version` selects `pkg/report/v210/sarif` or `pkg/report/v22/sarif`.
- [x] `tool.driver.rules` holds only the active rules.
- [x] Each result has a repository-relative URI, 1-based lines and columns, a `ruleId` and a `level`.
- [x] `rep.Validate()` runs before each write.
- [ ] Both self-scan SARIF reports exit 0.
- [ ] `README.md` passes `ste-check.py`.

## Notes
- go-sarif gives each SARIF 2.2 report a random `guid`. semscope sets a version 5 UUID of the report content instead, so the same scan gives the same report.
- The whole-repository scan found a defect: the Python dynamic-code check matched `re.compile`. It now matches plain builtin calls only.
- SARIF 2.2 is an OASIS draft. The help text says so.

## Related
- [Architecture overview](/architecture/semscope-architecture-overview--01a1e97295ff.md)
