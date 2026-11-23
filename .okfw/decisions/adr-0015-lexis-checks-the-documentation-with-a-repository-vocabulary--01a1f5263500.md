---
x-okfw-id: "01a1f526-3500-7b17-af40-7dd82ce851da"
title: "ADR-0015: lexis checks the documentation with a repository vocabulary"
type: "ADR"
tags:
  - "adr"
  - "docs"
  - "gates"
x-okfw-created: "2026-11-19T12:51:24-06:00"
x-okfw-timestamp: "2026-11-19T12:51:24-06:00"
description: "The docs-lexis gate scans the Markdown documentation with lexis and vocabulary.cue, fails on each finding and runs only when lexis is on the PATH."
---

# ADR-0015: lexis checks the documentation with a repository vocabulary

**Status:** Accepted
**Date:** 2026-11-19
**Source:** [ISSUE-015](/issues/issue-015-phase-13-lexis-checks-the-documentation--01a1f5265a5d.md). User request on 2026-11-19. lexis ADR-0009 records the same decision for lexis.

## Context
`scripts/ste-check.py` checks the semscope documentation against a part of ASD-STE100. lexis is the documentation linter of infrashift. Its `ste-100` ruleset has more checks: unapproved words, noun forms of verbs, -ing forms, marketing words and passive voice. On 2026-11-19, lexis gave 157 warnings in 52 files of the semscope documentation.

Many warnings name a term of the project, for example "unification" (CUE), "suppression" (an allow comment) or "divergence" (a parity record). A verb in place of such a term changes what the text says. Other warnings mark prose that a verb makes clearer.

lexis runs semscope in its own lint task (lexis ADR-0011). That task runs semscope only when semscope is on the PATH.

## Decision
- `vocabulary.cue` holds the technical terms of semscope. lexis loads it over the built-in ruleset: `--ruleset ste-100 --ruleset vocabulary.cue`. The vocabulary turns off the rules STE-100-06 and STE-100-07 for these terms only.
- Add a term only when it names a concept of the project, a tool or a standard. Rewrite other prose.
- The gate `docs-lexis` scans `README.md`, `CLAUDE.md`, `PLAN.md`, `docs/src/content/docs` and `.okfw` with `--fail-on note`. Each finding fails it.
- The gate skips the daily notes, `.okfw/ephemeral`, `index.md` and `log.md`. The daily notes are a journal. okfw writes `index.md` and `log.md`.
- `check` runs `docs-lexis` after `docs-style`.
- `docs-lexis` runs lexis only when `command -v` finds it. Otherwise it prints a message and passes. The variable `LEXIS` selects another binary.
- The gate reads the Markdown documentation only. It does not scan the Go comments in `cmd/`, `internal/` and `rules/`.

## Consequences
- The documentation gives 0 findings from lexis. The vocabulary removes 113 warnings. New text fixes the other 44.
- semscope has no hard dependency on lexis. lexis already runs semscope, so a hard dependency in both directions makes a cycle.
- On a host without lexis, `ste-check.py` checks the documentation alone. The gates on that host do not see a lexis finding.
- A new term in the documentation needs new text, or a vocabulary entry with a reason.
- `ste-check.py` and the `docs-style` gate do not change.

## Alternatives considered
- **Reword each finding, with no vocabulary.** Rejected: a verb in place of a project term, for example "unification", changes what the text says.
- **Fail on errors only.** Rejected: lexis reports noun forms, -ing forms and passive voice as warnings, so the gate then misses them. lexis ADR-0009 makes the same choice.
- **Make lexis a hard dependency of `check`.** Rejected: lexis runs semscope in its lint task, so each repository then needs the other to pass its gates.
- **Put `--allow-term` flags in the make files.** Rejected: a CUE file gives a reason for each term, and both make files read the same file.
- **Replace `ste-check.py` with lexis.** Rejected: a host without lexis then has no style gate.

## Related
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md): the Power of 10 terms of the vocabulary.
- [ADR-0006](/decisions/adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md): zmake runs the gates.
- [Wiki conventions](/wiki-conventions--01a1e97256ee.md): the style rules.
- [ISSUE-015](/issues/issue-015-phase-13-lexis-checks-the-documentation--01a1f5265a5d.md)

## Later change
- 2026-11-19: lexis is now lexiscope (lexis ADR-0012). The gate is `docs-lexiscope`, and the variable is `LEXISCOPE`. [ISSUE-016](/issues/issue-016-phase-14-the-documentation-gate-uses-the-name-lexiscope--01a1f548ca52.md) records the change.
