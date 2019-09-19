// Package evalfilter allows you to run simple  tests against objects or
// structs implemented in golang, via the use of user-supplied scripts
//
// Since the result of running tests against objects is a boolean/binary
// "yes/no" result it is perfectly suited to working as a filter.
package evalfilter

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Evaluator holds our object state
type Evaluator struct {
	// Debug is a flag which is used to indicate whether we perform
	// some minimal tracing to STDOUT during the course of our script
	// execution.
	Debug bool

	// Program is the script the user wishes to run.
	Program string

	// Functions contains references to helper functions which
	// have been made available to the user-script.
	Functions map[string]interface{}

	// Bytecode operations are stored here
	Bytecode []Operation
}

// Argument is an abstract argument-type.
//
// Our `if` operation applies an operator to a pair of operands.  The operands
// might be field-references, strings, numbers, or the result of function-calls.
//
// The differences are abstracted by this interface.
type Argument interface {

	// Value returns the value of the argument.
	//
	// Which might require use of the object.
	Value(self *Evaluator, obj interface{}) interface{}
}

// BooleanArgument holds a literal boolean value.
type BooleanArgument struct {
	// Content holds the value.
	Content bool
}

// Value returns the boolean content we're wrapping.
func (s *BooleanArgument) Value(self *Evaluator, obj interface{}) interface{} {
	return s.Content
}

// StringArgument holds a literal string.
type StringArgument struct {
	// Content holds the string literal.
	Content string
}

// Value returns the string content we're wrapping.
func (s *StringArgument) Value(self *Evaluator, obj interface{}) interface{} {
	return s.Content
}

// FieldArgument holds a reference to an object's field value.
type FieldArgument struct {
	// Field the name of the structure/object field we return.
	Field string
}

// Value returns the value of the field from the specified object.
func (f *FieldArgument) Value(self *Evaluator, obj interface{}) interface{} {

	ref := reflect.ValueOf(obj)
	field := reflect.Indirect(ref).FieldByName(f.Field)

	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		return field.Int()
	case reflect.Float32, reflect.Float64:
		return field.Float()
	case reflect.String:
		return field.String()
	case reflect.Bool:
		if field.Bool() {
			return "true"
		}
		return "false"
	}
	return nil
}

// FunctionArgument holds a reference to a function invokation.
type FunctionArgument struct {
	// Name of the function to invoke
	Function string

	// Optional arguments to function.
	Arguments []Argument
}

// Value returns the result of calling the function we're wrapping.
func (f *FunctionArgument) Value(self *Evaluator, obj interface{}) interface{} {
	res, ok := self.Functions[f.Function]
	if !ok {
		fmt.Printf("Unknown function: %s\n", f.Function)
		os.Exit(1)
	}

	//
	// Are we running with debugging?
	//
	if self.Debug {
		fmt.Printf("Calling function: %s\n", f.Function)
	}

	out := res.(func(eval *Evaluator, obj interface{}, args ...interface{}) interface{})

	//
	// Call the function.
	//
	ret := (out(self, obj, f.Arguments))

	//
	// Log the result?
	//
	if self.Debug {
		fmt.Printf("\tReturn: %v\n", ret)
	}

	return ret

}

//
// Bytecode functions
//

// Operation is the abstract interface all operations much implement.
type Operation interface {

	// Run runs the operation.
	//
	// Return values:
	//   return - If true we're returning
	//
	//   value  - The value we terminate with
	//
	//   error  - An error occurred
	//
	Run(self *Evaluator, obj interface{}) (bool, bool, error)
}

// IfOperation holds state for the `if` operation
type IfOperation struct {
	// Left argument
	Left Argument

	// Right argument.
	Right Argument

	// Test-operation
	Op string

	// Operations to be carried out if the statement matches.
	True []Operation
}

