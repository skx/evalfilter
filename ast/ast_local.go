package ast

import (
	"bytes"

	"github.com/skx/evalfilter/v2/token"
)

// LocalVariable is used for declaring a local variable.
//
// A local-variable may only be defined within the body
// of a function, and if used will only have scope until
// the function terminates.
type LocalVariable struct {
	Token token.Token
}

func (lv *LocalVariable) expressionNode() {}

// TokenLiteral returns the literal token.
func (lv *LocalVariable) TokenLiteral() string { return lv.Token.Literal }

// String returns this object as a string.
func (lv *LocalVariable) String() string {
	var out bytes.Buffer
	out.WriteString("local ")
	out.WriteString(lv.Token.Literal)
	out.WriteString(";\n")
	return out.String()
}
