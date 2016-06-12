package highlight

var HTML = Lexer{
	Name:      "HTML",
	MimeTypes: []string{"text/html", "application/xhtml+xml"},
	Filenames: []string{"*.html", "*.htm", "*.xhtml"},
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

func init() {
	Register(HTML)
}
