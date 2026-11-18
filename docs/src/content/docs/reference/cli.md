---
title: CLI
description: The commands, flags, reports and exit codes of the semscope command line.
---

## Commands

| Command | Effect |
|---|---|
| `semscope scan [paths...]` | Scan the paths with the active rules. The default path is `.`. |
| `semscope rules list` | List the built-in rulesets. |
| `semscope rules inspect <ruleset> [overlay...]` | Show the sources, the exclude map, the active rules, the omitted rules and the effective parameters. |
| `semscope version`, `semscope --version`, `semscope -v` | Print the version, the commit, the build date, the Go version and the SARIF versions. |
| `semscope help [command]`, `semscope [command] --help` | Print the help text of a command. |

## `scan` flags

| Flag | Value | Default |
|---|---|---|
| `--ruleset`, `-r` | A built-in ruleset name, a `.cue` file or a directory. Repeat the flag to add overlays. | `nasa-p10` |
| `--enable-rule` | Rule IDs, comma-separated | none |
| `--omit-rule`, `--disable-rule` | Rule IDs, comma-separated | none |
| `--set` | `RULE_ID.param=value`. A list value is comma-separated. Repeatable. | none |
| `--exclude` | A `path.Match` pattern relative to the scan root. Repeatable. | none |
| `--format` | `text`, `json` or `sarif` | `text` |
| `--sarif-version` | `2.1.0` or `2.2` (an OASIS draft) | `2.1.0` |
| `--output`, `-o` | A file path | stdout |
| `--fail-level` | `error`, `warning`, `note` or `none` | `error` |

## `rules inspect` flags

`rules inspect` takes `--enable-rule`, `--omit-rule`, `--disable-rule`, `--set` and `--exclude`. They have the same effect as on `scan`.
The first argument is the base ruleset. Each next argument is an overlay. An argument has the same form as a `--ruleset` value.

```sh
semscope rules inspect nasa-p10
semscope rules inspect nasa-p10 .semscope.cue --set NASA-P10-04.max_lines=40
```

## Flag semantics

- semscope unifies the `--ruleset` values in the sequence of the flags.
- A built-in name has priority over a path with the same name.
- `--set` parses the value with the type of the parameter: an integer, a boolean, a string or a comma-separated list.
- `--enable-rule` and `--omit-rule` win over the `enabled` field of each source.
- An unknown rule ID or an unknown parameter gives an error and exit status 2. semscope does not ignore it.
- A comma in a `--ruleset` or `--set` value does not split the value. A comma in `--enable-rule` or `--omit-rule` separates two IDs.
- `--exclude` adds a pattern with the reason `--exclude flag`. When a ruleset also excludes the pattern, the reason of the ruleset stays.

## Reports

| Format | Content |
|---|---|
| `text` | One line for each finding: `path:line:column: severity: RULE-ID: message`, then a summary line |
| `json` | `tool`, `version`, `summary`, `assertions` and `findings` |
| `sarif` | One SARIF run. The driver lists the active rules. semscope validates the report against the schema before it writes it. |

[Output Formats](../../guides/output-formats/) gives the fields of each report.

## Exit status

| Code | Meaning |
|---|---|
| 0 | No finding has the `--fail-level` severity or a higher severity. |
| 1 | One or more findings have the `--fail-level` severity or a higher severity. |
| 2 | Bad usage, a bad ruleset, a bad pattern, or work that exceeds a limit. |

## Version output

```console
$ semscope version
semscope 5bdd6ba
commit: 5bdd6ba8b4052a2e42063578e45531a97096c7b3
build date: 2026-11-18T05:52:54Z
go: go1.27.3
sarif: 2.1.0, 2.2
```

The `build` task sets the version from `git describe` through `-ldflags`.
Without these values, semscope reads the module and VCS data of the Go build info. A field without a value gives `dev` or `unknown`.
