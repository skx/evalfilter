// This file contains the implementation for the eval operation, which
// is not explicitly named.
//
// Given the script:
//
//  print "hello\n";
//  foo();
//
// The `foo` function call is evaluated and executed, via an instance of
// this object.

package runtime

import "github.com/skx/evalfilter/environment"

// EvalOperation holds state for the evaluation of a function-call.
type EvalOperation struct {
	// Value holds the function object to be evaluated, including arguments
	// which will be passed to that function.
	Value Argument
}

// Run runs the eval operation, the result is discarded.
func (eo *EvalOperation) Run(env *environment.Environment, obj interface{}) (bool, bool, error) {

	// Here we make the call, by evaluating the result.
	eo.Value.(*FunctionArgument).Value(env, obj)
	return false, false, nil
}
