package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// IfExpression holds an if-statement.
type IfExpression struct {
	// Token is the actual token
	Token token.Token

	// Condition is the thing that is evaluated to determine
	// which block should be executed.
	Condition Expression

	// Consequence is the set of statements executed if the
	// condition is true.
	Consequence *BlockStatement

	// Alternative is the set of statements executed if the
	// condition is not true (optional).
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns this object as a string.
func (ie *IfExpression) String() string {
	if ie == nil {
		return ""
	}

	var out bytes.Buffer
	out.WriteString("\nif (")
	out.WriteString(ie.Condition.String())
	out.WriteString(") ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}