// Run executes an if statement.
func (i *IfOperation) Run(e *Evaluator, obj interface{}) (bool, bool, error) {

	// Run the if-statement.
	res, err := i.doesMatch(e, obj)

	// Was there an error?
	if err != nil {
		return false, false, fmt.Errorf("failed to run if-test %s", err)
	}

	//
	// No error - and we got a match.
	//
	if res {

		//
		// The test matches so we should now handle
		// all the things that are in the `true`
		// list.
		//
		for _, t := range i.True {

			//
			// Process each operation.
			//
			// If this was a return statement then we return
			//
			ret, val, err := t.Run(e, obj)
			if ret {
				return ret, val, err
			}

		}

		//
		// At this point we've matched, and we've run
		// the statements in the block.
		//
		return false, false, nil
	}

	return false, false, nil
}

// doesMatch runs the actual comparision for an if statement
//
// We return "true" if the statement matched, and the return should
// be executed.  Otherwise we return false.
func (i *IfOperation) doesMatch(e *Evaluator, obj interface{}) (bool, error) {

	if e.Debug {
		fmt.Printf("IF %v %s %v;\n", i.Left.Value(e, obj), i.Op, i.Right.Value(e, obj))

	}

	//
	// Expand the left & right sides of the conditional
	//
	lVal := i.Left.Value(e, obj)
	rVal := i.Right.Value(e, obj)

	//
	// Convert to strings, in case they're needed for the early
	// operations.
	//
	lStr := fmt.Sprintf("%v", lVal)
	rStr := fmt.Sprintf("%v", rVal)

	//
	// Basic operations
	//

	// Equality - string and number.
	if i.Op == "==" {
		return (lStr == rStr), nil
	}

	// Inequality - string and number.
	if i.Op == "!=" {
		return (lStr != rStr), nil
	}

	// String-contains
	if i.Op == "~=" {
		return strings.Contains(lStr, rStr), nil
	}

	// String does not contain
	if i.Op == "!~" {
		return !strings.Contains(lStr, rStr), nil
	}

	//
	// All remaining operations are numeric, so we need to convert
	// the values into numbers.
	//
	// Call them `a` and `b`.
	//
	var a float64
	var b float64
	var err error

	//
	// Convert
	//
	a, err = i.toNumberArg(lVal)
	if err != nil {
		return false, err
	}
	b, err = i.toNumberArg(rVal)
	if err != nil {
		return false, err
	}

	//
	// Now operate.
	//
	if i.Op == ">" {
		return (a > b), nil
	}
	if i.Op == ">=" {
		return (a >= b), nil
	}
	if i.Op == "<" {
		return (a < b), nil
	}
	if i.Op == "<=" {
		return (a <= b), nil
	}

	//
	// Invalid operator?
	//
	return false, fmt.Errorf("unknown operator %v", i.Op)
}

// toNumberArg tries to convert the given interface to a float64 value.
func (i *IfOperation) toNumberArg(value interface{}) (float64, error) {

	// string?
	_, ok := value.(string)
	if ok {
		a, _ := strconv.ParseFloat(value.(string), 32)
		return a, nil
	}

	// int
	_, ok = value.(int)
	if ok {
		return (float64(value.(int))), nil
	}

	// float?
	_, ok = value.(int64)
	if ok {
		return (float64(value.(int64))), nil
	}

	return 0, fmt.Errorf("failed to convert %v to number", value)
}

// ReturnOperation holds state for the `return` operation
type ReturnOperation struct {
	// Value holds the value which will be returned.
	Value bool
}

// Run handles the return operation.
func (r *ReturnOperation) Run(e *Evaluator, obj interface{}) (bool, bool, error) {
	return true, r.Value, nil
}

// PrintOperation holds state for the `print` operation.
type PrintOperation struct {
	// Values are the various values to be printed.
	Values []Argument
}

// Run runs the print operation.
func (p *PrintOperation) Run(e *Evaluator, obj interface{}) (bool, bool, error) {
	for _, val := range p.Values {
		fmt.Printf("%v", val.Value(e, obj))
	}
	return false, false, nil
}

