---
id: "01a1e972-95ed-7298-847a-65bed36c7f51"
title: "Development workflow"
type: "Guide"
tags:
  - "conventions"
  - "workflow"
  - "meta"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "The loop for each semscope task: implement, run the zmake gates until green, update the wiki, then commit."
---

# Development workflow

Every change follows this loop. Do not report or commit work while a gate fails.

1. **Pick the task.** Read `PLAN.md`. Take the first unchecked task. Find the phase issue with `okfw search --path issues --tag open`.
2. **Tag the issue.** At the start of a phase, swap the tag `open` for `in-progress`.
3. **Search the wiki.** Read the ADRs and specs for the task before you design.
4. **Do only that task.** Keep `cmd/semscope/main.go` a thin entry point.
5. **Apply the Power of 10 rules.** Read [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md). Use bounded loops and explicit stacks, not recursion.
6. **Run the gates.** Run `zmake check`. Then run `zmake -f Makefile check`.
7. **Repeat until green.** Fix each failure and run both gates again. Do not weaken or delete tests.
8. **Update the docs site.** Add or refresh the `docs/` pages for the delivered features. Update the roadmap page. Run `zmake docs-build`.
9. **Record decisions.** Write an ADR for each decision when you make it. File each bug or design debt as an issue.
10. **Close the task.** Tick the `PLAN.md` checkbox.
11. **Close the phase.** Tick the acceptance criteria of the issue. Add a **Resolution** section. Swap the tag `in-progress` for `closed`.
12. **Journal.** Add an entry to the `okfw daily` note: what changed, what you learned, what is next.
13. **Commit.** Make one commit for each task. Commit the code, the wiki, the docs pages and `PLAN.md` together.

## Gates
`zmake check` runs these gates. [ADR-0006](/decisions/adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md) records why zmake runs them.

| Gate | Command | Checks |
|---|---|---|
| lint | `gofmt -l`, `nasa-lint.py`, `nasa-lint-go.py` and its fixture tests, `go vet`, `golangci-lint run` | Format, Power of 10 rules, static analyzers |
| test | `go test -race ./...` | Unit tests with the race detector |
| wiki | `okfw validate --all`, `okfw validate --okf` | Frontmatter schema, OKF conformance, links |
| docs-style | `scripts/ste-check.py` | ASD-STE100 style |
| docs-lexiscope | `lexiscope scan --ruleset ste-100 --ruleset vocabulary.cue --fail-on note`, when lexiscope is on the PATH | ASD-STE100 style of the Markdown files, with the repository vocabulary |
| selfscan | `semscope scan --ruleset nasa-p10` (from Task 4.1) | semscope on its own sources |

[ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md) records the `docs-lexiscope` gate. `lexis` is a short alias of lexiscope. `zmake check` does not build the docs site. `zmake docs-build` builds it with bun ([ADR-0011](/decisions/adr-0011-document-semscope-with-astro-starlight-on-github-pages--01a1ed3cd653.md)).

The ZMakefile (zmake, with `@name:` tasks) and the Makefile (GNU Make syntax) hold the same tasks. zmake reads the ZMakefile first, also with `--legacy`. Run `zmake -f Makefile check` to run the Makefile.

See [Wiki conventions](/wiki-conventions--01a1e97256ee.md).
