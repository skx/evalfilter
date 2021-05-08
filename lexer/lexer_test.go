package lexer

import (
	"testing"

	"github.com/skx/evalfilter/v2/token"
)

func TestTokenPosition(t *testing.T) {
	input := `


a = { "Name": "Steve",
      "Age": 2020 - 1976,
      "Location": "Helsinki", }


k = keys(a);

printf("Iterating over the hash via the output of keys():\n");
foreach name in k {
}

printf("Iterating via foreach key,value:\n");
foreach key, value in a {
  printf("\t%s -> %s\n", key, string(value) )
}

printf("I am %s - %d\n", a["Name"], a["Age"] );
printf("My hash is %s\n", string(a));


print( "Hello, world!\n" );

return false;

if ( √9 == 3 ) { print("Mathematics is fun\n"); }

return false;
{"Author":"Bob",
 "Channel": "tech_alerts",
 "Message": "Database on Fire!"}



if ( Channel !~ /_Alerts$/i ) {
    print("The channel " + Channel + " is not important\n");
    return false;
}


staff = [ "Alice", "Bob", "Chris", "Diane" ];

if ( Author in staff ) {
    print( "Ignoring message from staff-member ", Author, "\n" );
    return false;
}


if ( Message ~= /backup complete/ ) {
   return false;
}

if ( Channel == "dev_alerts" ) {
   print( "development-environment is not worth waking somebody over\n" );
   return false;
}

if ( weekday(now()) == "Saturday" || weekday(now()) == "Sunday" ) {

    print("This is the weekend, triggering call ..\n" );
    return true;
}

if ( hour(now()) <= 7 || hour(now()) >= 18 ) {
   print( "Outside working hours on a week-day - triggering call\n" );
   return true;
}


print( "Working hours, on a working day, ignoring the event\n");

return false;


if ( 1 + 2 * 3 == 7 ) {
   print("OK\n");
}

return( true );

i = 3;

print( "Starting value: ", i, "\n");

while( i < 10 ) {
   print("In loop, value: ", i, "\n");
   i++
}

print( "Completed.  Value is: ", i, "\n");
return 0;


x = 1;
y = 2;

printf("Start: x => %d, y => %d, z is not defined (%v)\n", x, y, z );

tmp(x);

printf("End: x => %d, y => %d, z is defined (%d)\n", x, y, z );

if ( test ) {
}
else {
  printf("The local variable did not leak, and is null as expected: %v\n", test);
}


return 1;





function tmp(x) {

   x = 10;

   y = 100;

   z = 1000;

   local test;
   test = 33;

   printf(" Inside the function we have a local-variable: test=>%d\n", test);
   test = test * 2;
   printf(" Inside the function the local-variable can be updated: test=>%d\n", test);


   printf(" At function end: x => %d, y => %d, z => %d, test => %d\n", x, y, z, test );


input = [ "Hello", "my", "name", "is", "Steve" ];

print("Showing array items:\n");
foreach index,entry in input {
   printf("\t%d:%s\n", index, entry );
}


print( "Showing sorted array items:\n");
sorted = sort(input);
foreach index,entry in sorted {
   printf("\t%d:%s\n", index, entry );
}

print( "Showing reversed array items:\n");
reversed = reverse(input);
foreach entry in reversed {
   print("\t", entry, "\n" );
}

print( "\n" );
print( "Original input:", input, "\n");
print( "Sorted  result:", sorted, "\n");
print( "Reversed  list:", reversed, "\n");


return false;

function test( name ) {

  switch( name ) {
    case "Ste" + "ve" {
	printf("I know you %s - expression-match!\n", name );
    }
    case "Steven" {
	printf("I know you %s - literal-match!\n", name );
    }
    case /^steve$/ {
        printf("I know you %s - regexp-match!\n", name );
    }
    default {
    }
  }
}


foreach name in [ "Steve", "Steven", "steve", "steven", "bob", "test" ] {
  test( name );
}
`

	// current line/position
	line := 0

	// lex
	l := New(input)
	tok := l.NextToken()
	for tok.Type != token.EOF {

		if tok.Line < line {
			t.Fatalf("Token %v went backwards!", tok)
		}
		line = tok.Line

		tok = l.NextToken()
	}

}

