package stack

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

// Test a new stack is empty
func TestStackStartsEmpty(t *testing.T) {
	s := NewStack()
	if !s.Empty() {
		t.Errorf("New stack is non-empty")
	}
	if s.Size() != 0 {
		t.Errorf("New stack is non-empty")
	}
}

// Test we can add/remove a value
func TestStack(t *testing.T) {
	s := NewStack()

	s.Push(&object.String{Value: "Steve Kemp"})
	if s.Empty() {
		t.Errorf("Stack should not be empty after adding item.")
	}
	if s.Size() != 1 {
		t.Errorf("stack has a size-mismatch")
	}

	val, err := s.Pop()

	if err != nil {
		t.Errorf("Received an unexpected error popping from the stack")
	}
	if !s.Empty() {
		t.Errorf("Stack should be empty now.")
	}
	if s.Size() != 0 {
		t.Errorf("stack has a size-mismatch")
	}

	if val.Inspect() != "Steve Kemp" {
		t.Errorf("Stack push/pop mismatch")
	}
}

// Test we can add/remove a value
func TestStackOrder(t *testing.T) {
	s := NewStack()

	s.Push(&object.String{Value: "Steve Kemp"})
	s.Push(&object.String{Value: "Adam Ant"})
	if s.Empty() {
		t.Errorf("Stack should not be empty after adding item.")
	}
	if s.Size() != 2 {
		t.Errorf("stack has a size-mismatch")
	}

	val, err := s.Pop()
	if err != nil {
		t.Errorf("Received an unexpected error popping from the stack")
	}
	if s.Size() != 1 {
		t.Errorf("stack has a size-mismatch")
	}

	if val.Inspect() != "Adam Ant" {
		t.Errorf("Stack push/pop mismatch:%s", val.Inspect())
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Received an unexpected error popping from the stack")
	}
	if s.Size() != 0 {
		t.Errorf("stack has a size-mismatch")
	}

	if val.Inspect() != "Steve Kemp" {
		t.Errorf("Stack push/pop mismatch:%s", val.Inspect())
	}
}

// Popping from an empty stack should fail
func TestEmptyStack(t *testing.T) {
	s := NewStack()

	_, err := s.Pop()

	if err == nil {
		t.Errorf("should receive an error popping an empty stack!")
	}
}