// New returns a new evaluation object, which can be used to apply
// the specified script to an object/structure.
func New(input string) *Evaluator {
	e := &Evaluator{
		Debug:     false,
		Functions: make(map[string]interface{}),
		Program:   input,
	}

	// Add default functions
	e.AddFunction("len",
		func(eval *Evaluator, obj interface{}, args ...interface{}) interface{} {

			//
			// Each argument is an array of args.
			//
			for _, arg := range args {

				//
				// The args
				//
				for _, n := range arg.([]Argument) {

					//
					// Get the first
					//
					str := n.Value(eval, obj)

					//
					// Return the length
					//
					return (utf8.RuneCountInString(fmt.Sprintf("%v", str)))

				}
			}
			return 0
		})

	e.AddFunction("trim",
		func(eval *Evaluator, obj interface{}, args ...interface{}) interface{} {

			//
			// Each argument is an array of args.
			//
			for _, arg := range args {

				//
				// The args
				//
				for _, n := range arg.([]Argument) {

					//
					// Get the first
					//
					str := n.Value(eval, obj)

					//
					// Return the trimmed version
					//
					return (strings.TrimSpace(fmt.Sprintf("%v", str)))

				}
			}
			return 0
		})

	// The environmental variable ${EVAL_FILTER_DEBUG} enables
	// the use of tracing.
	if os.Getenv("EVAL_FILTER_DEBUG") != "" {
		e.Debug = true
	}
	return e
}

// AddFunction adds a function to our runtime.
func (e *Evaluator) AddFunction(name string, fun interface{}) {
	e.Functions[name] = fun
}

// Look at the given token, and parse it as an operation.
//
// This is abstracted into a routine of its own so that we can
// either parse the stream of tokens for the full-script, or parse
// the block which is used inside an IF statement.
func (e *Evaluator) parseOperation(tok Token, l *Lexer) (Operation, error) {

	switch tok.Type {

	//
	// `if`
	//
	case IF:
		return e.parseIF(l)

	//
	// `return`
	//
	case RETURN:

		// Get the value this token returns
		val := l.NextToken()

		// The next token should be a semi-colon
		tmp := l.NextToken()
		if tmp.Type != SEMICOLON {
			return nil, fmt.Errorf("expected ';' after return-value")

		}
		// Update our bytecode
		return &ReturnOperation{Value: val.Literal == "true"}, nil

	//
	// `print`
	//
	case PRINT:

		//
		// Here are the arguments we're going to be printing.
		//
		var tmp []Argument

		for {
			//
			// We keep printing output until we hit
			// a semi-colon, or the end of the file.
			//
			n := l.NextToken()
			if n.Type == SEMICOLON || n.Type == EOF {
				break
			}

			//
			// Skip over any commas
			//
			if n.Type == COMMA {
				continue
			}

			//
			// Convert the token to an argument.
			//
			obj := e.tokenToArgument(n, l)

			//
			// Add it to our list.
			//
			tmp = append(tmp, obj)

		}

		//
		// Now record the print operation.
		//
		return &PrintOperation{Values: tmp}, nil

	}
	return nil, fmt.Errorf("failed to parse token type %s : %s", tok.Type, tok)
}

// parse is an internal method which reads the script we've been
// given in our constructor and writes out a series of operations
// to be carried out to our `Bytecode` array.
//
// This is simple because we have no control-flow, and no need to
// worry about nested-blocks, variables, etc.
func (e *Evaluator) parse() error {

	//
	// Create a lexer to process our script.
	//
	l := NewLexer(e.Program)

	//
	// Process all the tokens forever, until we hit the end of file.
	//
	tok := l.NextToken()

	for tok.Type != EOF {

		//
		// Parse the next statement.
		//
		op, err := e.parseOperation(tok, l)
		if err != nil {
			return err
		}

		//
		// Append it to our list.
		//
		e.Bytecode = append(e.Bytecode, op)

		//
		// Proceed onto the next token.
		//
		tok = l.NextToken()
	}

	//
	// Parsed with no error.
	//
	return nil
}

