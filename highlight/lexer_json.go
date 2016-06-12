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
