// Package evalfilter allows you to run simple tests against objects or
// structs implemented in golang, via the use of user-supplied scripts.
//
// Since the result of running tests against objects is a binary
// "yes/no" result it is perfectly suited to working as a filter.
//
// In short this allows you to provide user-customization of your host
// application, but it is explicitly not designed to be a general purpose
// embedded scripting language.
package evalfilter

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/skx/evalfilter/environment"
	"github.com/skx/evalfilter/parser"
	"github.com/skx/evalfilter/runtime"
)

// Evaluator holds our object state.
type Evaluator struct {

	// Bytecode operations are stored here.
	Bytecode []runtime.Operation

	// Environment holds our environment reference
	Env *environment.Environment

	// Parser holds our parser reference
	Parser *parser.Parser
}

// New returns a new evaluation object, which can be used to apply
// the specified script to an object/structure.
func New(input string) *Evaluator {

	// Create a stub object.
	e := &Evaluator{
		Env:    environment.New(),
		Parser: parser.New(input),
	}

	// Add default functions
	e.AddFunction("len",
		func(env *environment.Environment, obj interface{}, args []runtime.Argument) interface{} {

			len := 0

			//
			// Loop over the arguments.
			//
			for _, arg := range args {

				//
				// Get the string.
				//
				str := fmt.Sprintf("%v", arg.Value(env, obj))

				//
				// Add the length.
				//
				len += utf8.RuneCountInString(str)

			}
			return len
		})

	e.AddFunction("trim",
		func(env *environment.Environment, obj interface{}, args []runtime.Argument) interface{} {

			//
			// We loop over the args.
			//
			for _, arg := range args {

				//
				// Get the first, as a string-value
				//
				str := fmt.Sprintf("%v", arg.Value(env, obj))

				//
				// Return the trimmed version
				//
				return (strings.TrimSpace(str))

			}
			return ""
		})

	// Return the configured object.
	return e
}

// AddFunction adds a function to our runtime.
//
// Once a function has been added it may be used by the filter script.
func (e *Evaluator) AddFunction(name string, fun interface{}) {
	e.Env.AddFunction(name, fun)
}

// SetVariable adds, or updates, a variable which will be available
// to the filter script.
//
// Variables in the filter script should be prefixed with `$`,
// for example a variable set as `time` will be accessed via the
// name `$time`.
func (e *Evaluator) SetVariable(name string, value interface{}) {
	e.Env.SetVariable(name, value)
}

// Run executes the user-supplied script against the specified object.
//
// This function can be called multiple times, and doesn't require
// reparsing the script to complete the operation.
func (e *Evaluator) Run(obj interface{}) (bool, error) {

	//
	// Parse the script into operations, unless we've already done so.
	//
	if len(e.Bytecode) == 0 {
		var err error

		e.Bytecode, err = e.Parser.Parse()
		if err != nil {
			return false, err
		}
	}

	//
	// Run the parsed bytecode-operations from our program list,
	// until we hit a return, or the end of the list.
	//
	for _, op := range e.Bytecode {

		//
		// Run the opcode
		//
		ret, val, err := op.Run(e.Env, obj)

		//
		// Should we return?  If so do that.
		//
		if ret {
			return val, err
		}
	}

	//
	// If we reach this point we've processed a script which did not
	// hit a bare return-statement.
	//
	return false, fmt.Errorf("script failed to terminate with a return statement")
}
