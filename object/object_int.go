package object

import (
	"fmt"
)

// Integer wraps int64 and implements the Object interface.
type Integer struct {
	// Value holds the integer value this object wraps
	Value int64
}

// Inspect returns a string-representation of the given object.
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

// Type returns the type of this object.
func (i *Integer) Type() Type {
	return INTEGER_OBJ
}