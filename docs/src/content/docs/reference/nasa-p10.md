---
title: nasa-p10 Rules
description: The ten rules of the built-in ruleset nasa-p10, their parameters, and the parity with the oracle scripts.
---

The built-in ruleset `nasa-p10` holds the NASA/JPL Power of 10 rules of Gerard J. Holzmann. The source is "The Power of 10" (2006).
semscope adapts each rule to Go, C, C++, Zig, Python and Rust.
The source of the ruleset is `rules/nasa_p10/nasa_p10.cue`.

## Summary

| Rule | Title | Check kind | Default severity | Languages |
|---|---|---|---|---|
| NASA-P10-01 | Simple control flow | `no_goto_or_recursion` | error | all six |
| NASA-P10-02 | Fixed loop bounds | `bounded_loops` | error | all six |
| NASA-P10-03 | No dynamic allocation after initialization | `no_dynamic_alloc` | warning | Go, C, C++, Zig, Rust |
| NASA-P10-04 | Short functions | `max_function_lines` | error | all six |
| NASA-P10-05 | Assertion density | `min_assertions` | error | all six |
| NASA-P10-06 | Smallest scope | `narrowest_scope` | error | all six |
| NASA-P10-07 | Check return values | `checked_returns` | error | all six |
| NASA-P10-08 | Limited preprocessor | `restricted_preprocessor` | error | all six |
| NASA-P10-09 | Restricted pointers | `max_pointer_indirection` | error | Go, C, C++, Zig, Rust |
| NASA-P10-10 | Zero warnings and no unsafe constructs | `ast_query` | error | all six |

Each tunable value is a CUE default. An overlay or `--set` can change it. See [Rulesets & Overlays](../../guides/rulesets/).

## NASA-P10-01: Simple control flow

No `goto`, no `setjmp` or `longjmp`, and no direct or indirect recursion.
semscope builds a call graph for each language and finds each cycle with an iterative depth-first search.

| Field | Default |
|---|---|
| `check.banned_calls.c`, `check.banned_calls.cpp` | `setjmp`, `longjmp`, `sigsetjmp`, `siglongjmp`, `_setjmp`, `_longjmp` |

## NASA-P10-02: Fixed loop bounds

No loop without a condition that can end it, for example `for {}`, `while (true)`, `while True:` or `loop {}`.

## NASA-P10-03: No dynamic allocation after initialization

No allocator call outside an init function. The check uses call names only, so the default severity is `warning`.

| Field | Default |
|---|---|
| `check.alloc_calls.go` | `make`, `new`, `append` |
| `check.alloc_calls.c` | `malloc`, `calloc`, `realloc`, `aligned_alloc`, `strdup`, `strndup` |
| `check.alloc_calls.cpp` | The C list and `new` |
| `check.alloc_calls.zig` | `alloc`, `create`, `dupe`, `dupeZ`, `realloc`, `alignedAlloc`, `allocSentinel` |
| `check.alloc_calls.rust` | `Box::new`, `Vec::new`, `Vec::with_capacity`, `String::new`, `String::from`, `vec!`, `Rc::new`, `Arc::new` |
| `check.init_functions` | `main`, `init`, `init*`, `Init*`, `new`, `New*`, `setup*`, `Setup*` |

## NASA-P10-04: Short functions

Each function has `max_lines` lines or fewer, from the signature to the closing brace.

| Parameter | Type | Default |
|---|---|---|
| `max_lines` | int | `60` |

## NASA-P10-05: Assertion density

An average of `min_assertions` or more assertions for each function. Assertions check programmer errors only.
In the `average` mode, semscope reports one finding for each language, without a location.
In the `per_function` mode, semscope reports each function with too few assertions.

| Parameter | Type | Default |
|---|---|---|
| `min_assertions` | int | `2` |
| `mode` | `"average"` or `"per_function"` | `"average"` |

