// Package lexer contains our lexer.
//
// The lexer returns tokens from a (string) input.  These tokens are then
// parsed as a program to generate an AST, which is used to emit bytecode
// instructions ready for evaluation.
package lexer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/skx/evalfilter/v2/token"
)

// Lexer holds our object-state.
type Lexer struct {
	// The current character position
	position int

	// The next character position
	readPosition int

	// The current character
	ch rune

	// A rune slice of our input string
	characters []rune

	// Previous token.
	prevToken token.Token

	// Line contains our current line-number
	line int

	// column contains the place within the line where we are.
	column int
}

// New creates a Lexer instance from the given string
func New(input string) *Lexer {

	// Line counting starts at one.
	l := &Lexer{characters: []rune(input), line: 1}
	l.readChar()
	return l
}

// read forward one character.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.characters) {
		l.ch = rune(0)
	} else {
		l.ch = l.characters[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	// Line counting
	if l.ch == rune('\n') {
		l.column = 0
		l.line++
	} else {
		l.column++
	}
}

// NextToken reads and returns the next token, skipping any intervening
// white space, and swallowing any comments, in the process.
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
			tok = token.Token{Type: token.AND, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		}
	case rune('|'):
		if l.peekChar() == rune('|') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		}

	case rune('='):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}

	case rune(';'):
		tok = l.newToken(token.SEMICOLON, l.ch)

	case rune('('):
		tok = l.newToken(token.LPAREN, l.ch)

	case rune(')'):
		tok = l.newToken(token.RPAREN, l.ch)

	case rune(','):
		tok = l.newToken(token.COMMA, l.ch)

	case rune('.'):
		if l.peekChar() == rune('.') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.DOTDOT, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.PERIOD, l.ch)
		}

	case rune('+'):
		if l.peekChar() == rune('+') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.PLUSPLUS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.PLUSEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.PLUS, l.ch)
		}

	case rune('%'):
		tok = l.newToken(token.MOD, l.ch)

	case rune('√'):
		tok = l.newToken(token.SQRT, l.ch)

	case rune('{'):
		tok = l.newToken(token.LBRACE, l.ch)

	case rune('}'):
		tok = l.newToken(token.RBRACE, l.ch)

	case rune('['):
		tok = l.newToken(token.LSQUARE, l.ch)

	case rune(']'):
		tok = l.newToken(token.RSQUARE, l.ch)

	case rune('-'):
		if l.peekChar() == rune('-') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MINUSMINUS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.MINUSEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.MINUS, l.ch)
		}

	case rune('/'):

		// slash is mostly division, but could
		// be the start of a regular expression

		// We exclude:
		//   ( a + b ) / c   -> RPAREN
		//   a / c           -> IDENT
		//   foo[3] / 3      -> INDEX
		//   3.2 / c         -> FLOAT
		//   1 / c           -> INT
		//
		if l.prevToken.Type == token.RPAREN ||
			l.prevToken.Type == token.IDENT ||
			l.prevToken.Type == token.RSQUARE ||
			l.prevToken.Type == token.FLOAT ||
			l.prevToken.Type == token.INT {

			if l.peekChar() == rune('=') {
				ch := l.ch
				l.readChar()
				tok = token.Token{Type: token.SLASHEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
			} else {
				tok = l.newToken(token.SLASH, l.ch)
			}
		} else {
			str, err := l.readRegexp()
			if err == nil {
				tok.Column = l.column
				tok.Line = l.line
				tok.Literal = str
				tok.Type = token.REGEXP

			} else {
				tok.Column = l.column
				tok.Line = l.line
				tok.Literal = err.Error()
				tok.Type = token.ILLEGAL
			}
			return tok
		}
	case rune('*'):
		if l.peekChar() == rune('*') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.POW, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASTERISKEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.ASTERISK, l.ch)
		}

	case rune('?'):
		tok = l.newToken(token.QUESTION, l.ch)
	case rune(':'):
		tok = l.newToken(token.COLON, l.ch)

	case rune('<'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.LTEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.LT, l.ch)
		}

	case rune('>'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.GTEQUALS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			tok = l.newToken(token.GT, l.ch)
		}

	case rune('~'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.CONTAINS, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		}

	case rune('!'):
		if l.peekChar() == rune('=') {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.NOTEQ, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
		} else {
			if l.peekChar() == rune('~') {
				ch := l.ch
				l.readChar()
				tok = token.Token{Type: token.MISSING, Literal: string(ch) + string(l.ch), Line: l.line, Column: l.column}
			} else {
				tok = l.newToken(token.BANG, l.ch)
			}
		}

	case rune('"'):
		str, err := l.readString('"')
		if err == nil {
			tok.Column = l.column
			tok.Line = l.line
			tok.Literal = str
			tok.Type = token.STRING
		} else {
			tok.Column = l.column
			tok.Line = l.line
			tok.Literal = err.Error()
			tok.Type = token.ILLEGAL
		}

	case rune('\''):
		str, err := l.readString('\'')

		if err == nil {
			tok.Column = l.column
			tok.Line = l.line
			tok.Literal = str
			tok.Type = token.STRING
		} else {
			tok.Column = l.column
			tok.Line = l.line
			tok.Literal = err.Error()
			tok.Type = token.ILLEGAL
		}

	case rune(0):
		tok.Literal = ""
		tok.Type = token.EOF

	default:
		if isDigit(l.ch) {

			tok := l.readDecimal()
			l.prevToken = tok
			tok.Column = l.column
			tok.Line = l.line
			return tok
		}

		tok.Literal = l.readIdentifier()
		if len(tok.Literal) > 0 {
			tok.Type = token.LookupIdentifier(tok.Literal)
			l.prevToken = tok
			tok.Column = l.column
			tok.Line = l.line
			return tok
		}
		tok.Type = token.ILLEGAL
		tok.Literal = fmt.Sprintf("invalid character for indentifier '%c'", l.ch)
		tok.Column = l.column
		tok.Line = l.line
		l.readChar()
		return tok

	}

	l.readChar()

	l.prevToken = tok

	return tok
}

