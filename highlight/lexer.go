package highlight

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"path"
	"regexp"
)

// Lexer defines a simple state-based lexer.
type Lexer struct {
	StateMap
	Name      string
	Filenames []string
	MimeTypes []string
}

// StateMap is a map of states to their names. A StateMap should typically
// contain at least one "root" State.
type StateMap map[string]State

// State is a list of matching Rules.
type State []*Rule

// Rule describes the conditions required to match some subject text.
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

// Tokenize reads from the given input and emits tokens to the output channel.
// Will end on any error from the reader, including io.EOF to signify the end
// of input.
func (l Lexer) Tokenize(r io.Reader, ch chan<- Token) error {
	var err error
	var line string

	br := bufio.NewReader(r)

	stack := &Stack{"root"}

	for true {
		// Read next line if we've reached the end of the current one
		if line == "" {
			line, err = br.ReadString('\n')
			if err != nil {
				break
			}
		}

		// Match current state against current line
		stateName := stack.Peek()
		state := l.StateMap[stateName]
		n, tokens, matchedRule := state.Match(line)
		for _, t := range tokens {
			ch <- t
		}

		if matchedRule == nil {
			// Didn't match at all, reset to root state
			stack.Empty()
			stack.Push("root")
		} else if matchedRule.State == "#pop" {
			// Pop current state
			stack.Pop()
		} else if matchedRule.State != "" {
			// Push next state
			stack.Push(matchedRule.State)
		}

		// Consume matched part of line
		line = line[n:]
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

// Match tests the subject text against all rules within the State. If a match
// is found, it returns the number of characters consumed, a series of tokens
// consumed from the subject text, and the specific Rule that was succesfully
// matched against.
//
// If no rule was matched, the output will match the entire length of the
// input text, emitted as an Error token with no corresponding Rule.
func (s *State) Match(subject string) (int, []Token, *Rule) {
	var earliestPos int = len(subject)
	var earliestRule *Rule

	for _, rule := range *s {
		pos := rule.Find(subject)

		if pos < 0 {
			// no match; try next rule
			continue
		} else if pos < earliestPos {
			earliestPos = pos
			earliestRule = rule
		}
	}

	if earliestPos > 0 {
		// Return part of subject that doesn't match
		return earliestPos, []Token{{subject, Error}}, nil
	}

	// Return matching part
	n, tokens := earliestRule.Match(subject)
	return n, tokens, earliestRule
}

// Find returns the first position in subject where this Rule will
// match, or -1 if no match was found.
func (m *Rule) Find(subject string) int {
	// Compile (and cache) regexp if not done already
	if m.re == nil {
		m.re = regexp.MustCompile(m.Regexp)
	}

	if indices := m.re.FindStringIndex(subject); indices == nil {
		return -1
	} else {
		return indices[0]
	}
}

// Match attempts to match against the beginning of the given search string.
// Returns the number of characters matched, and an array of tokens.
//
// If the regular expression contains groups, they will be matched with the
// corresponding token type in `Rule.Types`. Any text inbetween groups will
// be returned using the token type defined by `Rule.Type`.
func (m *Rule) Match(subject string) (int, []Token) {
	// Compile (and cache) regexp if not done already
	if m.re == nil {
		m.re = regexp.MustCompile(m.Regexp)
	}

	// Find match group and sub groups, returns an array of start/end offsets
	// e.g. f(r/a(b+)c/g, "abbbc") = [0, 5, 1, 4]
	indices := m.re.FindStringSubmatchIndex(subject)

	if indices == nil || indices[0] != 0 {
		// Didn't match start of subject
		return -1, nil
	}

	// Get position after final matched character
	n := indices[1]

	if m.Types == nil {
		// No groups in expression; return single token and type
		return n, []Token{{subject[:n], m.Type}}
	}

	// Multiple groups; construct array of group values and tokens
	tokens := []Token{}
	var start, end int = 0, 0
	for i := 2; i < len(indices); i += 2 {
		// Extract text between submatches
		sep := subject[end:indices[i]]
		if sep != "" {
			// Append to output
			tokens = append(tokens, Token{sep, m.Type})
		}
		// Extract submatch text
		start, end = indices[i], indices[i+1]
		tokenType := m.Types[(i-2)/2]
		tokens = append(tokens, Token{subject[start:end], tokenType})
	}

	return n, tokens
}
