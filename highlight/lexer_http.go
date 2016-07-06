package highlight

import ()

var LexerHTTP = Lexer{
	Name: "http",
	States: StatesSpec{
		"root": {
			{Regexp: `^(HTTP)(/)([0-9\.]+)( )([0-9]+)(.*)(\r\n)$`,
				SubTypes: []TokenType{Entity, Punctuation, Entity,
					Whitespace, Number, Whitespace, String, Whitespace},
				State: "headers"},
		},
		"headers": {
			{Regexp: `^(.*?)(:)(\s*)`,
				SubTypes: []TokenType{Attribute, Assignment, Whitespace},
				State:    "headerValue"},
			{Regexp: `^\r\n$`, State: "#pop #pop"},
		},
		"headerValue": {
			{Regexp: `\r\n$`, State: "#pop", Type: Whitespace},
			{Regexp: `[^;]+?`, Type: Text},
			{Regexp: `;`, Type: Punctuation},
			{Regexp: `\r\n$`, State: "#pop"},
		},
	},
	Filters: []Filter{},
}

func init() {
	Register(LexerHTTP.Name, LexerHTTP)
}