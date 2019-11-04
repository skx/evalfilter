package object

// Environment stores our functions, variables, constants, etc.
type Environment struct {
	// store holds variables set by the user-script.
	store map[string]Object

	// functions holds golang function pointers, as set by
	// by the host-application.
	functions map[string]interface{}
}

// NewEnvironment creates new environment
func NewEnvironment() *Environment {
	str := make(map[string]Object)
	fun := make(map[string]interface{})
	return &Environment{store: str, functions: fun}
}

// Get returns the value of a given variable, by name.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

// Set stores the value of a variable, by name.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}

// SetFunction makes a (golang) function available to the script.
func (e *Environment) SetFunction(name string, fun interface{}) interface{} {
	e.functions[name] = fun
	return fun
}

// GetFunction allows a function to be retrieved, by name.
func (e *Environment) GetFunction(name string) (interface{}, bool) {
	fun, ok := e.functions[name]
	return fun, ok
}
