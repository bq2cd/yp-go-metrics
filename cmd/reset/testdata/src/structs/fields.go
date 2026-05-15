package structs

// generate:reset
type StructAsField struct {
	ref SelfReferential
	S   Standalone
	SP  *Standalone
	SPP **Standalone
}
