// Package lexer implements a simple lexer, which will parse
// user-scripts to be executed by evalfilter.
package lexer

import (
	"errors"
	"strings"

	"github.com/skx/evalfilter/token"
)

// Lexer holds our object-state.
type Lexer struct {
	// The current character position
	position int

	// The next character position
	readPosition int

	// The current character
	ch rune

	// The input string we're reading from.
	characters []rune
}

// NewLexer creates a Lexer instance from the specified string input.
func NewLexer(input string) *Lexer {
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// read one forward character
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken will read the next token, skipping any white space, and ignoring
// comments - which begin with `//`.
func (l *Lexer) NextToken() token.Token {

	var tok token.Token
	l.skipWhitespace()

	// skip single-line comments
	if l.ch == rune('/') && l.peekChar() == rune('/') {
		l.skipComment()
		return (l.NextToken())
	}

	switch l.ch {

	case rune('"'):
		str, err := l.readString()

		if err == nil {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	case rune(';'):
		tok.Literal = ";"
		tok.Type = token.SEMICOLON
	case rune(','):
		tok.Literal = ","
		tok.Type = token.COMMA
	case rune('('):
		tok.Literal = "("
		tok.Type = token.LBRACKET
	case rune(')'):
		tok.Literal = ")"
		tok.Type = token.RBRACKET
	default:
		if l.ch == '-' || isDigit(l.ch) {
			return l.readDecimalNumber()
		}

		//
		// Here we have something that could be
		// a literal, or could be a function-call
		//
		// peek at the next token to decide.
		//
		tok.Literal = l.readIdentifier()

		if l.ch == '(' {
			tok.Type = token.FUNCALL
		} else {
			if strings.HasPrefix(tok.Literal, "$") {
				tok.Type = token.VARIABLE
				tok.Literal = tok.Literal[1:]
			} else {
				tok.Type = token.LookupIdentifier(tok.Literal)
			}
		}
		return tok
	}
	l.readChar()
	return tok
}

// read and return the name of an identifier
func (l *Lexer) readIdentifier() string {
	id := ""
	for isIdentifier(l.ch) {
		id += string(l.ch)
		l.readChar()
	}
	return id
}

// skip white space, consuming input as we go.
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read a quote-terminated string
func (l *Lexer) readString() (string, error) {
	out := ""

	for {
		l.readChar()
		if l.ch == '"' {
			break
		}
		if l.ch == rune(0) {
			return "", errors.New("unterminated string")
		}

		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {

			// Line ending with "\" + newline
			if l.peekChar() == '\n' {
				// consume the newline.
				l.readChar()
				continue
			}

			l.readChar()

			if l.ch == rune('n') {
				l.ch = '\n'
			}
			if l.ch == rune('r') {
				l.ch = '\r'
			}
			if l.ch == rune('t') {
				l.ch = '\t'
			}
			if l.ch == rune('"') {
				l.ch = '"'
			}
			if l.ch == rune('\\') {
				l.ch = '\\'
			}
		}
		out = out + string(l.ch)

	}

	return out, nil
}

// read a decimal / floating-point number
func (l *Lexer) readDecimalNumber() token.Token {

	//
	// Read an integer-number.
	//
	integer := l.readNumber()

	//
	// Now we might expect further digits, after a dot:
	//
	//   .[digits]  -> Which converts us from an int to a float.
	//
	if l.ch == rune('.') && isDigit(l.peekChar()) {
		//
		// OK here we think we've got a float.
		//
		l.readChar()
		fraction := l.readNumber()
		return token.Token{Type: token.NUMBER, Literal: integer + "." + fraction}
	}

	//
	// OK just an integer.
	//
	return token.Token{Type: token.NUMBER, Literal: integer}
}

// read a numeric digit
func (l *Lexer) readNumber() string {
	str := ""

	for isDigit(l.ch) || (len(str) == 0 && l.ch == '-') {
		str += string(l.ch)
		l.readChar()
	}
	return str
}

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate whether `ch` is a character permitted in an identifier or not.
func isIdentifier(ch rune) bool {
	return !isWhitespace(ch) && !isEmpty(ch) && !isSpecial(ch)
}

// is the character white space?
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is this a special character?
func isSpecial(ch rune) bool {
	return ch == rune(',') || ch == rune(';') || ch == rune('(') || ch == rune(')')
}

// is this character empty?
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// is this character a digit?
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
