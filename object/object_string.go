package object

import "unicode/utf8"

// String wraps string and implements the Object interface.
type String struct {
	// Value holds the string value this object wraps.
	Value string

	// Offset holds our iteration-offset
	offset int
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

// Reset implements the Iterable interface, and allows the contents
// of the string to be reset to allow re-iteration.
func (s *String) Reset() {
	s.offset = 0
}

// Next implements the Iterable interface, and allows the contents
// of our string to be iterated over.
func (s *String) Next() (Object, int, bool) {

	if s.offset < utf8.RuneCountInString(s.Value) {
		s.offset++

		// Get the characters as an array of runes
		chars := []rune(s.Value)

		// Now index
		val := String{Value: string(chars[s.offset-1])}

		return &val, s.offset - 1, true
	}

	return nil, 0, false
}
