package primitives

// generate:reset
type StringValuesAndPointers struct {
	s, S string
	p, P *string
	PP   **string
}

// generate:reset
type StringStructAlias = StringValuesAndPointers // not generated

// generate:reset
type StringStructType StringValuesAndPointers // not generated
