package ast

import (
	"bytes"
	"strings"

	"github.com/skx/evalfilter/v2/token"
)

// ArrayLiteral holds an inline array
type ArrayLiteral struct {
	// Token is the token
	Token token.Token

	// Elements holds the members of the array.
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the literal token.
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

// String returns this object as a string.
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := make([]string, 0)
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("];\n")
	return out.String()
}

// IndexExpression holds an index-expression
type IndexExpression struct {
	// Token is the actual token
	Token token.Token

	// Left is the thing being indexed.
	Left Expression

	// Index is the value we're indexing
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns this object as a string.
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}
