// Package token holds the definitions of the various token-types
// we require to lex our source code.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined TokenTypes
const (
	COMMA     = "COMMA"
	EOF       = "EOF"
	FUNCALL   = "FUNCALL"
	IDENT     = "IDENT"
	ILLEGAL   = "ILLEGAL"
	LBRACKET  = "LBRACKET"
	NUMBER    = "NUMBER"
	RBRACKET  = "RBRACKET"
	SEMICOLON = "SEMICOLON"
	STRING    = "STRING"
	VARIABLE  = "VARIABLE"

	// Our keywords.
	ELSE   = "else"
	FALSE  = "false"
	IF     = "if"
	PRINT  = "print"
	RETURN = "return"
	TRUE   = "true"
)

// keywords holds our reversed keywords
var keywords = map[string]Type{
	"else":   ELSE,
	"false":  FALSE,
	"if":     IF,
	"print":  PRINT,
	"return": RETURN,
	"true":   TRUE,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
