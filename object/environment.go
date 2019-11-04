package object

// Environment stores our functions, variables, constants, etc.
type Environment struct {
	// store holds variables set by the user-script.
	store map[string]Object

	// functions holds golang function pointers, as set by
	// by the host-application.
	functions map[string]interface{}

	// outer holds any parent environment.  Our env. allows
	// nesting to implement scope.
	outer *Environment
}

// NewEnvironment creates new environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	f := make(map[string]interface{})
	return &Environment{store: s, functions: f, outer: nil}
}

// Get returns the value of a given variable, by name.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set stores the value of a variable, by name.
func (e *Environment) Set(name string, val Object) Object {
	//
	// Store the (updated) value.
	//
	e.store[name] = val
	return val
}

// SetFunction makes a (golang) function available to the script.
func (e *Environment) SetFunction(name string, fun interface{}) {
	e.functions[name] = fun
}

// GetFunction allows a function to be retrieved, by name.
func (e *Environment) GetFunction(name string) interface{} {
	return (e.functions[name])
}
