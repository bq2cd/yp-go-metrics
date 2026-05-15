package maps

// generate:reset
type maybeAmaP struct { // deliberate name choice to verify receiver name generation
	one map[int]struct{}
	two map[string]bool
}

// generate:reset
type maybeAmaPpointers struct { // deliberate name choice to verify receiver name generation
	oneP *map[int]struct{}
	twoP **map[string]bool
}
