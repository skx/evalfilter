package evalfilter

// Type is a string
type Type string

// Token struct represent the lexer token
type Token struct {
	Type    Type
	Literal string
}

// pre-defined TokenTypes
const (
	EOF       = "EOF"
	IDENT     = "IDENT"
	ILLEGAL   = "ILLEGAL"
	NUMBER    = "NUMBER"
	STRING    = "STRING"
	SEMICOLON = "SEMICOLON"

	// Our keywords.
	FALSE  = "false"
	IF     = "if"
	PRINT  = "print"
	RETURN = "return"
	TRUE   = "true"
)

// keywords holds our reversed keywords
var keywords = map[string]Type{
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
