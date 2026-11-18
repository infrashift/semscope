# semscope

semscope scans source code with composable CUE rulesets. The built-in ruleset `nasa-p10` holds the NASA/JPL Power of 10 rules for safety-critical code. semscope reads Go, C, C++, Zig, Python and Rust. It writes the findings as text, JSON, SARIF 2.1.0 or SARIF 2.2.

## Build

semscope needs Go 1.26 or later and a C compiler, because cgo builds the Tree-sitter grammars. The gates also need Python 3, golangci-lint v2, okfw and zmake.

1. Run `zmake build`.
2. Run `./bin/semscope --version`.
3. Run `zmake install` to copy the binary to `~/.local/bin/semscope`. Set `PREFIX` to use another directory.

## Scan

```sh
semscope scan .
semscope scan --format json -o findings.json src
semscope scan --format sarif -o report.sarif src
semscope scan --format sarif --sarif-version 2.2 -o report-v22.sarif src
semscope scan --fail-level warning cmd internal
```

| Exit status | Meaning |
|---|---|
| 0 | No finding has the `--fail-level` severity or a higher severity. |
| 1 | One or more findings have the `--fail-level` severity or a higher severity. |
| 2 | Bad usage, a bad ruleset, or work that exceeds a limit. |

The text report gives one line for each finding: `path:line:column: severity: RULE-ID: message`. SARIF 2.2 is an OASIS draft. The default SARIF version is 2.1.0.

## Rulesets

A ruleset is a CUE value. The flag `--ruleset` (`-r`) takes a built-in name, a `.cue` file or a directory. Repeat the flag to add overlays. The loader unifies the values in sequence.

```cue
// overlay.cue
package overlay

rules: "NASA-P10-04": params: max_lines: 40
rules: "NASA-P10-05": enabled: false
rules: "CUSTOM-01": {
	languages: ["go"]
	title:     "No fmt.Println"
	message:   "Write output through the report package."
	check: {kind: "token_ban", patterns: go: ["\\bfmt\\.Println\\s*\\("]}
}
exclude: vendor: "third-party code"
```

```sh
semscope scan -r nasa-p10 -r overlay.cue .
semscope scan --omit-rule NASA-P10-05 --set NASA-P10-04.max_lines=40 .
semscope scan --exclude vendor --exclude 'gen/*.go' .
semscope rules list
semscope rules inspect nasa-p10 overlay.cue
```

- An overlay can add a rule, omit a rule, change a parameter, narrow the languages of a rule and exclude paths.
- An unknown rule ID or parameter in a flag gives an error.
- `--exclude` adds a path pattern to the exclude map. The pattern has the `path.Match` syntax and is relative to the scan root.
- An `ast_query` rule runs one Tree-sitter query for each language. Each `@violation` capture gives a finding.
- A `token_ban` rule bans regular expressions in the source without comments and strings.
- With these two check kinds, a user adds a code rule in CUE without a change to the Go code.

## The nasa-p10 ruleset

| Rule | Check |
|---|---|
| NASA-P10-01 | No `goto`, `setjmp` or `longjmp`, and no direct or indirect recursion. |
| NASA-P10-02 | No loop without a bound, for example `for {}`, `while (true)` or `loop {}`. |
| NASA-P10-03 | No allocator calls outside the init functions. The default level is warning. |
| NASA-P10-04 | Functions of `max_lines` (60) lines or fewer. |
| NASA-P10-05 | An average of `min_assertions` (2) or more assertions for each function. |
| NASA-P10-06 | No mutable global state. |
| NASA-P10-07 | No discarded call results and no swallowed errors. Set `go_types` to `true` for a Go check with type information. |
| NASA-P10-08 | No token pasting, variadic macros, deep conditionals or dynamic code. |
| NASA-P10-09 | One level of pointer indirection at most, and no function pointers. |
| NASA-P10-10 | No unsafe constructs and no warning suppressions without a reason. |

The checks of NASA-P10-01, -02, -04, -05, -07 and -09 match `scripts/nasa-lint.py` for Zig and `scripts/nasa-lint-go.py` for Go. The parity tests in `internal/engine` compare them line by line.

The Go check with type information finds each call, `defer` and `go` statement that drops an `error` result. It needs the go command and a `go.mod`. The parameter `go_ignore_calls` names calls that never fail.

```sh
semscope scan --set NASA-P10-07.go_types=true .
```

## Suppressions

Write `// semscope: allow(NASA-P10-07) <reason>` on the line of the finding or on the line above. semscope also reads the tag `nasa-lint: allow(R7)` of the oracle scripts.

## Development

- `zmake check` runs every gate: the oracle scripts, gofmt, go vet, golangci-lint, the race tests, the wiki checks, the STE check and the self-scan.
- `zmake -f Makefile check` runs the same gates from the GNU Make file.
- `zmake parity` compares semscope with the oracles on real code: zmake, the Zig standard library and six Go modules. `scripts/parity_corpora.py` finds them.
- `PLAN.md` is the phase checklist. The OKF wiki in `.okfw/` records the design decisions and the specs.
