// The cgo include of ../../src/parser.c is outside this package, so the Go
// build cache does not see a new parser.c. The hash below changes this file
// with each new parser.c, and a semscope test checks it:
// parser.c sha256: 175d799a6aa10770ac3162c3409f217ff0b866c0104fa285ad17b1e1af64d4a3
package tree_sitter_zig

// #cgo CFLAGS: -std=c11 -fPIC
// #include "../../src/parser.c"
import "C"

import "unsafe"

// Get the tree-sitter Language for this grammar.
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_zig())
}
