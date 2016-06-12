package highlight

/*var JSON = Lexer{
	StateMap: StateMap{
		"root": {
			{Regexp: "{", Type: Separator, State: "object"},
			{Regexp: "\\[", Type: Separator, State: "array"},
			{Regexp: "\\s*", Type: Text},
		},
		"object": {
			{Regexp: "\".*\"", Type: Entity},
			{Regexp: ":", Type: Separator},
			{Regexp: "}", Type: Separator, State: "#pop"},
			{Regexp: "\\s*", Type: Text},
		},
		"array": {
			{Regexp: "\".*\"", Type: Entity},
			{Regexp: ",", Type: Separator},
			{Regexp: "\\]", Type: Separator, State: "#pop"},
			{Regexp: "\\s*", Type: Text},
		},
	},
}*/

var HTML = Lexer{
	StateMap: StateMap{
		"root": {
			{Regexp: "[^<&]", Type: Text, Extend: true},
			{Regexp: "<!--", Type: Comment, State: "comment"},
			{Regexp: "<![^>]*>", Type: Entity},
			{Regexp: "(</?)([\\w-]*:?[\\w-]+)(\\s*)(>)",
				Types: []TokenType{Punctuation, Entity, Text, Punctuation}},
			{Regexp: "(<)([\\w-]*:?[\\w-]+)(\\s*)",
				Types: []TokenType{Punctuation, Entity, Text},
				State: "tag"},
		},
		"comment": {
			{Regexp: "-->", Type: Comment, State: "#pop", Extend: true},
			{Regexp: "[^-]+", Type: Comment, Extend: true},
		},
		"tag": {
			{Regexp: "[\\w-]+\\s+", Type: Attribute, Extend: true},
			{Regexp: "([\\w-]+)(=)(\\s*)",
				Types:  []TokenType{Attribute, Operator, Text},
				Extend: true,
				State:  "tagAttr"},
			{Regexp: "\\s+", Type: Entity, Extend: true},
			{Regexp: "(/?)(\\s*)(>)",
				Types:  []TokenType{Punctuation, Entity, Punctuation},
				State:  "#pop",
				Extend: true},
		},
		"tagAttr": {
			{Regexp: "\"[^\"]*\"", Type: String, Extend: true, State: "#pop"},
			{Regexp: "'[^']*'", Type: String, Extend: true, State: "#pop"},
			{Regexp: "\\w+", Type: String, Extend: true, State: "#pop"},
		},
	},
}
