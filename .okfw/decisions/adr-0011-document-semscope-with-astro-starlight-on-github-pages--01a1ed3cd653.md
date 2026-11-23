---
x-okfw-id: "01a1ed3c-d653-7875-a819-fa41b951c621"
title: "ADR-0011: Document semscope with Astro Starlight on GitHub Pages"
type: "ADR"
tags:
  - "adr"
  - "docs"
x-okfw-created: "2026-11-17T23:59:09-06:00"
x-okfw-timestamp: "2026-11-17T23:59:09-06:00"
description: "An Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by GitHub Actions; the docs toolchain stays outside the Go build and the check gate."
---

# ADR-0011: Document semscope with Astro Starlight on GitHub Pages

**Status:** Accepted
**Date:** 2026-11-17
**Source:** User request, Phase 8, [ISSUE-010](/issues/issue-010-phase-8-documentation-site-on-github-pages--01a1ed3cf4ce.md)

## Context
The README and the wiki describe semscope for its developers. Users need a public site with guides and a reference. The sibling repository zmake publishes an Astro Starlight site to GitHub Pages (zmake ADR-0009). It uses bun 1.3.9 and a GitHub Actions workflow.

## Decision
- The site lives in `docs/`. It uses Astro 7, Starlight 0.41 and bun 1.3.9, the same versions as the zmake site. `docs/bun.lock` pins the packages.
- The docs toolchain is not a dependency of the semscope binary. The Go build does not read `docs/`. `docs/node_modules` holds no Go files, so `go vet ./...` and golangci-lint do not read it.
- The landing page is the Starlight `splash` template in `docs/src/content/docs/index.mdx`. The site has no custom landing page and no image assets. The Starlight pages live at the site root.
- The `docs-dev` task serves the site with live reload. The `docs-build` task builds it with `--frozen-lockfile`, the same build as the Pages deploy. The `ZMakefile` and the `Makefile` hold both tasks ([ADR-0006](/decisions/adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md)).
- `.github/workflows/docs-release.yml` builds the site with `withastro/action` and deploys it with `actions/deploy-pages`. A pull request only builds the site.
- The Pages source of the repository is "GitHub Actions".
- `scripts/ste-check.py` already reads `docs/src/content/docs`, so the `docs-style` gate checks the pages against ASD-STE100.
- `zmake check` does not build the site, so the Go gates do not need bun.

## Consequences
- Each phase updates the docs pages. The [Development workflow](/development-workflow--01a1e97295ed.md) has a step for it.
- The site is at https://infrashift.github.io/semscope/. The Astro `base` is `/semscope`, so a link in a page is relative or includes the base path.
- The examples on the pages come from real runs of semscope. A change to a report format needs a change to the pages.
- The self-scan does not read `docs/`, because semscope skips `node_modules` and the site has no source files of the six languages.

## Alternatives considered
- **A custom landing page, the same as zmake.** Rejected by the user: it needs about 600 lines of CSS and a logo, and the splash template gives the same links.
- **GitHub's Jekyll build of `docs/`.** Rejected: no search, no sidebar, and it conflicts with the Actions deploy.
- **Docs pages in the wiki only.** Rejected: the wiki holds knowledge for developers, not a guide for users.
