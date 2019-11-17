// Package environment holds the state for our objects.
package environment

import (
	"github.com/skx/evalfilter/v2/object"
)

// Environment stores our functions, variables, constants, etc.
type Environment struct {
	// store holds variables set by the user-script.
	store map[string]object.Object

	// functions holds golang function pointers, as set by
	// by the host-application.
	functions map[string]interface{}
}

// New creates a new environment, which is used for storing variable
// contents, and available golang functions which have been made available
// to the scripting environment.
func New() *Environment {
	str := make(map[string]object.Object)
	fun := make(map[string]interface{})
	return &Environment{store: str, functions: fun}
}

// Get returns the value of a given variable, by name.
func (e *Environment) Get(name string) (object.Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

// Set stores the value of a variable, by name.
func (e *Environment) Set(name string, val object.Object) object.Object {
	e.store[name] = val
	return val
}

// SetFunction makes a (golang) function available to the scripting
// environment.
func (e *Environment) SetFunction(name string, fun interface{}) interface{} {
	e.functions[name] = fun
	return fun
}

// GetFunction allows a function to be retrieved, by name.
//
// Functions retrieved are only those which have been previously added
// via `SetFunction`.
func (e *Environment) GetFunction(name string) (interface{}, bool) {
	fun, ok := e.functions[name]
	return fun, ok
}
