package ast

import "github.com/skx/evalfilter/token"

// BooleanLiteral holds a boolean type
type BooleanLiteral struct {
	// Token holds the actual token
	Token token.Token

	// Value stores the bools' value: true, or false.
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }

// String returns this object as a string.
func (bl *BooleanLiteral) String() string { return bl.Token.Literal }
