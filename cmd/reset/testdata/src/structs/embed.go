package structs

type Standalone struct{} // not generated

// generate:reset
type EmbeddedValue struct {
	Standalone

	V int
}

// generate:reset
type EmbeddedPointer struct {
	*Standalone

	V int
}
