---
id: "01a1f4bf-5f46-7bf7-ad49-0ad7692cdef3"
title: "ADR-0013: semscope uses the ztree runtime"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "ztree"
  - "dependency"
created: "2026-11-19T10:59:04-06:00"
timestamp: "2026-11-19T10:59:04-06:00"
description: "internal/parser and internal/engine use the ztree Go binding in place of go-tree-sitter, and the gates build libztree.a first."
---

# ADR-0013: semscope uses the ztree runtime

**Status:** Accepted
**Date:** 2026-11-19
**Source:** [ISSUE-012](/issues/issue-012-phase-10-the-ztree-runtime--01a1f4bf5f39.md). User decision on 2026-11-19: merge the branch `ztree-runtime`. ztree ADR-0007.

## Context
[ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md) uses `github.com/tree-sitter/go-tree-sitter` and six grammar modules. ztree is a batch Tree-sitter runtime in Zig. It follows the Power of 10 rules and holds the six grammars. ztree ADR-0007 makes ztree the Tree-sitter service of semscope and lexis.

## Decision
- `internal/parser` and `internal/engine` use `github.com/infrashift/ztree/bindings/go`. The binding keeps the go-tree-sitter API that semscope uses, so the change is small.
- `go.mod` points a `replace` directive to `../ztree/bindings/go`. cgo links `../ztree/zig-out/lib/libztree.a`.
- The task `ztree` in the `ZMakefile` and the `Makefile` runs `zig build lib --release` in `ZTREE_DIR`. The tasks `build`, `test`, `lint`, `selfscan` and `parity` run it first.
- The grammars come from ztree. go.mod has no grammar modules.

## Consequences
- ztree has no `Close` methods. Go owns the memory of a tree and a query.
- `Parser.Parse` of ztree returns an error. semscope returns it with the path of the file.
- The build needs a ztree checkout next to semscope and Zig 0.17.1.
- `zig build lib` writes the SHA-256 of `libztree.a` into the binding (ztree ADR-0016). The Go build cache then links a new library.
- `zmake check`, `zmake -f Makefile check` and `zmake parity` pass. The parity test passes on each corpus.
- `third_party/tree-sitter-zig` has no user in the build. [ISSUE-013](/issues/issue-013-the-unused-zig-grammar-in-third-party--01a1f4bf5f3f.md) tracks it.

## Alternatives considered
- **Keep go-tree-sitter.** Rejected: semscope and lexis then use two runtimes and two sets of grammars. ztree gives the same trees as the reference runtime on its corpora.
- **A tagged ztree module in place of the `replace` directive.** Rejected for now: ztree has no release.

## Related
- [ADR-0012](/decisions/adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md): the Zig grammar of ztree.
- [Architecture overview](/architecture/semscope-architecture-overview--01a1e97295ff.md)
