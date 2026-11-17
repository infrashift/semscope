#!/usr/bin/env python3
"""Check the Go sources against the scriptable NASA/JPL Power of 10 rules.

This script is the Go port of nasa-lint.py. ADR-0003 adapts the rules to Go.
The script checks:
  R1  no `goto`, and no direct or indirect recursion
  R2  no `for {`, `for ;; {`, `for true {` or `for init; ; post {` loop:
      every loop needs a visible bound
  R4  functions of 60 lines or fewer (the `func` line to the closing brace);
      a one-line function `func f() int { return 1 }` is a function too
  R5  an average of 2 or more assertions (`assert.That(`) per function
  R7  no discarded call results: `_ = f(...)`, or `_` as the last result of a
      call; the call must be the whole right side, so `v, _ := f().(T)` is
      not a finding
  R9  no function types of our own: `type T func(`, `type T = func(`, and
      func-typed fields, variables, parameters (also `...func(`), results and
      elements (`[]func(`, `map[K]func(`, `chan func(`)
The call graph of R1 has one node per package function and per method. An
edge goes to a function of the same package (`f(`) or to a method of the same
receiver type (`r.m(` where `r` is the receiver). A method call on an
expression, as in `g().m(` or `p[0].m(`, and a method declaration of an
interface in a function body have no edge. Other calls have no edge, the
same as nasa-lint.py. Go forbids import cycles, so a cycle of direct calls
stays inside one package.
Test files (`_test.go`) and `testdata`, `vendor` and hidden directories are
exempt. Test helpers (a name that starts with `test` or `Test`, or ends with
`Test`) are exempt from R1, R4 and R5. The package `assert` defines the
assertion primitive and is exempt from R5.
Suppress one finding with a comment on the same line or the line above:
`// nasa-lint: allow(R7) <reason>`. R5 has no suppression.
Run from the repository root:
  python3 -I scripts/nasa-lint-go.py [--verbose] [--root DIR] [DIR ...]
The default directories are cmd, internal and rules.
With --verbose, the script lists each function with fewer than 2 assertions.
"""
import pathlib
import re
import sys

MAX_FN_LINES = 60
MIN_ASSERT_AVG = 2.0
DEFAULT_DIRS = ("cmd", "internal", "rules")
SKIP_DIRS = {"testdata", "vendor"}
ASSERT_PKG = "assert"

