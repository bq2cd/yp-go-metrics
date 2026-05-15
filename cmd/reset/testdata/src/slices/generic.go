package slices

// generate:reset
type GenericSlices[A, B any, C comparable, D interface{ Dummy() }] struct {
	one   []A
	two   []B
	three []C
	four  []D
}

// generate:reset
type GenericSlicePointers[T comparable] struct {
	ptr  *[]T
	PPtr **[]T
}
