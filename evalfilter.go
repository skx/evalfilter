// Package evalfilter allows running a user-supplied script against an object.
//
// We're constructed with a program, and internally we parse that to an
// abstract syntax-tree, then we walk that tree to generate a series of
// bytecodes.
//
// The bytecode is then executed via the VM-package.
package evalfilter

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/skx/evalfilter/v2/code"
	"github.com/skx/evalfilter/v2/environment"
	"github.com/skx/evalfilter/v2/lexer"
	"github.com/skx/evalfilter/v2/object"
	"github.com/skx/evalfilter/v2/parser"
	"github.com/skx/evalfilter/v2/vm"
)

// Flags which can be optionally passed to Prepare.
const (
	// Don't run the optimizer when generating bytecode.
	NoOptimize byte = iota
)

// Eval is our public-facing structure which stores our state.
type Eval struct {
	// Script holds the script the user submitted in our constructor.
	Script string

	// Environment
	environment *environment.Environment

	// constants compiled
	constants []object.Object

	// bytecode we generate
	instructions code.Instructions

	// the machine we drive
	machine *vm.VM
}

// New creates a new instance of the evaluator.
func New(script string) *Eval {

	//
	// Create our object.
	//
	e := &Eval{
		environment: environment.New(),
		Script:      script,
	}

	//
	// Return it.
	//
	return e
}

// Prepare is the second function the caller must invoke, it compiles
// the user-supplied program to its final-form.
//
// Internally this compilation process walks through the usual steps,
// lexing, parsing, and bytecode-compilation.
func (e *Eval) Prepare(flags ...[]byte) error {

	//
	// Default to optimizing the bytecode.
	//
	optimize := true

	//
	// But let flags change our behaviour.
	//
	for _, arg := range flags {
		for _, val := range arg {
			if val == NoOptimize {
				optimize = false
			}
		}
	}

	//
	// Create a lexer.
	//
	l := lexer.New(e.Script)

	//
	// Create a parser using the lexer.
	//
	p := parser.New(l)

	//
	// Parse the program into an AST.
	//
	program := p.ParseProgram()

	//
	// Were there any errors produced by the parser?
	//
	// If so report that.
	//
	if len(p.Errors()) > 0 {
		return fmt.Errorf("\nErrors parsing script:\n" +
			strings.Join(p.Errors(), "\n"))
	}

	//
	// Compile the program to bytecode
	//
	err := e.compile(program)

	//
	// If there were errors then return them.
	//
	if err != nil {
		return err
	}

	//
	// Attempt to optimize the code, running multiple passes until no
	// more changes are possible.
	//
	// We do this so that each optimizer run only has to try one thing
	// at a time.
	//
	if optimize {
		e.optimize()
	}

	//
	// Now we're done, construct a VM with the bytecode and constants
	// we've created - as well as any function pointers and variables
	// which we were given.
	//
	e.machine = vm.New(e.constants, e.instructions, e.environment)

	//
	// All done; no errors.
	//
	return nil
}

// Dump causes our bytecode to be dumped.
//
// This is used by the `evalfilter` CLI-utility, but it might be useful
// to consumers of our library.
func (e *Eval) Dump() error {

	i := 0
	fmt.Printf("Bytecode:\n")

	for i < len(e.instructions) {

		// opcode
		op := e.instructions[i]

		// opcode length
		opLen := code.Length(code.Opcode(op))

		// opcode as a string
		str := code.String(code.Opcode(op))

		fmt.Printf("  %06d\t%14s", i, str)

		// show arg
		if op < byte(code.OpCodeSingleArg) {

			arg := binary.BigEndian.Uint16(e.instructions[i+1 : i+3])
			fmt.Printf("\t%d", arg)

			//
			// Show the values, as comments, to make the
			// bytecode more human-readable.
			//
			if code.Opcode(op) == code.OpConstant {

				v := e.constants[arg]
				s := strings.ReplaceAll(v.Inspect(), "\n", "\\n")

				fmt.Printf("\t// load constant: \"%s\"", s)
			}
			if code.Opcode(op) == code.OpLookup {
				fmt.Printf("\t// lookup field: %v", e.constants[arg])
			}
			if code.Opcode(op) == code.OpCall {
				fmt.Printf("\t// call function with %d arg(s)", arg)
			}
		}

		fmt.Printf("\n")

		i += opLen
	}

	// Show constants, if any are present.
	if len(e.constants) > 0 {
		fmt.Printf("\n\nConstants:\n")
		for i, n := range e.constants {

			s := strings.ReplaceAll(n.Inspect(), "\n", "\\n")

			fmt.Printf("  %06d Type:%s Value:\"%s\"\n", i, n.Type(), s)
		}
	}

	return nil
}

// Execute executes the program which the user passed in the constructor,
// and returns the object that the script finished with.
//
// This function is very similar to the `Run` method, however the Run
// method only returns a binary/boolean result, and this method returns
// the actual object your script returned with.
//
// Use of this method allows you to receive the `3` that a script
// such as `return 1 + 2;` would return.
func (e *Eval) Execute(obj interface{}) (object.Object, error) {

	//
	// Launch the program in the VM.
	//
	out, err := e.machine.Run(obj)

	//
	// Error executing?  Report that.
	//
	if err != nil {
		return &object.Null{}, err
	}

	//
	// Return the resulting object.
	//
	return out, nil
}

// Run executes the program which the user passed in the constructor.
//
// The return value, assuming no error, is a binary/boolean result which
// suits the use of this package as a filter.
//
// If you wish to return the actual value the script returned then you can
// use the `Execute` method instead.  That doesn't attempt to determine whether
// the result of the script was "true" or not.
func (e *Eval) Run(obj interface{}) (bool, error) {

	//
	// Execute the script, getting the resulting error
	// and return object.
	//
	out, err := e.Execute(obj)

	//
	// Error? Then return that.
	//
	if err != nil {
		return false, err
	}

	//
	// Otherwise case the resulting object into
	// a boolean and pass that back to the caller.
	//
	return out.True(), nil
}

// AddFunction exposes a golang function from your host application
// to the scripting environment.
//
// Once a function has been added it may be used by the filter script.
func (e *Eval) AddFunction(name string, fun interface{}) {
	e.environment.SetFunction(name, fun)
}

// SetVariable adds, or updates a variable which will be available
// to the filter script.
func (e *Eval) SetVariable(name string, value object.Object) {
	e.environment.Set(name, value)
}

// GetVariable retrieves the contents of a variable which has been
// set within a user-script.
//
// If the variable hasn't been set then the null-value will be returned.
func (e *Eval) GetVariable(name string) object.Object {
	value, ok := e.environment.Get(name)
	if ok {
		return value
	}
	return &object.Null{}
}
