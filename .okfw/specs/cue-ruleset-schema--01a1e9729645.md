---
id: "01a1e972-9645-7396-bccb-45faa9d82beb"
title: "CUE ruleset schema"
type: "Spec"
tags:
  - "spec"
  - "ruleset"
  - "cue"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:29:28-06:00"
description: "The CUE schema of a semscope ruleset and the steps that compose a base ruleset, overlays and overrides."
status: "stable"
---

# CUE ruleset schema

`internal/ruleset/schema.cue` holds the schema. The binary embeds it. Task 3.1 implements it.

## Shape
- `#Ruleset`: the optional fields `name`, `description` and `exclude`, and the map `rules: [ID=string]: #Rule & {id: ID}`. The map key sets the rule ID.
- `exclude`: a map from a path pattern to a reason, for example `exclude: scripts: "tooling"`. A pattern uses `path.Match` syntax on paths relative to the scan root. It matches a path or a parent directory of the path. A map composes over overlays, and each entry needs a reason.
- `#Rule`: `id`, `enabled?` (`bool`), `severity?` (`"error" | "warning" | "note"`), `languages` (one or more), `title`, `message`, `params` and `check`.
- The loader sets `enabled` to `true` and `severity` to `"error"` when a rule omits them.
- `#Language`: `"go" | "c" | "cpp" | "zig" | "python" | "rust" | "csharp"`.
- `params`: a map from a name to `int`, `string`, `bool` or `[...string]`.
- `#PerLanguage`: a map from a language to a list of strings.
- Each definition is a closed struct. An unknown field gives the error `field not allowed`.

## Check kinds
`check.kind` selects the checker. Some kinds need parameters. Some kinds have extra fields.

| Kind | Rule | Required parameters | Extra fields |
|---|---|---|---|
| `no_goto_or_recursion` | R1 | none | `banned_calls?: #PerLanguage` |
| `bounded_loops` | R2 | none | none |
| `no_dynamic_alloc` | R3 | none | `alloc_calls?: #PerLanguage`, `init_functions?: [...string]` |
| `max_function_lines` | R4 | `max_lines: int & >0` | none |
| `min_assertions` | R5 | `min_assertions: int & >=0`, `mode: "average" \| "per_function"` | `assert_calls?: #PerLanguage` |
| `narrowest_scope` | R6 | none | none |
| `checked_returns` | R7 | `go_types?: bool`, `go_ignore_calls?: [...string]` ([ADR-0009](/decisions/adr-0009-type-information-for-the-go-check-of-nasa-p10-07--01a1eba86090.md)) | `must_check_calls?: #PerLanguage` |
| `restricted_preprocessor` | R8 | `max_conditional_depth: int & >=0` | `banned_directives?: #PerLanguage` |
| `max_pointer_indirection` | R9 | `max_pointer_depth: int & >=0`, `allow_function_pointers: bool` | none |
| `ast_query` | any | none | `query: [#Language]: string`; each `@violation` capture gives a finding |
| `token_ban` | any | none | `patterns: #PerLanguage` of regular expressions |

## Sources
- A `--ruleset` value is a built-in name, a `.cue` file or a directory. A built-in name has priority over a path with the same name.
- A directory gives its `.cue` files in name sequence. The loader skips other files.
- The loader compiles each source alone. A source cannot refer to the schema definitions. Unification with `#Ruleset` applies them.
- Limits: 64 source files, 1 MiB for each file, 4096 rules, 64 parameters for each rule, 256 `--set` values.

## Composition
1. Read the base ruleset.
2. Unify each overlay with the result, in the sequence of the flags.
3. Apply each `--set` value with `FillPath`. The value takes the type of the current parameter. A list value is comma-separated.
4. Check the result. Each rule must be concrete after the defaults apply.
5. Decode the rules and sort them by ID. Decode the exclude map and add each `--exclude` pattern to it.
6. Apply `--enable-rule` and `--omit-rule`. These flags win over the `enabled` field of any source.

## CUE defaults
CUE unification does not replace a concrete value: `max_lines: 60` and `max_lines: 40` conflict. So the base ruleset declares each tunable value as a default, for example `max_lines: *60 | int`. An overlay with `max_lines: 40` then selects 40.

- The same applies to lists. `nasa-p10` declares `languages: *["go", "c"] | [...]`, so an overlay can narrow the list.
- Two different defaults cancel each other. `(*"error" | "warning") & (*"warning" | "error")` has no default, so the value stays incomplete.
- For this reason the schema gives no defaults for `enabled` and `severity`. The loader applies them in Go. A base ruleset can declare its own default, for example `severity: *"warning" | "error" | "note"`.

## Related
- [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md): composable CUE rulesets.
- [CLI contract](/specs/cli-contract--01a1e9729633.md)
- [ISSUE-003](/issues/issue-003-phase-3-composable-cue-ruleset-engine--01a1e97296e2.md)
