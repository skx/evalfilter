package object

// ReturnValue wraps Object and implements the Object interface.
type ReturnValue struct {
	// Value is the object that is to be returned
	Value Object
}

// Type returns the type of this object.
func (rv *ReturnValue) Type() Type {
	return RETURN_VALUE_OBJ
}

// Inspect returns a string-representation of the given object.
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}
