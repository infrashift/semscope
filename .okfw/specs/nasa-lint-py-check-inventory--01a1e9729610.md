---
id: "01a1e972-9610-7350-8eab-f33d69563cd8"
title: "nasa-lint.py check inventory"
type: "Spec"
tags:
  - "spec"
  - "nasa-p10"
  - "oracle"
  - "zig"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "Every check, threshold, pattern and quirk of scripts/nasa-lint.py, the Zig oracle for the nasa-p10 ruleset."
---

# nasa-lint.py check inventory

`scripts/nasa-lint.py` comes from zmake without changes. It checks Zig only. semscope uses it as the Zig oracle: the parity tests compare its output with the `nasa-p10` findings of semscope. [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md) defines parity.

## Interface
- Input: every `*.zig` file under `src/` in the current directory, in sorted sequence. The script has no path argument.
- Arguments: `--verbose` only. It adds an `info:` line for each function with fewer than 2 assertions.
- Output: one line `<path>:<line>: R<n>: <message>` for each finding, then a summary line.
- Exit status: 1 when the script reports an error, else 0.
- Constants: `MAX_FN_LINES = 60` and `MIN_ASSERT_AVG = 2.0`.

## Code stripping
`strip_code` cleans each line in three steps. It blanks a line that starts with `\\` (a multiline string). It removes the string and character literals. It removes the `//` comment.

## Test exemption
- A block that starts with `test ... {` is exempt from every rule. `block_end` finds the end of the block by brace counting.
- A function with a name that starts with `test` or ends with `Test` is a test helper. The script omits it from the function list, so it is exempt from R1, R4 and R5.
- Quirk: the line rules R2, R7 and R9 still apply inside a test helper.

## Function detection
- Pattern: `^\s*(pub\s+)?(export\s+)?(inline\s+)?fn\s+NAME\s*\(`. The qualifiers must occur in this sequence.
- The body opens on the first line that ends with `{`, within 20 lines of the `fn` line. A line that ends with `;` first marks a prototype. The script skips a prototype.
- The body ends on the line where the brace depth returns to zero.

## R1: recursion
- The call graph has one node for each pair of file name and function name. The file name has no directory.
- Quirk: two functions with the same name in one file share one node. Two files with the same name in different directories share nodes.
- `CALL_RE` finds the calls in the body lines. It also matches the name of a builtin such as `@intCast(`.
- An unqualified call, or a call through `self`, `Self` or `@This`, gives an edge only to a function of the same file.
- A call `alias.f(` gives an edge only when `alias` comes from `const alias = @import("x.zig")`. The node key uses the import string, not the resolved path.
- The script drops an edge to an unknown node. It ignores the keywords `if`, `while`, `for`, `switch`, `return`, `catch`, `fn`, `try`, `orelse`, `and` and `or`.
- An iterative three-color depth-first search visits the nodes in sorted sequence. Each back edge gives one finding at the first node of the cycle.

## R4: function length
- The length counts from the `fn` line to the line of the closing brace. Both lines count, and the signature lines count.
- A function with more than 60 lines gives a finding at the `fn` line.

## R5: assertion density
- The script counts the matches of `\bassert\s*\(` in the body lines. The body lines exclude the line of the opening brace and the line of the closing brace.
- Quirk: an assertion on the same line as a brace does not count. A one-line function counts zero assertions.
- The check is one average over all functions that are not exempt. An average below 2.0 gives one finding `src: R5: ...` without a line number.
- R5 has no suppression.

## Line rules
The script applies these patterns to each stripped line outside the test blocks. Each pattern gives at most one finding on a line.

| Rule | Pattern | Message |
|---|---|---|
| R2 | `\bwhile\s*\(\s*true\s*\)` | unbounded loop `while (true)` |
| R7 | `\bcatch\s*\{\s*\}` | swallowed error `catch {}` |
| R7 | `\belse\s*\|_\|\s*\{\s*\}` | swallowed error `else \|_\| {}` |
| R7 | `\bcatch\s+unreachable\b` | `catch unreachable` hides an error |
| R7 | `(?<![\w.])_\s*=\s*(?:try\s+)?[\w.]+\s*\(` | discarded call result `_ = f(...)` |
| R9 | `\*\s*const\s+fn\b` or `\*\s*fn\s*\(` | function pointer type |

Quirk: the R7 discard pattern does not match a builtin call such as `_ = @intCast(x)`.

## Suppression
A comment `nasa-lint: allow(R<n>) <reason>` on the line or on the line above suppresses one finding. The tag works for the line rules and for R4 on the `fn` line. For R1, the tag goes on the `fn` line of the first node of the cycle.

## Related
- [nasa-lint-go.py check inventory](/specs/nasa-lint-go-py-check-inventory--01a1e9729621.md)
- [ADR-0005](/decisions/adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md): parity with the oracles.
