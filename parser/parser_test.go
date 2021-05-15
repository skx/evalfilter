package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/skx/evalfilter/v2/ast"
	"github.com/skx/evalfilter/v2/lexer"
	"github.com/skx/evalfilter/v2/token"
)

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.errors
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(p.errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestArray(t *testing.T) {
	input := `[1, 2*2, 3+3]`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	array, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not ast.ArrayLiteral. got=%T", stmt.Expression)
	}
	if len(array.Elements) != 3 {
		t.Fatalf("len(array.Elements) not 3. got=%d", len(array.Elements))
	}
	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}

func TestAssign(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	for _, test := range []TestCase{
		{input: "a = 3;", error: false},
		{input: "3 = 3;", error: true},
	} {

		l := lexer.New(test.input)
		p := New(l)
		p.ParseProgram()

		if test.error {

			if len(p.errors) == 0 {
				t.Fatalf("expected to see an error, but didn't")
			}
		} else {

			if len(p.errors) > 0 {
				t.Fatalf("shouldn't have seen an error, but did: %s", p.errors[0])
			}
		}
	}
}

func TestFloatLiteralExpression(t *testing.T) {
	input := `5.2;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statemtnets[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	integer, ok := stmt.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("exp is not *ast.FloatLiteral. got=%T", stmt.Expression)
	}
	if integer.Value != 5.2 {
		t.Errorf("float.Value not %f. got=%f", 5.2, integer.Value)
	}
	if integer.TokenLiteral() != "5.2" {
		t.Errorf("float.TokenLiteral not %s. got=%s", "5.2", integer.TokenLiteral())
	}
}

func testFloatLiteral(t *testing.T, exp ast.Expression, v float64) bool {
	float, ok := exp.(*ast.FloatLiteral)
	if !ok {
		t.Errorf("exp not *ast.FloatLiteral. got=%T", exp)
		return false
	}
	if float.Value != v {
		t.Errorf("float.Value not %f. got=%f", v, float.Value)
		return false
	}
	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case float32:
		return testFloatLiteral(t, exp, float64(v))
	case float64:
		return testFloatLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar",
			ident.TokenLiteral())
	}
}

func TestParsingHashLiteral(t *testing.T) {
	input := `{"one":1, "two":2, "three":3}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got=%T", stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
		}
		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	input := "{}"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp isn not ast.HashLiteral. got=%T",
			stmt.Expression)
	}
	if len(hash.Pairs) != 0 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
}

