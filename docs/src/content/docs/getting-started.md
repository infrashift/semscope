---
title: Getting Started
description: Build semscope from source, install the binary and run the first scan.
---

semscope has no binary release. Build it from source on Linux.

## Prerequisites

- Go 1.26 or later.
- A C compiler and Zig 0.17.1. cgo links `libztree.a` of ztree, the Tree-sitter runtime of semscope.
- zmake or GNU Make to run the build tasks. The commands below use zmake.

## Build and install

1. Clone the repository: `git clone https://github.com/infrashift/semscope`.
2. Clone ztree into the same parent directory: `git clone https://github.com/infrashift/ztree`.
3. Go into the directory: `cd semscope`.
4. Run `zmake build`. The task builds `libztree.a` in `../ztree`, then writes the binary to `bin/semscope`.
5. Run `./bin/semscope --version` to check the build.
6. Run `zmake install` to copy the binary to `~/.local/bin/semscope`.

Set `ZTREE_DIR` when ztree is in a different directory.
Set `PREFIX` to install into a different directory, for example `zmake install PREFIX=/usr/local`.
With GNU Make, use `make build` and `make install`.

The version output names the commit, the build date, the Go version and the SARIF versions:

```console
$ semscope --version
semscope 5bdd6ba
commit: 5bdd6ba8b4052a2e42063578e45531a97096c7b3
build date: 2026-11-18T05:52:54Z
go: go1.27.3
sarif: 2.1.0, 2.2
```

## The first scan

This Go file breaks four of the Power of 10 rules:

```go
// main.go
package main

import "os"

var counter int

func walk(n int) int {
	if n == 0 {
		return 0
	}
	return walk(n - 1)
}

func main() {
	for {
		counter++
		os.Remove("tmp.txt")
	}
}
```

Scan the directory. Without `--ruleset`, semscope reads the built-in ruleset `nasa-p10`.

```console
$ semscope scan .
-: error: NASA-P10-05: go: 0 assertions in 2 functions (average 0.00, min 2)
main.go:5:5: error: NASA-P10-06: package variable `counter`
main.go:7:1: error: NASA-P10-01: recursion: walk -> walk
main.go:15:2: error: NASA-P10-02: `for` loop without a condition
semscope: 4 finding(s) in 1 file(s): 4 error(s), 0 warning(s), 0 note(s)
$ echo $?
1
```

Each line gives `path:line:column: severity: RULE-ID: message`.
The assertion average applies to all the functions of a language, so its line starts with `-`.
The exit status is 1, because one or more findings have the `error` level.

The call `os.Remove` drops an error. The Go check with type information finds it:

```console
$ semscope scan --set NASA-P10-07.go_types=true .
...
main.go:17:3: error: NASA-P10-07: unchecked error of `os.Remove`
```

## Next steps

- [Scanning](../guides/scanning/): the paths, the files and the exit status.
- [Rulesets & Overlays](../guides/rulesets/): tune `nasa-p10` or add a rule.
- [nasa-p10 Rules](../reference/nasa-p10/): what each rule checks.
