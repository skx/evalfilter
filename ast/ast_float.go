package ast

import "github.com/skx/evalfilter/v2/token"

// FloatLiteral holds a floating-point number
type FloatLiteral struct {
	// Token is the literal token
	Token token.Token

	// Value holds the floating-point number.
	Value float64
}

func (fl *FloatLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

// String returns this object as a string.
func (fl *FloatLiteral) String() string { return fl.Token.Literal }
