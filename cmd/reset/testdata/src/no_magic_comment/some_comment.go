package nomagiccomment

// a comment
type StructWithComment struct {
	a, b int
}

// alias comment
type AliasStructWithComment = StructWithComment

type (
	// another comment
	GroupStructWithComment struct {
		a, b int
	}
	GroupTypeWithoutCommentB StructWithComment
)

// group comment
type (
	CGroupStructWithoutComment struct {
		a, b int
	}
	// nested comment
	CGroupStructWithComment struct {
		a, b int
	}
)
