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

import "fmt"

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {

	// Type contains the type of the token
	// "AND", "ASSIGN", "STRING", etc.
	Type Type

	// Literal contains the literal text of the token
	Literal string

	// Line contains the line within the input where the
	// token was found.
	Line int

	// Column contains the position where the token (start)
	// was found
	Column int
}

// Our known token-types
const (
	AND            = "&&"
	ASSIGN         = "="
	ASTERISK       = "*"
	ASTERISKEQUALS = "*="
	BANG           = "!"
	CASE           = "case"
	COLON          = ":"
	COMMA          = ","
	CONTAINS       = "~="
	DEFAULT        = "DEFAULT"
	DOTDOT         = ".."
	ELSE           = "ELSE"
	EOF            = "EOF"
	EQ             = "=="
	FALSE          = "FALSE"
	FLOAT          = "FLOAT"
	FOR            = "FOR"
	FOREACH        = "FOREACH"
	FUNCTION       = "FUNCTION"
	GT             = ">"
	GTEQUALS       = ">="
	IDENT          = "IDENT"
	IF             = "IF"
	ILLEGAL        = "ILLEGAL"
	IN             = "IN"
	INT            = "INT"
	LBRACE         = "{"
	LOCAL          = "LOCAL"
	LPAREN         = "("
	LSQUARE        = "["
	LT             = "<"
	LTEQUALS       = "<="
	MINUS          = "-"
	MINUSEQUALS    = "-="
	MINUSMINUS     = "--"
	MISSING        = "!~"
	MOD            = "%"
	NOTEQ          = "!="
	OR             = "||"
	PERIOD         = "."
	PLUS           = "+"
	PLUSPLUS       = "++"
	PLUSEQUALS     = "+="
	POW            = "**"
	QUESTION       = "?"
	RBRACE         = "}"
	REGEXP         = "REGEXP"
	RETURN         = "RETURN"
	RPAREN         = ")"
	RSQUARE        = "]"
	SEMICOLON      = ";"
	SLASH          = "/"
	SLASHEQUALS    = "/="
	SQRT           = "√"
	STRING         = "STRING"
	SWITCH         = "switch"
	TRUE           = "TRUE"
	WHILE          = "WHILE"
)

// reversed keywords
var keywords = map[string]Type{
	"case":     CASE,
	"default":  DEFAULT,
	"else":     ELSE,
	"false":    FALSE,
	"for":      FOR,
	"foreach":  FOREACH,
	"function": FUNCTION,
	"if":       IF,
	"in":       IN,
	"local":    LOCAL,
	"return":   RETURN,
	"switch":   SWITCH,
	"true":     TRUE,
	"while":    WHILE,
}

// LookupIdentifier used to determinate whether identifier is keyword nor not
func LookupIdentifier(identifier string) Type {
	if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENT
}

// Position returns a report of the current token's position, reporting on
// the line-number and column-number of the token.
func (t Token) Position() string {
	return (fmt.Sprintf("line %d, column %d", t.Line, t.Column))
}
