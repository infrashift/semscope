---
x-okfw-id: "01a1f548-ca52-7082-b13d-815e00b1d9b4"
title: "ISSUE-016: Phase 14: the documentation gate uses the name lexiscope"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-14"
  - "docs"
  - "gates"
  - "closed"
x-okfw-created: "2026-11-19T13:29:10-06:00"
x-okfw-timestamp: "2026-11-19T13:29:45-06:00"
description: "The docs-lexis gate becomes docs-lexiscope, because the documentation linter lexis now has the official name lexiscope."
---

# ISSUE-016: Phase 14: the documentation gate uses the name lexiscope

**Status/Phase:** closed · Phase 14

## Description
The documentation linter lexis has the official name lexiscope from lexis commit `c8963c9` (lexis ADR-0012). `lexis` stays as a short alias: `~/.local/bin/lexis` is a symbolic link to `~/.local/bin/lexiscope`. The repository moves to `github.com/infrashift/lexiscope`. The site moves to <https://infrashift.github.io/lexiscope/>. The gate of [ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md) still uses the old name.

## Acceptance Criteria
- [x] The `ZMakefile` and the `Makefile` have the gate `docs-lexiscope` and the variable `LEXISCOPE`, with the default `lexiscope`.
- [x] `zmake -n check`, `make -n check` and `zmake -f Makefile -n check` print the same commands.
- [x] `LEXISCOPE=no-such-tool zmake docs-lexiscope` prints a message and passes.
- [x] `CLAUDE.md`, `README.md`, `vocabulary.cue`, the development workflow, the wiki conventions, the architecture overview and the site pages use the name lexiscope.
- [x] ADR-0015 has a dated note about the change. The other records keep the old name.
- [x] `zmake check` and `zmake -f Makefile check` pass. The lexiscope scan gives 0 errors, 0 warnings and 0 notes.

## Notes
- The records keep the old name: the ADRs, the closed issues, the old daily entries and the text of the complete phases of `PLAN.md`.
- The file names of the records do not change.

## Resolution
Task 14.1 renames the gate `docs-lexis` to `docs-lexiscope` and the variable `LEXIS` to `LEXISCOPE` in the `ZMakefile` and the `Makefile`. The three dry runs of `check` print the same commands. Task 14.2 puts the name lexiscope in the current documents and adds a dated note to ADR-0015. The lexiscope scan gives 0 errors, 0 warnings and 0 notes in 55 files.

## Related
- [ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md)
- [ISSUE-015](/issues/issue-015-phase-13-lexis-checks-the-documentation--01a1f5265a5d.md)
- [Development workflow](/development-workflow--01a1e97295ed.md)
