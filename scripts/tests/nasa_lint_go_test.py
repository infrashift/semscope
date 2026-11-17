#!/usr/bin/env python3
"""Run nasa-lint-go.py on each fixture and compare the output with expected.txt.

Each directory in scripts/testdata/nasa-lint-go is one fixture. The file
expected.txt holds the exact output. The exit code must be 1 when the output
reports an error, else 0.
Run from the repository root: python3 -I scripts/tests/nasa_lint_go_test.py
"""
import pathlib
import subprocess
import sys

SCRIPT = pathlib.Path("scripts/nasa-lint-go.py")
FIXTURES = pathlib.Path("scripts/testdata/nasa-lint-go")


def check(case):
    """Return a list of failure messages for one fixture directory."""
    run = subprocess.run([sys.executable, "-I", str(SCRIPT), "--root", str(case)],
                         capture_output=True, text=True, check=False)
    want = (case / "expected.txt").read_text(encoding="utf-8")
    failures = []
    if run.stdout != want:
        failures.append(f"{case.name}: output differs\n--- want\n{want}--- got\n{run.stdout}")
    want_code = 0 if "-go: 0 error(s)" in want else 1
    if run.returncode != want_code:
        failures.append(f"{case.name}: exit code {run.returncode}, want {want_code}")
    if run.stderr:
        failures.append(f"{case.name}: stderr: {run.stderr}")
    return failures


def main():
    cases = sorted(p for p in FIXTURES.iterdir() if p.is_dir())
    failures = [f for case in cases for f in check(case)]
    for f in failures:
        print(f)
    print(f"nasa-lint-go tests: {len(cases)} fixture(s), {len(failures)} failure(s)")
    return 1 if failures or not cases else 0


if __name__ == "__main__":
    sys.exit(main())
