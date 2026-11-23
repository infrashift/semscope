---
x-okfw-id: "01a1ebe0-5590-74a1-a6dc-5918c6528c12"
title: "ADR-0010: A patched Zig grammar for Zig 0.17"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "zig"
  - "dependency"
x-okfw-created: "2026-11-17T17:38:30-06:00"
x-okfw-timestamp: "2026-11-17T17:38:30-06:00"
description: "semscope keeps a patched copy of tree-sitter-zig v1.1.2 without the keywords async, await and usingnamespace, because no release reads Zig 0.17."
status: "deprecated"
---

# ADR-0010: A patched Zig grammar for Zig 0.17

**Status:** Superseded by [ADR-0014](/decisions/adr-0014-the-zig-grammar-lives-in-ztree-only--01a1f4d211fc.md). The grammar lives in ztree only.
**Date:** 2026-11-17
**Source:** User request: the Zig limits of [ISSUE-008](/issues/issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md) (`PLAN.md` Task 7.2).

## Context
Zig 0.17 has no keywords `async`, `await` and `usingnamespace`. The Zig tokenizer of the 0.17.1 standard library lists the keywords. `tree-sitter-zig` v1.1.2 is the last release, and it still reserves the three words. Upstream master also keeps them. The 0.17 standard library uses `await` as an identifier, for example `await(ev, future, result, result_alignment)`. The parse error makes `fn cancel` of `Io/Dispatch.zig` 275 lines long, so NASA-P10-04 gives a false finding.

## Decision
- Keep a patched copy of the v1.1.2 grammar in `third_party/tree-sitter-zig`. The `replace` directive in `go.mod` selects it.
- The patch removes three rules from `grammar.js`: `async_expression`, `await_expression` and `using_namespace_declaration`.
- The patch lets `asm_clobbers` read an anonymous struct literal, the clobber form of Zig 0.15 ([ISSUE-009](/issues/issue-009-the-zig-grammar-does-not-read-asm-clobbers-of-zig-0-15--01a1ebef12f0.md)).
- The patch changes no other rule.
- Generate `src/parser.c` with `tree-sitter-cli` 0.25.10, the version of the go-tree-sitter runtime.
- Base the copy on v1.1.2, not on master. Master also puts call arguments in an `arguments` node, and the adapter reads the v1.1.2 node shapes.
- The self-scan excludes `third_party`. golangci-lint and `nasa-lint-go.py` do not read it, because it is a separate Go module.

## Consequences
- Zig code before 0.15 that uses `async`, `await` or `usingnamespace` gives parse errors. semscope targets current Zig.
- The repository holds a generated parser of about 6 MB.
- The Go build cache does not see a new `parser.c`, because the binding includes it from outside its package. `binding.go` holds the SHA-256 of `parser.c`, so a new parser changes `binding.go` too. `TestZigParserHash` fails when the hash is old.
- An update of the grammar is a manual task: apply the patch to the new `grammar.js` and generate again. `third_party/tree-sitter-zig/README.md` gives the steps.
- Upstream can take the patch. A pull request is outside this task.
- [ADR-0012](/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md) replaces this grammar with the grammar of ztree. It adds empty containers and three rare forms.

## Alternatives considered
- **Wait for a release.** Rejected: master does not remove the keywords.
- **Rewrite `await` in the source before the parse.** Rejected: it changes the columns of the findings and hides real syntax.
- **Take upstream master and adapt the adapter.** Rejected for now: a second change of node shapes in the same task.
