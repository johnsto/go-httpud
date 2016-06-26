package highlight_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	. "bitbucket.org/johnsto/go-httpud/highlight"
)

func TestFilters(t *testing.T) {
	for _, item := range []struct {
		Name    string
		Filters Filters
		Input   []Token
		Output  []Token
		Error   error
	}{{
		Name:    "no filters",
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
		Name: "error",
		Filters: Filters{FilterFunc(func(l Lexer, in <-chan Token,
			out chan<- Token) error {
			return io.EOF
		})},
		Input: []Token{
			{Value: "a", Type: Text},
		},
		Output: []Token{},
		Error:  io.EOF,
	}, {
		Name:    "MergeTokensFilter",
		Filters: Filters{MergeTokensFilter},
		Input: []Token{
			{Value: "a", Type: Text},
			{Value: "b", Type: Text},
			{Value: "c", Type: Text},
		},
		Output: []Token{
			{Value: "abc", Type: Text},
		},
	}, {
		Name:    "RemoveEmptiesFilter",
		Filters: Filters{RemoveEmptiesFilter},
		Input: []Token{
			{Value: "a", Type: Text},
			{Value: "", Type: Text},
			{Value: "c", Type: Text},
		},
		Output: []Token{
			{Value: "a", Type: Text},
			{Value: "c", Type: Text},
		},
	}, {
		Name:    "MergeTokensFilter -> RemoveEmptiesFilter",
		Filters: Filters{RemoveEmptiesFilter, MergeTokensFilter},
		Input: []Token{
			{Value: "a", Type: Text},
			{Value: "", Type: Text},
			{Value: "c", Type: Text},
		},
		Output: []Token{
			{Value: "ac", Type: Text},
		},
	}} {
		err := testFilters(t, item.Filters, item.Input, item.Output, item.Name)
		assert.Equal(t, item.Error, err, item.Name)
	}
}

func testFilters(t *testing.T, filters Filters, input, expected []Token,
	name string) error {
	var lexer Lexer
	in, out := make(chan Token), make(chan Token)
	finished := make(chan chan Token)
	done := make(chan bool, 1)
	errChan := make(chan error, 1)

	// Producer emits tokens into `in` channel
	go func() {
		for _, token := range input {
			select {
			case in <- token:
			case <-done:
			}
		}
		close(in)
		finished <- in
	}()

	// Filter tokens
	go func() {
		err := filters.Filter(lexer, in, out)
		errChan <- err
		done <- true
		close(out)
		finished <- out
	}()

	// Consumer reads tokens and compares against input
	go func() {
		i := 0
		for token := range out {
			assert.Equal(t, expected[i], token, name)
			i++
		}
		assert.Equal(t, len(expected), i, name)
	}()

	<-finished
	<-finished
	return <-errChan
}
