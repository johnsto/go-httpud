package highlight

import (
	"strings"
)

var LexerJSON = Lexer{
	Name:      "JSON",
	MimeTypes: []string{"application/json"},
	Filenames: []string{"*.json"},
	States: StatesSpec{
		"root": {
			{Include: "value"},
		},
		"whitespace": {
			{Regexp: "\\s+", Type: Whitespace},
		},
		"boolean": {
			{Regexp: "(true|false|null)", Type: Constant},
		},
		"number": {
			// -123.456e+78
			{Regexp: "-?[0-9]+\\.?[0-9]*[eE][\\+\\-]?[0-9]+", Type: Number},
			// -123.456
			{Regexp: "-?[0-9]+\\.[0-9]+", Type: Number},
			// -123
			{Regexp: "-?[0-9]+", Type: Number},
		},
		"string": {
			{Regexp: `(")(")`,
				SubTypes: []TokenType{Punctuation, Punctuation}},
			{Regexp: `(")((?:\\\"|[^\"])*?)(")`,
				SubTypes: []TokenType{Punctuation, String, Punctuation}},
		},
		"value": {
			{Include: "whitespace"},
			{Include: "boolean"},
			{Include: "number"},
			{Include: "string"},
			{Include: "array"},
			{Include: "object"},
		},
		"object": {
			{Regexp: "{", Type: Punctuation, State: "objectKey"},
		},
		"objectKey": {
			{Include: "whitespace"},
			{Regexp: `(")((?:\\\"|[^\"])*?)(")(\s*)(:)`,
				SubTypes: []TokenType{Punctuation, Attribute, Punctuation,
					Whitespace, Assignment},
				State: "objectValue"},
			{Regexp: "}", Type: Punctuation, State: "#pop"},
		},
		"objectValue": {
			{Include: "whitespace"},
			{Include: "value"},
			{Regexp: ",", Type: Punctuation, State: "#pop"},
			{Regexp: "}", Type: Punctuation, State: "#pop #pop"},
		},
		"array": {
			{Regexp: "\\[", Type: Punctuation, State: "arrayValue"},
		},
		"arrayValue": {
			{Include: "whitespace"},
			{Include: "value"},
			{Regexp: ",", Type: Punctuation},
			{Regexp: "\\]", Type: Punctuation, State: "#pop"},
		},
	},
	Filters: []Filter{
		RemoveEmptiesFilter,
		&FormatterJSON{Indent: "  "},
	},
}

type FormatterJSON struct {
	Indent string
}

func (f *FormatterJSON) Filter(lexer Lexer, in <-chan Token,
	out chan<- Token) error {
	//var lastState string
	indents := 0
	for token := range in {
		indent := strings.Repeat(f.Indent, indents)

		switch token.Type {
		case Whitespace:
			// we'll add our own whitespace, thanks!
			continue
		case Assignment:
			switch token.Value {
			case ":":
				out <- token
				out <- Token{Type: Whitespace, Value: " "}
			default:
				out <- token
			}
		case Punctuation:
			switch token.Value {
			case ",":
				out <- token
				out <- Token{Type: Whitespace, Value: "\n"}
				out <- Token{Type: Whitespace, Value: indent}
			case "{":
				fallthrough
			case "[":
				out <- token
				out <- Token{Type: Whitespace, Value: "\n"}
				indents++
				indent = strings.Repeat(f.Indent, indents)
				out <- Token{Type: Whitespace, Value: indent}
			case "}":
				fallthrough
			case "]":
				out <- Token{Type: Whitespace, Value: "\n"}
				indents--
				indent = strings.Repeat(f.Indent, indents)
				out <- Token{Type: Whitespace, Value: indent}
				out <- token
			case "\"":
				out <- token
			default:
				out <- token
			}
		default:
			out <- token
		}

		//lastState = token.State
	}
	return nil
}
func init() {
	Register(LexerJSON)
}
