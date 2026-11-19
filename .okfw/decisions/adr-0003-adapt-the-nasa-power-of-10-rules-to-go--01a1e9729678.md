---
id: "01a1e972-9678-7816-8b9f-b715aa629b9a"
title: "ADR-0003: Adapt the NASA Power of 10 rules to Go"
type: "ADR"
tags:
  - "adr"
  - "safety"
  - "lint"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "semscope follows the Power of 10 rules where Go allows; nasa-lint-go.py, golangci-lint and the self-scan check them, with a recorded R3 deviation."
---

# ADR-0003: Adapt the NASA Power of 10 rules to Go

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request (`PLAN.md` section 0.2). zmake ADR-0003 adapts the rules to Zig. Gerard J. Holzmann, "The Power of 10" (2006).

## Context
The rules target C. Go has garbage collection, closures and interfaces, and no `assert` statement. `nasa-lint.py` checks Zig only. Some rules need changes for Go.

## Decision
| Rule | semscope adaptation | Check |
|---|---|---|
| R1 Simple control flow | No `goto`. No recursion: tree walks use a Tree-sitter cursor or an explicit bounded stack. The cycle search is iterative. | `nasa-lint-go.py`, self-scan |
| R2 Fixed loop bounds | No `for {}`. Loops iterate over a slice or count toward a limit. A loop that reaches its limit returns an error. | `nasa-lint-go.py` |
| R3 No allocation after initialization | Deviation: Go, cobra, CUE and Tree-sitter allocate at run time. `NewEngine` sizes the worker pool, the stacks and the result buffers. Hard limits cap the rest. | Review, self-scan overlay |
| R4 Short functions | 60 lines or fewer, from the `func` line to the closing brace. | `nasa-lint-go.py`, `funlen` |
| R5 Assertion density | An average of two or more `assert.That` calls per function. The package `assert` is exempt. | `nasa-lint-go.py` |
| R6 Smallest scope | No mutable package variables. | `gochecknoglobals` |
| R7 Check return values | Check each error. No `_ = f()`. No blank last result. | `nasa-lint-go.py`, `errcheck` with `check-blank` |
| R8 Limited preprocessor | No `//go:linkname`. Build tags for platforms only. Generated code is the same on each run. | Review |
| R9 Restricted pointers | No function types, func-typed fields or dispatch tables. Cobra callbacks are function literals that cobra requires. | `nasa-lint-go.py`, `depguard` (no `reflect`) |
| R10 Zero warnings | `go vet` and `golangci-lint` report zero issues. No `unsafe`. | `zmake lint` |

**Assertions:** `assert.That(cond, msg)` panics when `cond` is false. Assertions check programmer errors only. Bad input returns an error. `forbidigo` bans `panic` outside the package `assert`.

**R5 mode:** `nasa-lint.py` checks an average, not a minimum for each function. semscope keeps the average for parity. The ruleset parameter `mode` also offers `per_function`.

**Suppression:** `// nasa-lint: allow(R<n>) <reason>` on the line or the line above. Each suppression needs a reason.

## Consequences
- The code is more verbose: explicit stacks, bounds and assertions.
- An input that exceeds a limit fails with a clear error.
- The Go gate exempts the test files. golangci-lint still checks them.
- The self-scan turns off NASA-P10-03 through the overlay `.semscope.cue`. The overlay records the reason.
- The overlay also excludes `scripts`: the Python gate scripts. `nasa-lint.py` stays byte-identical, because it is the Zig oracle.
- Each package variable of semscope carries `semscope: allow(NASA-P10-06) <reason>`: two embedded files, one compiled pattern and one sentinel error.
- The self-scan and `nasa-lint-go.py` report the same R5 counts and the same R7 lines for the semscope sources.

## Alternatives considered
- **Literal R3: no allocation after initialization.** Rejected: Go cannot satisfy it. The runtime, the parser and the CUE evaluator allocate.
- **A minimum of two assertions in each function.** Rejected for the gate: trivial functions then need assertions without value. The `per_function` mode stays available.
- **Rewrite `nasa-lint.py` for Go.** Rejected: the script is then no longer a faithful Zig oracle.

## Related
- [nasa-lint-go.py check inventory](/specs/nasa-lint-go-py-check-inventory--01a1e9729621.md)
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md)
