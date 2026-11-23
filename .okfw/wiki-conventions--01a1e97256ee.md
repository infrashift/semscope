---
x-okfw-id: "01a1e972-56ee-78a9-a97a-44f30e4f489c"
title: "Wiki conventions"
type: "Guide"
tags:
  - "conventions"
  - "meta"
x-okfw-created: "2026-11-17T06:19:07-06:00"
x-okfw-timestamp: "2026-11-17T06:19:07-06:00"
description: "Layout, writing style and authoring rules for the architecture notes, specs, ADRs, issues and daily notes in this wiki."
---

# Wiki conventions

This guide defines how the semscope OKF wiki (`.okfw/`) stores project knowledge.

- Create concepts with `okfw create` or the okfw MCP `create` tool. Do not write frontmatter by hand.
- Search before you write. Link to related concepts.
- Keep one idea per concept. Give each concept a one-sentence `description`.
- Do not start a `--body` with `# Title`. okfw writes the H1 from the title.

## Layout
| Path | Type | Contents |
|---|---|---|
| `architecture/` | `Architecture` | System overview, package boundaries, data flow. |
| `specs/` | `Spec` | Contracts: the oracle check inventories, the CLI, the CUE ruleset schema. The spec states intent. The code implements it. |
| `decisions/` | `ADR` | Architecture decision records. |
| `issues/` | `Issue` (kind `issue`) | Tracked work: one issue for each `PLAN.md` phase, plus bugs and design debt. |
| `daily/` | daily note | `okfw daily`: what changed, what you learned, what is next. |

## PLAN.md and issues
`PLAN.md` is the phase checklist. Each phase has a mirror issue (ISSUE-001 to ISSUE-005). Each phase heading in `PLAN.md` has an `Issue:` line that links to its issue. When a phase is complete, tick the `PLAN.md` checkboxes and the acceptance criteria of the issue.

## ADRs (`decisions/`)
- Title: `ADR-NNNN: <decision>`. Number ADRs in sequence. Do not reuse a number.
- Command: `okfw create "ADR-NNNN: <title>" --path decisions --type ADR --tags adr,<topic> -d "<one sentence>" --body "<markdown>"`
- Body: **Status** (Proposed, Accepted, or Superseded by ADR-XXXX), **Date**, **Source**, then **Context**, **Decision**, **Consequences** and **Alternatives considered**.
- To supersede an ADR, write the new ADR with a link to the old ADR. Update the Status line of the old ADR. Run `okfw set-lifecycle <id> deprecated`.
- Write an ADR before you add a Go module dependency.

## Issues (`issues/`)
- Title: `ISSUE-NNN: <summary>`. Number issues in sequence.
- Command: `okfw create "ISSUE-NNN: <title>" --kind issue --path issues --type Issue --tags issue,open,phase-<n>,<area> -d "<summary>" --body "<markdown>"`
- Body: a **Status/Phase** line, **Description**, **Acceptance Criteria** (checkboxes), **Notes** and **Related**.
- okfw has no status field for issues. Tags carry the state: `open`, then `in-progress`, then `closed`. Use `okfw remove-tags` and `okfw add-tags`.
- Keep the Status line the same as the tags. Add a **Resolution** section when you close an issue.
- Find open work with `okfw search --path issues --tag open`.

## Writing style
All documentation follows ASD-STE100 Simplified Technical English. This includes the wiki, `README.md`, `PLAN.md`, the Go comments and the CLI help strings.

- Use active voice and the simple present tense: "The loader rejects the rule."
- Start each procedural step with an imperative verb: Run, Read, Inspect, Verify, Add, Set.
- Keep procedural steps to 20 words or fewer. Keep descriptive sentences to 25 words or fewer.
- Use a verb, not a noun form: "The engine checks the file", not "The check of the file occurs".
- Use one term for one concept. This wiki uses these terms:
  - *ruleset*: a CUE value with a `rules` map. *base ruleset*: the first ruleset that the loader reads, for example `nasa-p10`.
  - *overlay*: a user CUE file that the loader unifies with the base ruleset.
  - *rule*, *check* (the analyzer kind of a rule), *parameter* (a value in `params`) and *finding* (one violation at one location).
  - *oracle*: `scripts/nasa-lint.py` for Zig, or `scripts/nasa-lint-go.py` for Go.
  - *gate*: one check in `zmake check`. *task* and *phase*: units of work in `PLAN.md`.
- Do not use the words `should`, `will`, `would`, `please`, `simply`, `just`, `very`, `really`, `easily` or `your`. Do not use marketing language.
- Run `zmake docs-style`. The check (`scripts/ste-check.py`) reports banned words and long sentences as errors. It reports possible passive voice as warnings. The gate allows zero warnings.
- Run `zmake docs-lexiscope`. lexiscope also reports noun forms, -ing forms and unapproved words. Each finding fails the gate. Add a term to `vocabulary.cue` only when it names a concept of the project, a tool or a standard ([ADR-0015](/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md)).

## Housekeeping
- Use `okfw move` and `okfw delete`. Do not use `mv` or `rm`. The okfw commands repair links and update the `index.md` listings.
- `status` is the OKF lifecycle field (`draft`, `stable` or `deprecated`). It is not the task state. Mark superseded knowledge `deprecated`. Do not delete it.
- Do not edit `index.md` or `log.md`. okfw generates them.
- Run `zmake wiki` before you commit wiki changes.

See [Development workflow](/development-workflow--01a1e97295ed.md).
