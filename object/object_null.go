package object

// Null wraps nothing and implements our Object interface.
type Null struct{}

// Type returns the type of this object.
func (n *Null) Type() Type {
	return NULL
}

// Inspect returns a string-representation of the given object.
func (n *Null) Inspect() string {
	return "null"
}

// Is this value "true"?
func (n *Null) True() bool {
	return false
}
