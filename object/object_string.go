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

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (s *String) ToInterface() interface{} {
	return s.Value
}
