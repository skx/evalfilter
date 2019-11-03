// Package stack implements a stack which is used for our virtual
// machine.
package stack

import (
	"errors"

	"github.com/skx/evalfilter/object"
)

// Stack holds return-addresses when the `call` operation is being
// completed.  It can also be used for storing ints.
type Stack struct {
	// The entries on our stack
	entries []object.Object
}

//
// Stack functions
//

// NewStack creates a new stack object.
func NewStack() *Stack {
	return &Stack{}
}

// Empty returns true if the stack is empty.
func (s *Stack) Empty() bool {
	return (len(s.entries) <= 0)
}

// Size retrieves the length of the stack.
func (s *Stack) Size() int {
	return (len(s.entries))
}

// Push adds a value to the stack.
func (s *Stack) Push(value object.Object) {
	s.entries = append(s.entries, value)
}

// Pop removes a value from the stack.
func (s *Stack) Pop() (object.Object, error) {
	if s.Empty() {
		return nil, errors.New("Pop from an empty stack")
	}

	// get the last entry.
	result := s.entries[len(s.entries)-1]

	// remove it
	s.entries = s.entries[:len(s.entries)-1]

	return result, nil
}
