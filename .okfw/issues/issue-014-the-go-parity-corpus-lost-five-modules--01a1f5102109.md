---
id: "01a1f510-2109-72f1-9ff7-63701f14d04e"
title: "ISSUE-014: The Go parity corpus lost five modules"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "parity"
  - "scripts"
  - "phase-12"
  - "closed"
created: "2026-11-19T12:27:17-06:00"
timestamp: "2026-11-19T12:27:17-06:00"
description: "scripts/parity_corpora.py listed go-tree-sitter after semscope removed it, so go list failed and the Go parity corpus held the fixture only."
---

# ISSUE-014: The Go parity corpus lost five modules

**Status/Phase:** closed · Phase 12

## Description
`scripts/parity_corpora.py` gives the default corpora of the parity test ([ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)). Its list of Go modules held `github.com/tree-sitter/go-tree-sitter`. Since [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md) ([ISSUE-012](/issues/issue-012-phase-10-the-ztree-runtime--01a1f4bf5f39.md)), semscope does not use that module. The script sent all six modules to one `go list -m` command. That command failed with "not a known dependency", and the script then gave the fixture `corpus-forms` only. So `zmake parity` did not compare cobra, pflag, go-sarif, CUE and x/tools, and it gave no warning.

## Acceptance Criteria
- [x] Remove `go-tree-sitter` from `GO_MODULES`.
- [x] Run one `go list -m` for each module, so the script skips only a module that it cannot find.
- [x] `python3 -I scripts/parity_corpora.py go` gives six corpora.
- [x] `zmake parity` passes on the nine corpora with the ztree runtime.
- [x] Update the corpus list of ADR-0005.

## Resolution
Task 12.1 makes the two changes to the script. `zmake parity` passes on the three Zig corpora and on the six Go corpora. The Go corpora are the fixture, cobra 1.10.2, pflag 1.0.10, go-sarif 3.3.2, CUE 0.15.4 and x/tools 0.52.0. The record for x/tools in `parity-divergences.json` stays valid. The ztree runtime causes no new divergence.