// return new token
func (l *Lexer) newToken(tokenType token.Type, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Line: l.line, Column: l.column}
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

// read a regexp, including flags.
func (l *Lexer) readRegexp() (string, error) {
	out := ""

	for {
		l.readChar()

		if l.ch == rune(0) {
			return "", fmt.Errorf("unterminated regular expression")
		}
		if l.ch == '/' {

			// consume the terminating "/".
			l.readChar()

			// prepare to look for flags
			flags := ""

			// two flags are supported:
			//   i -> Ignore-case
			//   m -> Multiline
			//
			// We need to consume all letters, so we can
			// alert on illegal ones.
			for unicode.IsLetter(l.ch) {

				// save the char - unless it is a repeat
				if !strings.Contains(flags, string(l.ch)) {

					// we're going to sort the flags
					tmp := strings.Split(flags, "")
					tmp = append(tmp, string(l.ch))
					flags = strings.Join(tmp, "")

				}

				// read the next
				l.readChar()
			}

			for _, c := range flags {
				switch c {
				case 'i', 'm':
					// nop
				default:
					return "", fmt.Errorf("illegal regexp flag '%c' in string '%s'", c, flags)
				}
			}
			// convert the regexp to go-lang
			if len(flags) > 0 {
				out = "(?" + flags + ")" + out
			}
			break
		}
		if l.ch == '\\' {
			// Skip the escape-marker, and read the
			// escaped character literally.
			l.readChar()
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

// determinate ch is identifier or not.  Identifiers may be alphanumeric,
// but they must start with a letter.  Here that works because we are only
// called if the first character is alphabetical.
func isIdentifier(ch rune) bool {
	if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '$' || ch == '_' {
		return true
	}
	return false
}

// is white space
func isWhitespace(ch rune) bool {
	return ch == rune(' ') || ch == rune('\t') || ch == rune('\n') || ch == rune('\r')
}

// is Digit
func isDigit(ch rune) bool {
	return rune('0') <= ch && ch <= rune('9')
}
