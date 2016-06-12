package highlight

type TokenType string

const (
	Error       TokenType = "error"
	Comment               = "comment"
	Text                  = "text"
	Entity                = "entity"
	Attribute             = "attribute"
	String                = "string"
	Separator             = "separator"
	Operator              = "operator"
	Punctuation           = "punctuation"
)
