package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// TernaryExpression holds a ternary-expression.
type TernaryExpression struct {
	// Token is the actual token.
	Token token.Token

	// Condition is the thing that is evaluated to determine
	// which expression should be returned
	Condition Expression

	// IfTrue is the expression to return if the condition is true.
	IfTrue Expression

	// IFFalse is the expression to return if the condition is not true.
	IfFalse Expression
}

func (te *TernaryExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (te *TernaryExpression) TokenLiteral() string { return te.Token.Literal }

// String returns this object as a string.
func (te *TernaryExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(te.Condition.String())
	out.WriteString(" ? ")
	out.WriteString(te.IfTrue.String())
	out.WriteString(" : ")
	out.WriteString(te.IfFalse.String())
	out.WriteString(")")

	return out.String()
}
