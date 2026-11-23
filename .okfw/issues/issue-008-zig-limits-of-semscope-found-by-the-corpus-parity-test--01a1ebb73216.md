---
x-okfw-id: "01a1ebb7-3216-795d-a02a-e7293e52af32"
title: "ISSUE-008: Zig limits of semscope found by the corpus parity test"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "parser"
  - "zig"
  - "parity"
  - "phase-6"
  - "closed"
x-okfw-created: "2026-11-17T16:53:34-06:00"
x-okfw-timestamp: "2026-11-17T16:53:34-06:00"
description: "The Zig grammar v1.1.2 lags Zig 0.17, and the Zig call graph gives same-name functions of one file one node."
---

# ISSUE-008: Zig limits of semscope found by the corpus parity test

**Status/Phase:** closed · Phase 7

## Description
The corpus parity test of Task 6.3 found two semscope defects on the Zig 0.17.1 standard library. Phase 7 fixes both.

1. **The grammar lags Zig 0.17.** `tree-sitter-zig` v1.1.2 parses `await` as a keyword. In Zig 0.17 it is an identifier. The error recovery makes `fn cancel` of `Io/Dispatch.zig` and `Io/Uring.zig` longer than 270 lines, so NASA-P10-04 reports them. The real functions have 11 lines.
2. **Same-name functions share one node.** The call graph keys a Zig function by file and name, the same as `nasa-lint.py`. `Mutable.toFloat`, `Const.toFloat` and `Managed.toFloat` of `math/big/int.zig` are one node. A cycle through one of them can show at another.

## Fixed in Task 6.3
- A callee without a name, for example the range `(n - 1)..`, failed an assertion. The parser now skips it.
- R2 did not find `while (true)` as an expression.
- R7 did not find `const a, _ = f()`, `_ = f() orelse x`, `_ = f().?` or `else |_| {}` of an if expression.
- R7 reported `catch {}` at the start of the call, not at the keyword `catch`.
- R1 read the decl literal `.init()` as a call of the function `init` of the file.
- Go: R7 reported the conversion `var _ I = (*T)(nil)`, and R1 read `p[0].Msg()` as a call through the receiver `p`.

## Progress
- **Containers (done).** The Zig adapter gives each function a scope: the names of the enclosing containers, and `F()` for each enclosing function F. The node key is the scope and the name, for example `Mutable.toFloat` or `ArrayList().append`. A plain call `f()` looks for f in the container of the caller, then in each outer container, then in the file. A call through `self`, `Self` or `@This()` looks in the container of the caller only. The engine test `TestZigContainerScopes` holds each case. On the Zig standard library, the cycle `Const.toFloat -> Const.toFloat` now shows at `Const.toFloat`, not at `Mutable.toFloat`.

- **Grammar (done).** `third_party/tree-sitter-zig` holds v1.1.2 without the rules for `async`, `await` and `usingnamespace`. The `replace` directive in `go.mod` selects it. `TestZig017Identifiers` fails with the release grammar and passes with the patched grammar. On the Zig standard library, `fn cancel` of `Io/Dispatch.zig` and `Io/Uring.zig` has its real length again. The two line records for NASA-P10-04 went stale, and semscope finds two more functions.

## Acceptance Criteria
- [x] Read Zig 0.17: no release does, so semscope keeps a patched grammar ([ADR-0010](/decisions/adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md)).
- [x] Key a Zig function by its container, and resolve a call through the scope of the call.

## Resolution
Tasks 7.1 and 7.2 are complete. No divergence record names ISSUE-008 now. The Zig standard library has records for oracle defects only, and [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) lists them.

## Related
- [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md)
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md)
