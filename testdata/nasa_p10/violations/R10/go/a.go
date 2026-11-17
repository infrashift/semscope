package a

import "unsafe"

//nolint
func f() uintptr { return unsafe.Sizeof(0) }

//nolint:errcheck // the close error has no use here
func g() {}
