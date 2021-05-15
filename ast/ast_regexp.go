package ast

import (
	"fmt"

	"github.com/skx/evalfilter/v2/token"
)

// RegexpLiteral holds a regular-expression.
type RegexpLiteral struct {
	// Token is the token
	Token token.Token

	// Value is the value of the regular expression.
	Value string

	// Flags contains any flags associated with the regexp.
	Flags string
}

func (rl *RegexpLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (rl *RegexpLiteral) TokenLiteral() string { return rl.Token.Literal }

// String returns this object as a string.
func (rl *RegexpLiteral) String() string {
	if rl == nil {
		return ""
	}

	return (fmt.Sprintf("/%s/%s", rl.Value, rl.Flags))
}
