package highlight_test

import (
	"io"
	"testing"

	. "bitbucket.org/johnsto/go-httpud/highlight"
	"github.com/stretchr/testify/assert"
)

func TestLexerJSON(t *testing.T) {
	type simpleToken struct {
		Value string
		Type  TokenType
	}

	states, err := LexerJSON.States.Compile()
	assert.Nil(t, err, "JSON lexer should compile")

	for _, item := range []struct {
		State   string
		Length  int
		Subject string
		Tokens  []Token
	}{
		{"boolean", 4, "true", []Token{{Value: "true", Type: Constant}}},
		{"boolean", 5, "false", []Token{{Value: "false", Type: Constant}}},
		{"boolean", 4, "null", []Token{{Value: "null", Type: Constant}}},
		{"number", 1, "0", []Token{{Value: "0", Type: Number}}},
		{"number", 2, "-0", []Token{{Value: "-0", Type: Number}}},
		{"number", 3, "0.0", []Token{{Value: "0.0", Type: Number}}},
		{"number", 4, "-0.0", []Token{{Value: "-0.0", Type: Number}}},
		{"number", 5, "1.2e3", []Token{{Value: "1.2e3", Type: Number}}},
		{"number", 6, "-1.2e3", []Token{{Value: "-1.2e3", Type: Number}}},
		{"number", 7, "-1.2e-4", []Token{{Value: "-1.2e-4", Type: Number}}},
		{"string", 2, `""`, []Token{
			{Value: `"`, Type: Punctuation},
			{Value: `"`, Type: Punctuation},
		}},
		{"string", 4, `"  "`, []Token{
			{Value: `"`, Type: Punctuation},
			{Value: `  `, Type: String},
			{Value: `"`, Type: Punctuation},
		}},
		{"string", 5, `"xyz"`, []Token{
			{Value: `"`, Type: Punctuation},
			{Value: `xyz`, Type: String},
			{Value: `"`, Type: Punctuation},
		}},
		{"string", 9, `"\"xyz\""`, []Token{
			{Value: `"`, Type: Punctuation},
			{Value: `\"xyz\"`, Type: String},
			{Value: `"`, Type: Punctuation},
		}},
		{"string", 8, `"\t\n\r"`, []Token{
			{Value: `"`, Type: Punctuation},
			{Value: `\t\n\r`, Type: String},
			{Value: `"`, Type: Punctuation},
		}},
	} {
		n, _, tokens, err := states.Get(item.State).Match(item.Subject)
		assert.Nil(t, err, item.Subject)
		assert.Equal(t, item.Length, n, item.Subject)
		assert.Equal(t, item.Tokens, tokens, item.Subject)
	}

	for _, item := range []struct {
		State   string
		Subject string
		Tokens  []simpleToken
	}{
		{"array", "[]", []simpleToken{
			{"[", Punctuation},
			{"]", Punctuation},
		}},
		{"array", "[null]", []simpleToken{
			{"[", Punctuation},
			{"null", Constant},
			{"]", Punctuation},
		}},
		{"array", "[123]", []simpleToken{
			{"[", Punctuation},
			{"123", Number},
			{"]", Punctuation},
		}},
		{"array", "[ ]", []simpleToken{
			{"[", Punctuation},
			{" ", Text},
			{"]", Punctuation},
		}},
		{"array", "[1,2]", []simpleToken{
			{"[", Punctuation},
			{"1", Number},
			{",", Punctuation},
			{"2", Number},
			{"]", Punctuation},
		}},
		{"array", "[1,\n2]", []simpleToken{
			{"[", Punctuation},
			{"1", Number},
			{",", Punctuation},
			{"\n", Text},
			{"2", Number},
			{"]", Punctuation},
		}},
		{"object", "{}", []simpleToken{
			{"{", Punctuation},
			{"}", Punctuation},
		}},
		{"object", `{"key":"value"}`, []simpleToken{
			{"{", Punctuation},
			{`"`, Punctuation},
			{"key", String},
			{`"`, Punctuation},
			{"", Text},
			{":", Assignment},
			{`"`, Punctuation},
			{"value", String},
			{`"`, Punctuation},
			{"}", Punctuation},
		}},
		{"object", `{ "key" : "value" }`, []simpleToken{
			{"{", Punctuation},
			{" ", Text},
			{`"`, Punctuation},
			{"key", String},
			{`"`, Punctuation},
			{" ", Text},
			{":", Assignment},
			{" ", Text},
			{`"`, Punctuation},
			{"value", String},
			{`"`, Punctuation},
			{" ", Text},
			{"}", Punctuation},
		}},
		{"object", `{"a":"b","c":"d"}`, []simpleToken{
			{"{", Punctuation},
			{`"`, Punctuation},
			{"a", String},
			{`"`, Punctuation},
			{"", Text},
			{":", Assignment},
			{`"`, Punctuation},
			{"b", String},
			{`"`, Punctuation},
			{`,`, Punctuation},
			{`"`, Punctuation},
			{"c", String},
			{`"`, Punctuation},
			{"", Text},
			{":", Assignment},
			{`"`, Punctuation},
			{"d", String},
			{`"`, Punctuation},
			{"}", Punctuation},
		}},
	} {
		tokens, err := LexerJSON.TokenizeString(item.Subject)
		assert.Equal(t, io.EOF, err)
		assert.Equal(t, len(item.Tokens), len(tokens), item.Subject)
		for i, token := range tokens {
			actualToken := Token{
				Value: token.Value,
				Type:  token.Type,
			}
			expectedToken := Token{
				Value: item.Tokens[i].Value,
				Type:  item.Tokens[i].Type,
			}
			if i < len(item.Tokens) {
				assert.Equal(t, expectedToken, actualToken, item.Subject)
			}
		}
	}
}
