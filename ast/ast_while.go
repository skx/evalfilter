package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// WhileStatement holds a while-statement.
type WhileStatement struct {
	// Token is the actual token
	Token token.Token

	// Condition is the thing that is evaluated to determine
	// whether the block should be executed.
	Condition Expression

	// Body is the set of statements executed if the
	// condition is true.
	Body *BlockStatement
}

func (ws *WhileStatement) expressionNode() {}

// TokenLiteral returns the literal token.
func (ie *WhileStatement) TokenLiteral() string { return ie.Token.Literal }

// String returns this object as a string.
func (ie *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while (")
	out.WriteString(ie.Condition.String())
	out.WriteString(") {")
	out.WriteString(ie.Body.String())
	out.WriteString("}")
	return out.String()
}
