package ast

import "github.com/skx/evalfilter/token"

// Boolean holds a boolean type
type Boolean struct {
	// Token holds the actual token
	Token token.Token

	// Value stores the bools' value: true, or false.
	Value bool
}

func (b *Boolean) expressionNode() {}

// TokenLiteral returns the literal token.
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// String returns this object as a string.
func (b *Boolean) String() string { return b.Token.Literal }
