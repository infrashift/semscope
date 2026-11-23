---
x-okfw-id: "01a1e972-9610-7350-8eab-f33d69563cd8"
title: "nasa-lint.py check inventory"
type: "Spec"
tags:
  - "spec"
  - "nasa-p10"
  - "oracle"
  - "zig"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T06:19:23-06:00"
description: "Every check, threshold, pattern and quirk of scripts/nasa-lint.py, the Zig oracle for the nasa-p10 ruleset."
---

# nasa-lint.py check inventory

`scripts/nasa-lint.py` comes from zmake. It checks Zig only. semscope uses it as the Zig oracle: the parity tests compare its output with the `nasa-p10` findings of semscope. [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md) defines parity. [ISSUE-007](/issues/issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) fixed the defects that the Zig standard library showed. The fixture `scripts/testdata/nasa-lint/corpus-forms` holds each fixed form.

## Interface
- Input: every `*.zig` file under `src/` in the current directory, in sorted sequence. The script has no path argument.
- Arguments: `--verbose` only. It adds an `info:` line for each function with fewer than 2 assertions.
- Output: one line `<path>:<line>: R<n>: <message>` for each finding, then a summary line.
- Exit status: 1 when the script reports an error, else 0.
- Constants: `MAX_FN_LINES = 60` and `MIN_ASSERT_AVG = 2.0`.

## Code stripping
`strip_code` cleans each line in three steps. It blanks a line that starts with `\\` (a multiline string). It removes the string and character literals. It removes the `//` comment.

## Test exemption
- A block that starts with `test ... {` is exempt from every rule. `block_end` counts the braces to find the end of the block.
- A function with a name that starts with `test` or ends with `Test` is a test helper. It is exempt from R1, R4 and R5.
- Quirk: the line rules R2, R7 and R9 still apply inside a test helper.

## Function detection
- Pattern: `[pub] [export|extern] [inline|noinline] fn NAME(`. A name is an identifier or a quoted name `@"name"`.
- The body opens on the first line that ends with `{` after the parameter list closes. A `struct {` parameter does not open the body. A line that ends with `;` first marks a prototype. The script skips a prototype.
- A one-line function, for example `inline fn leave() void {}`, opens and closes on its line.
- The body ends where the last `{` of the opening line closes. A brace before it, as in `}) void {` or `}!T) {`, does not count.
- Each line belongs to the innermost function that holds it. The calls and the assertions of a nested function do not count for the enclosing function.

## R1: recursion
- A node is the file path from the root and the function key. The key holds the names of the enclosing `const Name = struct {` containers and `F()` for each enclosing function F, for example `Mutable.toFloat` or `ArrayList().append`.
- A plain call `f(` looks for f in the container of the caller, then in each outer container, then in the file.
- A call through `self.`, `Self.` or `@This().` looks in the container of the caller only.
- A call `alias.f(` of `const alias = @import("x.zig")` goes to the function f of x.zig. The path of x.zig is relative to the directory of the file.
- A builtin `@f(`, a decl literal `.f(`, a method call on an expression `x().f(` and the name of a nested `fn f(` give no edge. The keywords `if`, `while`, `for`, `switch`, `return`, `catch`, `fn`, `try`, `orelse`, `and` and `or` give no edge.
- An iterative three-color depth-first search visits the nodes in sorted sequence. Each back edge gives one finding at the first node of the cycle. semscope uses the same keys and the same sequence, so the findings match.

## R4: function length
- The length counts from the `fn` line to the line of the closing brace. Both lines count, and the signature lines count.
- A function with more than 60 lines gives a finding at the `fn` line.

## R5: assertion density
- The script counts the matches of `\bassert\s*\(` on the lines of each function, also on the lines of the braces.
- The primitive `fn assert` does not count as a function, and its declaration is not an assertion.
- The check is one average over all functions that are not exempt. An average below 2.0 gives one finding `src: R5: ...` without a line number.
- R5 has no suppression.

## Line and text rules
The line rules read each stripped line outside the test blocks. The text rules read the whole stripped file, so a match can span lines. A text rule gives the finding on the line where the match starts.

| Rule | Kind | Pattern | Message |
|---|---|---|---|
| R2 | line | `\bwhile\s*\(\s*true\s*\)` | unbounded loop `while (true)` |
| R9 | line | `\*\s*const\s+fn\b` or `\*\s*fn\s*\(` | function pointer type |
| R7 | line | `_ = f(`, `_, x = f(`, `const a, _ = f(`, also with `try` and `@This().` | discarded call result `_ = f(...)` |
| R7 | text | `\bcatch\s*\{\s*\}` | swallowed error `catch {}` |
| R7 | text | `\belse\s*\|_\|\s*\{\s*\}` | swallowed error `else \|_\| {}` |
| R7 | text | `\bcatch\s+unreachable\b` | `catch unreachable` hides an error |

- For the discard rule, the call must be the whole value: `f()`, `a.f().g()`, `f() catch x` or `f().?`, but not `f().field`. A builtin call such as `_ = @intCast(x)` is not a discard.
- A block with a comment only is empty after the script removes the comments, so `catch { // reason }` is a swallowed error.

## Suppression
A comment `nasa-lint: allow(R<n>) <reason>` on the line or on the line above suppresses one finding. The tag works for the line rules, the text rules and for R4 on the `fn` line. For R1, the tag goes on the `fn` line of the first node of the cycle.

## Related
- [nasa-lint-go.py check inventory](/specs/nasa-lint-go-py-check-inventory--01a1e9729621.md)
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md): parity with the oracles.
