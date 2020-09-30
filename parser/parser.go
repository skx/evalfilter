// Package parser consumes tokens from the lexer and returns a
// program as a set of AST-nodes.
//
// Later we walk the AST tree and generate a series of bytecode
// instructions.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skx/evalfilter/v2/ast"
	"github.com/skx/evalfilter/v2/lexer"
	"github.com/skx/evalfilter/v2/token"
)

// prefix Parse function
// infix parse function
// postfix parse function
type (
	prefixParseFn  func() ast.Expression
	infixParseFn   func(ast.Expression) ast.Expression
	postfixParseFn func() ast.Expression
)

// Here we define values for precedence, lowest to highest.
const (
	_ int = iota
	LOWEST
	TERNARY // ? :
	ASSIGN  // =
	COND    // OR or AND
	EQUALS  // == or !=
	CMP
	LESSGREATER // > or <
	SUM         // + or -
	PRODUCT     // * or /
	POWER       // **
	MOD         // %
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index], map[key]
)

// precedence contains the precedence for each token-type, which
// is part of the magic of a Pratt-Parser.
var precedences = map[token.Type]int{
	token.QUESTION:       TERNARY,
	token.ASSIGN:         ASSIGN,
	token.DOTDOT:         ASSIGN,
	token.EQ:             EQUALS,
	token.NOTEQ:          EQUALS,
	token.LT:             LESSGREATER,
	token.LTEQUALS:       LESSGREATER,
	token.GT:             LESSGREATER,
	token.GTEQUALS:       LESSGREATER,
	token.CONTAINS:       LESSGREATER,
	token.MISSING:        LESSGREATER,
	token.IN:             LESSGREATER,
	token.PLUSEQUALS:     SUM,
	token.PLUS:           SUM,
	token.MINUS:          SUM,
	token.MINUSEQUALS:    SUM,
	token.SLASH:          PRODUCT,
	token.SLASHEQUALS:    PRODUCT,
	token.ASTERISK:       PRODUCT,
	token.ASTERISKEQUALS: PRODUCT,
	token.POW:            POWER,
	token.MOD:            MOD,
	token.AND:            COND,
	token.OR:             COND,
	token.LPAREN:         CALL,
	token.LSQUARE:        INDEX,
}

// Parser is the object which maintains our parser state.
//
// We consume tokens, produced by our lexer, and so we need to
// keep track of our current token, the next token, and any
// errors we've seen, for example.
type Parser struct {
	// l is our lexer
	l *lexer.Lexer

	// prevToken holds the previous token from our lexer.
	// (used for "++" + "--")
	prevToken token.Token

	// curToken holds the current token from our lexer.
	curToken token.Token

	// peekToken holds the next token which will come from the lexer.
	peekToken token.Token

	// errors holds parsing-errors.
	errors []string

	// prefixParseFns holds a map of parsing methods for
	// prefix-based syntax.
	prefixParseFns map[token.Type]prefixParseFn

	// infixParseFns holds a map of parsing methods for
	// infix-based syntax.
	infixParseFns map[token.Type]infixParseFn

	// postfixParseFns holds a map of parsing methods for
	// postfix-based syntax.
	postfixParseFns map[token.Type]postfixParseFn

	// are we inside a ternary expression?
	//
	// Nested ternary expressions are illegal so we
	// need to keep track of this.
	tern bool

	// Are we inside a function?
	function bool
}

