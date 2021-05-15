package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// AssignStatement is used for a assignment statement.
type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) expressionNode() {}

// TokenLiteral returns the literal token.
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }

// String returns this object as a string.
func (as *AssignStatement) String() string {
	if as == nil {
		return ""
	}

	var out bytes.Buffer
	out.WriteString(as.Name.String())
	out.WriteString("=")
	out.WriteString(as.Value.String())
	return out.String()
}
