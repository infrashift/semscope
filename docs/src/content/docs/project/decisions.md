---
title: Design Decisions
description: A summary of the semscope architecture decision records (ADRs) in the project wiki.
---

The project wiki in `.okfw/decisions/` holds the full ADRs. Each ADR records the
context, the decision, the consequences and the rejected options. This page
gives a summary of each.

### ADR-0001: Adopt an OKF wiki for project knowledge

A local-first OKF wiki in `.okfw/` holds the architecture, the specs, the
decisions and the issues. The `okfw` tool manages it. The layout copies the
zmake wiki. `PLAN.md` stays the phase checklist.

### ADR-0002: Composable CUE rulesets with a nasa-p10 baseline

A ruleset is a CUE value. The binary embeds the schema and the `nasa-p10`
ruleset. An overlay extends a ruleset. The CLI omits rules with `--omit-rule`
and tunes parameters with `--set`. The check kinds `ast_query` and `token_ban`
let a user write a new code rule in CUE alone. See
[Rulesets & Overlays](../../guides/rulesets/).

### ADR-0003: Adapt the NASA Power of 10 rules to Go

The semscope code follows the Power of 10 rules where Go allows:

- No recursion. Tree walks use a Tree-sitter cursor or an explicit bounded stack.
- Loops iterate over a slice or count toward a limit.
- Functions of 60 lines or fewer, with an average of two or more `assert.That` calls.
- No mutable package variables, no function types and no `unsafe`.

Go, cobra, CUE and Tree-sitter allocate memory at run time. This is a recorded
deviation from R3. The self-scan overlay `.semscope.cue` turns off NASA-P10-03.

### ADR-0004: Tree-sitter parses all six languages

`internal/parser` uses the cgo Tree-sitter bindings for Go, C, C++, Zig,
Python and Rust. Each language has one adapter. The adapter gives a flat
record of the file: functions, calls, loops, assertions, discards, pointer
types and directives.

### ADR-0005: Two oracles define nasa-p10 parity

`scripts/nasa-lint.py` is the Zig oracle. `scripts/nasa-lint-go.py` is the Go
oracle. On the fixtures, semscope and the oracle report the same findings. On
a real corpus, each divergence needs a fix or a record with a reason. semscope
does not copy a known oracle defect.

### ADR-0006: zmake runs the development gates

A `ZMakefile` and a `Makefile` hold the same tasks. `zmake check` runs the
`ZMakefile`. `zmake -f Makefile check` runs the `Makefile`. Each task commit
needs both gates green. See [Make Targets](../../reference/make-targets/).

### ADR-0007: cobra parses the command line

semscope uses `github.com/spf13/cobra`. `--ruleset` and `--set` keep a comma
in the value. `--enable-rule` and `--omit-rule` split a value at each comma.
The function `run` prints each error once and maps it to an exit code.

### ADR-0008: go-sarif writes the SARIF reports

semscope writes SARIF 2.1.0 and 2.2 through `github.com/owenrumney/go-sarif/v3`.
It validates each report against its JSON schema before it writes the report.
A report has no absolute paths. See [Output Formats](../../guides/output-formats/).

### ADR-0009: Type information for the Go check of NASA-P10-07

When the parameter `go_types` is `true`, semscope loads the Go packages with
`golang.org/x/tools/go/packages`. It reports each call, `defer` and `go`
statement that drops an `error` result. The default is `false`, so the parity
with the Go oracle does not change.

### ADR-0010: A patched Zig grammar for Zig 0.17

Zig 0.17 has no keywords `async`, `await` and `usingnamespace`. No release of
`tree-sitter-zig` removes them. semscope kept a patched copy of v1.1.2 in
`third_party/tree-sitter-zig`. ADR-0014 replaces this decision.

### ADR-0011: Document semscope with Astro Starlight on GitHub Pages

This site lives in `docs/`. It uses Astro, Starlight and bun 1.3.9, the same
versions as the zmake site. A GitHub Actions workflow deploys it to GitHub
Pages. `zmake check` does not build the site, so the Go gates do not need bun.

### ADR-0012: The Zig grammar of ztree

semscope takes the patched Zig grammar of ztree. It adds four changes to the
patch of ADR-0010. A container can be empty. A container field can be an error
set. A slice can hold an `if` type. The asm clobbers can be an identifier. The
generated parser is the same as the parser of ztree, so both give the same
trees.
ADR-0014 replaces this decision: the grammar now lives in ztree only.

### ADR-0013: semscope uses the ztree runtime

`internal/parser` and `internal/engine` use the Go binding of ztree, a
Tree-sitter runtime in Zig. The grammars also come from ztree. The task
`ztree` builds `libztree.a` before the Go tasks. The build needs a ztree
checkout next to semscope.

### ADR-0014: The Zig grammar lives in ztree only

semscope removes its copy of the patched Zig grammar. The parser takes all
grammars from ztree, so the copy had no user. A change of the Zig grammar
occurs in ztree only. This decision deprecates ADR-0010 and ADR-0012.
