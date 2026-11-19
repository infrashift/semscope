#!/usr/bin/env python3
"""Print the default corpus list of the parity test for one language.

The output is one line of NAME=DIRECTORY entries, separated by ":". The
parity test reads it from SEMSCOPE_PARITY_CORPUS (zig) and
SEMSCOPE_PARITY_GO_CORPUS (go). The records of
testdata/nasa_p10/parity-divergences.json use the names.
  zig: the fixture scripts/testdata/nasa-lint/corpus-forms, zmake (the
       directory ../zmake), and the standard library of the zig command,
       with the name zig-std-<version>
  go:  the fixture scripts/testdata/nasa-lint-go/corpus-forms, and the
       source modules of semscope dependencies in the module cache, with the
       name <module path>@<version>
A corpus that is not on the disk is skipped.
Run from the repository root: python3 -I scripts/parity_corpora.py zig|go
"""
import pathlib
import re
import shutil
import subprocess
import sys

GO_MODULES = (
    "github.com/spf13/cobra",
    "github.com/spf13/pflag",
    "github.com/owenrumney/go-sarif/v3",
    "cuelang.org/go",
    "golang.org/x/tools",
)


def run(args):
    """Return the stdout of args, or "" when the command is not available or fails."""
    if shutil.which(args[0]) is None:
        return ""
    out = subprocess.run(args, capture_output=True, text=True, check=False)
    return out.stdout if out.returncode == 0 else ""


def zig_corpora():
    entries = [f"zig-corpus-forms={pathlib.Path('scripts/testdata/nasa-lint/corpus-forms').resolve()}"]
    zmake = pathlib.Path("../zmake").resolve()
    if (zmake / "src").is_dir():
        entries.append(f"zmake={zmake}")
    env = run(["zig", "env"])
    std = re.search(r'\.std_dir\s*=\s*"([^"]+)"', env)
    version = run(["zig", "version"]).strip()
    if std and version and pathlib.Path(std.group(1)).is_dir():
        entries.append(f"zig-std-{version}={std.group(1)}")
    return entries


def go_corpora():
    entries = [f"corpus-forms={pathlib.Path('scripts/testdata/nasa-lint-go/corpus-forms').resolve()}"]
    # One `go list` for each module: when one module is not a dependency,
    # `go list` fails, and only that module is skipped (ISSUE-014).
    for module in GO_MODULES:
        listing = run(["go", "list", "-m", "-f", "{{.Path}}@{{.Version}}={{.Dir}}", module])
        entries += [line for line in listing.split("\n") if "=" in line and not line.endswith("=")]
    return entries


def main():
    if len(sys.argv) != 2 or sys.argv[1] not in ("zig", "go"):
        print("usage: parity_corpora.py zig|go", file=sys.stderr)
        return 2
    print(":".join(zig_corpora() if sys.argv[1] == "zig" else go_corpora()))
    return 0


if __name__ == "__main__":
    sys.exit(main())
