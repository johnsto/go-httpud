package highlight_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "bitbucket.org/johnsto/go-httpud/highlight"
)

func TestFilters(t *testing.T) {
	for _, item := range []struct {
		Filters Filters
		Input   []Token
		Output  []Token
	}{{
		Filters: Filters{},
		Input: []Token{
			{Value: "a", Type: Text},
			{Value: "b", Type: Text},
			{Value: "c", Type: Text},
		},
		Output: []Token{
			{Value: "a", Type: Text},
			{Value: "b", Type: Text},
			{Value: "c", Type: Text},
		},
	}, {
		Filters: Filters{MergeTokensFilter},
		Input: []Token{
			{Value: "a", Type: Text},
			{Value: "b", Type: Text},
			{Value: "c", Type: Text},
		},
		Output: []Token{
			{Value: "abc", Type: Text},
		},
	}} {
		testFilters(t, item.Filters, item.Input, item.Output)
	}
}

func testFilters(t *testing.T, filters Filters, input, expected []Token) {
	var lexer Lexer
	in, out := make(chan Token), make(chan Token)
	done := make(chan chan Token)

	// Producer emits tokens into `in` channel
	go func() {
		for _, token := range input {
			in <- token
		}
		close(in)
		done <- in
	}()

	// Filter tokens
	go func() {
		filters.Filter(lexer, in, out)
		close(in)
	}()

	// Consumer reads tokens and compares against input
	go func() {
		i := 0
		for token := range out {
			assert.Equal(t, expected[i], token)
			i++
		}
		assert.Equal(t, len(expected), i)
		done <- out
	}()

	assert.Equal(t, in, <-done)
	assert.Equal(t, out, <-done)
}
