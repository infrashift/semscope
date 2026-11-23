---
x-okfw-id: "01a1fecc-5cce-79e5-8ee3-c5d5fa771324"
title: "ADR-0016: Tree-sitter parses seven languages"
type: "ADR"
tags:
  - "adr"
  - "parser"
  - "tree-sitter"
x-okfw-created: "2026-11-21T09:49:28-06:00"
x-okfw-timestamp: "2026-11-21T09:49:28-06:00"
description: "internal/parser reads Go, C, C++, Zig, Python, Rust and C# through the ztree runtime, with one adapter for each language; nasa-p10 applies to C# except NASA-P10-03."
---

# ADR-0016: Tree-sitter parses seven languages

**Status:** Accepted
**Date:** 2026-11-21
**Source:** User request on 2026-11-21: semscope checks C# files. User decision on 2026-11-21: NASA-P10-03 does not apply to C#. [ISSUE-017](/issues/issue-017-phase-15-c-sharp-support--01a1fecc5cbb.md) records the work.

## Context
[ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md) fixes six languages: Go, C, C++, Zig, Python and Rust. ztree adds the C# grammar tree-sitter-c-sharp v0.24.0 in its ADR-0019, with `LangCSharp` in its Go binding. semscope reads only the languages of its parser, so it skips each `.cs` file.

C# has a garbage collector. Each class instance comes from `new`, so a check of NASA-P10-03 reports most C# functions. Python has no NASA-P10-03 check for the same reason. C# has pointers and function pointers (`delegate*`) in `unsafe` code, so NASA-P10-09 applies.

## Decision
This ADR replaces ADR-0004. The decisions of ADR-0004 stay, with C# as the seventh language:
- `internal/parser` reads Go, C, C++, Zig, Python, Rust and C# through the ztree runtime ([ADR-0013](/decisions/adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md)). The language name is `csharp`, and the file extension is `.cs`.
- Each language has one adapter. `internal/parser/lang_csharp.go` is the C# adapter. It gives the same flat record as the other adapters: functions, calls, loops, assertions, discards, pointer types and directives.
- Tree walks use a `TreeCursor`, without recursion. One query language (`ast_query`) works for all seven languages.
- The key of a C# function is `Type.name`, with the innermost class, struct, record or interface. A plain call and a call through `this` name a function of the type of the caller. A call through a type name names a function of that type.
- `nasa-p10` applies to C#, except NASA-P10-03. The C# data of the rules:
  - NASA-P10-05: the assertion calls `Debug.Assert` and `Trace.Assert`, also with a namespace.
  - NASA-P10-06: a `static` field that is not `readonly` or `const`.
  - NASA-P10-07: `_ = F();`, and an empty `catch` block with no type or the type `Exception`.
  - NASA-P10-08: the banned directives `#define` and `#undef`, and the depth of `#if`.
  - NASA-P10-10: each `unsafe` block, each `unsafe` modifier and each `#pragma warning disable` without a reason in a comment.
- A C# file with a name that ends in `Test.cs` or `Tests.cs` is a test file. A method with a test attribute of MSTest, NUnit or xUnit is exempt. semscope does not enter a class with the attribute `[TestClass]` or `[TestFixture]`.

## Consequences
- semscope needs a `libztree.a` with the C# grammar: ztree commit of Phase 22 or later.
- C# has no oracle. Fixtures alone test the C# checks, as for C, C++, Python and Rust.
- A scan does not skip `bin` and `obj`. A .NET project adds these directories to the exclude map. A built-in skip of `bin` also skips the scripts of other languages.
- A call in an interpolated string (`$"{F()}"`) gives no call record, because the walker blanks each string.
- ztree does not compile a supertype pattern (ztree ADR-0007). The C# grammar has nine visible supertypes, for example `type`, `expression` and `statement`. A C# query names the concrete node types.
- An overlay can add `"csharp"` to the languages of NASA-P10-03 with its own `alloc_calls`.

## Alternatives considered
- **All ten rules for C#.** Rejected by the user: NASA-P10-03 then reports most C# code.
- **NASA-P10-03 and NASA-P10-09 off, as for Python.** Rejected by the user: C# has pointers in `unsafe` code.
- **The key `name` only, as for Python and Rust.** Rejected: two methods with the same name in two classes of one file then share one node of the call graph.

## Related
- [ADR-0004](/decisions/adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md): the six languages, replaced by this ADR.
- [Architecture overview](/architecture/semscope-architecture-overview--01a1e97295ff.md)

