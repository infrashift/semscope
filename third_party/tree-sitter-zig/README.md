# tree-sitter-zig, patched for Zig 0.17

This directory holds a patched copy of the Tree-sitter Zig grammar. semscope uses it in place of the release module. The `replace` directive in the semscope `go.mod` selects it. ADR-0010 in the semscope wiki records the decision.

## Base
- Repository: https://github.com/tree-sitter-grammars/tree-sitter-zig
- Version: v1.1.2 (commit `b670c8df85`), the last release.
- License: MIT. See `LICENSE`.

## Patch
Zig 0.17 has no keywords `async`, `await` and `usingnamespace`. The release grammar still reserves them. Zig 0.17 code uses `await` as an identifier, for example in `std.Io`, so the release grammar gives parse errors. Zig 0.15 also writes the clobbers of `asm` as a struct literal, for example `: .{ .rcx = true }`. The release grammar reads strings only. The patch changes four rules of `grammar.js`:

1. It removes `async_expression` and its reference in `expression`.
2. It removes `await_expression` and its reference in `expression`.
3. It removes `using_namespace_declaration` and its reference in the container members.
4. `asm_clobbers` also reads an anonymous struct literal (ISSUE-009 in the semscope wiki).

The patch changes no other rule. Upstream master after v1.1.2 also puts call arguments in an `arguments` node. This copy does not take that change, so the node shapes stay the same as in v1.1.2.

## Files
- `grammar.js`: the patched grammar.
- `src/`: the generated parser. Do not edit it.
- `bindings/go/binding.go`: the Go binding of v1.1.2.
- `go.mod`: the module path of the release, so the `replace` directive works.

## Generate the parser
Run these commands in this directory:

```sh
npx --yes tree-sitter-cli@0.25.10 generate
```

The command writes `src/parser.c`, `src/grammar.json`, `src/node-types.json` and `src/tree_sitter/`. Then do these steps:

1. Run `sha256sum src/parser.c`.
2. Put the hash on the line `// parser.c sha256:` of `bindings/go/binding.go`. The Go build cache does not see a new `parser.c` without a change to `binding.go`. The semscope test `TestZigParserHash` checks the hash.
3. Run `zmake check` and `zmake parity` in the semscope root.
