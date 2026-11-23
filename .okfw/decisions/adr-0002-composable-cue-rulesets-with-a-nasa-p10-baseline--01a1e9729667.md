---
x-okfw-id: "01a1e972-9667-7a01-8e28-5a934c609607"
title: "ADR-0002: Composable CUE rulesets with a nasa-p10 baseline"
type: "ADR"
tags:
  - "adr"
  - "ruleset"
  - "cue"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T06:19:23-06:00"
description: "Rulesets are CUE values that users extend with overlays, omissions and parameter overrides; nasa-p10 is the first built-in ruleset."
---

# ADR-0002: Composable CUE rulesets with a nasa-p10 baseline

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request (`PLAN.md`, Phase 3).

## Context
semscope needs rules that users can add, omit and tune without a change to the Go code. The first ruleset is the NASA Power of 10. zmake checks these rules with `nasa-lint.py`.

## Decision
- A ruleset is a CUE value. [CUE ruleset schema](/specs/cue-ruleset-schema--01a1e9729645.md) defines it.
- The semscope binary embeds the schema and the `nasa-p10` ruleset.
- A user extends a ruleset with overlays: CUE files that the loader unifies with the base ruleset.
- The CLI omits rules with `--omit-rule` and tunes parameters with `--set`.
- A rule names one check kind. The kinds `ast_query` and `token_ban` let a user write a new code rule in CUE alone.
- New dependency: `cuelang.org/go` (v0.15).

## Consequences
- CUE unification gives type checks and composition without custom merge code.
- The base ruleset declares each tunable value as a CUE default. Otherwise an overlay value conflicts with the base value.
- An unknown rule ID or parameter gives an error, so a typo does not disable a check without notice.
- The binary grows by the size of the CUE evaluator.

## Alternatives considered
- **YAML or JSON rulesets.** Rejected: no schema, no defaults and no unification. semscope then needs custom merge rules for each field.
- **Rules as Go plugins.** Rejected: a user cannot add a rule without a Go build. Go plugins need the same toolchain as the host.
- **Rego (OPA).** Rejected: Rego queries data, not syntax trees. It also has no typed parameters with defaults.

## Related
- [CLI contract](/specs/cli-contract--01a1e9729633.md)
- [ADR-0003](/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md)
