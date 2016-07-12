package lexers

import . "bitbucket.org/johnsto/go-httpud/highlight"

var CSS = Lexer{
	Name:      "CSS",
	MimeTypes: []string{"text/css"},
	Filenames: []string{"*.css"},
	States: StatesSpec{
		"root": {
			{Include: "whitespace"},
			{Include: "singleLineComment"},
			{Include: "multiLineComment"},
			{Include: "selector"},
			{Regexp: `[\-a-zA-Z][a-zA-Z0-9-]+?`,
				Type: Attribute, State: "rules"},
		},
		"rules": {
			{Include: "whitespace"},
			{Include: "singleLineComment"},
			{Include: "multiLineComment"},
			{Regexp: `[\-a-zA-Z][a-zA-Z0-9-]+?\w*:`,
				Type: Attribute, State: "ruleValue"},
		},
		"ruleValue": {
			{Regexp: `;`, Type: "Punctuation", State: "#pop"},
			{Regexp: `.*`, Type: "Text"},
		},
		"selector": {
			{Regexp: `\s+`, Type: Entity},
			{Regexp: `{`, Type: Punctuation, State: "declaration"},
		},
		"declaration": {
			{Include: "whitespace"},
			{Include: "singleLineComment"},
			{Include: "multiLineComment"},
			{Regexp: `[a-zA-Z0-9_-]+?\w+:`, Type: Entity, State: "declarationValue"},
			{Regexp: `}`, Type: Punctuation, State: "#pop"},
		},
		"declarationValue": {
			{Regexp: `;`, Type: Punctuation, State: "#pop"},
		},
		"whitespace": {
			{Regexp: `[ \r\n\f\t]+`, Type: Whitespace},
		},
		"singleLineComment": {
			{Regexp: `\/\/.*$`, Type: Comment},
		},
		"multiLineComment": {
			{Regexp: `\/\*`, Type: Comment, State: "multiLineCommentContents"},
		},
		"multiLineCommentContents": {
			{Regexp: `\*\/`, Type: Comment, State: "#pop"},
			{Regexp: `.*`, Type: Comment},
		},
	},
}

func init() {
	Register(CSS.Name, CSS)
}
