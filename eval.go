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

	"github.com/skx/evalfilter/ast"
	"github.com/skx/evalfilter/code"
	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/object"
	"github.com/skx/evalfilter/parser"
	"github.com/skx/evalfilter/vm"
)

// Eval is our public-facing structure which stores our state.
type Eval struct {
	// Script holds the script the user submitted in our constructor.
	Script string

	// Environment
	environment *object.Environment

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
		environment: object.NewEnvironment(),
		Script:      script,
	}

	//
	// Add our default functions.
	//
	e.AddFunction("len", fnLen)
	e.AddFunction("lower", fnLower)
	e.AddFunction("match", fnMatch)
	e.AddFunction("print", fnPrint)
	e.AddFunction("trim", fnTrim)
	e.AddFunction("type", fnType)
	e.AddFunction("upper", fnUpper)

	//
	// Return it.
	//
	return e
}

// Prepare is the second function the caller must invoke, it compiles
// the user-supplied program to its final-form.
func (e *Eval) Prepare() error {

	//
	// Create a parser
	//
	p := parser.New(lexer.New(e.Script))

	//
	// Parse the program.
	//
	program := p.ParseProgram()

	//
	// Where there any errors produced by the parser?
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
	err := e.Compile(program)

	//
	// If there were errors then return them.
	//
	if err != nil {
		return err
	}

	//
	// Otherwise construct a VM and save it.
	//
	e.machine = vm.New(e.constants, e.instructions, e.environment)

	//
	// All done.
	return nil
}

// Dump causes our bytecode to be dumped
func (e *Eval) Dump() error {

	i := 0
	fmt.Printf("Bytecode:\n")

	for i < len(e.instructions) {

		// opcode
		op := e.instructions[i]

		// as a string
		str := code.String(code.Opcode(op))

		fmt.Printf("%06d\t%s\t", i, str)

		// show arg
		if op < byte(code.OpCodeSingleArg) {

			arg := code.ReadUint16(e.instructions[i+1:])
			fmt.Printf("%d", arg)

			i += 2

			//
			// Show the values, as comments, to make the
			// bytecode more human-readable.
			//
			if code.Opcode(op) == code.OpConstant {
				fmt.Printf("\t\t// load constant: %v", e.constants[arg])
			}
			if code.Opcode(op) == code.OpLookup {
				fmt.Printf("\t\t// lookup field: %v", e.constants[arg])
			}
			if code.Opcode(op) == code.OpCall {
				fmt.Printf("\t\t\t// call function with %d arguments", arg)
			}
		}

		fmt.Printf("\n")

		i++
	}

	// constants
	fmt.Printf("\n\nConstants:\n")
	for i, n := range e.constants {
		fmt.Printf("%d - %v\n", i, n)
	}
	return nil
}

// Run takes the program which was passed in the constructor, and
// executes it.  The supplied object will be used for performing
// dynamic field-lookups, etc.
func (e *Eval) Run(obj interface{}) (bool, error) {

	//
	// Launch the program in the VM.
	//
	out, err := e.machine.Run(obj)

	//
	// Show the result.
	//
	//	fmt.Printf("Result: %v\n", out)

	//
	// Is the return-value an error?  If so report that.
	//
	if err != nil {
		return false, err
	}
	if out.Type() == object.ERROR_OBJ {
		return false, fmt.Errorf("%s", out.Inspect())
	}

	//
	// Otherwise convert the result to a boolean, and return it.
	//
	return e.isTruthy(out), err

}

// AddFunction adds a function to our runtime.
//
// Once a function has been added it may be used by the filter script.
func (e *Eval) AddFunction(name string, fun interface{}) {
	e.environment.SetFunction(name, fun)
}

// SetVariable adds, or updates, a variable which will be available
// to the filter script.
func (e *Eval) SetVariable(name string, value object.Object) {
	e.environment.Set(name, value)
}

// GetVariable retrieves the contents of a variable which has been
// set within a user-script.
//
// If the variable hasn't been set then the null-value will be returned
func (e *Eval) GetVariable(name string) object.Object {
	value, ok := e.environment.Get(name)
	if ok {
		return value
	}
	return &object.Null{}
}

