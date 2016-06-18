package highlight

type States interface {
	Get(name string) State
	Compile() (States, error)
}

type StatesSpec map[string][]RuleSpec

func (m StatesSpec) Get(name string) State {
	return nil
}

func (m StatesSpec) Compile() (States, error) {
	sm := &StateMap{}
	for name, specs := range m {
		rules := make(State, 0, len(specs))
		for _, spec := range specs {
			rules = append(rules, spec.Compile(sm))
		}
		(*sm)[name] = rules
	}
	return sm, nil
}

func (m StatesSpec) MustCompile() States {
	states, err := m.Compile()
	if err != nil {
		panic(err)
	}
	return states
}

type StateMap map[string]State

func (m StateMap) Get(name string) State {
	return m[name]
}

func (m StateMap) Compile() (States, error) {
	return nil, nil
}

// State is a list of matching Rules.
type State []Rule

func (s State) Find(subject string) (int, Rule) {
	var earliestPos int = len(subject)
	var earliestRule Rule

	for _, rule := range s {
		pos, matchedRule := rule.Find(subject)

		if pos < 0 {
			// no match; try next rule
			continue
		} else if pos < earliestPos {
			earliestPos = pos
			earliestRule = matchedRule
		}
	}

	if earliestPos > 0 {
		// Return part of subject that doesn't match
		return earliestPos, nil
	}

	return earliestPos, earliestRule
}

// Match tests the subject text against all rules within the State. If a match
// is found, it returns the number of characters consumed, a series of tokens
// consumed from the subject text, and the specific Rule that was succesfully
// matched against.
//
// If the start of the subject text can not be matched against any known rule,
// it will be emitted as an "Error" token and a nil Rule.
func (s State) Match(subject string) (int, Rule, []Token) {
	var earliestPos int = len(subject)
	var earliestRule Rule

	for _, rule := range s {
		pos, matchedRule := rule.Find(subject)

		if pos < 0 {
			// no match; try next rule
			continue
		} else if pos < earliestPos {
			earliestPos = pos
			earliestRule = matchedRule
		}
	}

	if earliestPos > 0 {
		// Return part of subject that doesn't match
		return earliestPos, nil, []Token{{Value: subject, Type: Error}}
	}

	// Return matching part
	return earliestRule.Match(subject)
}
