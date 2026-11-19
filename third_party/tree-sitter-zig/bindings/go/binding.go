// The cgo include of ../../src/parser.c is outside this package, so the Go
// build cache does not see a new parser.c. The hash below changes this file
// with each new parser.c, and a semscope test checks it:
// parser.c sha256: 1eb3ae82f298dfa64801d7e1698a33bcd0d4417c961255a7959840976de8e512
package tree_sitter_zig

// #cgo CFLAGS: -std=c11 -fPIC
// #include "../../src/parser.c"
import "C"

import "unsafe"

// Get the tree-sitter Language for this grammar.
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_zig())
}
