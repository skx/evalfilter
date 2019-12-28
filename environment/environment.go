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

	// Holder for objects.
	str := make(map[string]object.Object)

	// Holder for variables
	fun := make(map[string]interface{})

	// Create the environment object
	env := &Environment{store: str, functions: fun}

	// Register our default functions.
	env.SetFunction("len", fnLen)
	env.SetFunction("lower", fnLower)
	env.SetFunction("match", fnMatch)
	env.SetFunction("print", fnPrint)
	env.SetFunction("trim", fnTrim)
	env.SetFunction("type", fnType)
	env.SetFunction("upper", fnUpper)
	env.SetFunction("string", fnString)
	env.SetFunction("int", fnInt)
	env.SetFunction("float", fnFloat)

	//
	// These all refer to time.Time fields.
	//
	// (Though they will work on any object which
	// is an integer.  Because when we examine time.Time
	// fields via reflection we convert them to Unix epoch
	// seconds.)
	//

	// 10:11:12, etc.
	env.SetFunction("hour", fnHour)
	env.SetFunction("minute", fnMinute)
	env.SetFunction("seconds", fnSeconds)

	// 10/03/1976, etc.
	env.SetFunction("day", fnDay)
	env.SetFunction("month", fnMonth)
	env.SetFunction("year", fnYear)

	// "Saturday", "Sunday", etc.
	env.SetFunction("weekday", fnWeekday)

	// All done.
	return env
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
