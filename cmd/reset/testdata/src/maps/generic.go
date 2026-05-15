package maps

// generate:reset
type genericOmap[K comparable, V interface{ Dummy() }, S ~[]V] struct { // deliberate name choice to verify receiver name generation
	One map[K]V
	Two map[K]S
}

type IMapPointers[K comparable, V any] interface{}

// generate:reset
type GenericMapPointers[K comparable, V IMapPointers[K, V]] struct {
	p  *map[K]V
	pp **map[K]V
}
