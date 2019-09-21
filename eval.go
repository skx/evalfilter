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
	"os"
	"strings"
	"unicode/utf8"

	"github.com/skx/evalfilter/environment"
	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/runtime"
	"github.com/skx/evalfilter/token"
)

// Evaluator holds our object state.
type Evaluator struct {

	// Bytecode operations are stored here.
	Bytecode []runtime.Operation

	// Environment holds our environment reference
	Env *environment.Environment

	// Program is the script the user wishes to run.
	Program string
}

// New returns a new evaluation object, which can be used to apply
// the specified script to an object/structure.
func New(input string) *Evaluator {

	// Create a stub object.
	e := &Evaluator{
		Env:     environment.New(),
		Program: input,
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

// Look at the given token, and parse it as an operation.
//
// This is abstracted into a routine of its own so that we can
// either parse the stream of tokens for the full-script, or parse
// the blocks which is used for `if` statements.
func (e *Evaluator) parseOperation(tok token.Token, l *lexer.Lexer) (runtime.Operation, error) {

	switch tok.Type {

	//
	// `eval`
	//
	case token.FUNCALL:

		//
		// Eval is a special case, because we're basically making
		// a function-call but throwing away the result.
		//

		//
		// Parse the function into a callable argument.
		//
		arg := e.tokenToArgument(tok, l)

		//
		// Now append the eval-operation
		//
		return &runtime.EvalOperation{Value: arg}, nil

	//
	// `if`
	//
	case token.IF:

		// The `if` statement is our most complex case, and it
		// will not get simpler, so it is moved into its own
		// routine.
		return e.parseIF(l)

	//
	// `return`
	//
	case token.RETURN:

		// Get the value this token returns
		val := l.NextToken()

		// The token after that should be a semi-colon.
		tmp := l.NextToken()
		if tmp.Type != token.SEMICOLON {
			return nil, fmt.Errorf("expected ';' after return-value")

		}

		// Return the operation.
		return &runtime.ReturnOperation{Value: val.Literal == "true"}, nil

	//
	// `print`
	//
	case token.PRINT:

		//
		// Here are the arguments we're going to be printing.
		//
		var tmp []runtime.Argument

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
		return &runtime.PrintOperation{Values: tmp}, nil

	}

	//
	// If we hit this point we've received input that we don't
	// recognize - either because it was invalid, or because we've
	// become unsynced in our token-stream.
	//
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
func (e *Evaluator) parseIF(l *lexer.Lexer) (runtime.Operation, error) {

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
	var left runtime.Argument
	var right runtime.Argument
	var op string

	//
	// skip the (
	//
	skip := l.NextToken()
	if skip.Literal != "(" {
		return &runtime.IfOperation{}, fmt.Errorf("expected '(' got %v", skip)
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
		return &runtime.IfOperation{}, fmt.Errorf("expected ')' got %v", skip)
	}

block:
	// skip the {
	skip = l.NextToken()
	if skip.Literal != "{" {
		return &runtime.IfOperation{}, fmt.Errorf("expected '{' got %v", skip)
	}

	//
	// The list of statements to execute when the if-statement matches,
	// or fails to match.
	//
	var True []runtime.Operation
	var False []runtime.Operation

	// Now we should parse the statement.
	b := l.NextToken()

true_body:
	stmt, err := e.parseOperation(b, l)
	if err != nil {
		return &runtime.IfOperation{}, err
	}

	True = append(True, stmt)

	b = l.NextToken()
	if b.Literal != "}" {
		goto true_body
	}

	//
	// Now look for else
	//
	el := l.NextToken()
	if el.Type != token.ELSE {
		l.Rewind(el)

		return &runtime.IfOperation{Left: left, Right: right, Op: op,
			True:  True,
			False: False}, nil
	}

	// skip the {
	skip = l.NextToken()
	if skip.Literal != "{" {
		return &runtime.IfOperation{}, fmt.Errorf("expected '{' after 'else' got %v", skip)
	}

	// Now we should parse the statement.
	b = l.NextToken()

false_body:
	stmt, err = e.parseOperation(b, l)
	if err != nil {
		return &runtime.IfOperation{}, err
	}

	False = append(False, stmt)

	b = l.NextToken()
	if b.Literal != "}" {
		goto false_body
	}

	return &runtime.IfOperation{Left: left, Right: right, Op: op,
		True:  True,
		False: False}, nil

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
func (e *Evaluator) tokenToArgument(tok token.Token, lexer *lexer.Lexer) runtime.Argument {
	var tmp runtime.Argument

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
		var args []runtime.Argument

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

		// Skip the optional, but expected, trailing ";"
		skip := lexer.NextToken()
		if skip.Type != token.SEMICOLON {
			lexer.Rewind(skip)
		}

		tmp = &runtime.FunctionArgument{Function: tok.Literal,
			Arguments: args}
	case token.IDENT:
		tmp = &runtime.FieldArgument{Field: tok.Literal}
	case token.VARIABLE:
		tmp = &runtime.VariableArgument{Name: tok.Literal}
	case token.STRING, token.NUMBER:
		tmp = &runtime.StringArgument{Content: tok.Literal}
	case token.FALSE:
		tmp = &runtime.BooleanArgument{Content: false}
	case token.TRUE:
		tmp = &runtime.BooleanArgument{Content: true}

	default:
		fmt.Printf("Failed to convert token %v to object - token-type was %s\n", tok, tok.Type)
		os.Exit(1)
	}

	return tmp
}
