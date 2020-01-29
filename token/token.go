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

// Our known token-types
const (
	AND       = "&&"
	ASSIGN    = "="
	ASTERISK  = "*"
	BANG      = "!"
	COLON     = ":"
	COMMA     = ","
	CONTAINS  = "~="
	DOTDOT    = ".."
	ELSE      = "ELSE"
	EOF       = "EOF"
	EQ        = "=="
	FALSE     = "FALSE"
	FLOAT     = "FLOAT"
	FOREACH   = "FOREACH"
	GT        = ">"
	GTEQUALS  = ">="
	IDENT     = "IDENT"
	IF        = "IF"
	ILLEGAL   = "ILLEGAL"
	IN        = "IN"
	INT       = "INT"
	LBRACE    = "{"
	LPAREN    = "("
	LSQUARE   = "["
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
	QUESTION  = "?"
	RBRACE    = "}"
	REGEXP    = "REGEXP"
	RETURN    = "RETURN"
	RPAREN    = ")"
	RSQUARE   = "]"
	SEMICOLON = ";"
	SLASH     = "/"
	SQRT      = "√"
	STRING    = "STRING"
	TRUE      = "TRUE"
	WHILE     = "WHILE"
)

// reversed keywords
var keywords = map[string]Type{
	"else":    ELSE,
	"false":   FALSE,
	"foreach": FOREACH,
	"if":      IF,
	"in":      IN,
	"return":  RETURN,
	"true":    TRUE,
	"while":   WHILE,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}