// New returns a new parser.
//
// Once constructed it can be used to parse an input-program
// into an AST.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.Type]prefixParseFn)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.EOF, p.parseEOF)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.FOR, p.parseWhileStatement)
	p.registerPrefix(token.FOREACH, p.parseForEach)
	p.registerPrefix(token.FUNCTION, p.parseFunctionDefinition)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.ILLEGAL, p.parseIllegal)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.LOCAL, p.parseLocalVariable)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LSQUARE, p.parseArrayLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.REGEXP, p.parseRegexpLiteral)
	p.registerPrefix(token.SQRT, p.parsePrefixExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.SWITCH, p.parseSwitchStatement)
	p.registerPrefix(token.WHILE, p.parseWhileStatement)

	p.infixParseFns = make(map[token.Type]infixParseFn)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.ASSIGN, p.parseAssignExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.ASTERISKEQUALS, p.parseInfixExpression)
	p.registerInfix(token.CONTAINS, p.parseInfixExpression)
	p.registerInfix(token.DOTDOT, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTEQUALS, p.parseInfixExpression)
	p.registerInfix(token.IN, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LSQUARE, p.parseIndexExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LTEQUALS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MINUSEQUALS, p.parseInfixExpression)
	p.registerInfix(token.MISSING, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.NOTEQ, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.PLUSEQUALS, p.parseInfixExpression)
	p.registerInfix(token.POW, p.parseInfixExpression)
	p.registerInfix(token.QUESTION, p.parseTernaryExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.SLASHEQUALS, p.parseInfixExpression)

	p.postfixParseFns = make(map[token.Type]postfixParseFn)
	p.registerPostfix(token.MINUSMINUS, p.parsePostfixExpression)
	p.registerPostfix(token.PLUSPLUS, p.parsePostfixExpression)

	return p
}

