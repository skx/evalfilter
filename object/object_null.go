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

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (n *Null) True() bool {
	return false
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (n *Null) ToInterface() interface{} {
	return nil
}

// JSON converts this object to a JSON string.
func (n *Null) JSON() (string, error) {
	return "null", nil
}

// Ensure this object implements the expected interfaces.
var _ JSONAble = &Integer{}
