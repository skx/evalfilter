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
	Bytecode []interface{}
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
	Arguments []string
}

// Value returns the result of calling the function we're wrapping.
//
// TODO: Arguments are ignored
func (f *FunctionArgument) Value(self *Evaluator, obj interface{}) interface{} {
	res, ok := self.Functions[f.Function]
	if ok {
		out := res.(func() bool)
		return (out())

	}
	fmt.Printf("Unknown function: %s\n", f.Function)
	os.Exit(1)
	return false

}

//
// Bytecode functions
//

// IfOperation holds state for the `if` operation
type IfOperation struct {
	Left   Argument
	Right  Argument
	Op     string
	Return string
}

// ReturnOperation holds state for the `return` operation
type ReturnOperation struct {
	// Value holds the value which will be returned.
	Value bool
}

// PrintOperation holds state for the `print` operation.
type PrintOperation struct {
	// Values are the various values to be printed.
	Values []Argument
}

// New returns a new evaluation object, which can be used to apply
// the specified script to an object/structure.
func New(input string) *Evaluator {
	e := &Evaluator{
		Debug:     false,
		Functions: make(map[string]interface{}),
		Program:   input,
	}

	// The environmental variable ${EVAL_FILTER_DEBUG} enables
	// the use of tracing.
	if os.Getenv("EVAL_FILTER_DEBUG") != "" {
		e.Debug = true
	}
	return e
}

// AddFunction adds a function to our runtime.
func (e *Evaluator) AddFunction(name string, fun interface{}) {
	if !strings.HasSuffix(name, "()") {
		name += "()"
	}
	e.Functions[name] = fun
}

// parse is an internal method which reads the script we've been
// given in our constructor and writes out a series of operations
// to be carried out to `Bytecode`.
//
// This is simple because we have no control-flow, and no need to
// worry about nested-blocks, variables, etc.
func (e *Evaluator) parse() error {

	//
	// Create a lexer to process our script.
	//
	l := NewLexer(e.Program)

	//
	// Process all the tokens.
	//
	// We're a little fast & loose here.
	//
	tok := l.NextToken()

	for tok.Type != EOF {

		//
		// Return
		//
		switch tok.Type {

		case IF:
			err := e.parseIF(l)
			if err != nil {
				return err
			}

		case RETURN:

			// Get the value this token returns
			val := l.NextToken()

			// Update our bytecode
			e.Bytecode = append(e.Bytecode,
				&ReturnOperation{Value: val.Literal == "true"})

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
				// Convert the token to an argument.
				//
				obj := e.tokenToArgument(n)

				//
				// Add it to our list.
				//
				tmp = append(tmp, obj)

			}

			//
			// Now record the print operation.
			//
			e.Bytecode = append(e.Bytecode,
				&PrintOperation{Values: tmp})

		}

		tok = l.NextToken()
	}

	return nil
}

// parseIf is our biggest method; it parses the contents of an if-statement.
func (e *Evaluator) parseIF(l *Lexer) error {

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
		return fmt.Errorf("expected '(' got %v", skip)
	}

	//
	// Get the first operand.
	//
	t := l.NextToken()
	left = e.tokenToArgument(t)

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
	right = e.tokenToArgument(t)

	// skip the )
	skip = l.NextToken()
	if skip.Literal != ")" {
		return fmt.Errorf("expected ')' got %v", skip)
	}

