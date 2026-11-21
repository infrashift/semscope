---
title: Ruleset Schema
description: The CUE schema of a semscope ruleset, the check kinds and their parameters, and the loader limits.
---

`internal/ruleset/schema.cue` holds the schema. The semscope binary embeds it.
Each definition is a closed struct. An unknown field gives the error `field not allowed`.

## `#Ruleset`

| Field | Type | Notes |
|---|---|---|
| `name` | string | Optional |
| `description` | string | Optional |
| `exclude` | map from a pattern to a reason | Optional. See [Suppressions & Excludes](../../guides/suppressions-and-excludes/#exclude-patterns). |
| `rules` | map from a rule ID to a `#Rule` | The map key sets the `id` of the rule. |

## `#Rule`

| Field | Type | Default |
|---|---|---|
| `id` | string | The map key |
| `enabled` | bool | `true` |
| `severity` | `"error"`, `"warning"` or `"note"` | `"error"` |
| `languages` | a list of one or more `#Language` values | Required |
| `title` | string | Required |
| `message` | string | Required |
| `params` | map from a name to an int, a string, a bool or a list of strings | Empty |
| `check` | a struct with a `kind` field | Required |

`#Language` is one of `"go"`, `"c"`, `"cpp"`, `"zig"`, `"python"`, `"rust"` and `"csharp"`.
`#PerLanguage` is a map from a `#Language` to a list of strings.

The schema gives no default for `enabled` and `severity`. The loader applies the defaults in Go after the unification.
Two different CUE defaults cancel each other, so a default in the schema can make a value incomplete.
A base ruleset can declare its own default, for example `severity: *"warning" | "error" | "note"`.

## Check kinds

`check.kind` selects the checker. Some kinds need parameters in `params`. Some kinds have extra fields in `check`.

| Kind | Rule | Parameters | Extra `check` fields |
|---|---|---|---|
| `no_goto_or_recursion` | NASA-P10-01 | none | `banned_calls?: #PerLanguage` |
| `bounded_loops` | NASA-P10-02 | none | none |
| `no_dynamic_alloc` | NASA-P10-03 | none | `alloc_calls?: #PerLanguage`, `init_functions?: [...string]` |
| `max_function_lines` | NASA-P10-04 | `max_lines: int & >0` | none |
| `min_assertions` | NASA-P10-05 | `min_assertions: int & >=0`, `mode: "average" \| "per_function"` | `assert_calls?: #PerLanguage` |
| `narrowest_scope` | NASA-P10-06 | none | none |
| `checked_returns` | NASA-P10-07 | `go_types?: bool`, `go_ignore_calls?: [...string]` | `must_check_calls?: #PerLanguage` |
| `restricted_preprocessor` | NASA-P10-08 | `max_conditional_depth: int & >=0` | `banned_directives?: #PerLanguage` |
| `max_pointer_indirection` | NASA-P10-09 | `max_pointer_depth: int & >=0`, `allow_function_pointers: bool` | none |
| `ast_query` | any | none | `query: [#Language]: string` |
| `token_ban` | any | none | `patterns: #PerLanguage` |

- An `ast_query` rule runs one Tree-sitter query for each language. Each `@violation` capture gives a finding.
- A `token_ban` rule bans regular expressions. It reads the source without the comments and the strings.

## Defaults for overlays

CUE unification does not replace a concrete value. `max_lines: 60` and `max_lines: 40` conflict.
So a base ruleset declares each tunable value as a default, for example `max_lines: *60 | int`. An overlay with `max_lines: 40` then selects 40.

The same applies to lists. `nasa-p10` declares `languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]`, so an overlay can narrow the list.

## Sources

- A source is a built-in name, a `.cue` file or a directory.
- A directory gives its `.cue` files in name sequence.
- The loader compiles each source alone. A source cannot refer to the schema definitions. The unification with `#Ruleset` applies them.

## Limits

| Limit | Value |
|---|---|
| Source files | 64 |
| Size of each source file | 1 MiB |
| Rules | 4096 |
| Parameters of each rule | 64 |
| `--set` values | 256 |

A ruleset that exceeds a limit gives exit status 2.
