package highlight

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mime"
	"path"
	"strings"
)

// Lexer defines a simple state-based lexer.
type Lexer struct {
	Name      string
	States    States
	Filters   Filters
	Filenames []string
	MimeTypes []string
}

// Tokenize reads from the given input and emits tokens to the output channel.
// Will end on any error from the reader, including io.EOF to signify the end
// of input.
func (l Lexer) Tokenize(r io.Reader, emit func(Token) error) error {
	states, err := l.States.Compile()
	if err != nil {
		return err
	}

	br := bufio.NewReaderSize(r, 128)

	stack := &Stack{"root"}
	eol := false
	var subject = ""
	for {
		next, err := br.ReadString('\n')

		if err == bufio.ErrBufferFull {
			eol = false
		} else if err == io.EOF {
			eol = true
		} else if err != nil {
			// something bad happened....
			return emit(EndToken)
		} else {
			eol = strings.HasSuffix(subject, "\n")
		}

		subject = subject + next

		if subject == "" && err == io.EOF {
			emit(EndToken)
			return err
		}

		for subject != "" {
			// Match current state against current subject
			stateName := stack.Peek()
			state := states.Get(stateName)

			// Tokenize input
			n, rule, tokens, err := state.Match(subject)
			//fmt.Println(subject, n, rule, tokens)
			if err != nil {
				return emit(EndToken)
			}

			// No rules matched
			if rule == nil {
				if !eol {
					// Read more data for the current line
					break
				} else {
					// Emit entire subject asn an error
					tokens = []Token{{Value: subject, Type: Error}}
					n = len(subject)
				}
			}

			// Emit each token to the output
			for _, t := range tokens {
				t.State = stateName
				if err := emit(t); err != nil {
					log.Println("STOP4", err)
					emit(EndToken)
					return err
				}
			}

			// Update state
			if rule == nil {
				// Didn't match at all, reset to root state
				stack.Empty()
				stack.Push("root")
			} else {
				// Push new states as appropriate
				for _, state := range rule.Stack() {
					if state == "#pop" {
						stack.Pop()
					} else if state != "" {
						stack.Push(state)
					}
				}
			}

			if stack.Len() == 0 {
				return emit(EndToken)
			}

			// Consume matched part
			subject = subject[n:]
		}
	}

	return nil
}

// TokenizeString is a convenience method
func (l Lexer) TokenizeString(s string) ([]Token, error) {
	r := strings.NewReader(s)
	tokens := []Token{}
	err := l.Tokenize(r, func(t Token) error {
		tokens = append(tokens, t)
		return nil
	})
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
