# issues

<!-- okfw:generated — maintained by okfw; edits here are overwritten -->

## Concepts

* [ISSUE-001: Phase 1: Repository bootstrap, gates and OKF wiki](issue-001-phase-1-repository-bootstrap-gates-and-okf-wiki--01a1e97296be.md) - Bootstrap git, the zmake gates, the Go module, the oracle scripts, the STE gate and the wiki.
* [ISSUE-002: Phase 2: Version package and Cobra CLI](issue-002-phase-2-version-package-and-cobra-cli--01a1e97296d0.md) - Add internal/version and the cobra commands version, scan and rules with every ruleset flag.
* [ISSUE-003: Phase 3: Composable CUE ruleset engine](issue-003-phase-3-composable-cue-ruleset-engine--01a1e97296e2.md) - Add the CUE ruleset schema, the loader with overlays and overrides, and the nasa-p10 ruleset.
* [ISSUE-004: Phase 4: Bounded engine and nasa-p10 parity](issue-004-phase-4-bounded-engine-and-nasa-p10-parity--01a1e97296f3.md) - Add the Tree-sitter parsers, the bounded engine and the parity tests against both oracles.
* [ISSUE-005: Phase 5: SARIF reports and self-hosting gate](issue-005-phase-5-sarif-reports-and-self-hosting-gate--01a1e9729706.md) - Add the SARIF 2.1.0 and 2.2 writers and run the final self-hosting gate.
* [ISSUE-006: Phase 6: Exclude flag, Go type checks and parity corpora](issue-006-phase-6-exclude-flag-go-type-checks-and-parity-corpora--01a1eba38e60.md) - Add the --exclude flag, a type-aware Go check for NASA-P10-07 and more corpora for the parity test.
* [ISSUE-007: Oracle defects found by the corpus parity test](issue-007-oracle-defects-found-by-the-corpus-parity-test--01a1ebb7320f.md) - The corpus parity test of Task 6.3 found defects in nasa-lint.py and nasa-lint-go.py; this issue lists each class and its state.
* [ISSUE-008: Zig limits of semscope found by the corpus parity test](issue-008-zig-limits-of-semscope-found-by-the-corpus-parity-test--01a1ebb73216.md) - The Zig grammar v1.1.2 lags Zig 0.17, and the Zig call graph gives same-name functions of one file one node.
* [ISSUE-009: The Zig grammar does not read asm clobbers of Zig 0.15](issue-009-the-zig-grammar-does-not-read-asm-clobbers-of-zig-0-15--01a1ebef12f0.md) - Zig 0.15 writes asm clobbers as a struct literal; tree-sitter-zig reads strings only, so 26 syscall functions of the Zig standard library are lost.
* [ISSUE-010: Phase 8: Documentation site on GitHub Pages](issue-010-phase-8-documentation-site-on-github-pages--01a1ed3cf4ce.md) - Add an Astro Starlight documentation site in docs/, docs-dev and docs-build tasks, and a GitHub Actions workflow that deploys the site to GitHub Pages.
* [ISSUE-011: Phase 9: the Zig grammar of ztree](issue-011-phase-9-the-zig-grammar-of-ztree--01a1f3fdc1c1.md) - semscope takes the Zig grammar of ztree, so that both give the same trees for empty containers and three rare Zig forms.

