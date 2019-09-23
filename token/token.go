// Package token contains identifiers for the various logical things
// we find in our source-scripts.
package token

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined Type
const (
	EOF       = "EOF"
	IDENT     = "IDENT"
	ILLEGAL   = "ILLEGAL"
	INT       = "INT"
	FLOAT     = "FLOAT"
	ASSIGN    = "="
	PLUS      = "+"
	AND       = "&&"
	OR        = "||"
	COMMA     = ","
	SEMICOLON = ";"
	MINUS     = "-"
	BANG      = "!"
	ASTERISK  = "*"
	SLASH     = "/"
	LT        = "<"
	LT_EQUALS = "<="
	GT        = ">"
	GT_EQUALS = ">="
	CONTAINS  = "~="
	MISSING   = "!~"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	TRUE      = "TRUE"
	FALSE     = "FALSE"
	IF        = "IF"
	ELSE      = "ELSE"
	RETURN    = "RETURN"
	EQ        = "=="
	NOT_EQ    = "!="
	STRING    = "STRING"
	COLON     = ":"
	PERIOD    = "."
)

// reversed keywords
var keywords = map[string]Type{
	"else":   ELSE,
	"false":  FALSE,
	"if":     IF,
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
