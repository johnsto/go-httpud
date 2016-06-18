package highlight

import (
	"regexp"
	"strings"
)

type (
	Rule interface {
		Find(subject string) (int, Rule)
		Match(subject string) (int, Rule, []Token)
		Stack() []string
	}

	// Rule describes the conditions required to match some subject text.
	RuleSpec struct {
		// Regexp is the regular expression this rule should match against.
		Regexp string
		// Type is the token type for strings that match this rule.
		Type TokenType
		// SubTypes contains an ordered array of token types matching the order
		// of groups in the Regexp expression.
		SubTypes []TokenType
		// State indicates the next state to migrate to if this rule is
		// triggered.
		State string
		// Include specifies a state to run
		Include string
	}

	IncludeRule struct {
		StateMap  *StateMap
		StateName string
	}

	RegexpRule struct {
		Regexp     *regexp.Regexp
		Type       TokenType
		SubTypes   []TokenType
		NextStates []string
	}
)

func (rs RuleSpec) Compile(sm *StateMap) Rule {
	if rs.Include != "" {
		return IncludeRule{
			StateMap:  sm,
			StateName: rs.Include,
		}
	}
	return NewRegexpRule(rs.Regexp, rs.Type, rs.SubTypes,
		strings.Split(rs.State, " "))
}

func NewRegexpRule(re string, t TokenType, subTypes []TokenType,
	next []string) RegexpRule {
	return RegexpRule{
		Regexp:     regexp.MustCompile(re),
		Type:       t,
		SubTypes:   subTypes,
		NextStates: next,
	}
}

// Find returns the first position in subject where this Rule will
// match, or -1 if no match was found.
func (m RegexpRule) Find(subject string) (int, Rule) {
	if indices := m.Regexp.FindStringIndex(subject); indices == nil {
		return -1, nil
	} else {
		return indices[0], m
	}
}

// Match attempts to match against the beginning of the given search string.
// Returns the number of characters matched, and an array of tokens.
//
// If the regular expression contains groups, they will be matched with the
// corresponding token type in `Rule.Types`. Any text inbetween groups will
// be returned using the token type defined by `Rule.Type`.
func (r RegexpRule) Match(subject string) (int, Rule, []Token) {
	// Find match group and sub groups, returns an array of start/end offsets
	// e.g. f(r/a(b+)c/g, "abbbc") = [0, 5, 1, 4]
	indices := r.Regexp.FindStringSubmatchIndex(subject)

	if indices == nil || indices[0] != 0 || indices[1] == 0 {
		// Didn't match start of subject
		return -1, nil, nil
	}

	// Get position after final matched character
	n := indices[1]

	if r.SubTypes == nil {
		// No groups in expression; return single token and type
		return n, r, []Token{{
			Value: subject[:n],
			Type:  r.Type,
		}}
	}

	// Multiple groups; construct array of group values and tokens
	tokens := []Token{}
	var start, end int = 0, 0
	for i := 2; i < len(indices); i += 2 {
		// Extract text between submatches
		sep := subject[end:indices[i]]
		if sep != "" {
			// Append to output
			tokens = append(tokens, Token{
				Value: sep,
				Type:  r.Type,
			})
		}
		// Extract submatch text
		start, end = indices[i], indices[i+1]
		tokenType := r.SubTypes[(i-2)/2]
		tokens = append(tokens, Token{
			Value: subject[start:end],
			Type:  tokenType,
		})
	}

	return n, r, tokens
}

func (r RegexpRule) Stack() []string {
	return r.NextStates
}

func (r IncludeRule) Find(subject string) (int, Rule) {
	state := r.StateMap.Get(r.StateName)
	return state.Find(subject)
}

func (r IncludeRule) Match(subject string) (int, Rule, []Token) {
	state := r.StateMap.Get(r.StateName)
	return state.Match(subject)
}

func (r IncludeRule) Stack() []string {
	return nil
}
