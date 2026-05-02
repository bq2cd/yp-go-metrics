package structs

// generate:reset
type SelfReferential struct {
	left   *SelfReferential
	parent **SelfReferential
}
