---
x-okfw-id: "01a1ebb7-320f-7162-b57e-fadbae48ed50"
title: "ISSUE-007: Oracle defects found by the corpus parity test"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "parity"
  - "oracle"
  - "phase-6"
  - "closed"
x-okfw-created: "2026-11-17T16:53:34-06:00"
x-okfw-timestamp: "2026-11-17T16:53:34-06:00"
description: "The corpus parity test of Task 6.3 found defects in nasa-lint.py and nasa-lint-go.py; this issue lists each class and its state."
---

# ISSUE-007: Oracle defects found by the corpus parity test

**Status/Phase:** closed · Phase 7

## Description
Task 6.3 runs the oracles and semscope on real trees: zmake, the Zig 0.17.1 standard library, cobra, pflag, go-sarif, CUE, go-tree-sitter and x/tools. Each difference is a defect of one tool. This issue lists the defects of the oracles. [ISSUE-008](/issues/issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md) lists the defects of semscope that stay open.

## nasa-lint-go.py: fixed
`nasa-lint-go.py` is a semscope script, so Task 6.3 fixes it. The fixture `scripts/testdata/nasa-lint-go/corpus-forms` holds one case for each class.
- A one-line function `func f() int { return 1 }` was not a function. The next line that ends in `{` became its body.
- R9 did not find `...func(` parameters, `type T = func(`, `(T, func())` results, `[]func(`, `map[K]func(`, `chan func(` and `v.(func())`.
- R9 reported `F(x)(func() {})`: a call with a func literal is not a func-typed result.
- R2 did not find `for i := 0; ; i++ {`.
- R7 reported `v, _ := f().(T)`: the blank drops the ok of a type assertion, not a call result. R7 also dropped `defer func() { _ = c.Close() }()`.
- R1 read `g().m(` and `p[0].m(` as a call of the package function `m`. R1 also read an interface method `M() (T, bool)` in a function body as a call.

## nasa-lint.py: fixed in Task 7.3
Task 7.3 fixes each class below. The fixture `scripts/testdata/nasa-lint/corpus-forms` holds one case for each class.
- R1: `CALL_RE` reads `@sin(`, the decl literal `.init(` and `x.f().g(` as a call of a function of the same file. These false edges give 203 cycles that only the oracle reports.
- R4: the body search takes the first line that ends in `{`. A parameter type `struct {` or `union(enum) {` ends the function early. `FN_RE` has no `noinline` and no `@"name"`.
- R5: the oracle counts the assertions of a nested function also in each enclosing function. The oracle does not count an assertion on the line of `{` or `}`.
- R7: the line rules do not find `catch` and `unreachable` on two lines, `_, x = f()` or `_ = try @This().f()`. The oracle reports `const a, _ = f().field`, which drops a field, not a call result.

## Acceptance Criteria
- [x] Each class of `nasa-lint-go.py` has a fix and a fixture.
- [x] Each difference on the Zig standard library has a line record or a class record.
- [x] semscope fixes `nasa-lint.py` and removes the records of the classes.
- [x] zmake takes the fixed `nasa-lint.py`, so the two copies stay byte-identical ([ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)).

## Resolution
Task 7.3 changes `nasa-lint.py`:
- R1 uses the same node keys as semscope: the file path and the container path. A plain call resolves through the lexical scopes. Builtins, decl literals, method calls on expressions and nested `fn` names give no edge.
- Function detection reads `noinline`, `extern`, `@"name"` and one-line functions. A `struct {` parameter does not open the body. Each line belongs to the innermost function.
- R5 counts each assertion once, also on a brace line, and skips the primitive `fn assert`.
- The R7 rules for `catch` and `else |_|` read across lines. The discard rule reads the destructure forms and `@This().f(`, and needs the call as the whole value.

On zmake `src/`, the fixed script gives the same result as before: no errors, 306 functions and 614 assertions. On the Zig 0.17.1 standard library, all findings of R1, R2, R4, R7 and R9 now match semscope. The fix also showed three semscope defects, and Task 7.3 fixes them. The first is a prefix operator on a callee: `?Span(T)`, `!f(x)` and `*Q(T)`. The others are a `catch` block and an `else |_|` block with a comment only. One difference stays: [ISSUE-009](/issues/issue-009-the-zig-grammar-does-not-read-asm-clobbers-of-zig-0-15--01a1ebef12f0.md).

zmake took the fixed script in its commit `47f8ff7`. The two copies are byte-identical again, and zmake `make check` passes.

## Related
- [ISSUE-006](/issues/issue-006-phase-6-exclude-flag-go-type-checks-and-parity-corpora--01a1eba38e60.md)
