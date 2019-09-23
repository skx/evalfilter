// The implementation of our string-object.

package object

// String wraps string and implements the Object interface.
type String struct {
	// Value holds the string value this object wraps.
	Value string
}

// Type returns the type of this object.
func (s *String) Type() Type {
	return STRING_OBJ
}

// Inspect returns a string-representation of the given object.
func (s *String) Inspect() string {
	return s.Value
}
