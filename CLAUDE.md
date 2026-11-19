# semscope

semscope scans source code with composable CUE rulesets. `PLAN.md` is the phase checklist.

- Read the [Development workflow](.okfw/development-workflow--01a1e97295ed.md) before you start a task.
- Run `zmake check` and `zmake -f Makefile check` after each task. Commit only when both gates pass.
- Go code obeys the Power of 10 rules of [ADR-0003](.okfw/decisions/adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md). Use `assert.That` for programmer errors only.
- Write all prose in ASD-STE100. `scripts/ste-check.py` checks it.
- `zmake docs-lexiscope` checks the Markdown files with lexiscope and `vocabulary.cue`, when lexiscope is on the PATH ([ADR-0015](.okfw/decisions/adr-0015-lexis-checks-the-documentation-with-a-repository-vocabulary--01a1f5263500.md)). Each finding fails it.
- Add a term to `vocabulary.cue` only when it names a concept of the project, a tool or a standard. Rewrite other prose.

<!-- BEGIN okfw -->
## Project knowledge base (okfw)

semscope keeps a local-first OKF wiki in `.okfw/`: markdown concepts with
YAML frontmatter, linked to each other. The files are the whole database —
search and the link graph are computed from them on demand, so there is no
index to build and nothing to sync. Each directory's `index.md` is generated
and gives a cheap overview before you open anything (OKF section 8).

- **Search before you derive.** `okfw search "<query>"` / `okfw_search`. The
  wiki holds what the code does not say — why things are built the way they
  are, what was considered and rejected, dependency and security findings,
  operational context, open work. Check it before explaining a design,
  assessing risk, or running an analysis someone may already have recorded and
  adjudicated.
- **Browse top-down.** Each directory's `index.md` lists its concepts with
  their descriptions (`okfw_read {"ref":"<dir>/"}` over MCP). Open only what fits.
- **Search before you write**, too, so a new concept links to what already
  exists instead of duplicating it.
- **Link actively** — `[Title](/dir/file.md)` from the wiki root, or
  `./file.md`. Links are what makes the wiki more than a folder.
- **One idea per concept**, with a one-sentence `description`.
- **Use `okfw move` / `okfw delete`, never `mv` / `rm`.** `move` repairs every
  link to the concept and its own relative links, and both keep the `index.md`
  listings current; the shell commands break the links and leave the listings
  stale.
- **`status` is the OKF lifecycle field** (`draft`/`stable`/`deprecated`), not
  task state. Mark superseded knowledge rather than deleting it.
- **Custom metadata goes under `x:`** — the schema is closed and rejects any
  other unrecognized key.
- Don't edit `index.md` or `log.md`; okfw generates them.

Results are capped and say so: `truncated`, or `… N more`, means you did not
see everything — raise the limit or page with `--offset`/`"offset"`.

`okfw mcp` serves all of this as MCP stdio tools named `okfw_*`.
<!-- END okfw -->
