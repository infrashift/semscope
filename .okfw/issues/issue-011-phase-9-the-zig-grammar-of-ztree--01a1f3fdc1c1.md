---
id: "01a1f3fd-c1c1-7edd-9080-ce128e270612"
title: "ISSUE-011: Phase 9: the Zig grammar of ztree"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "parser"
  - "zig"
  - "phase-9"
  - "closed"
created: "2026-11-19T07:27:36-06:00"
timestamp: "2026-11-19T07:27:36-06:00"
description: "semscope takes the Zig grammar of ztree, so that both give the same trees for empty containers and three rare Zig forms."
---

# ISSUE-011: Phase 9: the Zig grammar of ztree

**Status/Phase:** closed · Phase 9

## Description
The Zig grammar of semscope ([ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md)) differs from the Zig grammar of ztree by four changes. A container can be empty. A container field can be an error set. A slice can hold an `if` type. The asm clobbers can be an identifier. semscope moves to the ztree runtime, so the two grammars must give the same trees.

## Acceptance Criteria
- [x] `third_party/tree-sitter-zig/grammar.js` is the same as the grammar of ztree.
- [x] The generated `parser.c` is the same as the parser of ztree, and `binding.go` holds its SHA-256.
- [x] A test holds the four changes and fails with the grammar of ADR-0010.
- [x] `zmake check`, `zmake -f Makefile check` and `zmake parity` pass.

## Resolution
Task 9.1 takes the `grammar.js` of ztree ([ADR-0012](/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md)). `tree-sitter-cli` 0.25.10 generates the same `parser.c`, `grammar.json` and `node-types.json` as in ztree. `TestZigEmptyContainersAndRareForms` holds the four changes. The parity test passes on each corpus.
