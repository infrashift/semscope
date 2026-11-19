---
id: "01a1f4bf-5f3f-7d7a-8037-1824c8c4e516"
title: "ISSUE-013: The unused Zig grammar in third_party"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "parser"
  - "zig"
  - "phase-11"
  - "closed"
created: "2026-11-19T10:59:04-06:00"
timestamp: "2026-11-19T10:59:04-06:00"
description: "third_party/tree-sitter-zig and its replace directive have no user in the build since semscope uses the ztree runtime."
---

# ISSUE-013: The unused Zig grammar in third_party

**Status/Phase:** closed · Phase 11

## Description
Since [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md), the grammars come from ztree. These parts of semscope have no user in the build:
- `third_party/tree-sitter-zig`, the patched Zig grammar of [ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md) and [ADR-0012](/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md).
- The `replace` directive for `github.com/tree-sitter-grammars/tree-sitter-zig` in `go.mod`.
- `TestZigParserHash`, which checks the hash in the copy's `binding.go`.

A later change of the Zig grammar occurs in ztree only. The copy then differs from the grammar that semscope uses.

## Acceptance Criteria
- [x] Remove `third_party/tree-sitter-zig`, its `replace` directive and `TestZigParserHash`. Or record a reason to keep them.
- [x] Mark ADR-0010 and ADR-0012 as superseded where they name the copy.
- [x] Update the site pages that name `third_party/tree-sitter-zig`.
- [x] `zmake check`, `zmake -f Makefile check` and `zmake parity` pass.

## Resolution
Task 11.1 removes the copy, the `replace` directive, `TestZigParserHash` and the `third_party` entry of `.semscope.cue` ([ADR-0014](/decisions/adr-0014-the-zig-grammar-lives-in-ztree-only--01a1f4d211fc.md)). ADR-0014 deprecates ADR-0010 and ADR-0012. The Zig tests of `internal/parser` now check the grammar of ztree, and they pass.
