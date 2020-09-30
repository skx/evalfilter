package ast

import (
	"bytes"
	"strings"

	"github.com/skx/evalfilter/v2/token"
)

// CaseExpression handles the case within a switch statement
type CaseExpression struct {
	// Token is the actual token
	Token token.Token

	// Default branch?
	Default bool

	// The thing we match
	Expr []Expression

	// The code to execute if there is a match
	Block *BlockStatement
}

func (ce *CaseExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (ce *CaseExpression) TokenLiteral() string { return ce.Token.Literal }

// String returns this object as a string.
func (ce *CaseExpression) String() string {
	var out bytes.Buffer

	if ce.Default {
		out.WriteString("default ")
	} else {
		out.WriteString("case ")

		tmp := []string{}
		for _, exp := range ce.Expr {
			tmp = append(tmp, exp.String())
		}
		out.WriteString(strings.Join(tmp, ","))
	}
	out.WriteString(ce.Block.String())
	return out.String()
}

// SwitchExpression handles a switch statement
type SwitchExpression struct {
	// Token is the actual token
	Token token.Token

	// Value is the thing that is evaluated to determine
	// which block should be executed.
	Value Expression

	// The branches we handle
	Choices []*CaseExpression
}

func (se *SwitchExpression) expressionNode() {}

// TokenLiteral returns the literal token.
func (se *SwitchExpression) TokenLiteral() string { return se.Token.Literal }

// String returns this object as a string.
func (se *SwitchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("\nswitch (")
	out.WriteString(se.Value.String())
	out.WriteString(")\n{\n")

	for _, tmp := range se.Choices {
		if tmp != nil {
			out.WriteString(tmp.String())
		}
	}
	out.WriteString("}\n")

	return out.String()
}