// Compile is core-code for converting the AST into a series of bytecodes.
func (e *Eval) Compile(node ast.Node) error {

	switch node := node.(type) {

	case *ast.Program:
		for _, s := range node.Statements {
			err := e.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := e.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.Boolean:
		if node.Value {
			e.emit(code.OpTrue)
		} else {
			e.emit(code.OpFalse)
		}

	case *ast.FloatLiteral:
		str := &object.Float{Value: node.Value}
		e.emit(code.OpConstant, e.addConstant(str))

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		e.emit(code.OpConstant, e.addConstant(integer))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		e.emit(code.OpConstant, e.addConstant(str))

	case *ast.ReturnStatement:
		err := e.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		e.emit(code.OpReturn)

	case *ast.ExpressionStatement:
		err := e.Compile(node.Expression)
		if err != nil {
			return err
		}

	case *ast.InfixExpression:
		err := e.Compile(node.Left)
		if err != nil {
			return err
		}

		err = e.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			e.emit(code.OpAdd)
		case "-":
			e.emit(code.OpSub)
		case "*":
			e.emit(code.OpMul)
		case "/":
			e.emit(code.OpDiv)
		case "%":
			e.emit(code.OpMod)
		case "**":
			e.emit(code.OpPower)
		case "<":
			e.emit(code.OpLess)
		case "<=":
			e.emit(code.OpLessEqual)
		case ">":
			e.emit(code.OpGreater)
		case ">=":
			e.emit(code.OpGreaterEqual)
		case "==":
			e.emit(code.OpEqual)
		case "!=":
			e.emit(code.OpNotEqual)
		case "~=":
			e.emit(code.OpMatches)
		case "!~":
			e.emit(code.OpNotMatches)
		case "&&":
			e.emit(code.OpAnd)
		case "||":
			e.emit(code.OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		err := e.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			e.emit(code.OpBang)
		case "-":
			e.emit(code.OpMinus)
		case "√":
			e.emit(code.OpRoot)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IfExpression:
		err := e.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit an `OpJumpIfFalse` with a bogus value
		jumpNotTruthyPos := e.emit(code.OpJumpIfFalse, 9999)

		err = e.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// Emit an `OpJump` with a bogus value
		jumpPos := e.emit(code.OpJump, 9999)

		afterConsequencePos := len(e.instructions)
		e.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative != nil {
			err := e.Compile(node.Alternative)
			if err != nil {
				return err
			}

		}

		afterAlternativePos := len(e.instructions)
		e.changeOperand(jumpPos, afterAlternativePos)

	case *ast.AssignStatement:

		// Get the value
		err := e.Compile(node.Value)
		if err != nil {
			return err
		}

		// Store the name
		str := &object.String{Value: node.Name.String()}
		e.emit(code.OpConstant, e.addConstant(str))

		// And make it work.
		e.emit(code.OpSet)

	case *ast.Identifier:
		str := &object.String{Value: node.Value}
		e.emit(code.OpLookup, e.addConstant(str))

	case *ast.CallExpression:

		//
		// call to print(1) will have the stack setup as:
		//
		//  1
		//  print
		//  call 1
		//
		// call to print( "steve", "kemp" ) will have:
		//
		//  "steve"
		//  "kemp"
		//  "print"
		//  call 2
		//
		args := len(node.Arguments)
		for _, a := range node.Arguments {

			err := e.Compile(a)
			if err != nil {
				return err
			}
		}

		// call - has the string on the stack
		str := &object.String{Value: node.Function.String()}
		e.emit(code.OpConstant, e.addConstant(str))

		// then a call instruction with the number of args.
		e.emit(code.OpCall, args)

	default:
		return fmt.Errorf("unknown node type %T %v", node, node)
	}
	return nil
}

// addConstant adds a constant to the pool
func (e *Eval) addConstant(obj object.Object) int {

	//
	// Look to see if the constant is present already
	//
	for i, c := range e.constants {

		//
		// If the existing constant has the same
		// type and value - then return the offset.
		//
		if c.Type() == obj.Type() &&
			c.Inspect() == obj.Inspect() {
			return i
		}
	}

	//
	// Otherwise this is a distinct constant and should
	// be added.
	//
	e.constants = append(e.constants, obj)
	return len(e.constants) - 1
}

// emit generates a bytecode operation, and adds it to our program-array.
func (e *Eval) emit(op code.Opcode, operands ...int) int {

	ins := make([]byte, 1)
	ins[0] = byte(op)

	if len(operands) == 1 {

		// Make a buffer for the arg
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(operands[0]))

		// append
		ins = append(ins, b...)
	}

	posNewInstruction := len(e.instructions)
	e.instructions = append(e.instructions, ins...)

	return posNewInstruction
}

// changeOperand is designed to patch the operand of
// and instruction.  It is basically used to rewrite the target
// of our jump instructions in the handling of `if`.
func (e *Eval) changeOperand(opPos int, operand int) {

	// get the opcode
	op := code.Opcode(e.instructions[opPos])

	// make a new buffer for the opcode
	ins := make([]byte, 1)
	ins[0] = byte(op)

	// Make a buffer for the arg
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(operand))

	// append argument
	ins = append(ins, b...)

	// replace
	e.replaceInstruction(opPos, ins)
}

// replaceInstruction rewrites the instruction at the given
// bytecode position.
func (e *Eval) replaceInstruction(pos int, newInstruction []byte) {
	ins := e.instructions

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

// isTruthy tests whether the given object is "true".
func (e *Eval) isTruthy(obj object.Object) bool {

	//
	// Is this a boolean object?
	//
	// If so look for the stored value.
	//
	switch tmp := obj.(type) {
	case *object.Boolean:
		return tmp.Value
	case *object.String:
		return (tmp.Value != "")
	case *object.Null:
		return false
	case *object.Integer:
		return (tmp.Value != 0)
	case *object.Float:
		return (tmp.Value != 0.0)
	default:
		return true
	}
}
