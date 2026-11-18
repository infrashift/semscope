---
title: Roadmap
description: The phases of the semscope implementation plan and their state.
---

`PLAN.md` in the repository is the phase checklist. Each phase has a mirror issue in the project wiki (`.okfw/issues/`).

| Phase | Content | State |
|---|---|---|
| 1 | Repository bootstrap, the gates and the OKF wiki | Complete |
| 2 | The version package and the cobra CLI | Complete |
| 3 | The composable CUE ruleset engine and the `nasa-p10` ruleset | Complete |
| 4 | The bounded analysis engine and the parity tests with the oracles | Complete |
| 5 | SARIF 2.1.0 and 2.2 reports and the self-scan gate | Complete |
| 6 | The `--exclude` flag, the Go check with type information and more parity corpora | Complete |
| 7 | Zig container keys, the Zig 0.17 grammar and the oracle defects | Complete |
| 8 | This documentation site on GitHub Pages | Complete |

## Releases

semscope has no binary release. Build it from source. See [Getting Started](../../getting-started/).

## Known limits

- The NASA-P10-03 check uses call names only, so its default severity is `warning`.
- The Go check with type information of NASA-P10-07 needs the `go` command and a `go.mod`. Its default is off.
- The Zig grammar is a patched copy of `tree-sitter-zig` v1.1.2, because no release reads Zig 0.17.
- The corpus parity test has one recorded divergence: an interface method in `golang.org/x/tools`.
