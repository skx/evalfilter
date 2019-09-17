package evalfilter

import (
	"errors"
	"strings"
)

// Lexer holds our object-state.
type Lexer struct {
	//current character position
	position int

	//next character position
	readPosition int

	//current character
	ch rune

	//rune slice of input string
	characters []rune
}

// NewLexer creates a Lexer instance from string input.
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

// NextToken to read next token, skipping the white space.
func (l *Lexer) NextToken() Token {

	var tok Token
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
			tok.Type = STRING
			tok.Literal = str
		} else {
			tok.Type = ILLEGAL
			tok.Literal = err.Error()
		}
	case rune(0):
		tok.Literal = ""
		tok.Type = EOF
	case rune(';'):
		tok.Literal = ";"
		tok.Type = SEMICOLON
	default:
		if isDigit(l.ch) {
			return l.readDecimal()
		}
		tok.Literal = l.readIdentifier()
		tok.Type = LookupIdentifier(tok.Literal)
		return tok
	}
	l.readChar()
	return tok
}

// read Identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isIdentifier(l.ch) {
		l.readChar()
	}
	return string(l.characters[position:l.position])
}

// skip white space
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

// read string
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

// read decimal
func (l *Lexer) readDecimal() Token {

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
		return Token{Type: NUMBER, Literal: integer + "." + fraction}
	}
	return Token{Type: NUMBER, Literal: integer}
}

// read number
func (l *Lexer) readNumber() string {
	str := ""

	for strings.Contains("0123456789", string(l.ch)) {
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

// determinate ch is identifier or not
func isIdentifier(ch rune) bool {
	return !isWhitespace(ch) && !isEmpty(ch) && ch != rune(';')
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is empty
func isEmpty(ch rune) bool {
	return rune(0) == ch
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
