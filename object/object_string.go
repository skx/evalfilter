package object

// String wraps string and implements the Object interface.
type String struct {
	// Value holds the string value this object wraps.
	Value string
}

// Type returns the type of this object.
func (s *String) Type() Type {
	return STRING
}

// Inspect returns a string-representation of the given object.
func (s *String) Inspect() string {
	return s.Value
}

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (s *String) True() bool {
	return (s.Value != "")
}
