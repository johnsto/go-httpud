package highlight

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"path"
	"regexp"
)

type Rule struct {
	// Regexp is the regular expression this rule should match against.
	Regexp string
	// Type is the token type for strings that match this rule.
	Type TokenType
	// Types contains an ordered array of token types matching the order
	// of groups in the Regexp expression.
	Types []TokenType
	// State indicates the next state to migrate to if this rule is triggered.
	State string
	// Extend indicates whether the matched value should be appended to
	// the previously-matched token if it's of the same type.
	Extend bool

	// re is the cached regular expression.
	re *regexp.Regexp
}

type State []*Rule

type StateMap map[string]State

type Lexer struct {
	StateMap
	Name      string
	Filenames []string
	MimeTypes []string
}

// Tokenize reads from the given input and emits tokens to the output channel.
// Will end on any error from the reader, including io.EOF to signify the end
// of input.
func (l Lexer) Tokenize(r io.Reader, ch chan<- Token) error {
	var err error
	var line string

	br := bufio.NewReader(r)

	stack := &Stack{"root"}
	token := Token{}

reading:
	for true {
		// Read next line if we've reached the end of the current one
		if line == "" {
			line, err = br.ReadString('\n')
			if err != nil {
				break
			}
		}

		// Get name of current state
		stateName := stack.Peek()
		state := l.StateMap[stateName]

		// Match text against each matcher within the current state
		for _, m := range state {
			n, toks, types := m.Match(line)

			if n < 0 {
				// No match; try next matcher
				continue
			}

			// Update state
			if m.State == "#pop" {
				stack.Pop()
			} else if m.State != "" {
				stack.Push(m.State)
			}

			// Get string that matched
			for i, value := range toks {
				tokType := types[i]

				if token.Type == tokType && m.Extend {
					// Extend existing token
					token.Value = token.Value + value
				} else {
					// Emit current token and create next one
					ch <- token
					if Debug {
						fmt.Printf("[%s] {%s} `%s`\n", stateName,
							token.Type, token.Value)
					}
					token = Token{
						Value: value,
						Type:  tokType,
					}
				}
			}

			// Strip matched part of line
			line = line[n:]

			// Move on to next token
			continue reading
		}

		// Nothing matched!

		if token.Value != "" {
			// Emit current token
			ch <- token
		}

		// Emit unmatched error token
		ch <- Token{Value: line, Type: Error}

		if Debug {
			fmt.Printf("[%s] {%s} `%s`\n", stateName,
				token.Type, token.Value)
			fmt.Printf("[%s] {%s} %s", stateName, line, Error)
		}

		line = "" // trigger next read
	}

	// Add final token
	if token.Value != "" {
		ch <- token
	}

	return err
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

// Match attempts to match against the beginning of the given search string.
// Returns the number of characters matched, an array of submatches and a
// corresponding array of token types.
func (m *Rule) Match(subject string) (int, []string, []TokenType) {
	// Compile (and cache) regexp if not done already
	if m.re == nil {
		m.re = regexp.MustCompile(m.Regexp)
	}

	// Find match group and sub groups, returns an array of start/end offsets
	// e.g. f(r/a(b+)c/g, "abbbc") = [0, 5, 1, 4]
	indices := m.re.FindStringSubmatchIndex(subject)

	if indices == nil || indices[0] != 0 {
		// Didn't match start of subject
		return -1, nil, nil
	}

	// Get position after final matched character
	n := indices[1]

	if m.Types == nil {
		// No groups in expression; return single token and type
		return n, []string{subject[:n]}, []TokenType{m.Type}
	}

	// Multiple groups; construct array of group values and tokens
	toks := []string{}
	types := []TokenType{}
	for i := 2; i < len(indices); i += 2 {
		start, end := indices[i], indices[i+1]
		toks = append(toks, subject[start:end])
		types = append(types, m.Types[(i-2)/2])
	}

	return n, toks, types
}
