// Package evalfilter allows you to run simple  tests against objects or
// structs implemented in golang, via the use of user-supplied scripts
//
// Since the result of running tests against objects is a boolean/binary
// "yes/no" result it is perfectly suited to working as a filter.
package evalfilter

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/token"
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

	// Variables contains references to variables set via
	// the golang host application.
	Variables map[string]interface{}
}

// New returns a new evaluation object, which can be used to apply
// the specified script to an object/structure.
func New(input string) *Evaluator {
	e := &Evaluator{
		Debug:     false,
		Functions: make(map[string]interface{}),
		Variables: make(map[string]interface{}),
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

// SetVariable adds, or updates, a variable which will be available
// to the filter script.
//
// Variables in the filter script should be prefixed with `$`,
// for example a variable set as `time` will be accessed via the
// name `$time`.
func (e *Evaluator) SetVariable(name string, value interface{}) {
	e.Variables[name] = value
}

// Look at the given token, and parse it as an operation.
//
// This is abstracted into a routine of its own so that we can
// either parse the stream of tokens for the full-script, or parse
// the block which is used inside an IF statement.
func (e *Evaluator) parseOperation(tok token.Token, l *lexer.Lexer) (Operation, error) {

	switch tok.Type {

	//
	// `if`
	//
	case token.IF:
		return e.parseIF(l)

	//
	// `return`
	//
	case token.RETURN:

		// Get the value this token returns
		val := l.NextToken()

		// The next token should be a semi-colon
		tmp := l.NextToken()
		if tmp.Type != token.SEMICOLON {
			return nil, fmt.Errorf("expected ';' after return-value")

		}
		// Update our bytecode
		return &ReturnOperation{Value: val.Literal == "true"}, nil

	//
	// `print`
	//
	case token.PRINT:

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
			if n.Type == token.SEMICOLON || n.Type == token.EOF {
				break
			}

			//
			// Skip over any commas
			//
			if n.Type == token.COMMA {
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
	l := lexer.NewLexer(e.Program)

	//
	// Process all the tokens forever, until we hit the end of file.
	//
	tok := l.NextToken()

	for tok.Type != token.EOF {

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
func (e *Evaluator) parseIF(l *lexer.Lexer) (Operation, error) {

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
		// I feel bad.  But not that bad.
		//
		// Here we skip parsing the right-operand
		// leaving `Right` and `Op` at their default
		// values
		//
		op = ""
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
func (e *Evaluator) tokenToArgument(tok token.Token, lexer *lexer.Lexer) Argument {
	var tmp Argument

	switch tok.Type {

	case token.FUNCALL:

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
			if t.Type == token.RBRACKET {
				break
			}

			// Ignore commas - and the opening bracket
			if t.Type == token.COMMA || t.Type == token.LBRACKET {
				continue
			}

			// Add tokens
			args = append(args, e.tokenToArgument(t, lexer))

		}
		tmp = &FunctionArgument{Function: tok.Literal,
			Arguments: args}
	case token.IDENT:
		tmp = &FieldArgument{Field: tok.Literal}
	case token.VARIABLE:
		tmp = &VariableArgument{Name: tok.Literal}
	case token.STRING, token.NUMBER:
		tmp = &StringArgument{Content: tok.Literal}
	case token.FALSE:
		tmp = &BooleanArgument{Content: false}
	case token.TRUE:
		tmp = &BooleanArgument{Content: true}

	default:
		fmt.Printf("Failed to convert token %v to object - token-type was %s\n", tok, tok.Type)
		os.Exit(1)
	}

	return tmp
}
