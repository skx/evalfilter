package lexer

import (
	"testing"

	"github.com/skx/evalfilter/token"
)

// TestSomeStrings tests that the input of a pair of strings is tokenized
// appropriately.
func TestSomeStrings(t *testing.T) {
	input := `"Steve" , "Kemp"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.COMMA, ","},
		{token.STRING, "Kemp"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestEscape ensures that strings have escape-characters processed.
func TestStringEscape(t *testing.T) {
	input := `"Steve\n\r\\" "Kemp\n\t\n" "Inline \"quotes\"."`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve\n\r\\"},
		{token.STRING, "Kemp\n\t\n"},
		{token.STRING, "Inline \"quotes\"."},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestIf ensures that an if statement can be handled.
func TestIf(t *testing.T) {
	input := `if ( true ) { print "OK"; } else { print "FAIL"; }`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IF, "if"},
		{token.LBRACKET, "("},
		{token.TRUE, "true"},
		{token.RBRACKET, ")"},
		{token.IDENT, "{"},
		{token.PRINT, "print"},
		{token.STRING, "OK"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "}"},
		{token.ELSE, "else"},
		{token.IDENT, "{"},
		{token.PRINT, "print"},
		{token.STRING, "FAIL"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "}"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestComments ensures that single-line comments work.
func TestComments(t *testing.T) {
	input := `// This is a comment
"Steve"
// This is another comment`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestUnterminated string ensures that an unclosed-string is an error
func TestUnterminatedString(t *testing.T) {
	input := `//
// This is a script
//
print "Steve`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.PRINT, "print"},
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNumber tests that we can parse numbers.
func TestNumber(t *testing.T) {
	input := `//
// This is a with numbers
//
1;
-10;
23;
34.54;
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.NUMBER, "1"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "-10"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "23"},
		{token.SEMICOLON, ";"},
		{token.NUMBER, "34.54"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestContinue checks we continue newlines.
func TestContinue(t *testing.T) {
	input := `//
// Long-lines are fine
//
print "This is a test \
which continues";
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.PRINT, "print"},
		{token.STRING, "This is a test which continues"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestRewind tests that we can rewind our input.
func TestRewind(t *testing.T) {
	input := `"Steve" , "Kemp"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.STRING, "Steve"},
		{token.COMMA, ","},
		{token.COMMA, ","},
		{token.STRING, "Kemp"},
		{token.EOF, ""},
	}

	rewound := false
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}

		if tok.Type == token.COMMA && !rewound {
			l.Rewind(tok)
			rewound = true
		}
	}
}

// TestTypes tests we can see ident, function calls , and variables.
func TestTypes(t *testing.T) {
	input := `foo("steve"); $type`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.FUNCALL, "foo"},
		{token.LBRACKET, "("},
		{token.STRING, "steve"},
		{token.RBRACKET, ")"},
		{token.SEMICOLON, ";"},
		{token.VARIABLE, "type"},
		{token.EOF, ""},
	}

	rewound := false
	l := NewLexer(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}

		if tok.Type == token.COMMA && !rewound {
			l.Rewind(tok)
			rewound = true
		}
	}
}