FN_RE = re.compile(
    r"^func\s*(?:\(\s*(?:(\w+)\s+)?\*?\s*(\w+)\s*(?:\[[^\]]*\])?\s*\))?\s*([A-Za-z_]\w*)\s*[\[(]"
)
PACKAGE_RE = re.compile(r"^package\s+(\w+)")
CALL_RE = re.compile(r"(?:\b(\w+)\s*\.\s*)?\b([A-Za-z_]\w*)\s*(?:\[[\w\s,.*\[\]]*\])?\s*\(")
ASSERT_RE = re.compile(r"\b" + ASSERT_PKG + r"\.\w+\s*\(")
RUNE_RE = re.compile(r"'(?:\\'|\\[^'\n]+|[^'\\\n])'")
KEYWORDS = {
    "break", "case", "chan", "const", "continue", "default", "defer", "else",
    "fallthrough", "for", "func", "go", "goto", "if", "import", "interface",
    "map", "package", "range", "return", "select", "struct", "switch", "type", "var",
}
FUNC_TYPE = r"(?:\.\.\.)?(?:map\[[^\]]*\]|\[\w*\])*\*?func\s*\("
LINE_RULES = [
    ("R1", re.compile(r"\bgoto\b"), "`goto` statement"),
    ("R2", re.compile(r"\bfor\s*\{"), "unbounded loop `for {`"),
    ("R2", re.compile(r"\bfor\s*;\s*;\s*\{"), "unbounded loop `for ;; {`"),
    ("R2", re.compile(r"\bfor\s+true\s*\{"), "unbounded loop `for true {`"),
    ("R2", re.compile(r"\bfor\b[^;{]*;\s*;"), "unbounded loop `for init; ; post {`"),
    ("R7", re.compile(r"(?<![\w.,]\s)(?<![\w.,])_\s*=\s*[\w.]*\w\s*(?:\[[^\]]*\])?\s*\("),
     "discarded call result `_ = f(...)`"),
    ("R7", re.compile(r"\b\w+\s*(?:,\s*\w+\s*)*,\s*_\s*:?=\s*[\w.]*\w\s*(?:\[[^\]]*\])?\s*\("),
     "discarded call result `x, _ := f(...)`"),
    ("R9", re.compile(r"^\s*type\s+\w+(?:\[[^\]]*\])?\s+(?:=\s*)?" + FUNC_TYPE), "function type declaration"),
    ("R9", re.compile(r"^\s*(?:var\s+)?(?!(?:" + "|".join(sorted(KEYWORDS)) + r")\b)\w+(?:\s*,\s*\w+)*\s+" + FUNC_TYPE),
     "func-typed field or variable"),
    ("R9", re.compile(r"[(,]\s*\w+(?:\s*,\s*\w+)*\s+" + FUNC_TYPE), "func-typed parameter"),
    ("R9", re.compile(r"\)\s+(?:\((?:[^()]*,\s*)?)?" + FUNC_TYPE), "func-typed result"),
    ("R9", re.compile(r"(?:\[\w*\]|map\[[^\]]*\]|\bchan\s+|\.\(\s*)\*?func\s*\("), "func-typed element"),
]
# A fallback rule reports a line only when no other rule of its R number
# reports the line: `for i := 0; ; i++ {`, but not `for ;; {` twice, and
# `make([]func(), n)`, `x := map[string]func(){` or `v.(func())`.
FALLBACK_RULES = {"unbounded loop `for init; ; post {`", "func-typed element"}
# A line that starts with a method declaration of an interface, for
# example `Deadline() (time.Time, bool)`, holds no call.
METHOD_SPEC_RE = re.compile(r"^\s*[A-Za-z_]\w*\s*\([^()]*\)\s+[\w*\[(]")


def strip_go(text):
    """Blank out comments and string, raw string and rune literals. Keep line breaks."""
    out, i, n = [], 0, len(text)
    while i < n:
        c = text[i]
        if text.startswith("//", i):
            j = text.find("\n", i)
            i = n if j < 0 else j
        elif text.startswith("/*", i):
            j = text.find("*/", i + 2)
            j = n if j < 0 else j + 2
            out.append(" " + "\n" * text.count("\n", i, j))
            i = j
        elif c == '"':
            j = i + 1
            while j < n and text[j] not in '"\n':
                j += 2 if text[j] == "\\" else 1
            out.append('""')
            i = j + 1 if j < n and text[j] == '"' else j
        elif c == "`":
            j = text.find("`", i + 1)
            j = n if j < 0 else j + 1
            out.append("``" + "\n" * text.count("\n", i, j))
            i = j
        elif c == "'" and RUNE_RE.match(text, i):
            out.append("''")
            i = RUNE_RE.match(text, i).end()
        else:
            out.append(c)
            i += 1
    return "".join(out).split("\n")


def whole_call(code, start):
    """Report whether the call that opens at code[start] is the whole right
    side: `x, _ := f()` or `_ = a.f().g()`, but not `x, _ := f().(T)`. A call
    that does not close on its line counts as whole."""
    i = start
    for _ in range(len(code)):
        depth, j = 0, i
        while j < len(code):
            depth += {"(": 1, ")": -1}.get(code[j], 0)
            if depth == 0:
                break
            j += 1
        if j >= len(code):
            return True
        rest = code[j + 1:]
        nxt = re.match(r"\s*(?:\.\s*\w+\s*(?:\[[^\]]*\])?\s*)?\(", rest)
        if not nxt:
            return rest.strip() == "" or rest.lstrip()[:1] in (";", "}")
        i = j + nxt.end()
    return True


