package object

import (
	"fmt"
	"hash/fnv"
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

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (f *Float) True() bool {
	return (f.Value > 0)
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (f *Float) ToInterface() interface{} {
	return f.Value
}

// Increase implements the Increment interface, and allows the postfix
// "++" operator to be applied to float-objects
func (f *Float) Increase() {
	f.Value++
}

// Decrease implements the Decrement interface, and allows the postfix
// "--" operator to be applied to float-objects
func (f *Float) Decrease() {
	f.Value--
}

// HashKey returns a hash key for the given object.
func (f *Float) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(f.Inspect()))
	return HashKey{Type: f.Type(), Value: h.Sum64()}
}

// JSON converts this object to a JSON string.
func (f *Float) JSON() (string, error) {
	return fmt.Sprintf("%f", f.Value), nil
}

// Ensure this object implements the expected interfaces.
var _ Decrement = &Float{}
var _ Hashable = &Float{}
var _ Increment = &Float{}
var _ JSONAble = &Float{}
