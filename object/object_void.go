package object

// Void wraps nothing and implements our Object interface.
type Void struct{}

// Type returns the type of this object.
func (v *Void) Type() Type {
	return VOID
}

// Inspect returns a string-representation of the given object.
func (v *Void) Inspect() string {
	return "void"
}

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (v *Void) True() bool {
	return false
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (v *Void) ToInterface() interface{} {
	return nil
}
