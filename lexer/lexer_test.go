package lexer

import (
	"testing"

	"github.com/skx/evalfilter/v2/token"
)

func TestNextToken1(t *testing.T) {
	input := `=+√%(){},;~= !~"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.SQRT, "√"},
		{token.MOD, "%"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
		{token.CONTAINS, "~="},
		{token.MISSING, "!~"},
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := New(input)
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

func TestNextToken2(t *testing.T) {
	input := `five=5;
ten =10;
result = add(five, ten);
!-/ *5;
5 < 10 > 5 ** 6;

if(5<10){
	return true;
}else{
	return false;
}
10 == 10;
10 != 9;
"foobar"
"foo bar";
{"foo":"bar"}
1.2
0.5
0.3
世界
'steve'
'
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.POW, "**"},
		{token.INT, "6"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.FLOAT, "1.2"},
		{token.FLOAT, "0.5"},
		{token.FLOAT, "0.3"},
		{token.IDENT, "世界"},
		{token.STRING, "steve"},
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := New(input)
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

func TestUnicodeLexer(t *testing.T) {
	input := `世界`
	l := New(input)
	tok := l.NextToken()
	if tok.Type != token.IDENT {
		t.Fatalf("token type wrong, expected=%q, got=%q", token.IDENT, tok.Type)
	}
	if tok.Literal != "世界" {
		t.Fatalf("token literal wrong, expected=%q, got=%q", "世界", tok.Literal)
	}
}

func TestSimpleComment(t *testing.T) {
	input := `=+// This is a comment
// This is still a comment
a = 1;
// This is a final
// comment on two-lines`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}
	l := New(input)
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

func TestIntegers(t *testing.T) {
	input := `10 20 33.3 "steve\
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.INT, "10"},
		{token.INT, "20"},
		{token.FLOAT, "33.3"},
		{token.ILLEGAL, "unterminated string"},
		{token.EOF, ""},
	}
	l := New(input)
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

// TestMoreHandling does nothing real, but it bumps our coverage!
func TestMoreHandling(t *testing.T) {
	input := `
t = true;
f = false;

if ( t && f ) { puts( "What?" ); }
if ( t || f ) { puts( "What?" ); }

a = 1;

if ( a<3 ) { puts( "Blah!"); }
if ( a>3 ) { puts( "Blah!"); }

b = 3;
if ( b <= 3  ) { puts "blah\n" }
if ( b >= 3  ) { puts "blah\n" }

a = "steve";
.
 a = "steve\n";
 a = "steve\t";
 a = "steve\r";
 a = "steve\\";
 a = "steve\"";
 c = 3.113£;
 a = "steve\
 kemp";
 b = "steve\
\n\a\b\r\n\t

\`
	l := New(input)
	tok := l.NextToken()
	for tok.Type != token.EOF {
		tok = l.NextToken()
	}
}

// TestEOF handles testing bound-checking
func TestEOF(t *testing.T) {
	input := ` "\`

	// Parse
	l := New(input)

	for {
		tk := l.NextToken()
		if tk.Type == token.EOF {
			break
		}
	}

	// Now we're at the end
	final := l.peekChar()
	if final != rune(0) {
		t.Fatalf("peeking past the EOF didn't give us nil")
	}
}

// TestLine handles testing that line-lengths work.
func TestLine(t *testing.T) {
	input := `
"line 1",
"line 2",
"line 3,\
`

	// Parse
	l := New(input)

	line := 1

	for {
		tk := l.NextToken()
		if tk.Type == token.EOF {
			break
		}
		if tk.Type == token.STRING {

			found := l.GetLine()

			if found != line {
				t.Fatalf("Invalid line number %d vs %d\n", line, l.GetLine())
			}
			line++
		}
	}

}
