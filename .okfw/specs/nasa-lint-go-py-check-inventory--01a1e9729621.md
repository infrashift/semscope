---
id: "01a1e972-9621-7880-958b-ec616f766d55"
title: "nasa-lint-go.py check inventory"
type: "Spec"
tags:
  - "spec"
  - "nasa-p10"
  - "oracle"
  - "go"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "Every check, threshold, pattern and exemption of scripts/nasa-lint-go.py, the Go oracle and Go gate for the nasa-p10 ruleset."
---

# nasa-lint-go.py check inventory

`scripts/nasa-lint-go.py` is the Go port of the [nasa-lint.py checks](/specs/nasa-lint-py-check-inventory--01a1e9729610.md). It has the same output format, constants and suppression tag. It is the Go oracle for the parity tests and the Go gate in `zmake lint`.

## Interface
- Input: every `*.go` file under `cmd/`, `internal/` and `rules/`, or under the directories in the arguments.
- Arguments: `--verbose`, `--root DIR` and a list of directories.
- Exempt: `_test.go` files, and the `testdata`, `vendor` and hidden directories.
- Golden tests: `scripts/tests/nasa_lint_go_test.py` runs the script on each fixture in `scripts/testdata/nasa-lint-go/`.

## Code stripping
`strip_go` reads the whole file once. It blanks out line comments, block comments, interpreted strings, raw strings and rune literals. It keeps the line breaks, so the line numbers stay correct.

## Function detection
- Pattern: `func`, an optional receiver such as `(r *T)`, the name, then `(` or `[`. Go declares functions at column 0 only.
- The body opens on the first line that ends with `{`, within 20 lines. An empty line or a new `func` line first marks a function without a body.
- A one-line function, for example `func (s byName) Len() int { return len(s) }`, opens and closes on its own line. Its body starts at the first `{` after the parameters, outside parentheses, that does not follow `struct` or `interface`.
- A test helper has a name that starts with `test` or `Test`, or ends with `Test`. It is exempt from R1, R4 and R5.

## R1: recursion and goto
- The node key is the package directory and the function name. A method uses `Type.name`.
- An unqualified call `f(` gives an edge to the package function `f`.
- A call `r.m(`, where `r` is the receiver name, gives an edge to the method `m` of the receiver type.
- A method call on an expression, as in `g().m(` or `p[0].m(`, gives no edge.
- A method declaration of an interface in a body, for example `Deadline() (time.Time, bool)`, is not a call.
- Other calls give no edge, the same as the Zig oracle. Go forbids import cycles, so a cycle of direct calls stays in one package.
- The cycle search is the same iterative three-color search as in the Zig oracle.
- A line rule flags each `goto` statement.

## R4 and R5
- R4: the same rule as in the Zig oracle. The length counts from the `func` line to the line of the closing brace.
- R5: the script counts the matches of `assert.<Name>(` in the body. The body starts after the last `{` of the opening line, so a one-line function counts its assertions.
- R5: the package `assert` defines the primitive and is exempt.

## Line rules
| Rule | Pattern (summary) | Message |
|---|---|---|
| R1 | `goto` | `goto` statement |
| R2 | `for {`, `for ;; {`, `for true {` | unbounded loop |
| R2 | `for init; ; post {`, a fallback rule | unbounded loop |
| R7 | `_ = f(...)` | discarded call result |
| R7 | `x, _ := f(...)` or `x, _ = f(...)`: a blank last result | discarded call result |
| R9 | `type T func(` or `type T = func(` | function type declaration |
| R9 | a field or variable with a `func`, `[]func` or `map[K]func` type | func-typed field or variable |
| R9 | a parameter with a function type, also `...func(` | func-typed parameter |
| R9 | a result with a function type, also `(T, func())` | func-typed result |
| R9 | `[]func(`, `map[K]func(`, `chan func(` or `.(func(`, a fallback rule | func-typed element |

- A fallback rule reports a line only when no other rule of its R number reports the line.
- For R7, the call must be the whole right side. A chain of calls, as in `_ = a.f().g()`, counts as one call. The R7 patterns do not flag `_, err := f()`, a map index `v, _ := m[k]` or a type assertion `v, _ := f().(T)`.
- The R9 patterns do not flag a function literal in an argument, as in `F(x)(func() {})`, a `go` statement or a `defer` statement.
- Task 6.3 added the one-line functions, the fallback rules and the R1 call filters. The fixture `scripts/testdata/nasa-lint-go/corpus-forms` holds each form. [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) lists the defects.

## Suppression
The tag is the same as in the Zig oracle: `// nasa-lint: allow(R<n>) <reason>` on the line or the line above.

## Related
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md): the Power of 10 rules in Go.
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md): parity with the oracles.
