---
id: "01a1ebef-12f0-7f1a-82e7-c58605dc59c7"
title: "ISSUE-009: The Zig grammar does not read asm clobbers of Zig 0.15"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "parser"
  - "zig"
  - "parity"
  - "phase-7"
  - "closed"
created: "2026-11-17T17:54:36-06:00"
timestamp: "2026-11-17T17:54:36-06:00"
description: "Zig 0.15 writes asm clobbers as a struct literal; tree-sitter-zig reads strings only, so 26 syscall functions of the Zig standard library are lost."
---

# ISSUE-009: The Zig grammar does not read asm clobbers of Zig 0.15

**Status/Phase:** closed · Phase 7

## Description
Zig 0.15 writes the clobbers of `asm` as a struct literal, for example `: .{ .rcx = true, .memory = true }`. The rule `asm_clobbers` of `tree-sitter-zig` reads strings only. The parse error drops the function, and the error recovery can drop the next function too. On the Zig 0.17.1 standard library, semscope does not see 26 syscall functions of `os/linux/alpha.zig`, `os/linux/riscv32.zig` and `os/linux/riscv64.zig`.

The fixed `nasa-lint.py` of [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) finds these functions. A stats record in `testdata/nasa_p10/parity-divergences.json` holds the difference: 12004 functions in the oracle, 11978 in semscope.

## Acceptance Criteria
- [x] The patched grammar of [ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md) reads a struct literal as the clobbers.
- [x] The stats record goes away, because the function counts match.

## Resolution
Task 7.4 lets `asm_clobbers` of the patched grammar read an anonymous struct literal. `TestZigAsmStructClobbers` holds the case. On the Zig 0.17.1 standard library, semscope and `nasa-lint.py` now count 12004 functions and 1871 assertions. The corpus has no divergence record.

The task also found a trap of the Go build cache. The binding includes `../../src/parser.c`, a file outside its package, so the cache did not see the new parser. `binding.go` now holds the SHA-256 of `parser.c`, and `TestZigParserHash` fails when the two do not agree.
