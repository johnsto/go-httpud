package highlight_test

import (
	"fmt"
	"testing"

	. "bitbucket.org/johnsto/go-httpud/highlight"
	"github.com/stretchr/testify/assert"
)

func TestRuleFind(t *testing.T) {
	assert.Equal(t, 0, (&Rule{Regexp: "a+b+c+"}).Find("aabbcc"))
	assert.Equal(t, 3, (&Rule{Regexp: "a+b+c+"}).Find("zzzaabbcc"))
	assert.Equal(t, -1, (&Rule{Regexp: "a+b+c+"}).Find("zzz"))
}

func TestRuleMatch(t *testing.T) {
	for _, item := range []struct {
		Regexp  string
		Type    TokenType
		Types   []TokenType
		Subject string
		Length  int
		Tokens  []Token
	}{
		// Non-matching
		{"ab+c", Text, nil, "", -1, nil},
		// Simple matching
		{"ab+c", Text, nil, "abc", 3, []Token{{"abc", Text}}},
		{"ab+c", Text, nil, "abbbc", 5, []Token{{"abbbc", Text}}},
		// Non-matching subgroup
		{"(b+)(c+)", Error, []TokenType{Text}, "bbb", -1, nil},
		// Simple matching subgroup
		{"(b+)(c+)", Error, []TokenType{Text, Text}, "bbcc", 4,
			[]Token{{"bb", Text}, {"cc", Text}}},
		{"(b+)(c+)", Error, nil, "bbcc", 4, []Token{{"bbcc", Error}}},
		// Subgroup with outliers
		{"a(b+)cc(d+)", Error, []TokenType{Text, Text}, "abbccddd", 8,
			[]Token{{"a", Error}, {"bb", Text}, {"cc", Error}, {"ddd", Text}}},
	} {
		rule := &Rule{
			Regexp: item.Regexp,
			Type:   item.Type,
			Types:  item.Types,
		}
		n, tokens := rule.Match(item.Subject)
		description := fmt.Sprintf("%s - %s", item.Regexp, item.Subject)
		assert.Equal(t, item.Length, n, description)
		assert.Equal(t, len(item.Tokens), len(tokens), description)
		assert.Equal(t, item.Tokens, tokens, description)
	}
}
