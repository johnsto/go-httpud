package highlight_test

import (
	"testing"

	. "bitbucket.org/johnsto/go-httpud/highlight"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	stack := Stack{}
	assert.Equal(t, "", stack.Peek())

	stack.Push("a")
	assert.Equal(t, "a", stack.Peek())

	stack.Push("b")
	assert.Equal(t, "b", stack.Peek())

	assert.Equal(t, "b", stack.Pop())
	assert.Equal(t, "a", stack.Peek())
	assert.Equal(t, "a", stack.Pop())

	assert.Equal(t, "", stack.Peek())
	assert.Equal(t, "", stack.Pop())

}
