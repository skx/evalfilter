package stack

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

// Test a new stack is empty
func TestStackStartsEmpty(t *testing.T) {
	s := New()
	if !s.Empty() {
		t.Errorf("New stack is non-empty")
	}
	if s.Size() != 0 {
		t.Errorf("New stack is non-empty")
	}
}

// Test we can add/remove a value
func TestStack(t *testing.T) {
	s := New()

	s.Push(&object.String{Value: "Steve Kemp"})
	if s.Empty() {
		t.Errorf("Stack should not be empty after adding item.")
	}
	if s.Size() != 1 {
		t.Errorf("stack has a size-mismatch")
	}

	out := s.Export()
	if len(out) != 1 {
		t.Errorf("Exported stack has wrong size")
	}
	if out[0] != "Steve Kemp" {
		t.Errorf("exported stack has the wrong value")
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
	out = s.Export()
	if len(out) != 0 {
		t.Errorf("Exported stack has wrong size")
	}

	if val.Inspect() != "Steve Kemp" {
		t.Errorf("Stack push/pop mismatch")
	}
}

// Test we can add/remove a value
func TestStackOrder(t *testing.T) {
	s := New()

	s.Push(&object.String{Value: "Steve Kemp"})
	s.Push(&object.String{Value: "Adam Ant\n"})
	if s.Empty() {
		t.Errorf("Stack should not be empty after adding item.")
	}
	if s.Size() != 2 {
		t.Errorf("stack has a size-mismatch")
	}

	// export to test
	out := s.Export()
	if len(out) != 2 {
		t.Errorf("Exported stack has wrong size")
	}

	// Test that the string is unescaped
	// i.e. "\" "n", rather than "\n".
	if out[0] != "Steve Kemp" {
		t.Errorf("exported stack has the wrong value")
	}
	if out[1] != "Adam Ant\\n" {
		t.Errorf("exported stack has the wrong value")
	}

	val, err := s.Pop()
	if err != nil {
		t.Errorf("Received an unexpected error popping from the stack")
	}
	if s.Size() != 1 {
		t.Errorf("stack has a size-mismatch")
	}

	if val.Inspect() != "Adam Ant\n" {
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
	s := New()

	_, err := s.Pop()

	if err == nil {
		t.Errorf("should receive an error popping an empty stack!")
	}
}
