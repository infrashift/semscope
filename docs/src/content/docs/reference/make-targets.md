---
title: Make Targets
description: The ZMakefile and Makefile tasks that build, test, lint and document semscope.
---

The semscope repository has two build files with the same tasks:

- `ZMakefile` holds the tasks in zmake syntax, with `@name:` tasks. `zmake check` reads it, because zmake finds a `ZMakefile` first.
- `Makefile` holds the same tasks in plain GNU Make syntax. `zmake -f Makefile check` and `make check` read it.

The `ZMakefile` declares each target as a task, with no `.PHONY` list:

```make
## check: run every gate
@check: lint test wiki docs-style docs-lexiscope selfscan

## docs-dev: run the Astro Starlight site locally (bun)
@docs-dev:
    cd docs && $(BUN) install && $(BUN) --bun run dev
```

| Target | Command |
|---|---|
| `zmake ztree` | `zig build lib --release` in `ZTREE_DIR` (default `../ztree`). `build`, `test`, `lint`, `selfscan` and `parity` run it first. |
| `zmake build` | `go build` → `bin/semscope`, with the version, the commit and the build date |
| `zmake install` | `build`, then copy `bin/semscope` to `$(PREFIX)/bin`. The default `PREFIX` is `~/.local`. |
| `zmake test` | `go test -race ./...` |
| `zmake lint` | `gofmt -l`, `nasa-lint.py`, `nasa-lint-go.py` and its fixture tests, `go vet` and `golangci-lint run` |
| `zmake fmt` | `gofmt -w` on the Go sources |
| `zmake docs-style` | `scripts/ste-check.py --strict`: checks the documentation against ASD-STE100 rules |
| `zmake docs-lexiscope` | `lexiscope scan` with `vocabulary.cue` on the Markdown files. Each finding fails it. It passes with a message when lexiscope is not on the PATH. |
| `zmake wiki` | `okfw validate --all` and `okfw validate --okf` on the project wiki |
| `zmake selfscan` | semscope scans its own sources with `nasa-p10` and the overlay `.semscope.cue` |
| `zmake check` | `lint`, `test`, `wiki`, `docs-style`, `docs-lexiscope` and `selfscan` |
| `zmake sarif` | Write the SARIF 2.1.0 and 2.2 reports of the repository |
| `zmake parity` | Compare semscope with the oracles on the Zig and Go corpora |
| `zmake docs-dev` | Serve this site with live reload (bun) |
| `zmake docs-dev-up` | Start the same server in the background (`astro dev --background`) |
| `zmake docs-dev-down` | Stop the server of `docs-dev` or `docs-dev-up` (`astro dev stop`) |
| `zmake docs-build` | Build this site into `docs/dist`, the same build as the Pages deploy |

`check` does not build the site, so the Go gates do not need bun.
`docs-style` and `docs-lexiscope` check the pages of this site, so `check` finds a style error in them.

Before a commit, run `zmake check` and `zmake -f Makefile check`. When you change the site, also run `zmake docs-build`.
Fix each failure and run the gates again until all pass.

## Variables

| Variable | Default | Use |
|---|---|---|
| `GO` | `go` | The Go command |
| `PYTHON` | `python3` | The Python interpreter for the scripts |
| `BUN` | `bun` | The bun command for the site |
| `PREFIX` | `$(HOME)/.local` | The install prefix |
| `GOLANGCI_LINT` | `$(go env GOPATH)/bin/golangci-lint` | The golangci-lint binary |
| `LEXISCOPE` | `lexiscope` | The [lexiscope](https://infrashift.github.io/lexiscope/) binary of `docs-lexiscope`. `lexis` is a short alias of lexiscope. |

Set a variable on the command line, for example `zmake install PREFIX=/usr/local`.
