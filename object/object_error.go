package object

// Error wraps string and implements Object interface.
type Error struct {
	// Message contains the error-message we're wrapping
	Message string
}

// Type returns the type of this object.
func (e *Error) Type() Type {
	return ERROR
}

// Inspect returns a string-representation of the given object.
func (e *Error) Inspect() string {
	return "ERROR: " + e.Message
}

// Is this value "true"?
func (e *Error) True() bool {
	return false
}
