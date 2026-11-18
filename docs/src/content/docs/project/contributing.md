---
title: Contributing
description: The development loop, the gates and the style rules for a change to semscope.
---

## Tools

- Go 1.26 or later and a C compiler.
- Python 3 for the oracle scripts and the style check.
- golangci-lint v2, okfw and zmake.
- bun 1.3.9 for this site.

## The development loop

Each change follows this loop. Do not commit work while a gate fails.

1. Read `PLAN.md`. Take the first unchecked task.
2. Search the project wiki for the ADRs and specs of the task: `okfw search "<query>"`.
3. Implement only that task.
4. Apply the Power of 10 rules of ADR-0003.
5. Update the pages of this site for the features of the task.
6. Run `zmake check`. Then run `zmake -f Makefile check`.
7. Fix each failure and run both gates again.
8. When the site changes, run `zmake docs-build`.
9. Write an ADR for each decision. File each bug or design debt as an issue.
10. Tick the `PLAN.md` checkbox. Add an entry to the `okfw daily` note.
11. Make one commit for the task, with the code, the wiki and `PLAN.md`.

## Gates

| Gate | Command | Checks |
|---|---|---|
| lint | `gofmt -l`, `nasa-lint.py`, `nasa-lint-go.py`, `go vet`, `golangci-lint run` | Format, Power of 10 rules, static analyzers |
| test | `go test -race ./...` | Unit tests with the race detector |
| wiki | `okfw validate --all`, `okfw validate --okf` | Frontmatter schema, OKF conformance, links |
| docs-style | `scripts/ste-check.py --strict` | ASD-STE100 style |
| selfscan | `semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning .` | semscope on its own sources |

## Power of 10 rules for the Go code

- Do not use recursion. Walk a tree with a Tree-sitter cursor or an explicit bounded stack.
- Give each loop a fixed bound. A loop that reaches its limit returns an error.
- Keep each function to 60 lines or fewer.
- Use an average of two or more `assert.That` calls for each function. Use `assert.That` for programmer errors only.
- Do not add mutable package variables, function types or `unsafe`.

## Documentation style

All Markdown files, wiki pages, help strings and code comments follow ASD-STE100 Simplified Technical English. `scripts/ste-check.py` checks them.

- Keep each procedural step to 20 words or fewer. Keep each descriptive sentence to 25 words or fewer.
- Use active voice. Name the actor: "The parser reads the source file."
- Do not use marketing words, for example `seamless`, `robust` or `powerful`.
- Do not use the words `should`, `will`, `would`, `please`, `simply` or `just`.
- Do not use the possessive `your`. Write `the ruleset file`, not `your ruleset file`.

## This site

The site lives in `docs/`. The pages are Markdown files in `docs/src/content/docs/`.

```sh
zmake docs-dev     # serve the site with live reload at http://localhost:4321/semscope/
zmake docs-build   # build the site into docs/dist
```

The sidebar is in `docs/astro.config.mjs`. Add each new page to it.
A push to `main` that changes `docs/` deploys the site. A pull request only builds it.
