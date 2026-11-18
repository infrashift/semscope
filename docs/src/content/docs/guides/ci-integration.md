---
title: CI Integration
description: Stop a CI build on a finding, and upload a SARIF report to GitHub code scanning.
---

semscope gives exit status 1 when a finding has the `--fail-level` severity or a higher severity. A CI job that runs `semscope scan` stops on that status.

## A gate in a Makefile

```make
## lint: check the Power of 10 rules
lint:
	semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning src
```

Keep the overlay in the repository, for example `.semscope.cue`. Then a developer and the CI job run the same rules.
Record the reason for each change to the base ruleset in a comment of the overlay.

## Fallback for a missing binary

The zmake repository runs semscope when it is on the `PATH`, and runs its oracle script when it is not:

```sh
if command -v semscope >/dev/null 2>&1; then
  semscope scan -r nasa-p10 -r .semscope.cue --fail-level warning src
else
  python3 -I scripts/nasa-lint.py
fi
```

## GitHub code scanning

GitHub code scanning reads SARIF 2.1.0. Upload the report with `github/codeql-action/upload-sarif`.
The job needs the permission `security-events: write`.

```yaml
name: semscope

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read
  security-events: write

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
        with:
          persist-credentials: false
      - uses: actions/setup-go@v6
        with:
          go-version: '1.26'
      - name: Build semscope
        run: |
          git clone --depth 1 https://github.com/infrashift/semscope "$RUNNER_TEMP/semscope"
          cd "$RUNNER_TEMP/semscope" && go build -trimpath -o "$RUNNER_TEMP/bin/semscope" ./cmd/semscope
          echo "$RUNNER_TEMP/bin" >> "$GITHUB_PATH"
      - name: Scan
        run: semscope scan --format sarif -o semscope.sarif --fail-level none .
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v4
        with:
          sarif_file: semscope.sarif
          category: semscope
```

The step `Scan` uses `--fail-level none`, so the upload step always runs. Code scanning then shows the findings on the pull request.
To also stop the job on a finding, add a second scan step without `--fail-level none` after the upload.

The SARIF locations are relative to the scan root, with the `uriBaseId` value `%SRCROOT%`. Scan from the repository root, so that GitHub can find each file.

## JSON for other tools

The JSON report gives a `summary` object with the counts of each severity. A script can read it with `jq`:

```sh
semscope scan --format json --fail-level none . | jq '.summary'
```

[Output Formats](../output-formats/) gives each field of the JSON and SARIF reports.
