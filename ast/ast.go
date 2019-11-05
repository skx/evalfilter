// Package ast contains the structures we use for generating an AST
// of a program.
//
// The AST is built up by our parser, via a series of tokens read
// from the user-script.  This is then walked to generate bytecode
// which is ultimately executed.
package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// Node reresents a node.
type Node interface {
	// TokenLiteral returns the literal of the token.
	TokenLiteral() string

	// String returns this object as a string.
	String() string
}

// Statement represents a single statement.
type Statement interface {
	// Node is the node holding the actual statement
	Node

	statementNode()
}

// Expression represents a single expression.
type Expression interface {
	// Node is the node holding the expression.
	Node
	expressionNode()
}

// Program represents a complete program.
type Program struct {
	// Statements is the set of statements which the program is comprised
	// of.
	Statements []Statement
}

// TokenLiteral returns the literal token of our program.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// String returns this object as a string.
func (p *Program) String() string {
	var out bytes.Buffer
	for _, stmt := range p.Statements {
		out.WriteString(stmt.String())
	}
	return out.String()
}

// Identifier holds a single identifier.
type Identifier struct {
	// Token is the literal token
	Token token.Token

	// Value is the name of the identifier
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the literal token.
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// String returns this object as a string.
func (i *Identifier) String() string {
	return i.Value
}

// ExpressionStatement is an expression
type ExpressionStatement struct {
	// Token is the literal token
	Token token.Token

	// Expression holds the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// String returns this object as a string.
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// PrefixExpression holds a prefix-based expression
type PrefixExpression struct {
	// Token holds the token.  e.g. "!"
	Token token.Token

	// Operator holds the operator being invoked (e.g. "!" ).
	Operator string

	// Right holds the thing to be operated upon
	Right Expression
}

func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// String returns this object as a string.
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression stores an infix expression.
type InfixExpression struct {
	// Token holds the literal expression
	Token token.Token

	// Left holds the left-most argument
	Left Expression

	// Operator holds the operation to be carried out (e.g. "+", "-" )
	Operator string

	// Right holds the right-most argument
	Right Expression
}

func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns this object as a string.
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// BlockStatement holds a group of statements, which are treated
// as a block.  (For example the body of an `if` expression.)
type BlockStatement struct {
	// Token holds the actual token
	Token token.Token

	// Statements contain the set of statements within the block
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// String returns this object as a string.
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	out.WriteString("\n{\n")
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	out.WriteString("\n}\n")
	return out.String()
}
