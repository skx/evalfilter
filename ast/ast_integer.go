package ast

import "github.com/skx/evalfilter/v2/token"

// IntegerLiteral holds an integer
type IntegerLiteral struct {
	// Token is the literal token
	Token token.Token

	// Value holds the integer.
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// String returns this object as a string.
func (il *IntegerLiteral) String() string { return il.Token.Literal }
