---
x-okfw-id: "01a1f4d2-11fc-735b-9136-8d066b7f9c4d"
title: "ADR-0014: The Zig grammar lives in ztree only"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "zig"
  - "dependency"
x-okfw-created: "2026-11-19T11:19:30-06:00"
x-okfw-timestamp: "2026-11-19T11:19:30-06:00"
description: "semscope removes its copy of the patched Zig grammar, because the parser takes all grammars from ztree since ADR-0013."
---

# ADR-0014: The Zig grammar lives in ztree only

**Status:** Accepted
**Date:** 2026-11-19
**Source:** [ISSUE-013](/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md). User request on 2026-11-19.

## Context
Since [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md), the parser takes all six grammars from ztree. The patched Zig grammar of [ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md) and [ADR-0012](/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md) stays in `third_party/tree-sitter-zig`, but no package imports it. The `replace` directive for it in `go.mod` selects a module that semscope does not require. `TestZigParserHash` checks a hash of a parser that semscope does not link.

ztree keeps the same grammar in `third_party/grammars/zig`, with the patch `zig-0.17.patch` and a `PIN` file (ztree ADR-0006 and ADR-0015). A change of the copy in semscope has no effect, and the copy can differ from the grammar that semscope uses.

## Decision
- Remove `third_party/tree-sitter-zig` and its `replace` directive.
- Remove `TestZigParserHash`. The test of ztree `TestLibraryHash` and the hash file of ztree ADR-0016 cover a new parser in `libztree.a`.
- Remove the `third_party` entry from the self-scan overlay `.semscope.cue`.
- The Zig tests of `internal/parser` stay. They now check the Zig grammar of ztree: `TestZig017Identifiers`, `TestZigAsmStructClobbers` and `TestZigEmptyContainersAndRareForms`.
- A change of the Zig grammar occurs in ztree only.

## Consequences
- The repository loses about 6 MB of generated parser code.
- This decision deprecates ADR-0010 and ADR-0012. Their reasons for the patch stay valid in ztree.
- The parity test and the Zig tests of semscope fail when a ztree grammar change breaks the adapter.

## Alternatives considered
- **Keep the copy as a reference.** Rejected: ztree holds the same grammar and patch, and a stale copy misleads a reader.
- **Keep the copy and check that it is the same as the grammar of ztree.** Rejected: the check guards a file that no build reads.

## Related
- [ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md): the ztree runtime.
- [ISSUE-013](/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md)
