// Package parser is the package which is responsible for parsing user-scripts.
//
// The input to the parser is the text of the script to be parsed, and the
// output will be a series of Operations - these operations live in the
// runtime package so they can be shared between this parser and the evaluator
// which actually executes them.
package parser

import (
	"fmt"
	"os"

	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/runtime"
	"github.com/skx/evalfilter/token"
)

// Parser holds the state for the parser.
type Parser struct {

	// Script is the text of the script the user wishes to run.
	Script string
}

// New creates a new parser-object, which will operate upon the script
// supplied by the caller.
func New(script string) *Parser {

	p := &Parser{
		Script: script,
	}

	return p
}

// Parse is the method which reads the script we've been given in our
// constructor and returns a series of operations to be carried out by
// the main Evaluator package.
//
// The parser is simple because we have no control-flow, and no need to
// worry about nested-blocks, variables, etc.
func (p *Parser) Parse() ([]runtime.Operation, error) {

	//
	// The operations we return
	//
	var ops []runtime.Operation

	//
	// Create a lexer to process our script.
	//
	l := lexer.NewLexer(p.Script)

	//
	// Process all the tokens forever, until we hit the end of file.
	//
	tok := l.NextToken()

	for tok.Type != token.EOF {

		//
		// Parse the next statement.
		//
		op, err := p.parseOperation(tok, l)
		if err != nil {
			return ops, err
		}

		//
		// Append it to our list.
		//
		ops = append(ops, op)

		//
		// Proceed onto the next token.
		//
		tok = l.NextToken()
	}

	//
	// Parsed with no error.
	//
	return ops, nil
}

// parseIf is our biggest method; it parses an if-expression.
func (p *Parser) parseIF(l *lexer.Lexer) (runtime.Operation, error) {

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
	// We build up a list of expressions
	//
	var expr []runtime.IfExpression
	exprType := "and"

	//
	// skip the (
	//
	skip := l.NextToken()
	if skip.Literal != "(" {
		return &runtime.IfOperation{}, fmt.Errorf("expected '(' got %v", skip)
	}

expr:
	//
	// Get the first operand.
	//
	t := l.NextToken()
	left = p.tokenToArgument(t, l)

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
		expr = append(expr, runtime.IfExpression{Left: left})
		goto block
	}

	//
	// OK we're in the three-argument form, so we
	// get the right operand.
	//
	t = l.NextToken()
	right = p.tokenToArgument(t, l)

	//
	// Add on the expression
	//
	expr = append(expr, runtime.IfExpression{Left: left, Right: right, Op: op})

	//
	// Loop?
	//
	skip = l.NextToken()
	if skip.Literal == ")" {
		goto block
	}
	if skip.Literal == "and" {
		exprType = "and"
		goto expr
	}
	if skip.Literal == "or" {
		exprType = "or"
		goto expr
	}
	return &runtime.IfOperation{}, fmt.Errorf("unterminated if expression: %v", skip)
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
	stmt, err := p.parseOperation(b, l)
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

		return &runtime.IfOperation{Expressions: expr,
			ExpressionType: exprType,
			True:           True,
			False:          False}, nil
	}

	// skip the {
	skip = l.NextToken()
	if skip.Literal != "{" {
		return &runtime.IfOperation{}, fmt.Errorf("expected '{' after 'else' got %v", skip)
	}

	// Now we should parse the statement.
	b = l.NextToken()

false_body:
	stmt, err = p.parseOperation(b, l)
	if err != nil {
		return &runtime.IfOperation{}, err
	}

	False = append(False, stmt)

	b = l.NextToken()
	if b.Literal != "}" {
		goto false_body
	}

	return &runtime.IfOperation{Expressions: expr,
		ExpressionType: exprType,
		True:           True,
		False:          False}, nil

}

// Look at the given token, and parse it as an operation.
//
// This is abstracted into a routine of its own so that we can
// either parse the stream of tokens for the full-script, or parse
// the blocks which is used for `if` statements.
func (p *Parser) parseOperation(tok token.Token, l *lexer.Lexer) (runtime.Operation, error) {

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
		arg := p.tokenToArgument(tok, l)

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
		return p.parseIF(l)

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
			obj := p.tokenToArgument(n, l)

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
func (p *Parser) tokenToArgument(tok token.Token, lexer *lexer.Lexer) runtime.Argument {
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
			args = append(args, p.tokenToArgument(t, lexer))

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
