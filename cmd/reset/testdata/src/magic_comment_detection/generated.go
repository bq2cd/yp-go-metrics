package magiccommentdetection

// Generation happens for all types with comments in this file
// as all comments are valid.

//generate:reset
type StructA struct {
	a int
}

// generate:reset
type StructB struct {
	B int
}

// Some big multi line comment here
// and there
//
// and here
//
//	generate:reset
//
// and magic comment in the middle
type StructMultiA struct {
	a int
}

//	generate:reset
//
// multi line comment follows magic comment
// and again
//
// and again
type StructMultiB struct {
	B int
}

// another multi
// line
// comment
//
// here
//
//generate:reset
type StructMultiC struct {
	c int
}

// multi line
// comment
// and
// magic comment is
// the
// last
// generate:reset
type StructMultiD struct {
	D int
}

type (
	// generate:reset
	GroupStructA struct {
		A int
	}
	GroupStructB struct { // not generated
		B int
	}
)

// multi line type group
//
//generate:reset
type (
	GroupStructC struct {
		C int
	}
	GroupStructD struct {
		D int
	}
	// generate:reset
	GroupStructE struct {
		E int
	}
)
