package ast

import (
	"bytes"
	"strings"

	"github.com/skx/evalfilter/v2/token"
)

// CallExpression holds the invokation of a method-call.
type CallExpression struct {
	// Token stores the literal token
	Token token.Token

	// Function is the function to be invoked.
	Function Expression

	// Arguments are the arguments to be applied
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// String returns this object as a string.
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := make([]string, 0)
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(");\n")
	return out.String()
}
