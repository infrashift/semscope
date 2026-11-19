---
id: "01a1f3fd-a243-760d-b470-fa45ce9c4a63"
title: "ADR-0012: The Zig grammar of ztree"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "zig"
  - "dependency"
created: "2026-11-19T07:27:28-06:00"
timestamp: "2026-11-19T07:27:28-06:00"
description: "semscope takes the patched Zig grammar of ztree, which adds empty containers and three rare forms to the patch of ADR-0010."
---

# ADR-0012: The Zig grammar of ztree

**Status:** Accepted
**Date:** 2026-11-19
**Source:** [ISSUE-011](/issues/issue-011-phase-9-the-zig-grammar-of-ztree--01a1f3fdc1c1.md). ztree ADR-0006 and ADR-0015.

## Context
[ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md) patches `tree-sitter-zig` v1.1.2 for Zig 0.17. ztree starts from the same patch and adds four changes:
- A `struct`, `enum`, `union` or `opaque` body can be empty (ztree ADR-0006). The release grammar puts a MISSING `identifier` in each empty body, such as `struct {}`.
- A container field can have an error set as its type, such as `w: error{A},` (ztree ADR-0015).
- A slice can hold an `if` type, such as `[]if (c) A else B` (ztree ADR-0015).
- The asm clobbers can be an identifier, such as `: clobbers);` (ztree ADR-0015).

Without these changes, semscope and ztree give different trees for the same Zig file. semscope moves to the ztree runtime, so the two grammars must agree.

## Decision
- Take the `grammar.js` of ztree in `third_party/tree-sitter-zig`. It holds the patch of ADR-0010 and the four changes.
- Generate `src/` with `tree-sitter-cli` 0.25.10, as ADR-0010 says. The generated `parser.c`, `grammar.json` and `node-types.json` are the same, byte for byte, as the files of ztree.
- Put the new SHA-256 of `parser.c` in `binding.go`.

## Consequences
- semscope reads empty containers and the three rare forms without parse errors.
- `TestZigEmptyContainersAndRareForms` holds the four changes. It fails with the grammar of ADR-0010.
- The parity test with the oracles passes on each corpus, also on the Zig 0.17.1 standard library and zmake.
- `node-types.json` keeps the same node types. Only the child types of `asm_clobbers`, `container_field` and `slice_type` change. The adapter reads none of these nodes.
- A later change of the Zig grammar starts in ztree. semscope then takes the same `grammar.js`.

## Alternatives considered
- **Keep the grammar of ADR-0010.** Rejected: semscope then sees a false `container_field` in each empty container, and the trees differ from the trees of ztree.
- **Take only the empty-container change.** Rejected: the three rare forms occur in the Zig 0.17.1 standard library, and the grammars must agree.
