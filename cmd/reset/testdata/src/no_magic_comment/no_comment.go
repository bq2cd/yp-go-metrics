package nomagiccomment

type StructWithoutComment struct {
	a, b int
}

type AliasStructWithoutComment = StructWithComment

type (
	GroupStructWithoutComment struct {
		a, b int
	}
	GroupTypeWithoutCommentA StructWithoutComment
)
