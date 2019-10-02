package ast

import (
	"strings"

	"github.com/skx/evalfilter/token"
)

// StringLiteral holds a string
type StringLiteral struct {
	// Token is the token
	Token token.Token

	// Value is the value of the string.
	Value string
}

func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// String returns this object as a string.
func (sl *StringLiteral) String() string {
	str := "\"" + sl.Token.Literal + "\""

	str = strings.ReplaceAll(str, "\n", "\\n")
	return str
}
