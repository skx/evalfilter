package ast

import (
	"strings"

	"github.com/skx/evalfilter/v2/token"
)

// RegexpLiteral holds a regular-expression.
type RegexpLiteral struct {
	// Token is the token
	Token token.Token

	// Value is the value of the regular expression.
	Value string
}

func (rl *RegexpLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (rl *RegexpLiteral) TokenLiteral() string { return rl.Token.Literal }

// String returns this object as a string.
func (rl *RegexpLiteral) String() string {

	start := "/"
	val := rl.Token.Literal
	end := "/"

	if strings.HasPrefix(rl.Token.Literal, "(?i)") {
		end = "/i"
		val = strings.TrimPrefix(val, "(?i)")
	}

	str := start + val + end
	return str
}
