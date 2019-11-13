package object

import (
	"strconv"
)

// Float wraps float64 and implements the Object interface.
type Float struct {
	// Value holds the float-value this object wraps.
	Value float64
}

// Inspect returns a string-representation of the given object.
func (f *Float) Inspect() string {
	return strconv.FormatFloat(f.Value, 'f', -1, 64)
}

// Type returns the type of this object.
func (f *Float) Type() Type {
	return FLOAT
}

// Is this value "true"?
func (f *Float) True() bool {
	return (f.Value != 0)
}
