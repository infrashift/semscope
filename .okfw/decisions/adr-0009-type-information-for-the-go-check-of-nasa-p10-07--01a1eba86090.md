---
x-okfw-id: "01a1eba8-6090-7d14-ba66-605bf411eb91"
title: "ADR-0009: Type information for the Go check of NASA-P10-07"
type: "ADR"
tags:
  - "adr"
  - "engine"
  - "go"
  - "nasa-p10"
  - "dependency"
x-okfw-created: "2026-11-17T16:37:22-06:00"
x-okfw-timestamp: "2026-11-17T16:37:22-06:00"
description: "When go_types is true, semscope loads the Go packages with golang.org/x/tools/go/packages and reports each call, defer and go statement that drops an error result."
---

# ADR-0009: Type information for the Go check of NASA-P10-07

**Status:** Accepted
**Date:** 2026-11-17
**Source:** The next steps of the daily note of 2026-11-17 (`PLAN.md` Task 6.2).

## Context
Rule 7 of the Power of 10 says: check the return value of each non-void function. The Go adapter reads Tree-sitter syntax only. It finds `_ = f()` and `x, _ := f()`, the same as `nasa-lint-go.py`. It cannot find the statement `f()` when `f` returns an error, because it does not know the result types. In semscope, golangci-lint (errcheck) closes this gap. For a user of semscope, nothing closes it.

## Decision
- Add the package `internal/gotypes`. It loads the Go packages with `golang.org/x/tools/go/packages` (v0.52.0) and walks the typed syntax trees.
- A drop is a call statement, a `defer` statement or a `go` statement whose call returns `error`, or an interface that includes `error`.
- The rule kind `checked_returns` gets two parameters:
  - `go_types` (bool) turns on the check. The default in `nasa-p10` is `false`.
  - `go_ignore_calls` (list) names calls that never return a non-nil error. An entry is a full function name, for example `(*strings.Builder).WriteString`, or a full name with the type of the first argument, for example `fmt.Fprintf(*strings.Builder)`. The errcheck tool uses the same form.
- The engine runs the check after the per-file checks. It reports a drop only in a scanned Go file that is not a test file. The suppression tags apply.
- The self-scan overlay `.semscope.cue` sets `go_types` to `true`.

## Consequences
- The check needs the go command and a `go.mod` at or above each scanned directory. A package that does not load gives an error and exit status 2.
- A load runs `go list` and type-checks the packages. A scan with `go_types` is slower.
- `nasa-lint-go.py` has no type information. The default `false` keeps the parity of [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md).
- `go/ast` and `go/types` walk the syntax with recursion inside the standard library. The semscope code has no recursion. [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md) applies to semscope code only.
- The default ignore list is short. A drop of `fmt.Println` is a finding.

## Alternatives considered
- **Run errcheck as a library.** Rejected: one more dependency, and its configuration does not compose with CUE overlays.
- **Use `go/importer` without `go/packages`.** Rejected: it does not find the packages of a module.
- **Turn on the check by default.** Rejected: the oracle cannot check it, and the check needs the go command.
