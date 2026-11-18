---
title: Scanning
description: How semscope finds the source files, which rules it runs, and how the fail level sets the exit status.
---

`semscope scan` reads the source files under each path and runs the active rules on them.

```sh
semscope scan                       # the current directory
semscope scan cmd internal          # two directories
semscope scan src/main.c            # one file
```

## Files and languages

semscope selects a parser from the file extension. It skips a file with a different extension.

| Language | Extensions |
|---|---|
| Go | `.go` |
| C | `.c`, `.h` |
| C++ | `.cc`, `.cpp`, `.cxx`, `.hh`, `.hpp`, `.hxx` |
| Zig | `.zig` |
| Python | `.py` |
| Rust | `.rs` |

A scan does not go into these directories:

- A hidden directory, for example `.git` or `.cache`.
- `vendor`, `node_modules` and `testdata`.
- The build output directories `zig-cache`, `zig-out`, `target` and `__pycache__`.

To skip more paths, add patterns to the exclude map. See [Suppressions & Excludes](../suppressions-and-excludes/).

## Test code

semscope does not check a test file. A test file is a Go file with a name that ends in `_test.go`.
A Python file with a name that starts with `test_` or ends in `_test.py` is also a test file.

In other files, a test function or a test helper is exempt from three checks: recursion, function length and the assertion average.

| Language | Test function or test helper |
|---|---|
| Go | A name that starts with `test` or `Test`, or ends in `Test` |
| Python | A name that starts with `test` |
| Zig | A name that starts with `test` or ends in `Test` |

## Rules and rulesets

Without `--ruleset`, semscope runs the built-in ruleset `nasa-p10`.
Give `--ruleset` (`-r`) one or more times to select a different ruleset or to add overlays:

```sh
semscope scan -r nasa-p10 -r overlay.cue .
semscope scan --omit-rule NASA-P10-03,NASA-P10-05 .
semscope scan --set NASA-P10-04.max_lines=40 .
```

[Rulesets & Overlays](../rulesets/) gives the full composition model.

## Fail level and exit status

The flag `--fail-level` sets the lowest severity that gives exit status 1. The default is `error`.

| `--fail-level` | Exit status 1 when a finding has the severity |
|---|---|
| `error` | `error` |
| `warning` | `error` or `warning` |
| `note` | `error`, `warning` or `note` |
| `none` | never |

| Exit status | Meaning |
|---|---|
| 0 | No finding has the `--fail-level` severity or a higher severity. |
| 1 | One or more findings have the `--fail-level` severity or a higher severity. |
| 2 | Bad usage, a bad ruleset, or work that exceeds a limit. |

The `--fail-level` flag does not change the report. A scan with `--fail-level none` reports all the findings and gives exit status 0.

## Errors

semscope does not ignore a bad flag. An unknown rule ID or an unknown parameter gives exit status 2:

```console
$ semscope scan --omit-rule NASA-P10-99 .
semscope: ruleset: --omit-rule: unknown rule "NASA-P10-99"
$ semscope scan --set NASA-P10-04.max_lines=abc .
semscope: ruleset: --set NASA-P10-04.max_lines=abc: want an integer, got "abc"
```

## Limits

The engine has fixed limits, for example the number of ruleset files and rules. Work that exceeds a limit stops the scan with exit status 2.
The [Ruleset Schema](../../reference/ruleset-schema/#limits) page lists the ruleset limits.
