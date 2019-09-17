// Package evalfilter allows you to run simple  tests against objects or
// structs implemented in golang, via the use of user-supplied scripts
//
// Since the result of running tests against objects is a boolean/binary
// "yes/no" result it is perfectly suited to working as a filter.
package main

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

// Bytecode functions

// IfOperation holds state for the `if` operation
type IfOperation struct {
	Left   Token
	Right  Token
	Op     Token
	Return Token
}

// ReturnOperation holds state for the `return` operation
type ReturnOperation struct {
	// Value holds the value of a number to be pushed upon the RPN stack.
	Value bool
}

// PrintOperation holds state for the `print` operation.
type PrintOperation struct {
	Values []Token
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

			var tmp []Token

			for {
				//
				// Keep printing output until we hit
				// a semi-colon, or the end of the file.
				//
				n := l.NextToken()
				if n.Type == SEMICOLON || n.Type == EOF {
					break
				}

				tmp = append(tmp, n)
			}

			e.Bytecode = append(e.Bytecode,
				&PrintOperation{Values: tmp})

		}

		tok = l.NextToken()
	}

	return nil
}

func (e *Evaluator) parseIF(l *Lexer) error {

	//
	// The general form is:
	//
	//  IF ( LEFT TEST RIGHT ) { RETURN YY; }
	//
	// e.g. "if ( Count == 3 ) { return true; }"
	//
	//
	var left Token
	var op Token
	var right Token

	// skip the (
	skip := l.NextToken()
	if skip.Literal != "(" {
		return fmt.Errorf("expected '(' got %v", skip)
	}

	//
	// Get the first operand.
	//
	left = l.NextToken()

	//
	// Get the operand.
	//
	op = l.NextToken()

	//
	// There are two forms of IF:
	//
	//   IF ( left OP right ) ..
	//
	// And
	//
	//   IF ( function() ) ..
	//
	// We tell them apart by looking above.
	//
	if op.Literal == ")" {

		//
		// Because we see ")" we assume we're
		// the single-form of the IF-statement.
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
		op.Literal = "=="

		//
		// Fake the right-value.
		//
		// NB: This works because we force user-added
		// functions to return boolean values.
		//
		right.Literal = "true"

		//
		// I feel bad.  But not that bad.
		//
		goto block
	}

	//
	// OK we're in the three-argument form, so we
	// get the right operand.
	//
	right = l.NextToken()

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
	val := l.NextToken()

	// skip the }
	skip = l.NextToken()

	// Skip optional ";" after return
	if skip.Literal == ";" {
		skip = l.NextToken()
	}
	if skip.Literal != "}" {
		return fmt.Errorf("expected '}' got %v", skip)
	}

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
	// Parse - the first time.
	//
	if len(e.Bytecode) == 0 {
		err := e.parse()
		if err != nil {
			return false, err
		}
	}

	//
	// Run the bytecode.
	//
	for _, op := range e.Bytecode {

		switch v := op.(type) {

		case *IfOperation:

			// Cast for neatness
			ifo := v

			// Run the iff.
			res, err := e.runIf(ifo.Left, ifo.Right, ifo.Op.Literal, ifo.Return.Literal, obj)

			// Was there an error?
			if err != nil {
				return false, fmt.Errorf("failed to run if-test %s", err)
			}

			// No error - and we got a match.
			if res {

				// Show that this matched
				if e.Debug {
					fmt.Printf("\tIF test matched\n")
				}

				// If it matched we return the stated value
				if ifo.Return.Literal == "true" {
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

				val := e.expandToken(val, obj)
				fmt.Printf("%v", val)
			}

		default:
			fmt.Printf("Unknown bytecode thing")
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
func (e *Evaluator) runIf(left Token, right Token, op string, res string, obj interface{}) (bool, error) {

	if e.Debug {
		fmt.Printf("IF %s %s %s Then return %s;\n", left.Literal, op, right.Literal, res)

	}

	//
	// Expand the left & right sides of the conditional
	//
	lVal := e.expandToken(left, obj)
	rVal := e.expandToken(right, obj)

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
	// Numeric operations.
	//

	//
	// Get the parameters as numbers.
	//
	var a float64
	var b float64
	var err error

	a, err = e.toNumberArg(lVal)
	if err != nil {
		return false, err
	}
	b, err = e.toNumberArg(rVal)
	if err != nil {
		return false, err
	}

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

	return false, fmt.Errorf("unknown operator %v", op)
}

// expandToken expands the given token, handling a literal string/number,
// requesting a field-lookup, or making a function-call.
func (e *Evaluator) expandToken(tok Token, obj interface{}) interface{} {

	// The value we return
	var val interface{}

	// Assume we're dealing with a literal string.
	val = tok.Literal

	// But lookup if it is a field-structure, or function-call
	if tok.Type == IDENT {

		// Is this a function-call?
		if strings.HasSuffix(tok.Literal, "()") {
			val = e.callFunction(tok.Literal)
		} else {
			// If not it is a field-lookup.
			val = e.getStructureField(tok.Literal, obj)
		}
	}

	return val
}

// Return the value of the given field from the object.
func (e *Evaluator) getStructureField(field string, obj interface{}) interface{} {
	r := reflect.ValueOf(obj)
	f := reflect.Indirect(r).FieldByName(field)

	switch f.Kind() {
	case reflect.Int, reflect.Int64:
		return f.Int()
	case reflect.Float32, reflect.Float64:
		return f.Float()
	case reflect.String:
		return f.String()
	case reflect.Bool:
		if f.Bool() {
			return "true"
		}
		return "false"
	}
	return nil
}

// callFunction invokes a function the user must have defined and passed to
// us via `AddFunction`.
func (e *Evaluator) callFunction(fun string) bool {

	res, ok := e.Functions[fun]
	if ok {
		out := res.(func() bool)
		return (out())

	}
	fmt.Printf("Unknown function: %s\n", fun)
	os.Exit(1)
	return false
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
