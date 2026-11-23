---
x-okfw-id: "01a1e97f-a733-70b6-ae9b-2e9ca6be9ec0"
title: "ADR-0007: cobra parses the command line"
type: "ADR"
tags:
  - "adr"
  - "cli"
  - "dependency"
x-okfw-created: "2026-11-17T06:33:39-06:00"
x-okfw-timestamp: "2026-11-17T06:33:39-06:00"
description: "semscope uses github.com/spf13/cobra for its commands and flags; command callbacks are function literals that ADR-0003 allows."
---

# ADR-0007: cobra parses the command line

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request (`PLAN.md` Task 2.2).

## Context
semscope has the commands `scan`, `rules list`, `rules inspect` and `version`, with repeatable flags, flag aliases and a `--version` flag. The [CLI contract](/specs/cli-contract--01a1e9729633.md) lists them.

## Decision
- Use `github.com/spf13/cobra` (v1.10.2) and its flag package `spf13/pflag`.
- `--ruleset` and `--set` use `StringArray`, so a comma in a path or in a list value does not split the value.
- `--enable-rule`, `--omit-rule` and `--disable-rule` use `StringSlice`, so one flag can hold a comma-separated list.
- Each command sets `SilenceUsage` and `SilenceErrors`. The function `run` prints each error once and maps it to an exit code.

## Consequences
- Each command has a `RunE` callback, a function literal that cobra requires. ADR-0003 allows a function literal as an argument or a field value of a library type.
- cobra adds the `-v` shorthand for `--version` when no other flag uses `v`.
- The default completion command is off.

## Alternatives considered
- **The standard `flag` package.** Rejected: no subcommands, no flag aliases and no repeatable string flags.
- **`urfave/cli`.** Rejected: cobra is the more common choice in Go tools, and the plan names it.
