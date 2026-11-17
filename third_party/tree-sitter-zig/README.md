# tree-sitter-zig, patched for Zig 0.17

This directory holds a patched copy of the Tree-sitter Zig grammar. semscope uses it in place of the release module. The `replace` directive in the semscope `go.mod` selects it. ADR-0010 in the semscope wiki records the decision.

## Base
- Repository: https://github.com/tree-sitter-grammars/tree-sitter-zig
- Version: v1.1.2 (commit `b670c8df85`), the last release.
- License: MIT. See `LICENSE`.

## Patch
Zig 0.17 has no keywords `async`, `await` and `usingnamespace`. The release grammar still reserves them. Zig 0.17 code uses `await` as an identifier, for example in `std.Io`, so the release grammar gives parse errors. The patch removes three rules from `grammar.js`:

1. `async_expression` and its reference in `expression`.
2. `await_expression` and its reference in `expression`.
3. `using_namespace_declaration` and its reference in the container members.

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

The command writes `src/parser.c`, `src/grammar.json`, `src/node-types.json` and `src/tree_sitter/`. Then run `zmake check` in the semscope root.
