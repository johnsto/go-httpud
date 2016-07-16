package output

import (
	"fmt"

	"bitbucket.org/johnsto/go-httpud/highlight"
)

type DebugOutputter struct {
}

func NewDebugOutputter() *DebugOutputter {
	return &DebugOutputter{}
}

func (o *DebugOutputter) Emit(t highlight.Token) (int, error) {
	return fmt.Printf("%24s\t%12s\t%#v\n", t.State, t.Type, t.Value)
}
