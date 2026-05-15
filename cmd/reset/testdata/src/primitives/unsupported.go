package primitives

import "unsafe"

// generate:reset
type UnsupportedValues struct { // not generated
	c1    complex64
	c2    complex128
	ptr   uintptr
	unptr unsafe.Pointer
}

// generate:reset
type UnsupportedPointers struct { // not generated
	c1    *complex64
	c2    *complex128
	ptr   *uintptr
	unptr *unsafe.Pointer
}
