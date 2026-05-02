package slices

// generate:reset
type PrimitiveSlices struct {
	_bools  []bool
	ints    []int
	Strings []string
}

// generate:reset
type PrimitiveSlicePointers struct {
	single *[]string
	Double **[]string
}
