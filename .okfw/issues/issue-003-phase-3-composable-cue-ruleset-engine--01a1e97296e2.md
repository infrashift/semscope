---
id: "01a1e972-96e2-749a-a7e3-93903179bbf7"
title: "ISSUE-003: Phase 3: Composable CUE ruleset engine"
type: "Issue"
kind: "issue"
tags:
  - "issue"
  - "open"
  - "phase-3"
  - "ruleset"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "Add the CUE ruleset schema, the loader with overlays and overrides, and the nasa-p10 ruleset."
---

# ISSUE-003: Phase 3: Composable CUE ruleset engine

**Status/Phase:** open · Phase 3

## Description
Add the CUE ruleset schema, the loader with overlays and overrides, and the embedded `nasa-p10` ruleset.

## Acceptance Criteria
- [ ] `internal/ruleset/schema.cue` matches the [CUE ruleset schema](/specs/cue-ruleset-schema--01a1e9729645.md).
- [ ] The loader unifies a base ruleset with overlays and applies `--set`, `--enable-rule` and `--omit-rule`.
- [ ] `rules/nasa_p10/nasa_p10.cue` defines NASA-P10-01 to NASA-P10-10 with parameters for each language.
- [ ] Tests: add `CUSTOM-01`, omit `NASA-P10-05` in two ways, and set `max_lines` to 40 in two ways.
- [ ] Tests: an unknown rule ID, an unknown parameter and a type mismatch each give an error.

## Notes
- The base ruleset declares each tunable value as a CUE default. See [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md).

## Related
- [ADR-0002](/decisions/adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md)
