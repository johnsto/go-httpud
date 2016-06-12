package highlight

// Stack is a simple stack of string values.
type Stack []string

// Push puts a new value on to the top of the stack
func (s *Stack) Push(v string) {
	*s = append(*s, v)
}

// Peek returns the item on the top of the stack, but does not pop it.
func (s Stack) Peek() string {
	return s[len(s)-1]
}

// Pop removes and returns  the item on the top of the stack.
func (s *Stack) Pop() string {
	top := s.Peek()
	*s = (*s)[0 : len(*s)-1]
	return top
}
