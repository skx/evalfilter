package object

// Regexp wraps string and implements the Object interface.
type Regexp struct {
	// Value holds the string value this object wraps.
	//
	// (Yes we're a regexp, but we pretend we're string!)
	Value string
}

// Type returns the type of this object.
func (r *Regexp) Type() Type {
	return REGEXP
}

// Inspect returns a string-representation of the given object.
func (r *Regexp) Inspect() string {
	return r.Value
}

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (r *Regexp) True() bool {
	return (r.Value != "")
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (r *Regexp) ToInterface() interface{} {
	return r.Value
}
