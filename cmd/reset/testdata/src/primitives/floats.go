package primitives

// generate:reset
type (
	Float64Values struct {
		a, B float64
	}

	Float64ORpointers struct { // deliberate struct name to validate generated receiver
		P **float64
	}

	Float32Values struct {
		A float32
	}

	Float32Pointers struct {
		p *float32
	}
)
