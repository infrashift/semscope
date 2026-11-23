---
x-okfw-id: "01a1e999-8cbd-7baf-9ea8-33bd9a65e301"
title: "ADR-0008: go-sarif writes the SARIF reports"
type: "ADR"
tags:
  - "adr"
  - "report"
  - "sarif"
  - "dependency"
x-okfw-created: "2026-11-17T07:01:56-06:00"
x-okfw-timestamp: "2026-11-17T07:01:56-06:00"
description: "semscope writes SARIF 2.1.0 and SARIF 2.2 through github.com/owenrumney/go-sarif/v3 and validates each report against the schema before it writes it."
---

# ADR-0008: go-sarif writes the SARIF reports

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request (`PLAN.md` Task 5.1).

## Context
Code scanning services read SARIF. `PLAN.md` asks for SARIF 2.1.0 and for SARIF 2.2, selected by `--sarif-version`, and for schema validation before each write.

## Decision
- Use `github.com/owenrumney/go-sarif/v3` (v3.3.2): the package `pkg/report/v210/sarif` for 2.1.0 and `pkg/report/v22/sarif` for 2.2.
- `internal/report` builds one run description that does not depend on the version. Two small functions map it to the types of each package.
- `Report.Validate()` checks each report against its JSON schema before the writer writes it. A report that fails the check gives an error and no output.
- The driver lists the active rules only, sorted by ID. Each result gives `ruleId`, `ruleIndex`, `level`, a message and a location.
- A location has a repository-relative `uri` with `uriBaseId` `%SRCROOT%`, and a 1-based region. A finding without a location, for example an assertion average, gives a result without locations.
- The report has no absolute paths, so the same scan gives the same report on each machine.

## Consequences
- SARIF 2.2 is an OASIS draft. The help text of `--sarif-version` says so. The default is 2.1.0.
- The library uses `reflect` inside. The `depguard` rule of ADR-0003 applies to semscope code only.
- The two version packages have separate types, so the mapping code exists twice. Each copy is short.

## Alternatives considered
- **Write the JSON by hand.** Rejected: no schema types and no validation.
- **Only SARIF 2.1.0.** Rejected: the plan asks for 2.2 too.
