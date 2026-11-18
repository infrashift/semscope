---
title: Output Formats
description: The text, JSON and SARIF reports of semscope, and the fields of each.
---

The flag `--format` selects the report: `text` (the default), `json` or `sarif`.
The flag `--output` (`-o`) writes the report to a file. Without it, semscope writes to stdout.

```sh
semscope scan .
semscope scan --format json -o findings.json src
semscope scan --format sarif -o report.sarif src
semscope scan --format sarif --sarif-version 2.2 -o report-v22.sarif src
```

Each report gives paths relative to the scan root. A report has no absolute paths, so the same scan gives the same report on each machine.

## Text

Each finding gives one line:

```text
path:line:column: severity: RULE-ID: message
```

A finding without a location starts with `-`. An example is the assertion average of NASA-P10-05, because it applies to all the functions of one language.
The last line counts the findings of each severity:

```text
-: error: NASA-P10-05: go: 0 assertions in 2 functions (average 0.00, min 2)
main.go:5:5: error: NASA-P10-06: package variable `counter`
main.go:7:1: error: NASA-P10-01: recursion: walk -> walk
main.go:15:2: error: NASA-P10-02: `for` loop without a condition
semscope: 4 finding(s) in 1 file(s): 4 error(s), 0 warning(s), 0 note(s)
```

## JSON

The JSON report holds the tool version, a summary, the assertion counts of each language and the findings:

```json
{
  "tool": "semscope",
  "version": "5bdd6ba",
  "summary": { "files": 1, "errors": 4, "warnings": 0, "notes": 0 },
  "assertions": {
    "go": { "functions": 2, "assertions": 0 }
  },
  "findings": [
    {
      "ruleId": "NASA-P10-05",
      "severity": "error",
      "message": "go: 0 assertions in 2 functions (average 0.00, min 2)"
    },
    {
      "ruleId": "NASA-P10-06",
      "severity": "error",
      "message": "package variable `counter`",
      "path": "main.go",
      "line": 5,
      "column": 5,
      "endLine": 5,
      "endColumn": 16
    }
  ]
}
```

A finding without a location has no `path`, `line` or `column` field.

## SARIF

semscope writes SARIF 2.1.0 by default. The flag `--sarif-version 2.2` selects SARIF 2.2.
SARIF 2.2 is an OASIS draft. GitHub code scanning reads SARIF 2.1.0.

semscope validates each SARIF report against its JSON schema before it writes the report. A report that fails the check gives an error and no output.

- The driver lists the active rules only, sorted by ID.
- Each result gives `ruleId`, `ruleIndex`, `level`, a message and a location.
- A location has a `uri` relative to the scan root, with the `uriBaseId` value `%SRCROOT%`.
- A region is 1-based.
- A finding without a location gives a result without locations.

```json
{
  "ruleId": "NASA-P10-06",
  "ruleIndex": 5,
  "level": "error",
  "message": { "text": "package variable `counter`" },
  "locations": [
    {
      "physicalLocation": {
        "artifactLocation": { "uri": "main.go", "uriBaseId": "%SRCROOT%" },
        "region": { "startLine": 5, "startColumn": 5, "endLine": 5, "endColumn": 16 }
      }
    }
  ]
}
```

The [CI Integration](../ci-integration/) guide uploads a SARIF report to GitHub code scanning.
