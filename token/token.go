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
	AND       = "&&"
	ASSIGN    = "="
	ASTERISK  = "*"
	BANG      = "!"
	COLON     = ":"
	COMMA     = ","
	CONTAINS  = "~="
	ELSE      = "ELSE"
	EOF       = "EOF"
	EQ        = "=="
	FALSE     = "FALSE"
	FLOAT     = "FLOAT"
	GT        = ">"
	GT_EQUALS = ">="
	IDENT     = "IDENT"
	IF        = "IF"
	ILLEGAL   = "ILLEGAL"
	INT       = "INT"
	LBRACE    = "{"
	LPAREN    = "("
	LT        = "<"
	LT_EQUALS = "<="
	MINUS     = "-"
	MISSING   = "!~"
	MOD       = "%"
	NOT_EQ    = "!="
	OR        = "||"
	PERIOD    = "."
	PLUS      = "+"
	POW       = "**"
	RBRACE    = "}"
	RETURN    = "RETURN"
	RPAREN    = ")"
	SEMICOLON = ";"
	SLASH     = "/"
	SQRT      = "√"
	STRING    = "STRING"
	TRUE      = "TRUE"
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
