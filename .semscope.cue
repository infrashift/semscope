// The overlay of the semscope self-scan (zmake selfscan).
// ADR-0003 records each change to nasa-p10 in this file.
package overlay

// R3 deviation: Go, cobra, CUE and Tree-sitter allocate memory at run time.
// semscope sizes its pools in engine.New and caps the rest with hard limits.
rules: "NASA-P10-03": enabled: false

// The Python gate scripts are tooling, not semscope code. scripts/nasa-lint.py
// stays byte-identical with zmake, because it is the Zig oracle (ADR-0005).
exclude: scripts: "Python gate tooling; nasa-lint.py stays byte-identical as the Zig oracle (ADR-0005)"

// semscope checks its own Go code with type information: each call, defer
// and go statement that drops an error is a finding (ADR-0009).
rules: "NASA-P10-07": params: go_types: true

// The patched Tree-sitter Zig grammar is third-party code: parser.c is
// generated, and the Go binding uses cgo and unsafe (ADR-0010).
exclude: third_party: "vendored Tree-sitter Zig grammar with a generated parser (ADR-0010)"
