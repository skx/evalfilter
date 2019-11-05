// Package token contains identifiers for the various things
// we find in our source-scripts.
//
// Our lexer will convert an input-script into a series of tokens,
// which will then be consumed by the parser and transformed into an
// AST.
//
// Once we have the AST we will then generate a series of bytecode
// instructions which will ultimately be executed by our virtual machine.
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
	COMMA     = ","
	CONTAINS  = "~="
	ELSE      = "ELSE"
	EOF       = "EOF"
	EQ        = "=="
	FALSE     = "FALSE"
	FLOAT     = "FLOAT"
	GT        = ">"
	GTEQUALS  = ">="
	IDENT     = "IDENT"
	IF        = "IF"
	ILLEGAL   = "ILLEGAL"
	INT       = "INT"
	LBRACE    = "{"
	LPAREN    = "("
	LT        = "<"
	LTEQUALS  = "<="
	MINUS     = "-"
	MISSING   = "!~"
	MOD       = "%"
	NOTEQ     = "!="
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