func TestWhile(t *testing.T) {
	input := `while ( true ) { a += 1;}`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.WHILE, "while"},
		{token.LPAREN, "("},
		{token.TRUE, "true"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "a"},
		{token.PLUSEQUALS, "+="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
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

func TestNextToken1(t *testing.T) {
	input := `-=*=..=+√%(){},;~= !~"`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.MINUSEQUALS, "-="},
		{token.ASTERISKEQUALS, "*="},
		{token.DOTDOT, ".."},
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
	input := `a = 1..4;
five=5;
ten =10;
result = add(five, ten);
!- *5;
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
{"foo" "bar"}
1.2
0.5
0.3
世界
'steve'
[]
?:
'
`
	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.DOTDOT, ".."},
		{token.INT, "4"},
		{token.SEMICOLON, ";"},
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
		{token.NOTEQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.FLOAT, "1.2"},
		{token.FLOAT, "0.5"},
		{token.FLOAT, "0.3"},
		{token.IDENT, "世界"},
		{token.STRING, "steve"},
		{token.LSQUARE, "["},
		{token.RSQUARE, "]"},
		{token.QUESTION, "?"},
		{token.COLON, ":"},
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
 c = 3.113/;
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
	input := `"line 1",
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

			if tk.Line != line {
				t.Fatalf("Token %v has wrong line number wanted %d got %d\n", tk, line, tk.Line)
			}
			line++
		}
	}

}

// TestRegexp ensures a simple regexp can be parsed.
func TestRegexp(t *testing.T) {
	input := `if ( f ~= /steve/i )
if ( f ~= /steve/m )
if ( f ~= /steve\//m )
if ( f ~= /steve/mi )
if ( f ~= /steve/miiiiiiiiiiiiiiiiimmmmmmmmmmmmmiiiii )
if ( f ~= /steve/fx )`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		// if ( f ~= /steve/i )
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.REGEXP, "(?i)steve"},
		{token.RPAREN, ")"},

		// if ( f ~= /steve/m )
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.REGEXP, "(?m)steve"},
		{token.RPAREN, ")"},

		// if ( f ~= /steve\//m )
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.REGEXP, "(?m)steve/"},
		{token.RPAREN, ")"},

		// if ( f ~= /steve/mi )
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.REGEXP, "(?mi)steve"},
		{token.RPAREN, ")"},

		//if ( f ~= /steve/miiiiiiiiiiiiiiiiimmmmmmmmmmmmmiiiii )`
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.REGEXP, "(?mi)steve"},
		{token.RPAREN, ")"},

		//if ( f ~= /steve/fx )`
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.ILLEGAL, "illegal regexp flag 'f' in string 'fx'"},
		{token.RPAREN, ")"},

		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%s] - tokentype wrong, expected=%q, got=%q", tt, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong, expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestIllegalRegexp is designed to look for an unterminated/illegal regexp
func TestIllegalRegexp(t *testing.T) {
	input := `if ( f ~= /steve )`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "f"},
		{token.CONTAINS, "~="},
		{token.ILLEGAL, "unterminated regular expression"},
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

// TestDiv is designed to test that a division is recognized; that it is
// not confused with a regular-expression.
func TestDiv(t *testing.T) {
	input := `a = b / c;
a = 3/4;
a /= 3;
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.IDENT, "b"},
		{token.SLASH, "/"},
		{token.IDENT, "c"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.ASSIGN, "="},
		{token.INT, "3"},
		{token.SLASH, "/"},
		{token.INT, "4"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "a"},
		{token.SLASHEQUALS, "/="},
		{token.INT, "3"},
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

// TestPostfix is designed to test postfix operations.
func TestPostfix(t *testing.T) {
	input := `a++;
b--;
`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.IDENT, "a"},
		{token.PLUSPLUS, "++"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "b"},
		{token.MINUSMINUS, "--"},
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

// TestIllegal is designed to catch illegal identifiers
func TestIllegal(t *testing.T) {
	input := `#`

	tests := []struct {
		expectedType    token.Type
		expectedLiteral string
	}{
		{token.ILLEGAL, "invalid character for indentifier '#'"},
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

// TestIdentifier is designed to show we can handle identifiers.
func TestIdentifiers(t *testing.T) {

	valid := []string{"foo_bar", "$foo", "bar_"}

	for _, input := range valid {
		l := New(input)

		tok := l.NextToken()
		if tok.Type != token.IDENT {
			t.Fatalf("token is not an identifier")
		}
		if tok.Literal != input {
			t.Fatalf("token didn't have a literal match")
		}
	}
}
