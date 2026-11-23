---
x-okfw-id: "01a1ed3c-f4ce-7b3a-85a8-4527064501bd"
title: "ISSUE-010: Phase 8: Documentation site on GitHub Pages"
type: "Issue"
x-okfw-kind: "issue"
tags:
  - "issue"
  - "phase-8"
  - "docs"
  - "closed"
x-okfw-created: "2026-11-17T23:59:17-06:00"
x-okfw-timestamp: "2026-11-18T00:00:07-06:00"
description: "Add an Astro Starlight documentation site in docs/, docs-dev and docs-build tasks, and a GitHub Actions workflow that deploys the site to GitHub Pages."
---

# ISSUE-010: Phase 8: Documentation site on GitHub Pages

**Status/Phase:** closed · Phase 8

## Description
semscope has no public documentation for users. This phase adds an Astro Starlight site in `docs/`, the same setup as the zmake site. The `docs-dev` and `docs-build` tasks build it with bun. A GitHub Actions workflow deploys it to GitHub Pages.

## Acceptance Criteria
- [x] `docs/` holds an Astro Starlight site with `base` `/semscope`. `docs/bun.lock` pins the packages.
- [x] The `ZMakefile` and the `Makefile` have the tasks `docs-dev` and `docs-build`. Both files give the same commands.
- [x] The site has a splash landing page, guides, a reference and project pages. The examples come from real runs of semscope.
- [x] `zmake docs-build` builds the site with no broken internal link.
- [x] `.github/workflows/docs-release.yml` builds the site on a pull request and deploys it on a push to `main`.
- [x] The `docs-style` gate checks the pages. Both gates pass.

## Notes
- The repository Pages source must be "GitHub Actions". A person with admin access sets it once in Settings → Pages.
- `zmake check` does not build the site, so the Go gates do not need bun.

## Resolution
Tasks 8.1, 8.2 and 8.3 are complete.

- `docs/` holds the site: a splash landing page and 13 pages. bun resolved Astro 7.4.12, Starlight 0.41.11 and sharp 0.35.5.
- `zmake docs-build` builds 15 HTML files. A link check of the build found no broken internal link or anchor.
- `zmake docs-dev` serves the site at http://localhost:4321/semscope/.
- The `ZMakefile` and the `Makefile` give the same commands for `docs-dev` and `docs-build`.
- `docs-release.yml` copies the zmake workflow. The first deploy needs the Pages source "GitHub Actions" and a push to `main`.
- The pages give two facts that the README does not give. semscope does not check a test file. A scan skips a fixed list of directories.

## Related
- [ADR-0011](/decisions/adr-0011-document-semscope-with-astro-starlight-on-github-pages--01a1ed3cd653.md)
- [ADR-0006](/decisions/adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md)
