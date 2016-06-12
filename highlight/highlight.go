package highlight

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
)

var Debug bool = false

type Rule struct {
	Regexp string
	Type   TokenType
	Types  []TokenType
	State  string
	Extend bool

	re *regexp.Regexp
}

type State []*Rule

type StateMap map[string]State

type Lexer struct {
	StateMap
}

type Token struct {
	Value string
	Type  TokenType
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

// Tokenise reads from the given input and emits tokens to the output channel.
// Will end on any error from the reader, including io.EOF to signify the end
// of input.
func (l Lexer) Tokenise(r io.Reader, ch chan<- Token) error {
	br := bufio.NewReader(r)

	stack := &Stack{"root"}
	token := Token{}

	var err error
	var line string

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
