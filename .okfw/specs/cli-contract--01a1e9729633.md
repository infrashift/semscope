---
id: "01a1e972-9633-7718-8eae-df9a2cd538ba"
title: "CLI contract"
type: "Spec"
tags:
  - "spec"
  - "cli"
created: "2026-11-17T06:19:23-06:00"
timestamp: "2026-11-17T07:07:37-06:00"
description: "The commands, flags, outputs and exit codes of the semscope command line."
status: "stable"
---

# CLI contract

Task 2.2 implements the commands and the flags. Task 4.1 adds the scan engine, `--fail-level` and the text and JSON reports. Task 5.1 adds SARIF 2.1.0 and 2.2 ([ADR-0008](/decisions/adr-0008-go-sarif-writes-the-sarif-reports--01a1e9998cbd.md)). `rules inspect` also lists the exclude patterns. The help text of each command has golden tests in `cmd/semscope/testdata/`.

## Commands
| Command | Effect |
|---|---|
| `semscope --version`, `semscope -v`, `semscope version` | Print the version, the commit, the build date and the SARIF versions. |
| `semscope scan [paths...]` | Scan the paths with the active rules. The default path is `.`. |
| `semscope rules list` | List the built-in rulesets. |
| `semscope rules inspect <ruleset>` | Show the active rules, the omitted rules and the effective parameters. |

## Scan flags
| Flag | Value | Default |
|---|---|---|
| `--ruleset`, `-r` | A built-in ruleset name, a `.cue` file or a directory. Repeatable. | `nasa-p10` |
| `--enable-rule` | Rule IDs, comma-separated | none |
| `--omit-rule`, `--disable-rule` | Rule IDs, comma-separated | none |
| `--set` | `RULE_ID.param=value`. Repeatable. | none |
| `--format` | `sarif`, `json` or `text` | `text` |
| `--sarif-version` | `2.1.0` or `2.2` | `2.1.0` |
| `--output`, `-o` | A file path | stdout |
| `--fail-level` | `error`, `warning`, `note` or `none` | `error` |

## Flag semantics
- The loader unifies all `--ruleset` values in the sequence of the flags. Without `--ruleset`, the loader reads `nasa-p10`.
- To extend `nasa-p10`, give `-r nasa-p10 -r overlay.cue`. A `.cue` file alone forms a complete ruleset.
- `--set` parses the value with the type of the parameter in the schema: int, bool, string, or a comma-separated list.
- An unknown rule ID or parameter gives an error. The CLI does not ignore it.
- SARIF 2.2 is an OASIS draft. The help text says so.

## Text report
Each finding gives one line: `path:line:column: severity: RULE-ID: message`. A finding without a location, for example an assertion average, starts with `-`. The last line counts the findings of each severity.

## Exit codes
| Code | Meaning |
|---|---|
| 0 | No findings at the `error` level. |
| 1 | One or more findings at the `--fail-level` severity or a higher severity. |
| 2 | Bad usage, a bad ruleset, or work that exceeds a bound. |

## Related
- [CUE ruleset schema](/specs/cue-ruleset-schema--01a1e9729645.md)
- [ISSUE-002](/issues/issue-002-phase-2-version-package-and-cobra-cli--01a1e97296d0.md)
