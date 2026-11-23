---
x-okfw-id: "01a1e972-9656-7537-bfdd-2fa979738ef3"
title: "ADR-0001: Adopt an OKF wiki (okfw) for project knowledge"
type: "ADR"
tags:
  - "adr"
  - "wiki"
x-okfw-created: "2026-11-17T06:19:23-06:00"
x-okfw-timestamp: "2026-11-17T06:19:23-06:00"
description: "semscope records decisions, specs and open work in a local-first OKF wiki in .okfw/."
---

# ADR-0001: Adopt an OKF wiki (okfw) for project knowledge

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request. zmake records the same decision in its own ADR-0001.

## Context
The code does not record why semscope works the way it does. Decisions, rejected options, oracle quirks and open work need a durable home that agents and people can search.

## Decision
semscope keeps a local-first OKF wiki in `.okfw/`. The `okfw` tool (v2026.10.2) manages it. The layout copies the zmake wiki. [Wiki conventions](/wiki-conventions--01a1e97256ee.md) defines the layout and the style rules.

## Consequences
- Agents search the wiki before they derive an answer. The okfw MCP server in `.mcp.json` and the `okfw` CLI give the same operations.
- `zmake wiki` checks the frontmatter, the OKF format and the links on each task.
- Each phase has a mirror issue. `PLAN.md` stays the checklist.

## Alternatives considered
- **A `docs/` folder of plain Markdown.** Rejected: no schema, no link checks and no search.
- **GitHub issues and the GitHub wiki.** Rejected: they live outside the repository and need network access.
