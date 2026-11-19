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
* [ADR-0008: go-sarif writes the SARIF reports](adr-0008-go-sarif-writes-the-sarif-reports--01a1e9998cbd.md) - semscope writes SARIF 2.1.0 and SARIF 2.2 through github.com/owenrumney/go-sarif/v3 and validates each report against the schema before it writes it.
* [ADR-0009: Type information for the Go check of NASA-P10-07](adr-0009-type-information-for-the-go-check-of-nasa-p10-07--01a1eba86090.md) - When go_types is true, semscope loads the Go packages with golang.org/x/tools/go/packages and reports each call, defer and go statement that drops an error result.
* [ADR-0010: A patched Zig grammar for Zig 0.17](adr-0010-a-patched-zig-grammar-for-zig-0-17--01a1ebe05590.md) - semscope keeps a patched copy of tree-sitter-zig v1.1.2 without the keywords async, await and usingnamespace, because no release reads Zig 0.17.
* [ADR-0011: Document semscope with Astro Starlight on GitHub Pages](adr-0011-document-semscope-with-astro-starlight-on-github-pages--01a1ed3cd653.md) - An Astro Starlight site in docs/, built with bun and deployed to GitHub Pages by GitHub Actions; the docs toolchain stays outside the Go build and the check gate.
* [ADR-0012: The Zig grammar of ztree](adr-0012-the-zig-grammar-of-ztree--01a1f3fda243.md) - semscope takes the patched Zig grammar of ztree, which adds empty containers and three rare forms to the patch of ADR-0010.
* [ADR-0013: semscope uses the ztree runtime](adr-0013-semscope-uses-the-ztree-runtime--01a1f4bf5f46.md) - internal/parser and internal/engine use the ztree Go binding in place of go-tree-sitter, and the gates build libztree.a first.
* [ADR-0014: The Zig grammar lives in ztree only](adr-0014-the-zig-grammar-lives-in-ztree-only--01a1f4d211fc.md) - semscope removes its copy of the patched Zig grammar, because the parser takes all grammars from ztree since ADR-0013.

