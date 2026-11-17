# decisions

<!-- okfw:generated — maintained by okfw; edits here are overwritten -->

## Concepts

* [ADR-0001: Adopt an OKF wiki (okfw) for project knowledge](adr-0001-adopt-an-okf-wiki-okfw-for-project-knowledge--01a1e9729656.md) - semscope records decisions, specs and open work in a local-first OKF wiki in .okfw/.
* [ADR-0002: Composable CUE rulesets with a nasa-p10 baseline](adr-0002-composable-cue-rulesets-with-a-nasa-p10-baseline--01a1e9729667.md) - Rulesets are CUE values that users extend with overlays, omissions and parameter overrides; nasa-p10 is the first built-in ruleset.
* [ADR-0003: Adapt the NASA Power of 10 rules to Go](adr-0003-adapt-the-nasa-power-of-10-rules-to-go--01a1e9729678.md) - semscope follows the Power of 10 rules where Go allows; nasa-lint-go.py, golangci-lint and the self-scan check them, with a recorded R3 deviation.
* [ADR-0004: Tree-sitter parses all six languages](adr-0004-tree-sitter-parses-all-six-languages--01a1e9729689.md) - internal/parser uses the cgo Tree-sitter bindings for Go, C, C++, Zig, Python and Rust, with one adapter for each language.
* [ADR-0005: Two oracles define nasa-p10 parity](adr-0005-two-oracles-define-nasa-p10-parity--01a1e972969b.md) - nasa-lint.py (Zig) and nasa-lint-go.py (Go) are the parity oracles; fixtures match exactly and corpus divergences need a recorded reason.
* [ADR-0006: zmake runs the development gates](adr-0006-zmake-runs-the-development-gates--01a1e97296ad.md) - A ZMakefile and a Makefile hold the same gate tasks; zmake runs both before each task commit.
* [ADR-0007: cobra parses the command line](adr-0007-cobra-parses-the-command-line--01a1e97fa733.md) - semscope uses github.com/spf13/cobra for its commands and flags; command callbacks are function literals that ADR-0003 allows.