// registerPrefix registers a function for handling a prefix-based statement
func (p *Parser) registerPrefix(tokenType token.Type, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers a function for handling a infix-based statement
func (p *Parser) registerInfix(tokenType token.Type, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// registerPostfix registers a function for handling a postfix-based statement
func (p *Parser) registerPostfix(tokenType token.Type, fn postfixParseFn) {
	p.postfixParseFns[tokenType] = fn
}

// Errors return stored errors
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError raises an error if the next token is not the expected type.
func (p *Parser) peekError(t token.Type) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead around %s", t, p.curToken.Type, p.curToken.Position())
	p.errors = append(p.errors, msg)
}

// nextToken moves to our next token from the lexer.
func (p *Parser) nextToken() {
	p.prevToken = p.curToken
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Parse is the main public-facing method to parse an input program.
//
// It will return any error-encountered in parsing the input, but
// to avoid confusion it will only return the first error.
//
// To access any subsequent errors please see `Errors`.
func (p *Parser) Parse() (*ast.Program, error) {

	// Parse
	a := p.ParseProgram()

	// Look for errors
	if len(p.errors) == 0 {
		return a, nil
	}

	// Only the first error matters.
	return a, fmt.Errorf("%s", p.Errors()[0])
}

// ParseProgram used to parse the whole program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF && p.curToken.Type != token.ILLEGAL {
		stmt := p.parseStatement()
		if stmt == nil {
			msg := fmt.Sprintf("unexpected nil statement around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}

	if p.curToken.Type == token.ILLEGAL {
		p.errors = append(p.errors, p.curToken.Literal)
	}
	return program
}

// parseStatement parses a single statement.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.RETURN:
		r := p.parseReturnStatement()
		if r == nil {
			msg := fmt.Sprintf("unexpected nil statement around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
		return r

	default:
		return p.parseExpressionStatement()
	}
}

// parseReturnStatement parses a return-statement.
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	p.nextToken()
	if p.curToken.Type != token.SEMICOLON {
		p.errors = append(p.errors, fmt.Sprintf("expected semicolon after return-value; found token '%v'", p.curToken))
		stmt.ReturnValue = nil
		return nil
	}

	return stmt
}

// Function called on error if there is no prefix-based parsing method
// for the given token.
func (p *Parser) noPrefixParseFnError(t token.Type) {
	msg := fmt.Sprintf("no prefix parse function for %s found around %s", t, p.curToken.Position())
	p.errors = append(p.errors, msg)
}

// parse Expression Statement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parse an expression.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	postfix := p.postfixParseFns[p.curToken.Type]
	if postfix != nil {
		msg := fmt.Sprintf("expected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return (postfix())
	}
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		msg := fmt.Sprintf("invalid token '%s' around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	leftExp := prefix()

	// Look for errors
	if leftExp == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)

		// Look for errors
		if leftExp == nil {
			msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
	}
	return leftExp
}

// report an error that we found an illegal state.
//
// This is generally seen with an unterminated string.
func (p *Parser) parseIllegal() ast.Expression {
	msg := fmt.Sprintf("illegal token hit parsing program %s around %s", p.curToken.Literal, p.curToken.Position())
	p.errors = append(p.errors, msg)
	return nil
}

// report an error if we hit an unexpected end of file.
func (p *Parser) parseEOF() ast.Expression {
	p.errors = append(p.errors, "unexpected end of file reached")
	return nil
}

// parseIdentifier parses an identifier.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseLocal parses something like "local x;"
func (p *Parser) parseLocalVariable() ast.Expression {

	if !p.function {
		msg := fmt.Sprintf("'local' may only be used inside a function, around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	// Skip over the `local`
	p.nextToken()

	// Ensure we got an ident.
	if !p.curTokenIs(token.IDENT) {
		msg := fmt.Sprintf("'local' may only be used with an IDENT, around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	return &ast.LocalVariable{Token: p.curToken}

}

// parseIntegerLiteral parses an integer literal.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// parseFloatLiteral parses a float-literal
func (p *Parser) parseFloatLiteral() ast.Expression {
	flo := &ast.FloatLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	flo.Value = value
	return flo
}

// parseSwitchStatement handles a switch statement
func (p *Parser) parseSwitchStatement() ast.Expression {

	// switch
	expression := &ast.SwitchExpression{Token: p.curToken}

	// look for (xx)
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	expression.Value = p.parseExpression(LOWEST)
	if expression.Value == nil {
		return nil
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Now we have a block containing blocks.
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	p.nextToken()

	// Process the block which we think will contain
	// various case-statements
	for !p.curTokenIs(token.RBRACE) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated switch statement")
			return nil
		}
		tmp := &ast.CaseExpression{Token: p.curToken}

		// Default will be handled specially
		if p.curTokenIs(token.DEFAULT) {

			// We have a default-case here.
			tmp.Default = true

		} else if p.curTokenIs(token.CASE) {

			// skip "case"
			p.nextToken()

			// Here we allow "case default" even though
			// most people would prefer to write "default".
			if p.curTokenIs(token.DEFAULT) {
				tmp.Default = true
			} else {

				// parse the match-expression.
				tmp.Expr = append(tmp.Expr, p.parseExpression(LOWEST))
				for p.peekTokenIs(token.COMMA) {

					// skip the comma
					p.nextToken()

					// setup the expression.
					p.nextToken()

					tmp.Expr = append(tmp.Expr, p.parseExpression(LOWEST))

				}
			}
		} else {
			// error - unexpected token
			p.errors = append(p.errors, fmt.Sprintf("expected case|default, got %s around position %s", p.curToken.Type, p.curToken.Position()))
			return nil
		}

		if !p.expectPeek(token.LBRACE) {

			msg := fmt.Sprintf("expected token to be '{', got %s instead", p.curToken.Type)
			p.errors = append(p.errors, msg)
			fmt.Printf("error\n")
			return nil
		}

		// parse the block
		tmp.Block = p.parseBlockStatement()

		if !p.curTokenIs(token.RBRACE) {
			msg := fmt.Sprintf("Syntax Error: expected token to be '}', got %s instead", p.curToken.Type)
			p.errors = append(p.errors, msg)
			fmt.Printf("error\n")
			return nil

		}
		p.nextToken()

		// save the choice away
		expression.Choices = append(expression.Choices, tmp)

	}

	// ensure we're at the the closing "}"
	if !p.curTokenIs(token.RBRACE) {
		return nil
	}

	// More than one default is a bug
	count := 0
	for _, c := range expression.Choices {
		if c.Default {
			count++
		}
	}
	if count > 1 {
		msg := "A switch-statement should only have one default block"
		p.errors = append(p.errors, msg)
		return nil

	}
	return expression

}

// parseBoolean parses a boolean token.
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parsePrefixExpression parses a prefix-based expression.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	// If there was an error parsing the target of our
	// prefix operation then we must abort.
	if expression.Right == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return expression
}

// parseInfixExpression parses an infix-based expression.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	// If there was an error parsing the second operand
	// then we must abort.
	if expression.Right == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return expression
}

// parsePostfixExpression parses a postfix-based expression.
func (p *Parser) parsePostfixExpression() ast.Expression {
	expression := &ast.PostfixExpression{
		Token:    p.prevToken,
		Operator: p.curToken.Literal,
	}
	return expression
}

// parseTernaryExpression parses a ternary expression
func (p *Parser) parseTernaryExpression(condition ast.Expression) ast.Expression {

	if p.tern {
		p.errors = append(p.errors, fmt.Sprintf("nested ternary expressions are illegal around %s", p.curToken.Position()))
		return nil
	}

	p.tern = true
	defer func() { p.tern = false }()

	expression := &ast.TernaryExpression{
		Token:     p.curToken,
		Condition: condition,
	}
	p.nextToken() //skip the '?'
	precedence := p.curPrecedence()
	expression.IfTrue = p.parseExpression(precedence)

	// error?
	if expression.IfTrue == nil {
		p.errors = append(p.errors, fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position()))
		return nil
	}

	if !p.expectPeek(token.COLON) { //skip the ":"
		p.errors = append(p.errors, fmt.Sprintf("missing colon in ternary expression around  %s", p.curToken.Position()))
		return nil
	}

	// Get to next token, then parse the else part
	p.nextToken()
	expression.IfFalse = p.parseExpression(precedence)

	// error?
	if expression.IfFalse == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	return expression
}

// parseGroupedExpression parses a grouped-expression.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if exp == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("expected ) but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return exp
}

// parseIfCondition parses an if-expression.
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	if expression == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("expected ( but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if expression.Condition == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("expected ) but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("expected { but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	expression.Consequence = p.parseBlockStatement()
	if expression.Consequence == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.expectPeek(token.LBRACE) {
			msg := fmt.Sprintf("expected { but got %s around %s", p.curToken.Literal, p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
		expression.Alternative = p.parseBlockStatement()
		if expression.Alternative == nil {
			msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
	}
	return expression
}

// parseForEach parses 'foreach x X { .. block .. }`
func (p *Parser) parseForEach() ast.Expression {
	expression := &ast.ForeachStatement{Token: p.curToken}

	// get the id
	p.nextToken()
	expression.Ident = p.curToken.Literal

	// If we find a "," we then get a second identifier too.
	if p.peekTokenIs(token.COMMA) {

		//
		// Generally we have:
		//
		//    foreach IDENT in THING { .. }
		//
		// If we have two arguments the first becomes
		// the index, and the second becomes the IDENT.
		//

		// skip the comma
		p.nextToken()

		if !p.peekTokenIs(token.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("second argument to foreach must be ident, got %v", p.peekToken))
			return nil
		}
		p.nextToken()

		//
		// Record the updated values.
		//
		expression.Index = expression.Ident
		expression.Ident = p.curToken.Literal

	}

	// The next token, after the ident(s), should be `in`.
	if !p.expectPeek(token.IN) {
		msg := fmt.Sprintf("missing 'in' in foreach statement around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()

	// get the thing we're going to iterate  over.
	expression.Value = p.parseExpression(LOWEST)
	if expression.Value == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	// parse the block
	p.nextToken()
	expression.Body = p.parseBlockStatement()

	return expression
}

// parseFunctionDefinition parses the definition of a function.
func (p *Parser) parseFunctionDefinition() ast.Expression {

	// We're inside a function
	p.function = true

	// skip the `function` keyword
	p.nextToken()

	// Define a function with the identifier
	lit := &ast.FunctionDefinition{Token: p.curToken}

	// Expect "("
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("expected ( but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	// Swallow all arguments until the closing ")"
	lit.Parameters = p.parseFunctionParameters()

	// Now we want "{"
	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("expected { but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	// And consume the function-body including the
	// closing "}".
	lit.Body = p.parseBlockStatement()

	// We're no longer inside a function
	p.function = false

	return lit
}

// parseFunctionParameters parses the parameters used for a function.
//
// Function parameters are untyped, so we're looking for "foo, bar, baz)".
func (p *Parser) parseFunctionParameters() []*ast.Identifier {

	// The argument-definitions.
	identifiers := make([]*ast.Identifier, 0)

	// Is the next parameter ")" ?  If so we're done. No args.
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()

	// Keep going until we find a ")"
	for !p.curTokenIs(token.RPAREN) {

		if p.curTokenIs(token.EOF) {
			p.errors = append(p.errors, "unterminated function parameters found end of file")
			return nil
		}

		// Get the identifier.
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
		p.nextToken()

		// Skip any comma.
		if p.curTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	return identifiers
}

// parseWhileStatement parses a while-statement.
func (p *Parser) parseWhileStatement() ast.Expression {
	expression := &ast.WhileStatement{Token: p.curToken}
	if expression == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.LPAREN) {
		msg := fmt.Sprintf("expected ( but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	if expression.Condition == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.RPAREN) {
		msg := fmt.Sprintf("expected ) but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	if !p.expectPeek(token.LBRACE) {
		msg := fmt.Sprintf("expected { but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	expression.Body = p.parseBlockStatement()
	return expression
}

// parseBlockStatement parses a block.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}
	p.nextToken()
	for !p.curTokenIs(token.RBRACE) {
		stmt := p.parseStatement()
		if stmt == nil {
			msg := fmt.Sprintf("unexpected nil statement around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
		block.Statements = append(block.Statements, stmt)
		p.nextToken()

		if p.curToken.Type == token.EOF || p.curToken.Type == token.ILLEGAL {
			p.errors = append(p.errors, "incomplete block statement")
			return nil
		}
	}
	return block
}

// parseStringLiteral parses a string-literal.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseRegexpLiteral parses a regular-expression.
func (p *Parser) parseRegexpLiteral() ast.Expression {

	flags := ""

	val := p.curToken.Literal
	if strings.HasPrefix(val, "(?") {
		val = strings.TrimPrefix(val, "(?")

		i := 0
		for i < len(val) {

			if val[i] == ')' {

				val = val[i+1:]
				break
			} else {
				flags += string(val[i])
			}

			i++
		}
	}
	return &ast.RegexpLiteral{Token: p.curToken, Value: val, Flags: flags}
}

// parseArrayLiteral parses an array literal.
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RSQUARE)
	return array
}

// parseHashLiteral parses a hash literal.
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)
	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(token.COLON) {
			return nil
		}
		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value
		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

// parse an array of expressions, as used for function-arguments.
func (p *Parser) parseExpressionList(end token.Type) []ast.Expression {
	list := make([]ast.Expression, 0)
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()

	// parse first item
	first := p.parseExpression(LOWEST)
	if first == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	list = append(list, first)

	// Keep going if we hit a comma
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		ent := p.parseExpression(LOWEST)
		if ent == nil {
			msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
			p.errors = append(p.errors, msg)
			return nil
		}
		list = append(list, ent)
	}
	if !p.expectPeek(end) {
		msg := fmt.Sprintf("expected EOF not found around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return list
}

// parseAssignExpression parses an assignment-statement.
func (p *Parser) parseAssignExpression(name ast.Expression) ast.Expression {
	stmt := &ast.AssignStatement{Token: p.curToken}
	if n, ok := name.(*ast.Identifier); ok {
		stmt.Name = n
	} else {
		msg := fmt.Sprintf("expected assign token to be IDENT, got %s instead around %s", name.TokenLiteral(), p.curToken.Position())
		p.errors = append(p.errors, msg)
	}

	// Skip over the `=`
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		msg := fmt.Sprintf("unexpected nil statement around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return stmt
}

// parseCallExpression parses a function-call expression.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseIndexExpression parse an array-index expression.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	// error?
	if exp.Index == nil {
		msg := fmt.Sprintf("unexpected nil expression around %s", p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}

	if !p.expectPeek(token.RSQUARE) {
		msg := fmt.Sprintf("expected ] but got %s around %s", p.curToken.Literal, p.curToken.Position())
		p.errors = append(p.errors, msg)
		return nil
	}
	return exp
}

// curTokenIs tests if the current token has the given type.
func (p *Parser) curTokenIs(t token.Type) bool {
	return p.curToken.Type == t
}

// peekTokenIs tests if the next token has the given type.
func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

// expectPeek validates the next token is of the given type,
// and advances if so.  If it is not an error is stored.
func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

// peekPrecedence looks up the next token precedence.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// curPrecedence looks up the current token precedence.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