def suppressed(lines, idx, rule):
    tag = f"nasa-lint: allow({rule})"
    return tag in lines[idx] or (idx > 0 and tag in lines[idx - 1])


def body_open(code, fn_idx):
    """Return the line where the body of the func on `fn_idx` opens, or None.
    A one-line function `func f() int { return 1 }` opens on its own line."""
    for i in range(fn_idx, min(fn_idx + 20, len(code))):
        s = code[i].strip()
        if s.endswith("{") or (i == fn_idx and s.endswith("}") and one_line_body(s) is not None):
            return i
        if i > fn_idx and (not s or s.startswith("func")):
            return None
    return None


def block_end(code, start):
    """Return the line that closes the block opened on line `start`."""
    depth = 0
    for i in range(start, len(code)):
        depth += code[i].count("{") - code[i].count("}")
        if depth <= 0:
            return i
    return len(code) - 1


def one_line_body(line):
    """Return the text after the body brace of a one-line func, or None.
    The body brace is the first `{` after the parameters, outside parentheses,
    that does not follow `struct` or `interface`."""
    depth, closed = 0, False
    for i, ch in enumerate(line):
        if ch == "(":
            depth += 1
        elif ch == ")":
            depth -= 1
            closed = closed or depth == 0
        elif ch == "{" and depth == 0 and closed and not re.search(r"\b(?:struct|interface)\s*$", line[:i]):
            return line[i + 1:]
    return None


def body_text(code, start, open_idx, end):
    first = code[open_idx]
    if open_idx == start == end:
        return [one_line_body(first.strip()) or ""]
    return [first[first.rfind("{") + 1:]] + code[open_idx + 1:end + 1]


def is_test_helper(name):
    return name.startswith(("test", "Test")) or name.endswith("Test")


def collect_files(root, dirs):
    files = []
    for d in dirs:
        base = root / d
        if not base.is_dir():
            continue
        for p in sorted(base.rglob("*.go")):
            rel = p.relative_to(root).parts
            if any(part in SKIP_DIRS or part.startswith(".") for part in rel[:-1]):
                continue
            if not p.name.endswith("_test.go"):
                files.append(p)
    return files


def scan_file(root, path):
    text = path.read_text(encoding="utf-8")
    lines = text.split("\n")
    code = strip_go(text)
    pkgdir = path.parent.relative_to(root).as_posix()
    pkgname = next((m.group(1) for m in map(PACKAGE_RE.match, code) if m), "")
    fns = []
    for i, c in enumerate(code):
        m = FN_RE.match(c)
        if not m:
            continue
        open_idx = body_open(code, i)
        if open_idx is None:
            continue
        end = block_end(code, open_idx)
        recv_type, name = m.group(2), m.group(3)
        fns.append({
            "name": name, "recv": m.group(1), "type": recv_type,
            "key": (pkgdir, f"{recv_type}.{name}" if recv_type else name),
            "start": i, "end": end, "body": body_text(code, i, open_idx, end),
            "helper": is_test_helper(name),
        })
    return {"path": path, "lines": lines, "code": code, "pkgdir": pkgdir,
            "pkgname": pkgname, "fns": fns}


def calls_of(fn, f):
    edges = set()
    for c in fn["body"]:
        if METHOD_SPEC_RE.match(c) and not c.rstrip().endswith("{"):
            continue
        for m in CALL_RE.finditer(c):
            qual, name = m.group(1), m.group(2)
            if name in KEYWORDS or (qual is None and c[:m.start()].rstrip().endswith(".")):
                continue
            if qual is None:
                edges.add((f["pkgdir"], name))
            elif fn["recv"] and qual == fn["recv"]:
                edges.add((f["pkgdir"], f"{fn['type']}.{name}"))
    return edges