| Language | Assertion calls (`check.assert_calls`) |
|---|---|
| Go | `assert.*` |
| C | `assert`, `static_assert`, `_Static_assert` |
| C++ | `assert`, `static_assert` |
| Zig | `assert` |
| Python | `assert` |
| Rust | `assert!`, `assert_eq!`, `assert_ne!`, `debug_assert!`, `debug_assert_eq!`, `debug_assert_ne!` |

## NASA-P10-06: Smallest scope

No mutable global state, for example a mutable package variable in Go.

## NASA-P10-07: Check return values

No discarded call results and no swallowed errors.
For C and C++, `check.must_check_calls` names the calls with a result to check, for example `malloc`, `fopen`, `read` and `pthread_create`.

| Parameter | Type | Default |
|---|---|---|
| `go_types` | bool | `false` |
| `go_ignore_calls` | list | `(*bytes.Buffer).Write…`, `(*strings.Builder).Write…`, `fmt.Fprint…(*bytes.Buffer)`, `fmt.Fprint…(*strings.Builder)` |

### The Go check with type information

When `go_types` is `true`, semscope loads the Go packages with type information. It reports each call statement, `defer` statement and `go` statement that drops an `error` result.

- The check needs the `go` command and a `go.mod`.
- It skips `_test.go` files.
- An entry of `go_ignore_calls` names a call that never returns a non-nil error. The entry is a full function name, for example `(*strings.Builder).WriteString`.
- An entry can also give the type of the first argument, for example `fmt.Fprintf(*strings.Builder)`. The errcheck tool uses the same form.

```sh
semscope scan --set NASA-P10-07.go_types=true .
```

The default is `false`, so the parity with the Go oracle does not change.

## NASA-P10-08: Limited preprocessor

Limit the preprocessor and code generation to includes and simple constants.

| Parameter | Type | Default |
|---|---|---|
| `max_conditional_depth` | int | `2` |

| Language | Banned directives (`check.banned_directives`) |
|---|---|
| Go | `go:linkname` |
| C, C++ | Token pasting, stringizing, variadic macros, `#undef` |
| Zig | `@cImport` |
| Python | `exec`, `eval`, `compile` |
| Rust | `macro_rules!` |

## NASA-P10-09: Restricted pointers

One level of pointer indirection at most, and no function pointers.

| Parameter | Type | Default |
|---|---|---|
| `max_pointer_depth` | int | `1` |
| `allow_function_pointers` | bool | `false` |

## NASA-P10-10: Zero warnings and no unsafe constructs

No unsafe constructs, and no warning suppression without a reason. The rule is an `ast_query` rule with one Tree-sitter query for each language.

| Language | Findings |
|---|---|
| Go | An import of `unsafe`. A `//nolint` comment without a reason. |
| C, C++ | Calls to `gets`, `strcpy`, `strcat`, `sprintf` and `vsprintf`. A `NOLINT` comment without a reason. |
| Zig | `@ptrFromInt`, `@intFromPtr`, `@ptrCast` and `@alignCast` |
| Python | An import of `ctypes`. A `# noqa` or `# type: ignore` comment without a reason. |
| Rust | Each `unsafe` block |

## Parity with the oracles

Two scripts are the oracles of `nasa-p10`:

- `scripts/nasa-lint.py` checks Zig. It stays byte-identical with the copy in the zmake repository.
- `scripts/nasa-lint-go.py` checks Go.

The checks of NASA-P10-01, -02, -04, -05, -07 and -09 match the oracles. The parity tests in `internal/engine` compare the findings line by line.
On the fixtures in `testdata/nasa_p10/`, semscope and the oracle report the same set of file, line and rule. The averages of NASA-P10-05 match to two decimals.
On a real corpus, each divergence needs a fix or a record with a reason in `testdata/nasa_p10/parity-divergences.json`.

The rules NASA-P10-03, -06, -08 and -10 have no oracle. Fixtures alone test them. For C, C++, Python and Rust, fixtures alone test all the rules.
