package magiccommentdetection

// Nothing is generated for types in this file
// even though some comments might be valid.

// generate:reset
type StructEmpty struct{}

// generate:reset
type AliasToStruct = StructA

// generate:reset
type TypeFromStruct StructB

// generate:reset
type Interface interface {
	Method()
}

// generate :reset
type StructInvalidA struct {
	A int
}

// generate : reset
type StructInvalidB struct {
	B int
}

// generate: reset
type StructInvalidC struct {
	C int
}

// //generate:reset
type StructInvalidD struct {
	D int
}

//generate:reset//
type StructInvalidE struct {
	E int
}

// valid comment but not valid types
//
//generate:reset
type (
	GroupAliasToStruct  = StructA
	GroupTypeFromStruct StructB
	//generate:reset
	GroupInterface interface {
		Method()
	}
	//generate:reset
	GroupTypeFromPrimitive int
)
