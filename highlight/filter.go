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

func (f Filters) Filter(lexer Lexer, in <-chan Token, out chan<- Token) error {
	errChan := make(chan error, len(f))
	defer close(errChan)

	var next chan Token
	var fin <-chan Token = in
	var fout chan<- Token = out

	wg := sync.WaitGroup{}
	wg.Add(len(f))
	for i, filter := range f {
		if i < len(f) {
			next = make(chan Token)
			fout = next
		} else {
			fout = out
		}

		go func(in <-chan Token, out chan<- Token) {
			if err := filter.Filter(lexer, in, out); err != nil {
				errChan <- err
			}
			close(out)
			wg.Done()
		}(fin, fout)

		fin = next
	}

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
			if token.Value == "" {
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
