---
id: "01a1e972-96ad-74b9-987e-04553c6aec1c"
title: "ADR-0006: zmake runs the development gates"
type: "ADR"
tags:
  - "adr"
  - "build"
  - "zmake"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T06:19:23-06:00"
description: "A ZMakefile and a Makefile hold the same gate tasks; zmake runs both before each task commit."
---

# ADR-0006: zmake runs the development gates

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request on 2026-11-17: "use zmake rather than make. Construct Makefile(s) and ZMakefiles."

## Context
semscope needs one command that runs every gate. zmake is the infrashift Make implementation. Its own repository keeps a ZMakefile and a Makefile with the same tasks.

## Decision
- `ZMakefile` holds the gates in zmake modern mode: `@name:` tasks and no `.PHONY` list.
- `Makefile` holds the same tasks in GNU Make syntax.
- `zmake check` runs the ZMakefile. `zmake -f Makefile check` runs the Makefile.
- Each task commit needs both gates green.

## Consequences
- zmake reads the ZMakefile first, also with `--legacy`. The flag `-f Makefile` selects the Makefile.
- GNU Make can also run the Makefile.
- The two files can drift apart. Both gates run on each task, so a drift in behavior shows at once.
- The files call `gofmt` from `$(go env GOROOT)/bin` and `golangci-lint` from `$(go env GOPATH)/bin`. These directories are not on the `PATH` of every shell.

## Alternatives considered
- **GNU Make only.** Rejected: the user requires zmake.
- **A Go task runner, for example mage.** Rejected: a second build language for the same tasks.

## Related
- [Development workflow](/development-workflow--01a1e97295ed.md)
