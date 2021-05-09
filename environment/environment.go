// Package environment holds the state for our objects.
//
// Storing references to our global-variables and our global list of
// built-in functions is trivial; we just use a map.
//
// To allow variables to persist only for the duration of our
// `foreach` iterator we need to have a notion of "scopes".  For that
// purpose we allow creating "temporary" storage for a particular
// scope.  When we enter a new scope we create a new store which
// shadows the global one.
//
// Leaving a scope means removing that temporary shadowing-store.
//
// As loops can be arbitrarily nested by the user we do not have
// any limit on the number of scopes/nested-states we support.  That's
// why we're using an unbound array, rather than just trying to save
// the previous content(s) of any index/variable at the start of the
// loop - that would have been much simpler to implement.
//
// This might be wrong and buggy, we'll see.  Reference to the
// problem https://github.com/skx/evalfilter/issues/123
package environment

import (
	"fmt"

	"github.com/skx/evalfilter/v2/object"
)

// Environment stores our functions, variables, constants, etc.
type Environment struct {

	// global is the storage for globally-scoped variables.
	global map[string]object.Object

	// local holds variables which are scoped for the
	// duration of `foreach` iterations only.  If we
	// were to support user-defined functions we'd also
	// use local-variables too.
	//
	// We create an entry here each time we enter a new scope,
	// removing it on exit.
	local []map[string]object.Object

	// functions holds golang function pointers, as set by
	// by the host-application.
	//
	// These are largely static, and always global.
	functions map[string]interface{}
}

// New creates a new environment, which is used for storing variable
// contents, and pointers to any golang functions which have been made
// available to the scripting environment by the host application.
func New() *Environment {

	// Holder for variables.
	global := make(map[string]object.Object)

	// Holder for function-pointers for all our builtins.
	functions := make(map[string]interface{})

	// Create the environment object.
	env := &Environment{global: global, functions: functions}

	// Now register our default functions.
	env.SetFunction("between", fnBetween)
	env.SetFunction("float", fnFloat)
	env.SetFunction("getenv", fnGetenv)
	env.SetFunction("int", fnInt)
	env.SetFunction("keys", fnKeys)
	env.SetFunction("len", fnLen)
	env.SetFunction("lower", fnLower)
	env.SetFunction("match", fnMatch)
	env.SetFunction("max", fnMax)
	env.SetFunction("min", fnMin)
	env.SetFunction("now", fnNow)
	env.SetFunction("print", fnPrint)
	env.SetFunction("printf", fnPrintf)
	env.SetFunction("reverse", fnReverse)
	env.SetFunction("sort", fnSort)
	env.SetFunction("split", fnSplit)
	env.SetFunction("sprintf", fnSprintf)
	env.SetFunction("string", fnString)
	env.SetFunction("time", fnNow)
	env.SetFunction("trim", fnTrim)
	env.SetFunction("type", fnType)
	env.SetFunction("upper", fnUpper)

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
//
// This has to test any locally-scoped storage as well as the global
// storage area.  Local values have precedence, and are walked backwards
// because they can be arbitrarily nested.
func (e *Environment) Get(name string) (object.Object, bool) {

	// Look for a local variable
	obj, ok := e.isLocal(name)
	if ok {
		return obj, ok
	}

	// There was no locally-scoped variable.
	//
	// Looking at the global-variable storage.
	//
	obj, ok = e.global[name]
	return obj, ok
}

// Is the variable locally scoped?
//
// This is a bit icky.  On the one hand we know that when a caller
// uses `SetLocal` they mean to set a local variable only, but on
// the other hand if a local-variable is available and the name
// shadows a global we *MUST* prefer the local.
//
// So we've abstracted the test here.  Sigh.
func (e *Environment) isLocal(name string) (object.Object, bool) {

	//
	// The order we search here is very important:
	//
	// We MUST look at the most-recent scopes before the older ones.
	//
	ln := len(e.local)
	for ln > 0 {
		cur := e.local[ln-1]
		obj, ok := cur[name]
		if ok {
			return obj, ok
		}
		ln--
	}
	return nil, false
}

// Set stores the value of a variable, by name.
//
// See also SetLocal for scoped-variables.
func (e *Environment) Set(name string, val object.Object) object.Object {

	//
	// If the variable is locally scoped then we MUST
	// redirect writes to that local variable.
	//
	// Without this things don't work as expected.
	//
	// Life is hard, when you're a scripting engine.
	//
	_, ok := e.isLocal(name)
	if ok {
		e.SetLocal(name, val)
		return val
	}

	//
	// OK we're storing globally.
	//
	e.global[name] = val
	return val
}

// AddScope sets up storage for a new scope, which can store an arbitrary
// number of local variables, these will be mass-discarded in the future
// via `RemoveScope`.
//
// When retrieving variable-contents we iterate over our locals, in the
// order most-recent to least-recent, before checking the global store.
func (e *Environment) AddScope() {

	// Create a new map to hold locally-scoped things.
	locals := make(map[string]object.Object)

	// Add it to our list of scopes.
	e.local = append(e.local, locals)
}

// RemoveScope removes the storage for the most recently added store.
func (e *Environment) RemoveScope() error {

	// Ensure we've got something to reset.
	if len(e.local) > 0 {

		// Remove the last entry and we're done
		e.local = e.local[:len(e.local)-1]
		return nil
	}

	// Otherwise we've found a bug.
	return fmt.Errorf("attempt to RemoveScope when no scopes are present")
}

// SetLocal stores the value of a variable, by name, but only for the local scope.
func (e *Environment) SetLocal(name string, val object.Object) object.Object {

	// If we're not in a wrapped environment that's a bug .. (!)
	if len(e.local) > 0 {

		//
		// We walk upwards to set the variable
		// the scope in which it occurs.
		//
		ln := len(e.local)
		for ln > 0 {
			cur := e.local[ln-1]
			_, ok := cur[name]
			if ok {

				// Found it.
				// Update the value, and return.
				e.local[ln-1][name] = val
				return val
			}
			ln--
		}

		// The variable wasn't found in any parent-scope.
		// Add it on this (bottom) layer.
		cur := e.local[len(e.local)-1]
		cur[name] = val

	}
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

// DeleteFunction allows a function to be disabled by name.
//
// This is used at the moment in our test-cases, however it could be
// used to allow you to disable any existing built-in functions you did
// not wish to expose to your scripting environment.
func (e *Environment) DeleteFunction(name string) {
	delete(e.functions, name)
}
