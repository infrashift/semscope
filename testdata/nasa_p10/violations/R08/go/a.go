package a

import _ "unsafe"

//go:linkname now runtime.nanotime
func now() int64

//go:generate echo ok
