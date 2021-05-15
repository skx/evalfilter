package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// ForeachStatement holds a foreach-statement.
type ForeachStatement struct {
	// Token is the actual token
	Token token.Token

	// Index is the variable we'll set with the index, for the blocks' scope
	//
	// This is optional.
	Index string

	// Ident is the variable we'll set with each item, for the blocks' scope
	Ident string

	// Value is the thing we'll range over.
	Value Expression

	// Body is the block we'll execute.
	Body *BlockStatement
}

func (fes *ForeachStatement) expressionNode() {}

// TokenLiteral returns the literal token.
func (fes *ForeachStatement) TokenLiteral() string { return fes.Token.Literal }

// String returns this object as a string.
func (fes *ForeachStatement) String() string {
	if fes == nil {
		return ""
	}

	var out bytes.Buffer
	out.WriteString("foreach ")
	out.WriteString(fes.Ident)
	out.WriteString(" ")
	out.WriteString(fes.Value.String())
	out.WriteString(fes.Body.String())
	return out.String()
}
