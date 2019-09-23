package object

// Environment stores our functions, variables, constants, etc.
type Environment struct {
	// store holds variables, including functions.
	store map[string]Object

	// outer holds any parent environment.  Our env. allows
	// nesting to implement scope.
	outer *Environment
}

// NewEnvironment creates new environment
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

// NewEnclosedEnvironment create new environment by outer parameter
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
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
