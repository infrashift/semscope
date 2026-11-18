---
id: "01a1ed3d-dbbd-7bfb-8d74-febc70f3c1e9"
title: "2026-11-18 Wednesday"
type: "daily-note"
kind: "daily-note"
tags:
  - "daily"
created: "2026-11-18T00:00:16-06:00"
timestamp: "2026-11-18T00:00:16-06:00"
---

# 2026-11-18 Wednesday

## Changed
- Opened Phase 8 and ISSUE-010. Added the Astro Starlight site in `docs/`, with the zmake setup and a splash landing page (Task 8.1, ADR-0011).
- Added the tasks `docs-dev` and `docs-build` to the `ZMakefile` and the `Makefile`, with the variable `BUN`.
- Wrote 13 pages: the guides, the reference and the project pages. Each example output comes from a real run of semscope (Task 8.2).
- Added `.github/workflows/docs-release.yml`. A pull request builds the site. A push to `main` deploys it (Task 8.3).
- Added a step for the docs site to the Development workflow. Closed ISSUE-010.

## Learned
- semscope does not parse a test file. A Go `_test.go` file and a Python `test_*.py` or `*_test.py` file give no findings.
- An exclude pattern matches a path or each parent directory of the path. The `*` wildcard does not match a `/`.
- The `docs-style` gate already read `docs/src/content/docs`, because the zmake copy of `ste-check.py` had the path.
- The Starlight build prints two warnings for the `i18n` collection and the 404 entry. The zmake build prints the same warnings.

## Next
- Set the Pages source of the GitHub repository to "GitHub Actions". Then push to `main` for the first deploy.

## Links Created
- [ADR-0011](/decisions/adr-0011-document-semscope-with-astro-starlight-on-github-pages--01a1ed3cd653.md)
- [ISSUE-010](/issues/issue-010-phase-8-documentation-site-on-github-pages--01a1ed3cf4ce.md)
