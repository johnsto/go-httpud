package highlight

import (
	"sync"
)

// Filter describes a type that is capable of filtering/processing tokens.
type Filter interface {
	// Filter reads tokens from `in` and outputs tokens to `out`, typically
	// modifying or filtering them along the way. The function should return
	// as soon as the input is exhausted (i.e. the channel is closed), or an
	// error is encountered.
	Filter(lexer Lexer, in <-chan Token, out chan<- Token) error
}

// FilterFunc is a helper type allowing filter functions to be used as
// filters.
type FilterFunc func(lexer Lexer, in <-chan Token, out chan<- Token) error

func (f FilterFunc) Filter(lexer Lexer, in <-chan Token, out chan<- Token) error {
	return f(lexer, in, out)
}

type Filters []Filter

// PushFunc returns a helper function allowing tokens to be pushed to an
// output. The returned function will run the passed Token through each filter
// and emit it to the output.
//
// It is the caller's responsibility to close the output channel when done.
func (fs Filters) PushFunc(lexer Lexer, out chan<- Token) func(token Token) error {
	in := make(chan Token)
	done := make(chan error)

	go func() {
		done <- fs.Filter(lexer, in, out)
		close(in)
	}()

	return func(token Token) error {
		select {
		case in <- token:
		case err := <-done:
			return err
		}
		return nil
	}
}

// Filter runs the input through each filter in series, emitting the final
// result to `out`. This function will return as soon as the last token has
// been processed, or iff an error is encountered by one of the filters.
//
// It is safe to close the output channel as soon as this function returns.
func (fs Filters) Filter(lexer Lexer, in <-chan Token, out chan<- Token) error {
	var prev <-chan Token
	var next chan Token

	// Return channel for errors encountered by any of the filters, must
	// be large enough to accept an error for each filter
	errChan := make(chan error, len(fs))
	defer close(errChan)

	prev = in

	wg := sync.WaitGroup{}

	// Execute each filter in turn, re-assigning `prev` each time so filters
	// feed into each other.
	for i, f := range fs {
		next = make(chan Token) // closed in goroutine
		wg.Add(1)
		go func(i int, f Filter, in <-chan Token, out chan<- Token) {
			if err := f.Filter(lexer, in, out); err != nil {
				// Emit filter error to error receiver
				errChan <- err
			}
			wg.Done()
			close(out)
		}(i, f, prev, next)
		prev = next
	}

	wg.Add(1)

	// Push output from final filter to output.
	go func(in <-chan Token, out chan<- Token) {
		for t := range in {
			out <- t
		}
		wg.Done()
	}(prev, out)

	wg.Wait()

	select {
	case err := <-errChan:
		// One of the filters returned an error, so return that
		return err
	default:
		// Filters finished
		return nil
	}
}

// PassthroughFilter simply emits each token to the output without
// modification.
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
			if token.Type == "" {
				// EOF
				break
			} else if token.Type == curr.Type {
				// Same as last token; combine
				curr.Value += token.Value
				continue
			} else if curr.Value != "" {
				out <- curr
			}
			curr = Token{
				Value: token.Value,
				Type:  token.Type,
				State: token.State,
			}
		}

		// Emit final pending token
		if curr.Value != "" {
			out <- curr
		}

		return nil
	})
