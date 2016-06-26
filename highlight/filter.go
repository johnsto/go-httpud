package highlight

import (
	"sync"
)

type Filter interface {
	Filter(lexer Lexer, in <-chan Token, out chan<- Token) error
}

type FilterFunc func(lexer Lexer, in <-chan Token, out chan<- Token) error

func (f FilterFunc) Filter(lexer Lexer, in <-chan Token, out chan<- Token) error {
	return f(lexer, in, out)
}

type Filters []Filter

func (fs Filters) Filter(lexer Lexer, in <-chan Token, out chan<- Token) error {
	var prev <-chan Token
	var next chan Token

	errChan := make(chan error, len(fs))

	prev = in

	wg := sync.WaitGroup{}
	for i, f := range fs {
		next = make(chan Token)
		wg.Add(1)
		go func(i int, f Filter, in <-chan Token, out chan<- Token) {
			if err := f.Filter(lexer, in, out); err != nil {
				errChan <- err
			}
			wg.Done()
			close(out)
		}(i, f, prev, next)
		prev = next
	}

	wg.Add(1)
	go func(in <-chan Token, out chan<- Token) {
		for t := range in {
			out <- t
		}
		wg.Done()
	}(prev, out)

	wg.Wait()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

var PassthroughFilter = FilterFunc(
	func(l Lexer, in <-chan Token, out chan<- Token) error {
		for token := range in {
			out <- token
		}
		return nil
	})

// RemoveEmptiesFilter removes empty (zero-length) tokens from the output.
var RemoveEmptiesFilter = FilterFunc(
	func(lexer Lexer, in <-chan Token, out chan<- Token) error {
		for token := range in {
			if token.Value != "" {
				out <- token
			}
		}
		return nil
	})

// MergeTokensFilter combines Tokens if they have the same type.
var MergeTokensFilter = FilterFunc(
	func(lexer Lexer, in <-chan Token, out chan<- Token) error {
		curr := Token{}

		for token := range in {
			if token.Type == curr.Type {
				curr.Value += token.Value
				continue
			}
			if curr.Value != "" {
				out <- curr
			}
			curr = Token{
				Value: token.Value,
				Type:  token.Type,
				State: token.State,
			}
		}

		if curr.Value != "" {
			out <- curr
		}

		return nil
	})