func TestParsingHashLiteralWithExpression(t *testing.T) {
	input := `{"one":0+1, "two":10 - 8, "three": 15/5}`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt := program.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not ast.HashLiteral. got=%T",
			stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d",
			len(hash.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not ast.StringLiteral. got=%T", key)
			continue
		}
		testFunc, ok := tests[literal.Value]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
			continue
		}
		testFunc(value)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `5;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statemtnets[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	integer, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp is not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if integer.Value != 5 {
		t.Errorf("integer.Value not %d. got=%d", 5, integer.Value)
	}
	if integer.TokenLiteral() != "5" {
		t.Errorf("integer.TokenLiteral not %s. got=%s", "5", integer.TokenLiteral())
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}
	return true
}

func TestIncompleteThings(t *testing.T) {
	input := []string{
		`if `,
		`if ( true `,
		`if ( true ) `,
		`if ( true ) { `,
		`if ( true ) { puts( "OK" ) ; } else { `,
		`if ( true ) { puts( "OK" ) ; } else while { `,
		`x = `,
		`function foo bar `,
		`function foo( a, b ="steve", `,
		`function foo() {`,
		`for (`,
		`h { "foo"`,
		`h { "foo": "test"`,
		`while (`,
		`3 + `,
		`switch `,
		`switch ( `,
		`switch (foo `,
		`switch (foo) `,
		`switch (foo) { case 3 {`,
		`switch (foo) { `,
		`a = b[3`,
		`a = ( 1 + ( 2 * 3 ) `,
		`if ( += 3`,
		`while ( true ) { 1; `,
		`while ( true ) { `,
		`while ( true ) `,
		`while ( true `,
		`while (  `,
		`while  `,
	}

	for _, str := range input {

		// Here we do this the typical way.
		l := lexer.New(str)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.errors) < 1 {
			t.Errorf("unexpected error-count, got %d  expected %d", len(p.errors), 1)
		}

		// Now we repeat with the new parser-method.
		l2 := lexer.New(str)
		p2 := New(l2)
		_, err := p2.Parse()

		if err == nil {
			t.Errorf("We expected an error, but saw none")
		}
	}
}

func TestIndex(t *testing.T) {
	input := "myArray[1+1]"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	stmt, _ := program.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp not *ast.IndexExpression. got=%T", stmt.Expression)
	}
	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}
	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

func TestParseIllegal(t *testing.T) {

	l := lexer.New("")
	p := New(l)

	// int
	p.curToken = token.Token{Literal: "xx", Type: token.ILLEGAL}
	p.parseIllegal()

	if len(p.errors) != 1 {
		fmt.Printf("weird error-count")
	}
}

func TestParseLocal(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	for _, test := range []TestCase{{input: "function foo() { local foo; }", error: false},
		{input: "function foo(){ local 3; }", error: true},
		{input: "function foo(){ local local; }", error: true},
		{input: "function foo() { local; }", error: true},
		{input: "local a;", error: true}} {

		// Do this the normal way
		l := lexer.New(test.input)
		p := New(l)
		p.ParseProgram()

		if test.error {

			if len(p.errors) == 0 {
				t.Fatalf("expected to see an error, but didn't")
			}
		} else {

			if len(p.errors) > 0 {
				t.Fatalf("shouldn't have seen an error, but did: %s", p.errors[0])
			}
		}

		// Now with our new API
		l2 := lexer.New(test.input)
		p2 := New(l2)
		_, err := p2.Parse()

		if test.error {
			if err == nil {
				t.Fatalf("expected to see an error, but didn't")
			}
		} else {

			if err != nil {
				t.Fatalf("shouldn't have seen an error, but did: %s", err.Error())
			}
		}

	}
}

func TestParseMissingPrefix(t *testing.T) {
	incomplete := `?`
	l := lexer.New(incomplete)
	p := New(l)
	p.ParseProgram()

	if len(p.errors) < 1 {
		t.Fatalf("expected a parse error - got none")
	}

	txt := p.errors[0]
	if !strings.Contains(txt, "no prefix parse function for ?") {
		t.Fatalf("parse error was not the error we expected, got %s", txt)
	}
}

func TestParsePrefixPostfix(t *testing.T) {
	input := []string{"if ( √9 == 3 ) { return true ; }",
		"if ( -1 == ( 0 - 1 ) ) { return true ; }",
		"if ( !false == true ) { return true; }",
		"a = 1; a++;",
		"a = 2; a--;",
	}

	for _, test := range input {

		l := lexer.New(test)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.errors) != 0 {
			t.Errorf("unexpected error on %s: %s", test, strings.Join(p.errors, ","))
		}
	}

}
func TestParseRegexp(t *testing.T) {

	input := []string{"if ( Content ~= /needle/ ) { true ; }",
		"if ( Content ~= /needle/i ) { true ; }",
		"if ( Content !~ /needle/i ) { true ; }",
		"if ( Content !~ /needle/ ) { true ; }"}

	for _, test := range input {

		l := lexer.New(test)
		p := New(l)
		_ = p.ParseProgram()

		if len(p.errors) != 0 {
			t.Errorf("unexpected error on %s", test)
		}
	}
}

func TestParseTernary(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	for _, test := range []TestCase{{input: "min  = ( 3 > 2 ) ? 3 : 2;", error: false},
		{input: "( 3 > 2 ) ? 3", error: true},
		{input: "a = subject ? ( subject ? subject : Subject ) : title ", error: true},
		{input: "( 3 > 2 ) ? 3 :", error: true},
		{input: "( 3 > 2 ) ? 3 : )", error: true}} {
		l := lexer.New(test.input)
		p := New(l)
		p.ParseProgram()

		if test.error {

			if len(p.errors) == 0 {
				t.Fatalf("expected to see an error, but didn't: %s", test.input)
			}
		} else {

			if len(p.errors) > 0 {
				t.Fatalf("shouldn't have seen an error, but did: %s", p.errors[0])
			}
		}
	}
}

func TestParseWhile(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	for _, test := range []TestCase{{input: "while (1) { };", error: false},
		{input: "while ( 1 ) { ", error: true},
		{input: "while ( 1 ) ", error: true},
		{input: "while false", error: true}} {
		l := lexer.New(test.input)
		p := New(l)
		p.ParseProgram()

		if test.error {

			if len(p.errors) == 0 {
				t.Fatalf("expected to see an error, but didn't: %s", test.input)
			}
		} else {

			if len(p.errors) > 0 {
				t.Fatalf("shouldn't have seen an error, but did: %s", p.errors[0])
			}
		}
	}
}

func TestParseForeach(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	for _, test := range []TestCase{
		// OK
		{input: "foreach x in y { }", error: false},
		{input: "foreach x,y in z { }", error: false},

		// bogus
		{input: "foreach x,y in z ", error: true},
		{input: "foreach x,2 in z ", error: true},
		{input: "foreach x in { } ", error: true},
		{input: "foreach x  { } ", error: true},
		{input: "foreach 3  { } ", error: true},
	} {

		l := lexer.New(test.input)
		p := New(l)
		p.ParseProgram()

		if test.error {

			if len(p.errors) == 0 {
				t.Fatalf("expected to see an error, but didn't")
			}
		} else {

			if len(p.errors) > 0 {
				t.Fatalf("shouldn't have seen an error, but did: %s", p.errors[0])
			}
		}
	}
}

func TestReturnStatement(t *testing.T) {
	input := `
return 993322;
return "steve";
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 2 {
		t.Errorf("program does not contain 2 statements, got=%d", len(program.Statements))
	}
	for _, stmt := range program.Statements {
		returnStatement, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement. got %T", stmt)
		}
		if returnStatement.TokenLiteral() != "return" {
			t.Errorf("returnStatement.TokenLiteral not 'return', got %q", returnStatement.TokenLiteral())
		}
	}

	incomplete := `return true`
	l = lexer.New(incomplete)
	p = New(l)
	p.ParseProgram()

	if len(p.errors) < 1 {
		t.Fatalf("expected a parse error - got none: %s", strings.Join(p.errors, ","))
	}

	txt := p.Errors()[0]
	if !strings.Contains(txt, "expected semicolon after return-value") {
		t.Fatalf("parse error was not the error we expected")
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q, got=%q", "hello world", literal.Value)
	}
}

// switch can only have a single `default` block
func TestCaseDefaults(t *testing.T) {

	type TestCase struct {
		input string
		error bool
	}

	tests := []TestCase{

		// OK
		{input: `a = 1
switch( a ) {
  case 3, 7 {
    a = 7
  }
  default {
    a = 3
  }
}
a
`, error: false},

		// Two defaults: error
		{input: `a = 1
switch( a ) {
  default {
    a = 3
  }
  case default {
    a = 4
  }
}
a
`, error: true},
		// Unexpected token: error
		{input: `a = 1
switch( a ) {
  while a < 3 {
  }
}

`, error: true},

		// Unexpected token: error
		{input: `a = 1
switch( a ) {
  case "foo" if
`, error: true},

		// Incomplete: error
		{input: `a = 1
switch( a ) {
  case "foo" {
     printf("OK\n");
  }
if
`, error: true},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)
		_, err := p.Parse()

		if test.error {
			if err == nil {
				t.Fatalf("expected error, got none")
			}
		} else {
			if err != nil {
				t.Fatalf("didn't expect an error, but saw one: %s\n", err.Error())
			}
		}
	}
}

func TestParseNumberLiteral(t *testing.T) {

	l := lexer.New("")
	p := New(l)

	// int
	p.curToken = token.Token{Literal: "0b33"}
	err := p.parseIntegerLiteral()
	if err != nil {
		t.Fatalf("expected error parsing '0b33', got none")
	}

	fmt.Printf("ERROR: %s\n", err)
	out := strings.Join(p.errors, ",")
	if !strings.Contains(out, "could not parse") {
		t.Fatalf("error was different than expected: %s", out)
	}

	// float
	p.curToken = token.Token{Literal: "0.3.2.1"}
	err = p.parseFloatLiteral()
	if err != nil {
		t.Fatalf("expected error parsing '0.3.2.1', got none")
	}

	out = strings.Join(p.errors, ",")
	if !strings.Contains(out, "could not parse") {
		t.Fatalf("error was different than expected: %s", out)
	}
}

// This function tests some cases the fuzz-testing evolved.
func TestFuzzerResults(t *testing.T) {

	inputs := []string{
		`0.1.`,
		`0=0.function((){`,
		`0=0.!0.function((){`,
		`{0.function((){`,
		`0=0.!function((){`,
		`0.function((){`,
		`0.!function((){`,
		`0=!0.function((){`,
		`0.!!function((){`,
		`{!0.function((){`,
		`!0.function((){`,
		`0.!0.function((){`,
		`0..0.function((){`,
		`0.!!!function((){`,
		`0.foreach 00in 0()=0`,
	}

	for _, input := range inputs {

		l := lexer.New(input)
		p := New(l)
		err := p.ParseProgram()

		if err == nil {
			t.Fatalf("expected error, got none")
		}
	}
}
