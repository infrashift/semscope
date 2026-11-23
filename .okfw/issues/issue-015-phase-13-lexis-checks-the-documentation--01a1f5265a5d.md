---
x-okfw-id: "01a1f526-5a5d-7d3e-a4c0-705941f412d4"
title: "ISSUE-015: Phase 13: lexis checks the documentation"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "docs"
  - "gates"
  - "phase-13"
  - "closed"
x-okfw-created: "2026-11-19T12:51:33-06:00"
x-okfw-timestamp: "2026-11-19T12:52:32-06:00"
description: "A repository vocabulary and new text bring the lexis findings of the documentation to 0, and the docs-lexis gate keeps them at 0."
---

# ISSUE-015: Phase 13: lexis checks the documentation

**Status/Phase:** closed · Phase 13

## Description
lexis checks Markdown files against ASD-STE100. On 2026-11-19, its scan of the semscope documentation gave 0 errors and 157 warnings in 52 files. `scripts/ste-check.py` does not check noun forms, -ing forms and some unapproved words, so the `docs-style` gate did not see these warnings. No gate keeps the count at 0.

## Acceptance Criteria
- [x] `vocabulary.cue` holds the technical terms of semscope, each with a reason.
- [x] New text fixes each other finding. The decisions and the facts of the ADRs do not change.
- [x] The `ZMakefile` and the `Makefile` have the gate `docs-lexis`. `check` runs it after `docs-style`.
- [x] `docs-lexis` runs lexis only when lexis is on the PATH, and each finding fails it.
- [x] [ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md) records the decision.
- [x] `CLAUDE.md`, the development workflow, the architecture overview and the site pages name the gate.
- [x] `zmake check` and `zmake -f Makefile check` pass. The lexis scan gives 0 errors, 0 warnings and 0 notes.

## Notes
- The gate reads the Markdown documentation only: `README.md`, `CLAUDE.md`, `PLAN.md`, `docs/src/content/docs` and `.okfw`. The Go comments stay out of scope.
- lexis does not check headings, tables, frontmatter and code spans.

## Resolution
Task 13.1 adds `vocabulary.cue` with 27 terms. The terms come from the Power of 10 rules, CUE, SARIF, the parity records and the project process. The vocabulary removes 113 of the 157 warnings. New text fixes the other 44: 25 -ing forms (STE-100-07), 12 noun forms (STE-100-06), 4 unapproved words (STE-100-05) and 3 marketing words (STE-100-09). Task 13.2 adds the gate `docs-lexis`. The lexis scan now gives 0 errors, 0 warnings and 0 notes in 52 files.

## Related
- [ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md)
- [Wiki conventions](/wiki-conventions--01a1e97256ee.md)