// parseIf is our biggest method; it parses an if-expression.
func (e *Evaluator) parseIF(l *Lexer) (Operation, error) {

	//
	// The general form is:
	//
	//  IF ( LEFT TEST RIGHT ) { RETURN YY; }
	//
	// e.g. "if ( Count == 3 ) { return true; }"
	//
	// However there is a second form which is designed for the use
	// of functions:
	//
	//   IF ( function() ) ..
	//
	// We tell them apart by looking at the tokens we receive.
	//
	var left Argument
	var right Argument
	var op string

	//
	// skip the (
	//
	skip := l.NextToken()
	if skip.Literal != "(" {
		return &IfOperation{}, fmt.Errorf("expected '(' got %v", skip)
	}

	//
	// Get the first operand.
	//
	t := l.NextToken()
	left = e.tokenToArgument(t, l)

	//
	// Get the operator.
	//
	t = l.NextToken()
	op = t.Literal

	//
	// In the general case we'd have:
	//
	//   IF ( LEFT OP RIGHT )
	//
	// But remember we also allow:
	//
	//   IF ( FUNCTION() )
	//
	// If we've been given the second form our `op` token will be `)`,
	// because the `OP` & `RIGHT` tokens will not be present.
	//
	// If that is the case we fake values.
	//
	if op == ")" {

		//
		// To avoid making changes we simply
		// FAKE the other arguments, because
		// saying "if ( foo() )" is logically
		// the same as saying:
		//
		//  if ( foo() == "true" )
		//

		//
		// Fake the operation.
		//
		op = "=="

		//
		// Fake the right-value.
		//
		// NB: This works because we force user-added
		// functions to return boolean values.
		//
		right = &StringArgument{Content: "true"}

		//
		// I feel bad.  But not that bad.
		//
		goto block
	}

	//
	// OK we're in the three-argument form, so we
	// get the right operand.
	//
	t = l.NextToken()
	right = e.tokenToArgument(t, l)

	// skip the )
	skip = l.NextToken()
	if skip.Literal != ")" {
		return &IfOperation{}, fmt.Errorf("expected ')' got %v", skip)
	}

block:
	// skip the {
	skip = l.NextToken()
	if skip.Literal != "{" {
		return &IfOperation{}, fmt.Errorf("expected '{' got %v", skip)
	}

	//
	// The list of statements to execute when the if-statement
	// matches
	//
	var Matches []Operation

	// Now we should parse the statement.
	b := l.NextToken()

body:
	stmt, err := e.parseOperation(b, l)
	if err != nil {
		return &IfOperation{}, err
	}

	Matches = append(Matches, stmt)

	b = l.NextToken()
	if b.Literal != "}" {
		goto body
	}
	return &IfOperation{Left: left, Right: right, Op: op, True: Matches}, nil
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
		err := e.parse()
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
		ret, val, err := op.Run(e, obj)

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

// tokenToArgument takes a given token, and converts it to an argument
// which can be evaluated.
//
// There is a minor complication here which is that when we see a
// token which represents a function-call we need to consume the
// arguments - recursively.
//
// This means we need a reference to our lexer, so we can fetch the
// next token(s).
//
func (e *Evaluator) tokenToArgument(tok Token, lexer *Lexer) Argument {
	var tmp Argument

	switch tok.Type {

	case FUNCALL:

		//
		// We've got a function.
		//
		// There are two cases:
		//
		//   Function()
		//
		// Or
		//
		//   Function( foo, bar , baz .. , bart )
		//
		// Either way we handle the parsing the same way, we
		// consume tokens forever until we hit the trailing `)`.
		//
		// If we find commas, which separate arguments, then we
		// discard them, otherwise we expand the tokens recursively.
		//
		// Recursive operations mean we can have a script which
		// runs `len(len(len(Name)))` if we wish.
		//
		var args []Argument

		for {
			t := lexer.NextToken()

			// Terminate when we find a right bracket
			if t.Type == RBRACKET {
				break
			}

			// Ignore commas - and the opening bracket
			if t.Type == COMMA || t.Type == LBRACKET {
				continue
			}

			// Add tokens
			args = append(args, e.tokenToArgument(t, lexer))

		}
		tmp = &FunctionArgument{Function: tok.Literal,
			Arguments: args}
	case IDENT:
		tmp = &FieldArgument{Field: tok.Literal}
	case STRING, NUMBER:
		tmp = &StringArgument{Content: tok.Literal}
	case FALSE:
		tmp = &BooleanArgument{Content: false}
	case TRUE:
		tmp = &BooleanArgument{Content: true}

	default:
		fmt.Printf("Failed to convert token %v to object - token-type was %s\n", tok, tok.Type)
		os.Exit(1)
	}

	return tmp
}
