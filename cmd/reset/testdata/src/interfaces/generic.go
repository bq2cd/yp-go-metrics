package interfaces

// generate:reset
type GenericInterface[A any, C comparable, R interface{ Reset() }, S ~[]A, M ~map[C]R] struct {
	a A
	c C
	r R
	s S
	m M
}

// generate:reset
type GenericInterfacePointers[A any, D interface{ Dummy() }, S ~[]A] struct {
	aP  *A
	dP  *D
	ddP **D
	ssP **S
}
