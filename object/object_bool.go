package object

import (
	"fmt"
)

// Boolean wraps bool and implements the Object interface.
type Boolean struct {
	// Value holds the boolean value we wrap.
	Value bool
}

// Type returns the type of this object.
func (b *Boolean) Type() Type {
	return BOOLEAN
}

// Inspect returns a string-representation of the given object.
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (b *Boolean) True() bool {
	return b.Value
}
