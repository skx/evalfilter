// Package lexer contains our simple lexer.
//
// The lexer returns tokens from a (string) input, as a series of Token
// objects.
package lexer

import (
	"errors"
	"fmt"

	"github.com/skx/evalfilter/token"
)

// Lexer holds our object-state.
type Lexer struct {
	// The current character position
	position int

	// The next character position
	readPosition int

	//The current character
	ch rune

	// A rune slice of our input string
	characters []rune
}

// New creates a Lexer instance from the given string
func New(input string) *Lexer {
	l := &Lexer{characters: []rune(input)}
	l.readChar()
	return l
}

// GetLine returns the rough line-number of our current position.
//
// This is used to report errors in a more humane manner.
func (l *Lexer) GetLine() int {
	line := 0
	chars := len(l.characters)
	i := 0

	for i < l.readPosition && i < chars {

		if l.characters[i] == rune('\n') {
			line++
		}

		i++
	}
	return line
}

// read one forward character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken reads and returns the next token, skipping any white space.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	// skip single-line comments
	if l.ch == rune('/') && l.peekChar() == rune('/') {
		l.skipComment()
		return (l.NextToken())
	}

	switch l.ch {
	case rune('&'):
		if l.peekChar() == rune('&') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch)}
		}
	case rune('|'):
		if l.peekChar() == rune('|') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch)}
		}

	case rune('='):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case rune(';'):
		tok = newToken(token.SEMICOLON, l.ch)
	case rune('('):
		tok = newToken(token.LPAREN, l.ch)
	case rune(')'):
		tok = newToken(token.RPAREN, l.ch)
	case rune(','):
		tok = newToken(token.COMMA, l.ch)
	case rune('.'):
		tok = newToken(token.PERIOD, l.ch)
	case rune('+'):
		tok = newToken(token.PLUS, l.ch)
	case rune('%'):
		tok = newToken(token.MOD, l.ch)
	case rune('√'):
		tok = newToken(token.SQRT, l.ch)
	case rune('{'):
		tok = newToken(token.LBRACE, l.ch)
	case rune('}'):
		tok = newToken(token.RBRACE, l.ch)
	case rune('-'):
		tok = newToken(token.MINUS, l.ch)
	case rune('/'):
		tok = newToken(token.SLASH, l.ch)
	case rune('*'):
		if l.peekChar() == rune('*') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.POW, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.ASTERISK, l.ch)
		}
	case rune('<'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case rune('>'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case rune('~'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.CONTAINS, Literal: string(ch) + string(l.ch)}
		}

	case rune('!'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: string(ch) + string(l.ch)}
		} else {
			if l.peekChar() == rune('~') {
				ch := l.ch
				l.readChar()
				tok = token.Token{Type: token.MISSING, Literal: string(ch) + string(l.ch)}
			} else {
				tok = newToken(token.BANG, l.ch)
			}
		}
	case rune('"'):
		str, err := l.readString('"')
		if err == nil {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	case rune('\''):
		str, err := l.readString('\'')

		if err == nil {
			tok.Type = token.STRING
			tok.Literal = str
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = err.Error()
		}
	case rune(':'):
		tok = newToken(token.COLON, l.ch)
	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isDigit(l.ch) {
			return l.readDecimal()
		}
		tok.Literal = l.readIdentifier()
		tok.Type = token.LookupIdentifier(tok.Literal)
		return tok
	}
	l.readChar()
	return tok
}

// return new token
func newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// readIdentifier is designed to read an identifier (name of variable,
// function, etc).
func (l *Lexer) readIdentifier() string {

	id := ""

	for isIdentifier(l.ch) {
		id += string(l.ch)
		l.readChar()
	}
	return id
}

// skip over any white space.
func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.ch) {
		l.readChar()
	}
}

// skip a comment (until the end of the line).
func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != rune(0) {
		l.readChar()
	}
	l.skipWhitespace()
}

// read a number.  We only care about numerical digits here, floats will
// be handled elsewhere.
func (l *Lexer) readNumber() string {

	id := ""

	for isDigit(l.ch) {
		id += string(l.ch)
		l.readChar()
	}
	return id
}

// read a decimal number, either int or floating-point.
func (l *Lexer) readDecimal() token.Token {

	//
	// Read an integer-number.
	//
	integer := l.readNumber()

	//
	// If the next token is a `.` we've got a floating-point number.
	//
	if l.ch == rune('.') && isDigit(l.peekChar()) {

		// Skip the period
		l.readChar()

		// Get the float-component.
		fraction := l.readNumber()
		return token.Token{Type: token.FLOAT, Literal: integer + "." + fraction}
	}

	//
	// Just an integer.
	//
	return token.Token{Type: token.INT, Literal: integer}
}

// read a string, deliminated by the given character.
func (l *Lexer) readString(delim rune) (string, error) {
	out := ""

	for {
		l.readChar()

		if l.ch == rune(0) {
			return "", fmt.Errorf("unterminated string")
		}
		if l.ch == delim {
			break
		}
		//
		// Handle \n, \r, \t, \", etc.
		//
		if l.ch == '\\' {

			// Line ending with "\" + newline
			if l.peekChar() == '\n' {
				// consume the newline.
				l.readChar()
				if l.ch == rune(0) {
					return "", fmt.Errorf("unterminated string")
				}
				continue
			}

			l.readChar()

			if l.ch == rune(0) {
				return "", errors.New("unterminated string")
			}
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

// peek character
func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.characters) {
		return rune(0)
	}
	return l.characters[l.readPosition]
}

// determinate ch is identifier or not
func isIdentifier(ch rune) bool {
	return !isDigit(ch) && !isWhitespace(ch) && !isBrace(ch) && !isOperator(ch) && !isComparison(ch) && !isCompound(ch) && !isBrace(ch) && !isParen(ch) && !isEmpty(ch)
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is operators
func isOperator(ch rune) bool {
	return ch == rune('+') || ch == rune('-') || ch == rune('/') || ch == rune('*')
}

// is comparison
func isComparison(ch rune) bool {
	return ch == rune('=') || ch == rune('!') || ch == rune('>') || ch == rune('<') || ch == rune('~')
}

// is compound
func isCompound(ch rune) bool {
	return ch == rune(',') || ch == rune(':') || ch == rune('"') || ch == rune(';')
}

// is brace
func isBrace(ch rune) bool {
	return ch == rune('{') || ch == rune('}')
}

// is parenthesis
func isParen(ch rune) bool {
	return ch == rune('(') || ch == rune(')')
}

// is empty
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
