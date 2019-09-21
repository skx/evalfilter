// Package environment contains a run-time environment for our
// evaluation-engine.
//
// The environment contains any functions and variables which have
// been set by the host application
package environment

// Environment is the runtime object which holds references
// to functions/variables set by the host application.
type Environment struct {

	// Functions contains references to helper functions which
	// have been made available to the user-script and which are
	// implemented outside this package in the golang host-application.
	Functions map[string]interface{}

	// Variables contains references to variables set via
	// the golang host-application.
	Variables map[string]interface{}
}

// New creates a new (empty) environment.
func New() *Environment {

	// Create a stub object.
	return &Environment{
		Functions: make(map[string]interface{}),
		Variables: make(map[string]interface{}),
	}
}

// AddFunction adds a function to our runtime.
//
// Once a function has been added it may be used by the filter script.
func (e *Environment) AddFunction(name string, fun interface{}) {
	e.Functions[name] = fun
}

// SetVariable adds, or updates, a variable which will be available
// to the filter script.
//
// Variables in the filter script should be prefixed with `$`,
// for example a variable set as `time` will be accessed via the
// name `$time`.
func (e *Environment) SetVariable(name string, value interface{}) {
	e.Variables[name] = value
}
