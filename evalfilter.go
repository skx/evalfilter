// Package evalfilter allows running a user-supplied script against an object.
//
// We're constructed with a program, and internally we parse that to an
// abstract syntax-tree, then we walk that tree to generate a series of
// bytecodes.
//
// The bytecode is then executed via the VM-package.
package evalfilter

import (
	"context"
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

	// context for handling timeout
	context context.Context

	// user-defined functions
	functions map[string]environment.UserFunction
}

// New creates a new instance of the evaluator.
func New(script string) *Eval {

	//
	// Create our object.
	//
	e := &Eval{
		environment: environment.New(),
		Script:      script,
		context:     context.Background(),
		functions:   make(map[string]environment.UserFunction),
	}

	//
	// Return it.
	//
	return e
}

// SetContext allows a context to be passed to the evaluator.
//
// The context is passed down to the virtual machine, which allows you to
// setup a timeout/deadline for the execution of user-supplied scripts.
func (e *Eval) SetContext(ctx context.Context) {
	e.context = ctx
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
	// If we've got the optimizer enabled then set the environment
	// variable, so that the virtual machine knows it should
	// run a series of optimizations.
	//
	if optimize {
		e.environment.Set("OPTIMIZE", &object.Boolean{Value: true})
	}

	//
	// Now we're done, construct a VM with the bytecode and constants
	// we've created - as well as any function pointers and variables
	// which we were given.
	//
	// The optimization will happen at this step, so that it is complete
	// before Execute/Run are invoked - and we only take the speed hit
	// once.
	e.machine = vm.New(e.constants, e.instructions, e.functions, e.environment)

	//
	// Setup our context
	//
	e.machine.SetContext(e.context)

	//
	// All done; no errors.
	//
	return nil
}

// dumper is the callback function which is invoked for dumping bytecode
func (e *Eval) dumper(offset int, opCode code.Opcode, opArg interface{}) (bool, error) {

	// Show the offset + instruction.
	fmt.Printf("  %04d\t%14s", offset, code.String(opCode))

	// Show the optional argument, if present.
	if opArg != nil {
		fmt.Printf("\t% 4d", opArg.(int))
	}

	// Some opcodes benefit from inline comments
	if code.Opcode(opCode) == code.OpConstant {
		v := e.constants[opArg.(int)]
		s := strings.ReplaceAll(v.Inspect(), "\n", "\\n")
		s = strings.ReplaceAll(s, "\r", "\\r")
		s = strings.ReplaceAll(s, "\t", "\\t")
		fmt.Printf("\t// push constant onto stack: \"%s\"", s)
	}
	if code.Opcode(opCode) == code.OpLookup {
		v := e.constants[opArg.(int)]
		s := strings.ReplaceAll(v.Inspect(), "\n", "\\n")
		s = strings.ReplaceAll(s, "\r", "\\r")
		s = strings.ReplaceAll(s, "\t", "\\t")
		fmt.Printf("\t// lookup field/variable: %s", s)
	}
	if code.Opcode(opCode) == code.OpCall {
		fmt.Printf("\t// call function with %d arg(s)", opArg.(int))
	}
	if code.Opcode(opCode) == code.OpPush {
		fmt.Printf("\t// Push %d to stack", opArg.(int))
	}
	fmt.Printf("\n")

	// Keep walking, no error.
	return true, nil
}

// Dump causes our bytecode to be dumped, along with the contents
// of the constant-pool
func (e *Eval) Dump() error {

	fmt.Printf("Bytecode:\n")

	// Use the walker to dump the bytecode.
	e.machine.WalkBytecode(e.dumper)

	// Show constants, if any are present.
	consts := e.constants
	if len(consts) > 0 {
		fmt.Printf("\n\nConstant Pool:\n")
		for i, n := range consts {
			s := strings.ReplaceAll(n.Inspect(), "\n", "\\n")
			fmt.Printf("  %04d Type:%s Value:\"%s\"\n", i, n.Type(), s)
		}
	}

	// Do we have user-defined functions?
	funs := e.functions
	if len(funs) > 0 {
		fmt.Printf("\nUser-defined functions:\n")
	}

	// For each function
	count := 0
	for name, obj := range funs {
		// Show brief information
		fmt.Printf(" function %s(%s)\n", name, strings.Join(obj.Arguments, ","))

		// Then dump the body.
		e.machine.WalkFunctionBytecode(name, e.dumper)

		// Put a newline between functions.
		if count < len(e.functions)-1 {
			fmt.Printf("\n")
		}
		count++
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