block:
	// skip the {
	skip = l.NextToken()
	if skip.Literal != "{" {
		return fmt.Errorf("expected '{' got %v", skip)
	}

	// The body should only contain a return-statement
	skip = l.NextToken()
	if skip.Type != RETURN {
		return fmt.Errorf("expected 'return' got %v", skip)
	}

	// Return value
	t = l.NextToken()
	val := t.Literal

	// skip the }
	skip = l.NextToken()

	// Skip optional ";" after return
	if skip.Literal == ";" {
		skip = l.NextToken()
	}
	if skip.Literal != "}" {
		return fmt.Errorf("expected '}' got %v", skip)
	}

	//
	// Record the IF-operation.
	//
	e.Bytecode = append(e.Bytecode,
		&IfOperation{Left: left, Right: right, Op: op, Return: val})
	return nil
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

		switch v := op.(type) {

		case *IfOperation:

			// Cast for neatness
			ifo := v

			// Run the if-statement.
			res, err := e.runIf(ifo.Left, ifo.Right, ifo.Op, ifo.Return, obj)

			// Was there an error?
			if err != nil {
				return false, fmt.Errorf("failed to run if-test %s", err)
			}

			//
			// No error - and we got a match.
			//
			if res {

				// Show that this matched
				if e.Debug {
					fmt.Printf("\tIF test matched\n")
				}

				// Return the value to the caller.
				if ifo.Return == "true" {
					return true, nil
				}
				return false, nil

			}

			// Show that IF-statement did not match
			if e.Debug {
				fmt.Printf("\tIF-statement did not match.\n")
			}

		case *ReturnOperation:

			return op.(*ReturnOperation).Value, nil

		case *PrintOperation:

			for _, val := range op.(*PrintOperation).Values {
				fmt.Printf("%v", val.Value(e, obj))
			}

		default:

			fmt.Printf("Unknown bytecode operation: %v", op)
		}
	}

	//
	// If we reach this point we've processed a script which did not
	// hit a bare return-statement.
	//
	return false, fmt.Errorf("script failed to terminate with a return statement")
}

// runIf runs an if comparison.
//
// We return "true" if the statement was true, and the return should
// be executed.  Otherwise we return false.
func (e *Evaluator) runIf(left Argument, right Argument, op string, res string, obj interface{}) (bool, error) {

	if e.Debug {
		fmt.Printf("IF %v %s %v Then return %s;\n", left, op, right, res)

	}

	//
	// Expand the left & right sides of the conditional
	//
	lVal := left.Value(e, obj)
	rVal := right.Value(e, obj)

	//
	// Basic operations
	//

	// Equality - string and number.
	if op == "==" {

		// Convert values to string, and compare.
		//
		// This allows "5 == "5""
		//
		return (fmt.Sprintf("%v", lVal) == fmt.Sprintf("%v", rVal)), nil
	}

	// Inequality - string and number.
	if op == "!=" {

		// Convert values to string, and compare.
		//
		// This allows "5 != "5""
		//
		return (fmt.Sprintf("%v", lVal) != fmt.Sprintf("%v", rVal)), nil
	}

	// String-contains
	if op == "~=" {

		src := fmt.Sprintf("%v", lVal)
		val := fmt.Sprintf("%v", rVal)

		return strings.Contains(src, val), nil
	}

	// String does not contain
	if op == "!~" {

		src := fmt.Sprintf("%v", lVal)
		val := fmt.Sprintf("%v", rVal)

		return !strings.Contains(src, val), nil
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
	a, err = e.toNumberArg(lVal)
	if err != nil {
		return false, err
	}
	b, err = e.toNumberArg(rVal)
	if err != nil {
		return false, err
	}

	//
	// Now operate.
	//
	if op == ">" {
		return (a > b), nil
	}
	if op == ">=" {
		return (a >= b), nil
	}
	if op == "<" {
		return (a < b), nil
	}
	if op == "<=" {
		return (a <= b), nil
	}

	//
	// Invalid operator?
	//
	return false, fmt.Errorf("unknown operator %v", op)
}

// toNumberArg tries to convert the given interface to a float64 value.
func (e *Evaluator) toNumberArg(value interface{}) (float64, error) {

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

// tokenToArgument takes a given token, and converts it to
// an argument.
//
// TODO: In the future this should parse a function and consume
// the arguments until we see ")".
//
func (e *Evaluator) tokenToArgument(tok Token) Argument {
	var tmp Argument

	switch tok.Type {

	case IDENT:
		//
		// TODO - handle functions more better.
		//
		// Specifically we'll have a function-token.
		//
		if strings.HasSuffix(tok.Literal, "()") {
			tmp = &FunctionArgument{Function: tok.Literal}
		} else {
			tmp = &FieldArgument{Field: tok.Literal}
		}
	case STRING:
		tmp = &StringArgument{Content: tok.Literal}
	case NUMBER:
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
