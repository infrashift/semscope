---
x-okfw-id: "01a1f4bf-5f39-739c-be17-aeae56288355"
title: "ISSUE-012: Phase 10: the ztree runtime"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "parser"
  - "ztree"
  - "phase-10"
  - "closed"
x-okfw-created: "2026-11-19T10:59:04-06:00"
x-okfw-timestamp: "2026-11-19T10:59:04-06:00"
description: "semscope parses with the ztree Go binding in place of go-tree-sitter, and the gates build libztree.a first."
---

# ISSUE-012: Phase 10: the ztree runtime

**Status/Phase:** closed · Phase 10

## Description
The local branch `ztree-runtime` moves semscope from go-tree-sitter to the ztree Go binding. The merge into `main` needs a record of the decision and a build task for `libztree.a`. Without that task, a new checkout cannot link.

## Acceptance Criteria
- [x] `main` holds the branch `ztree-runtime` (a fast-forward merge).
- [x] [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md) records the decision. [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md) names it.
- [x] The `ZMakefile` and the `Makefile` have a task `ztree`. The build, test, lint, self-scan and parity tasks run it first.
- [x] The README, the architecture overview and the site pages name ztree.
- [x] `zmake check`, `zmake -f Makefile check` and `zmake parity` pass.

## Resolution
Task 10.1 merges the branch. Task 10.2 adds the task `ztree` and the documentation. [ISSUE-013](/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md) tracks the Zig grammar copy that has no user now.
