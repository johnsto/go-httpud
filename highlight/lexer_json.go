package highlight

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
				SubTypes: []TokenType{Punctuation, String, Punctuation,
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
	/*Filters: []Filter{
		RemoveEmptiesFilter,
		MergeTokensFilter,
	},*/
}

func init() {
	Register(LexerJSON)
}
