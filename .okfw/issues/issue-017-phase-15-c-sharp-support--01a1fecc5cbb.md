---
id: "01a1fecc-5cbb-761b-8884-6c09466f405b"
title: "ISSUE-017: Phase 15: C# support"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "phase-15"
  - "parser"
  - "rules"
  - "closed"
created: "2026-11-21T09:49:28-06:00"
timestamp: "2026-11-21T09:49:53-06:00"
description: "semscope reads C# files and checks the nasa-p10 rules in them, except NASA-P10-03."
---

# ISSUE-017: Phase 15: C# support

**Status/Phase:** closed · Phase 15

## Description
ztree has a C# grammar from its Phase 22 (ztree ADR-0019). semscope does not read C# files: `#Language` has no `csharp`, `parser.Detect` maps no `.cs` extension, and `nasa-p10` has no C# data. [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md) fixes six languages.

## Acceptance Criteria
- [x] `parser.CSharp` reads `.cs` files with `ts.LangCSharp`. `Languages()` holds seven languages.
- [x] `internal/parser/lang_csharp.go` gives functions, calls and sites for the checks of NASA-P10-01, -02, -04 to -09.
- [x] `#Language` of the schema has `"csharp"`. `nasa-p10` lists `csharp` for each rule except NASA-P10-03, and has C# data for NASA-P10-05, -08 and -10.
- [x] Each rule directory of `testdata/nasa_p10/violations/` except R03 has a C# fixture, and `expected.txt` holds its findings.
- [x] `internal/parser/testdata/sample_a.cs` and its golden file cover the adapter.
- [x] [ADR-0016](/decisions/adr-0016-tree-sitter-parses-seven-languages--01a1fecc5cce.md) replaces ADR-0004. ADR-0004 has the status `deprecated`.
- [x] `zmake check` and `zmake -f Makefile check` pass.

## Resolution
- The C# adapter records methods, constructors, destructors, operators, accessors and local functions with a body. The key is `Type.name`. The recursion check resolves a plain call, a call through `this` and a call through a type name.
- The fixtures give 17 C# findings for nine rules. The sample gives 11 findings.
- A scan does not skip `bin` and `obj`. The scanning guide tells a .NET project to add them to the exclude map.
- The pragma query matches the text of the `preproc_pragma` node. A comment on the same line is a child of that node, so `[^/]*$` finds a pragma without a reason.

## Related
- [ADR-0016](/decisions/adr-0016-tree-sitter-parses-seven-languages--01a1fecc5cce.md)
- [Development workflow](/development-workflow--01a1e97295ed.md)

