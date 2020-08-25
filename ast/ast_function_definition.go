package ast

import (
	"bytes"
	"strings"

	"github.com/skx/evalfilter/v2/token"
)

// FunctionDefinition holds the definition of a user-defined function.
//
// This means a function which is defined within our scripting-language,
// rather than a function implemented in Golang and exported to the
// scripting environmnet.
type FunctionDefinition struct {

	// Token holds the name of the function.
	Token token.Token

	// Paremeters holds the function parameters.
	Parameters []*Identifier

	// Body holds the set of statements in the functions' body.
	Body *BlockStatement
}

func (fd *FunctionDefinition) expressionNode() {}

// TokenLiteral returns the literal token.
func (fd *FunctionDefinition) TokenLiteral() string {
	return fd.Token.Literal
}

// String returns this object as a string.
func (fd *FunctionDefinition) String() string {
	var out bytes.Buffer
	params := make([]string, 0)
	for _, p := range fd.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("function ")
	out.WriteString(fd.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(fd.Body.String())
	return out.String()
}
