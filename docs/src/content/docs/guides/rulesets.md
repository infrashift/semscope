---
title: Rulesets & Overlays
description: Compose a ruleset from a base and overlays, tune it with flags, and write a new rule in CUE.
---

A ruleset is a CUE value. It holds a map of rules and an optional exclude map.
The semscope binary embeds the schema and the built-in ruleset `nasa-p10`.
[Ruleset Schema](../../reference/ruleset-schema/) gives each field.

## Sources

The flag `--ruleset` (`-r`) takes one of these values:

- The name of a built-in ruleset, for example `nasa-p10`. `semscope rules list` lists them.
- A `.cue` file.
- A directory. semscope reads its `.cue` files in name sequence and skips other files.

A built-in name has priority over a path with the same name.
Without `--ruleset`, semscope reads `nasa-p10`.
A `.cue` file alone is a complete ruleset. To extend `nasa-p10`, give both: `-r nasa-p10 -r overlay.cue`.

## Composition

semscope builds the active ruleset in this sequence:

1. Read the first `--ruleset` value.
2. Unify each next `--ruleset` value with the result, in the sequence of the flags.
3. Apply each `--set` value.
4. Check the result against the schema.
5. Add each `--exclude` pattern to the exclude map.
6. Apply `--enable-rule` and `--omit-rule`. These flags win over the `enabled` field of each source.

CUE unification adds values. It does not replace a concrete value.
For this reason, `nasa-p10` declares each tunable value as a CUE default, for example `max_lines: *60 | int`.
An overlay with `max_lines: 40` selects 40 from the disjunction.

## An overlay

```cue
// overlay.cue
package overlay

// Allow 40 lines for each function.
rules: "NASA-P10-04": params: max_lines: 40

// Do not check the assertion average.
rules: "NASA-P10-05": enabled: false

// Check the loop bounds in Go files only.
rules: "NASA-P10-02": languages: ["go"]

// Do not scan generated code. Each pattern needs a reason.
exclude: gen: "generated code"
```

```sh
semscope scan -r nasa-p10 -r overlay.cue .
```

An overlay can do these changes:

- Change a parameter, a severity or the list of languages.
- Omit a rule with `enabled: false`.
- Add a rule.
- Add an exclude pattern.

An overlay cannot remove a field of the base. It also cannot add a field that the schema does not know. The schema closes each struct, so an unknown field gives the error `field not allowed`.

## Flags

The command-line flags change the ruleset without a file:

```sh
semscope scan --omit-rule NASA-P10-03,NASA-P10-05 .
semscope scan --enable-rule NASA-P10-05 -r nasa-p10 -r overlay.cue .
semscope scan --set NASA-P10-04.max_lines=40 --set NASA-P10-07.go_types=true .
```

- `--omit-rule` (or `--disable-rule`) and `--enable-rule` take rule IDs, comma-separated.
- `--set RULE_ID.param=value` parses the value with the type of the parameter: an integer, a boolean, a string or a comma-separated list.
- An unknown rule ID or parameter gives exit status 2.

## A new rule in CUE

Two check kinds let a ruleset add a code rule without a change to the Go code.

`token_ban` bans regular expressions. It reads the source without the comments and the strings.

`ast_query` runs one Tree-sitter query for each language. Each `@violation` capture gives a finding.

ztree does not compile a pattern of a supertype node. The C# grammar has visible supertypes, for example `type`, `expression` and `statement`. In a C# query, name the concrete node types, for example `predefined_type` in place of `type`.

```cue
package overlay

rules: "CUSTOM-01": {
	languages: ["go"]
	title:     "No fmt.Println"
	message:   "Write output through the report package."
	check: {kind: "token_ban", patterns: go: ["\\bfmt\\.Println\\s*\\("]}
}

rules: "CUSTOM-02": {
	languages: ["go"]
	severity:  "warning"
	title:     "No os.Exit"
	message:   "Return an error to main. Do not call os.Exit."
	check: {
		kind: "ast_query"
		query: go: """
			(call_expression function: (selector_expression
			  operand: (identifier) @pkg (#eq? @pkg "os")
			  field: (field_identifier) @violation (#eq? @violation "Exit")))
			"""
	}
}
```

Add the two rules to `overlay.cue`. A scan of a file that calls `fmt.Println` and `os.Exit` gives two more findings:

```console
$ semscope scan -r nasa-p10 -r overlay.cue .
...
main.go:25:2: error: CUSTOM-01: banned pattern `\bfmt\.Println\s*\(`
main.go:26:5: warning: CUSTOM-02: construct `Exit`
```

The rule `NASA-P10-10` of `nasa-p10` is an `ast_query` rule. Its queries are examples for each language.
Run `semscope rules inspect nasa-p10` to see the active rules.

## Inspect a ruleset

`semscope rules inspect` takes the same sources as `--ruleset`. It shows the sources, the exclude map, the active rules, the omitted rules and the effective parameters:

```console
$ semscope rules inspect nasa-p10 overlay.cue
ruleset: nasa-p10
sources: builtin:nasa-p10, overlay.cue
exclude: gen (generated code)

active rules: 11
...
omitted rules: 1
  NASA-P10-05  error  min_assertions  Assertion density
      languages: go, c, cpp, zig, python, rust, csharp
      params: min_assertions=2 mode="average"
```
