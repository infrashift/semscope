---
id: "01a1e972-9645-7396-bccb-45faa9d82beb"
title: "CUE ruleset schema"
type: "Spec"
tags:
  - "spec"
  - "ruleset"
  - "cue"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:21:28-06:00"
description: "The CUE schema of a semscope ruleset and the steps that compose a base ruleset, overlays and overrides."
status: "draft"
---

# CUE ruleset schema

This spec states the intended schema. Task 3.1 implements it in `internal/ruleset/schema.cue`.

## Shape
```cue
#Ruleset: {
	name:        string
	description: string
	rules: [ID=string]: #Rule & {id: ID}
}

#Rule: {
	id:        string
	enabled:   bool | *true
	severity:  *"error" | "warning" | "note"
	languages: [...#Language]
	title:     string
	message:   string
	params: [string]: int | string | bool | [...string]
	check: #Check
}

#Language: "go" | "c" | "cpp" | "zig" | "python" | "rust"
```

## Check kinds
`check.kind` selects the checker. Each kind has its own fields.

| Kind | Rule | Purpose |
|---|---|---|
| `no_goto_or_recursion` | R1 | `goto`, `setjmp`, `longjmp`, direct and indirect recursion |
| `bounded_loops` | R2 | Loops without a static bound |
| `no_dynamic_alloc` | R3 | Allocator calls outside the init functions |
| `max_function_lines` | R4 | Function length |
| `min_assertions` | R5 | Assertion density, `average` or `per_function` |
| `narrowest_scope` | R6 | Mutable global state |
| `checked_returns` | R7 | Discarded results and swallowed errors |
| `restricted_preprocessor` | R8 | Macros, conditional compilation, build directives |
| `max_pointer_indirection` | R9 | Pointer depth and function pointer types |
| `ast_query` | any | A Tree-sitter query; each `@violation` capture gives a finding |
| `token_ban` | any | Banned tokens or patterns in the stripped source |

## Composition
1. Read the base ruleset.
2. Unify each overlay with the result, in the sequence of the flags.
3. Apply each `--set` value with `FillPath`.
4. Apply `--enable-rule`, then `--omit-rule`.
5. Validate the result against `#Ruleset`. Each rule must be concrete.

## CUE defaults
CUE unification does not replace a concrete value: `max_lines: 60` and `max_lines: 40` conflict. So the base ruleset declares each tunable value as a default, for example `max_lines: *60 | int`. An overlay with `max_lines: 40` then selects 40. The same applies to `enabled: *true | bool`.

## Related
- [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md): composable CUE rulesets.
- [CLI contract](/specs/cli-contract--01a1e9729633.md)
- [ISSUE-003](/issues/issue-003-phase-3-composable-cue-ruleset-engine--01a1e97296e2.md)
