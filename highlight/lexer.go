package highlight

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"path"
	"strings"
)

// Lexer defines a simple state-based lexer.
type Lexer struct {
	States    States
	Filters   []Filter
	Name      string
	Filenames []string
	MimeTypes []string
}

// Tokenize reads from the given input and emits tokens to the output channel.
// Will end on any error from the reader, including io.EOF to signify the end
// of input.
func (l Lexer) Tokenize(r io.Reader, ch chan<- Token) error {
	var err error
	var subject string

	states, err := l.States.Compile()
	if err != nil {
		return err
	}

	br := bufio.NewReader(r)

	stack := &Stack{"root"}

	for {
		// Read next line if we've reached the end of the current one
		if subject == "" {
			subject, err = br.ReadString('\n')
			if subject == "" {
				break
			}

			if err != nil && err != io.EOF {
				break
			}
		}

		// Match current state against current subject
		stateName := stack.Peek()
		state := states.Get(stateName)
		n, rule, tokens, err := state.Match(subject)
		if err != nil {
			return err
		}

		if tokens == nil {
			// No match found; treat as error instead
			tokens = []Token{{Value: subject, Type: Error}}
		}

		// Emit each token to the output
		for _, t := range tokens {
			t.State = stateName
			ch <- t
		}

		// Update state
		if rule == nil {
			// Didn't match at all, reset to root state
			stack.Empty()
			stack.Push("root")
		} else {
			for _, state := range rule.Stack() {
				if state == "#pop" {
					stack.Pop()
				} else if state != "" {
					stack.Push(state)
				}
			}
		}

		// Consume matched part
		subject = subject[n:]
	}

	return err
}

func (l Lexer) TokenizeString(s string) ([]Token, error) {
	r := strings.NewReader(s)
	tokens := []Token{}
	ch := make(chan Token, 0)
	sem := make(chan bool)
	go func() {
		for token := range ch {
			tokens = append(tokens, token)
		}
		sem <- true
	}()
	err := l.Tokenize(r, ch)
	close(ch)
	<-sem
	return tokens, err
}

// AcceptsFilename returns true if this Lexer thinks it is suitable for the
// given filename. An error will be returned iff an invalid filename pattern
// is registered by the Lexer.
func (l Lexer) AcceptsFilename(name string) (bool, error) {
	for _, fn := range l.Filenames {
		if matched, err := path.Match(fn, name); err != nil {
			return false, fmt.Errorf("malformed filename pattern '%s' for "+
				"lexer '%s': %s", fn, l.Name, err)
		} else if matched {
			return true, nil
		}
	}
	return false, nil
}

// AcceptsMediaType returns true if this Lexer thinks it is suitable for the
// given meda (MIME) type. An error wil be returned iff the given mime type
// is invalid.
func (l Lexer) AcceptsMediaType(media string) (bool, error) {
	if mime, _, err := mime.ParseMediaType(media); err != nil {
		return false, err
	} else {
		for _, mt := range l.MimeTypes {
			if mime == mt {
				return true, nil
			}
		}
	}
	return false, nil

}
