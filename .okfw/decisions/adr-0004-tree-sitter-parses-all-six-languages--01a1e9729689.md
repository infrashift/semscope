---
id: "01a1e972-9689-7c77-9a6f-e7179bfdb535"
title: "ADR-0004: Tree-sitter parses all six languages"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "tree-sitter"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "internal/parser uses the cgo Tree-sitter bindings for Go, C, C++, Zig, Python and Rust, with one adapter for each language."
---

# ADR-0004: Tree-sitter parses all six languages

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User decision on 2026-11-17: built-in checks for all six languages.

## Context
The `nasa-p10` ruleset targets Go, C, C++, Zig, Python and Rust. The parity tests need Zig, because `nasa-lint.py` reads Zig only. The `ast_query` rules need one query language for all six languages.

## Decision
- `internal/parser` uses `github.com/tree-sitter/go-tree-sitter` (v0.25) for all six languages, Go included.
- The grammars: `tree-sitter-go`, `tree-sitter-c`, `tree-sitter-cpp`, `tree-sitter-python`, `tree-sitter-rust` and `tree-sitter-grammars/tree-sitter-zig`.
- Each language has one adapter. The adapter gives a flat record of the file: functions, calls, loops, assertions, discards, pointer types and directives.
- Tree walks use a `TreeCursor`. A cursor walks the tree without recursion.

## Consequences
- semscope needs cgo and a C compiler. `go test -race` needs cgo already.
- A cross-build needs a C cross-compiler.
- One query language (`ast_query`) works for all six languages.
- Go code gets no type information. R7 checks the syntax only. `errcheck` covers the typed case for the semscope sources.

## Alternatives considered
- **`go/ast` for Go and Tree-sitter for the rest.** Rejected: two parser paths, and `ast_query` then needs a second query language for Go.
- **Regular expressions, as in the oracles.** Rejected: no structure, so no scopes, no pointer depth and no reliable function bounds.
- **A pure-Go port of Tree-sitter.** Rejected: the official cgo bindings are the maintained option, and each grammar publishes Go bindings for them.

## Related
- [Architecture overview](/architecture/semscope-architecture-overview--01a1e97295ff.md)
- [ISSUE-004](/issues/issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md)
