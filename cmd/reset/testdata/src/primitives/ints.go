package primitives

//generate:reset
type IntFramework struct { // deliberate struct name to test receiver generation
	a, b, c int
	V64     int64
	V32     int32
	V16     int16
	V8      int8
}

//generate:reset
type unsignedIntFramework struct {
	a, b, c uint
	V64     uint64
	V32     uint32
	V16     uint16
	V8      uint8
}

// generate:reset
type IntPointers struct {
	A *int
	B **int
	C ***int
}