def find_cycles(graph):
    """Iterative three-color DFS. Return each back edge as a cycle path."""
    color, cycles = {}, []
    for root in sorted(graph):
        if color.get(root):
            continue
        stack = [(root, iter(sorted(graph[root])))]
        path = [root]
        color[root] = 1
        while stack:
            node, it = stack[-1]
            nxt = next(it, None)
            if nxt is None:
                color[node] = 2
                stack.pop()
                path.pop()
                continue
            if nxt not in graph:
                continue
            if color.get(nxt) == 1:
                cycles.append(path[path.index(nxt):] + [nxt])
            elif not color.get(nxt):
                color[nxt] = 1
                stack.append((nxt, iter(sorted(graph[nxt]))))
                path.append(nxt)
    return cycles


def parse_args(argv):
    verbose, root, dirs, i = False, pathlib.Path.cwd(), [], 0
    while i < len(argv):
        a = argv[i]
        if a == "--verbose":
            verbose = True
        elif a == "--root" and i + 1 < len(argv):
            root = pathlib.Path(argv[i + 1]).resolve()
            i += 1
        elif a.startswith("-"):
            raise SystemExit(f"nasa-lint-go: unknown argument {a!r}")
        else:
            dirs.append(a)
        i += 1
    return verbose, root, tuple(dirs) or DEFAULT_DIRS


def main():
    verbose, root, dirs = parse_args(sys.argv[1:])
    files = [scan_file(root, p) for p in collect_files(root, dirs)]
    errors, graph, where = [], {}, {}
    total_fns = total_asserts = 0

    def err(f, line, rule, msg):
        errors.append(f"{f['path'].relative_to(root)}:{line + 1}: {rule}: {msg}")

    for f in files:
        for i, c in enumerate(f["code"]):
            hits = set()
            for rule, rx, msg in LINE_RULES:
                m = rx.search(c)
                if not m or (msg in FALLBACK_RULES and rule in hits):
                    continue
                if rule == "R7" and not whole_call(c, m.end() - 1):
                    continue
                hits.add(rule)
                if not suppressed(f["lines"], i, rule):
                    err(f, i, rule, msg)
        for fn in f["fns"]:
            if fn["helper"]:
                continue
            length = fn["end"] - fn["start"] + 1
            if length > MAX_FN_LINES and not suppressed(f["lines"], fn["start"], "R4"):
                err(f, fn["start"], "R4", f"func {fn['name']} has {length} lines (max {MAX_FN_LINES})")
            graph.setdefault(fn["key"], set()).update(calls_of(fn, f))
            where[fn["key"]] = (f, fn["start"])
            if f["pkgname"] == ASSERT_PKG:
                continue
            n_asserts = sum(len(ASSERT_RE.findall(c)) for c in fn["body"])
            total_fns += 1
            total_asserts += n_asserts
            if verbose and n_asserts < MIN_ASSERT_AVG:
                print(f"{f['path'].relative_to(root)}:{fn['start'] + 1}: info: "
                      f"func {fn['name']} has {n_asserts} assertion(s)")
    for cyc in find_cycles(graph):
        f, line = where[cyc[0]]
        if not suppressed(f["lines"], line, "R1"):
            chain = " -> ".join(f"{d}:{name}" for d, name in cyc)
            err(f, line, "R1", f"recursion: {chain}")
    avg = total_asserts / total_fns if total_fns else 0.0
    if total_fns and avg < MIN_ASSERT_AVG:
        errors.append(f"{','.join(dirs)}: R5: {total_asserts} assertions in {total_fns} functions "
                      f"(average {avg:.2f}, min {MIN_ASSERT_AVG})")
    for e in errors:
        print(e)
    print(f"nasa-lint-go: {len(errors)} error(s); {total_fns} functions, "
          f"{total_asserts} assertions (average {avg:.2f})")
    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
