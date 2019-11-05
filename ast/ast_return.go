package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// ReturnStatement stores a return-statement
type ReturnStatement struct {
	// Token contains the literal token.
	Token token.Token

	// ReturnValue is the value whichis to be returned.
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the literal token.
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// String returns this object as a string.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.TokenLiteral())
	}
	out.WriteString(";")
	return out.String()
}
