package evalfilter

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

// TestLess tests uses `>` and `>=`.
func TestLess(t *testing.T) {

	// Dummy structure to test field-access.
	type Structure struct {
		Count int
	}

	// Instance of object
	var object Structure
	object.Count = 3

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( !true == false ) { return true; }`, Result: true},
		{Input: `if ( "steve" ~= /steve/i ) { return true; }`, Result: true},
		{Input: `if ( "steve" !~ /steve/i ) { return true; } return false;`, Result: false},
		{Input: `if ( !false == true ) { return true; }`, Result: true},
		{Input: `if ( -3 == -3 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 + 1.0 == 4 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 - 1.0 == 2 ) { return true; }`, Result: true},
		{Input: `if ( -3 + 2 - 1 - 1 == -3 ) { return true; }`, Result: true},
		{Input: `if ( 3 * 3 == 9 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 * 3.0 == 9.0 ) { return true; }`, Result: true},
		{Input: `if ( 12 / 4 == 3 ) { return true; }`, Result: true},
		{Input: `if ( 12.0 / 4.0 == 3.0 ) { return true; }`, Result: true},
		{Input: `if ( true && true ) { return true; }`, Result: true},
		{Input: `if ( false || true ) { return true; }`, Result: true},
		{Input: `if ( -3 == -3.0 ) { return true; }`, Result: true},
		{Input: `if ( -3.3 == -3.3 ) { return true; }`, Result: true},
		{Input: `if ( 1 < 4 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 < 4.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 < 4 ) { return true; } else { print( "help!"); return false;}`, Result: true},
		{Input: `if ( 1.4 < 2 ) { return false; }`, Result: false},
		{Input: `if ( 1 < 2.0 ) { return false; }`, Result: false},
		{Input: `if ( 3 <= 3 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 <= 3.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 <= 3 ) { return false; }`, Result: false},
		{Input: `if ( 65538 > 1 ) { return false; }`, Result: false},
		{Input: `if ( Count <= 3 ) { print( "", ""); return false; }`, Result: false},
		{Input: `if ( len("steve") <= 3 ) { return false; } else { return true; }`, Result: true},
		{Input: `if ( len("π") == 1) { return true; } else { return false; }`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare([]byte{NoOptimize})
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestMore tests uses `>` and `>=`.
func TestMore(t *testing.T) {

	// Dummy structure to test field-access.
	type Object struct {
		Count int
	}

	// Instance of object
	object := &Object{Count: 32}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( 1 > 4 ) { return true; } return false;`, Result: false},
		{Input: `if ( 1.1 > 1 ) { return true; }`, Result: true},
		{Input: `if ( Count > 1 ) { return true; }`, Result: true},
		{Input: `if ( 3 >= 3 ) { return true; }`, Result: true},
		{Input: `if ( 1 >= 3 ) { return false; } return true;`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestBroken confirms "Run" and "Execute" return errors.
func TestBroken(t *testing.T) {
	input := `

function top( ) {
   return 3;
}

function bottom( ) {
   return 0;
}

// Division by zero
out = top() / bottom();

return( true );
`

	obj := New(input)
	err := obj.Prepare()
	if err != nil {
		t.Fatalf("Failed to compile")
	}

	a, err := obj.Run(nil)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "division") {
		fmt.Printf("Wrong value")
	}
	if a != false {
		fmt.Printf("Wrong value")
	}

	b, err := obj.Execute(nil)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
	if !strings.Contains(err.Error(), "division") {
		fmt.Printf("Wrong value")
	}
	if b.Inspect() != "null" {
		fmt.Printf("Wrong value")
	}
}

// TestDump just dumps some bytecode
func TestDump(t *testing.T) {
	input := `

function hello( name ) {
   printf("Hello, %s\n", name );
}

function goodbye( name ) {
   printf("Goodbye, %s\n", name );
}

hello( "world" );
goodbye( "world" );

a = 3;
a += 3;

return( true );
`

	obj := New(input)

	ctx := context.Background()
	obj.SetContext(ctx)
	p := obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Dump
	obj.Dump()

	_, err := obj.Run(nil)
	if err != nil {
		t.Fatalf("unexpected error running code")
	}

	// Confirm a == 6
	val := obj.GetVariable("a")
	if val.Inspect() != "6" {
		t.Fatalf("Variable had wrong value: Got %v\n", val)
	}

	// Confirm bogus variables are null
	val = obj.GetVariable("aa")
	if val.Inspect() != "null" {
		t.Fatalf("Variable had wrong value: Got %v\n", val)
	}
}

// TestEq tests uses `==` and `!=`.
func TestEq(t *testing.T) {

	// Dummy structure to test field-access.
	type Object struct {
		Count float64
	}

	// Instance of object
	object := &Object{Count: 12.4}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( Count == 12.4 ) { return true; } return false;`, Result: true},
		{Input: `if ( len(trim(" steve " ) ) == 5 ) { return true; } return false;`, Result: true},
		{Input: `if ( len(trim("   ") ) == 0 ) { return true; } return false;`, Result: true},
		{Input: `if ( Count == 3 ) { return true; } return false;`, Result: false},
		{Input: `if ( Count != 1 ) { return true; }`, Result: true},
		{Input: `if ( Count != 12.4 ) { return false; } return true;`, Result: true},
		{Input: `if ( "Steve" == "Steve" ) { return true; }`, Result: true},
		{Input: `if ( "Steve" == "Kemp" ) { return true; } return false;`, Result: false},
		{Input: `a = "Steve"; a += " Kemp"; return ( a == "Steve Kemp" );`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script: %s", tst.Input)
		}
	}
}

// TestContains tests uses `~=` and `!~`.
func TestContains(t *testing.T) {

	// Dummy structure to test field-access.
	type Object struct {
		Greeting string
	}

	// Instance of object
	object := &Object{Greeting: "Hello World"}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( Greeting ~= /World/ ) { return true; } return false;`, Result: true},
		{Input: `if ( Greeting ~= /Moi/ ) { return true; } return false;`, Result: false},
		{Input: `if ( Greeting !~ /Cake/ ) { return true; } return false;`, Result: true},
		{Input: `if ( Greeting ~= /Cake/ ) { return true; } return false;`, Result: false},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestFunctionBool calls a function
func TestFunctionBool(t *testing.T) {

	// Dummy structure to test field-access.
	type Structure struct {
		Count int
	}

	// Instance of object
	o := &Structure{Count: 3}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( True() ) {  return true; } return false;`, Result: true},
		{Input: `if ( True() == false ) { return true; } return false;`, Result: false},
		{Input: `if ( True() != false ) { return true; } return false;`, Result: true},
		{Input: `if ( ! True() ) { return true; } else { return false; }`, Result: false},
		{Input: `function foo() { local x; x = true; return x; } ; if ( foo() ) { return true ; } return false;`, Result: true},
		{Input: `function foo() { local a ; a = 0; a += 3; } ; foo() ; if ( a ) { return true ; } return false;`, Result: false},
		{Input: `function foo() { local a ; a = 3; a -= 3; } ; foo() ; if ( a ) { return true ; } return false;`, Result: false},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		obj.AddFunction("True",
			func(args []object.Object) object.Object {
				return &object.Boolean{Value: true}
			})

		ret, err := obj.Run(o)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestFunctionInt calls a function
func TestFunctionInt(t *testing.T) {

	// Dummy structure to test field-access.
	type Structure struct {
		Count int
	}

	// Instance of object
	o := &Structure{Count: 3}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( Number() ) { return true; } return false;`, Result: true},
		{Input: `if ( Number() != 0 ) { return true; } return false;`, Result: true},
		{Input: `if ( Number() == 3 ) { return true; } return false;`, Result: true},
		{Input: `if ( ! Number() ) { return true; } else { return false; }`, Result: false},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		obj.AddFunction("Number",
			func(args []object.Object) object.Object {
				return &object.Integer{Value: 3}
			})

		ret, err := obj.Run(o)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestFunctionString calls a function
func TestFunctionString(t *testing.T) {

	// Dummy structure to test field-access.
	type Structure struct {
		Count int
	}

	// Instance of object
	o := &Structure{Count: 3}

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( Str() ) { return true; } return false;`, Result: true},
		{Input: `if ( Str() == "Steve" ) { return true; } return false;`, Result: true},
		{Input: `if ( Str() == "Bob" ) { return true; } return false;`, Result: false},
		{Input: `if ( ! Str() ) { return true; } else { return false; }`, Result: false},

		{Input: `if ( EmptyStr() ) { return true; } return false;`, Result: false},
		{Input: `if ( EmptyStr() == "Steve" ) { return true; } return false;`, Result: false},
		{Input: `if ( EmptyStr() == "" ) { return true; } return false;`, Result: true},
		{Input: `if ( ! EmptyStr() ) { return true; } else { return false; }`, Result: false},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		obj.AddFunction("Str",
			func(args []object.Object) object.Object {
				return &object.String{Value: "Steve"}
			})
		obj.AddFunction("EmptyStr",
			func(args []object.Object) object.Object {
				return &object.String{Value: ""}
			})

		ret, err := obj.Run(o)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script %s", tst.Input)
		}
	}
}

// TestBool tests a struct-member can be boolean
func TestBool(t *testing.T) {

	// Dummy structure to test field-access.
	type Object struct {
		Valid bool
	}

	// Instances of object
	var object Object
	object.Valid = true

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `if ( Valid ) { return true; } return false;`, Result: true},
		{Input: `if ( Valid == true ) { return true; } return false;`, Result: true},
		{Input: `if ( !Valid == false ) { return true; } return false;`, Result: true},
		{Input: `if ( !!Valid ) { return true; } return false;`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script; got %v expected %v", ret, tst.Result)
		}
	}
}

// TestVariable sets a variable.
func TestVariable(t *testing.T) {

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `name = "Steve"; if ( len( $name ) == 5 ) { return true; } return false;`, Result: true},
		{Input: `name = "Steve"; if ( len( name ) == 5 ) { return true; } return false;`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile:%s", p.Error())
		}

		obj.SetVariable("name", &object.String{Value: "Steve"})

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestOperations makes tests against operations of various types.
//
// This is because we differentiate between "int OP int", "float OP int",
// and "int OP float".
func TestOperations(t *testing.T) {

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{

		// int OP int
		{Input: `if ( 1 + 2 * 3 == 7 ) { return true; }`, Result: true},
		{Input: `a = 9; a /= 3 ; if ( a == 3 )  { return true; }`, Result: true},
		{Input: `a = 2; a *= 3 ; if ( a == 6 )  { return true; }`, Result: true},
		{Input: `if ( 1 % 3 == 1 ) { return true; }`, Result: true},
		{Input: `if ( 2 % 3 == 2 ) { return true; }`, Result: true},
		{Input: `if ( 3 % 3 == 0 ) { return true; }`, Result: true},
		{Input: `if ( 2 ** 3  == 8 ) { return true; }`, Result: true},
		{Input: `if ( √9 == 3 ) { return true; }`, Result: true},
		{Input: `if ( √9.0 == 3 ) { return true; }`, Result: true},

		// int OP float
		{Input: `if ( 1 + 2.0 == 3.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 - 1.0 == 0.0 ) { return true; }`, Result: true},
		{Input: `if ( 3 / 3.0 == 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 3 % 3.0 == 0 ) { return true; }`, Result: true},
		{Input: `if ( 1 ** 3.0 == 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 2 * 3.0 == 6.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 == 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 != 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 < 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 1 <= 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 10 > 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 10 >= 2.0 ) { return true; }`, Result: true},

		// float OP float
		{Input: `if ( 3.0 % 3.0 == 0 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 * 3 == 9.0 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 ** 3.0 == 1.0 ) { return true; }`, Result: true},
		// float - int
		{Input: `if ( 1.0 == 1 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 != 20 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 < 20 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 <= 1 ) { return true; }`, Result: true},
		{Input: `if ( 10.0 > 2 ) { return true; }`, Result: true},
		{Input: `if ( 10.0 >= 3 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 < 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 <= 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 10.0 > 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 10.0 >= 2.0 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 + 2 == 3.0 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 - 1 == 0.0 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 / 3 == 1.0 ) { return true; }`, Result: true},
		{Input: `if ( 3.0 % 3 == 0 ) { return true; }`, Result: true},
		{Input: `if ( 1.0 ** 3 == 1.0 ) { return true; }`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script: %s", tst.Input)
		}
	}
}

// TestRevFunction is a trivial test of a simple user-defined functions
func TestRevFunction(t *testing.T) {

	// Test structure
	type Tests struct {
		Name   string
		Result bool
	}

	// Some tests of pallendromes
	var tests = []Tests{
		{Name: "Steve", Result: false},
		{Name: "StetS", Result: true},
		{Name: "EE", Result: true},
		{Name: "EVE", Result: true},
		{Name: "LOL", Result: true},
	}

	for _, test := range tests {

		type Input struct {
			Name string
		}

		// Test-script
		src := `function rev(str) {
   local tmp;
   tmp = "";

   foreach char in str {
      tmp = char + tmp;
   }

   return tmp;
}

return ( rev(Name) == Name );
`
		obj := New(src)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		var i Input
		i.Name = test.Name

		// Run
		ret, err := obj.Run(i)
		if err != nil {
			t.Fatalf("Found unexpected error running test - %s\n", err.Error())
		}

		if test.Result != ret {
			t.Fatalf("Found unexpected result running reverse against %s.", test.Name)
		}
	}

}

// TestAndOr ensure that `and` + `or` work in the real world
// https://github.com/skx/evalfilter/issues/31
func TestAndOr(t *testing.T) {

	// Test structure
	type Params struct {
		Origin  string
		Country string
		Value   int
		Adults  int
	}

	// Instance of structure with known values
	var d Params
	d.Origin = "Mow"
	d.Country = "RU"
	d.Adults = 1
	d.Value = 100

	// Test-script
	src := `if ( (Origin == "MOW" || Country == "RU") && (Value > 100 || Adults == 1) ) { return true; } else { return false; }`
	obj := New(src)

	p := obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err := obj.Run(d)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

}

// TestArrayObject tests that using reflection to get array values works
// for basic types.
func TestArrayObject(t *testing.T) {

	// string-test
	type Parent struct {
		Name     string
		Children []string
	}

	// float-test
	type Prices struct {
		Year   int
		Prices []float32
	}

	// int-test
	type Years struct {
		Years []int
	}

	// bool-test
	type Alive struct {
		Dead []bool
	}

	// String-object
	var dad Parent
	dad.Name = "Homer"
	dad.Children = []string{"Bart", "Lisa", "Maggie"}

	// Float-object
	var chocolate Prices
	chocolate.Year = 2019
	chocolate.Prices = []float32{3.0, 32, 12.2, 332.2}

	// Int-object
	var years Years
	years.Years = []int{1988, 1999, 2020}

	// Bool-object
	var granny Alive
	granny.Dead = []bool{false, false, true, true}

	//
	// String
	//
	src := `if ( Children[0] == "Bart" ) { return true; } return false;`
	obj := New(src)

	p := obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err := obj.Run(dad)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Float32
	//
	src = `if ( Prices[0] == 3.0 ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(chocolate)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Int
	//
	src = `if ( Years[0] == 1988 && Years[1] == 1999 && Years[2] == 2020 ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(years)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Bool
	//
	src = `if ( Dead[0] == false && Dead[2] == true ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(granny)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}
}

// TestArrayMap tests that using reflection to get array values works
// for basic types.
func TestArrayMap(t *testing.T) {

	// string-test
	dad := make(map[string]interface{})
	dad["Name"] = "Homer"
	dad["Children"] = []string{"Bart", "Lisa", "Maggie"}

	// float-test
	chocolate := make(map[string]interface{})
	chocolate["Year"] = 2019
	chocolate["Prices"] = []float32{3.0, 32, 12.2, 332.2}

	// int-test
	years := make(map[string]interface{})
	years["Years"] = []int{1988, 1999, 2020}

	// bool-test
	granny := make(map[string]interface{})
	granny["Dead"] = []bool{false, false, true, true}

	//
	// String
	//
	src := `if ( Children[0] == "Bart" ) { return true; } return false;`
	obj := New(src)

	p := obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err := obj.Run(dad)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Float32
	//
	src = `if ( Prices[0] == 3.0 ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(chocolate)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Int
	//
	src = `if ( Years[0] == 1988 && Years[1] == 1999 && Years[2] == 2020 ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(years)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

	//
	// Bool
	//
	src = `if ( Dead[0] == false && Dead[2] == true ) { return true; } return false;`
	obj = New(src)

	p = obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err = obj.Run(granny)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}
}

// TestOptimizer is a simple test-case to confirm an issue is resolved
// https://github.com/skx/evalfilter/issues/82
func TestOptimizer(t *testing.T) {

	//
	// String
	//
	src := `
value = 0;

if ( 1 == 0 ) {
   print( "Weird output\n" );
   value = value + 1;
}

if ( 0 == 0 ) {
   print( "Expected output\n");
   value = value + 1;
}

if ( 1 != 1 ) {
   print( "Weird output\n" );
   value = value + 1;
}

print( "After" );

// This should match
if ( value == 1 ) { return true; }

return false;
`
	obj := New(src)

	p := obj.Prepare()
	if p != nil {
		t.Fatalf("Failed to compile")
	}

	// Run
	ret, err := obj.Run(nil)
	if err != nil {
		t.Fatalf("Found unexpected error running test - %s\n", err.Error())
	}

	if !ret {
		t.Fatalf("Found unexpected result running script.")
	}

}

// TestArrayIn checks our array-inclusion functionality is sane.
func TestArrayIn(t *testing.T) {

	type Test struct {
		Input  string
		Error  bool
		Result bool
	}

	tests := []Test{

		{Input: `
names = [ "Steve", "Kemp" ]
if ( "Steve" in names ) {
  return true;
} else {
  return false;
}
`, Result: true},

		{Input: `
return( "Steve" in [ "Steve", "Blah", "Kemp" ] );
`,
			Result: true},
		{Input: `return( "Steve" in "Steve" );`, Result: true},
		{Input: `return( "Steve" in "Steven" );`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(nil)
		if err != nil {
			if tst.Error == false {
				t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
			}
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script: %s", tst.Input)
		}
	}
}

// TestForeach tests our foreach loop, at least a little
func TestForeach(t *testing.T) {

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{

		{Input: `foreach item in ["Steve" ] { return true; }`,
			Result: true},
		{Input: `foreach item in ["Steve", "Kemp"] {
return len(item) == 5 ; };
return false;`,
			Result: true},
		{Input: `sum = 0 ; foreach item in [1, 2, 3, 4] {
sum = sum + item; }
return( sum == 10 );
`,
			Result: true},
		{Input: `sum = 0 ; foreach item in 1..4 {
sum = sum + item; }
return( sum == 10 );
`,
			Result: true},
		{Input: `
str = "狐犬"; len = 0;
foreach char in str { len++; }
return len == 2 ;
`,
			Result: true},
		{Input: `
str = "狐犬"; len = 2;
foreach char in str { len--; }
return len == 0 ;
`,
			Result: true},
		{Input: `
str = "狐犬"; len = 0;
foreach char in str { len++ }
foreach char in str { len++ }
return len == 4 ;
`,
			Result: true},
		{Input: `return( len( 1..10 ) == 10);`,
			Result: true},
		{Input: `return( len( 9..10 ) == 2);`,
			Result: true},
		{Input: `if( len( 2..10 )  == 2) { return false; } return true;`,
			Result: true},

		{Input: `
// ensure that we don't leak scoped-variables
name = "Steve";

foreach index,name in [ "Bob", "Chris" ] {
   index++;
   printf("%d:%s\n", index, name);
}
if ( name != "Steve") {    print( "Test failed: name is changed\n");  return false; }
if ( index ) {    print( "Test FAILED: index is set: ", index, "\n"); return false; }
return true;
`,
			Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile: %s - %s", tst.Input, p.Error())
		}

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running script: %s : %s", tst.Input, err.Error())
		}
		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script: %s", tst.Input)
		}
	}
}

// TestTernary checks our simple ternary expression(s)
func TestTernary(t *testing.T) {

	type Test struct {
		Input  string
		Error  bool
		Result bool
	}

	tests := []Test{

		{Input: `return( true ? true : false );`,
			Result: true},
		{Input: `return( false ? true : false );`,
			Result: false},
		{Input: `a = "Steve"; return( a == "Steve" ? true : false );`,
			Result: true},
		{Input: `return( true ? 3 + 3 == 6 : false);`, Result: true},
		{Input: `return( 3 + 3 == 7 ? true : false);`, Result: false},
		{Input: `return( ( 3 + 3 == 7 ) ? ( true ) : ( false ));`,
			Result: false},
		{Input: `
a = 1;
return( a == 1 ? true ? true : false : false );
`, Error: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			if tst.Error == false {
				t.Fatalf("Failed to compile: %s - %s", tst.Input, p.Error())
			} else {
				fmt.Printf("Received expected error: %s\n", p.Error())
			}

		}

		if tst.Error == false {
			ret, err := obj.Run(nil)
			if err != nil {
				t.Fatalf("Found unexpected error running script: %s : %s", tst.Input, err.Error())
			}
			if ret != tst.Result {
				t.Fatalf("Found unexpected result running script: %s", tst.Input)
			}
		}
	}
}

// TestStringIndex tests that we handle UTF characters.
func TestStringIndex(t *testing.T) {

	type Test struct {
		Input  string
		Result bool
	}

	tests := []Test{
		{Input: `return( "√√1"[0] == "√" );`, Result: true},
		{Input: `return( "√√1"[1] == "√" );`, Result: true},
		{Input: `return( "√√1"[2] == "1" );`, Result: true},
		{Input: `return( "Hachikō"[6] == "ō" );`, Result: true},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile")
		}

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
		}
	}
}

// TestNumberTruth ensures that we're only expecting "true" results
// from positive values.  Negative/Zero values are not true
//
//  i.e.
//    if ( 1 ) { .. executes .. }
//    if ( -1 ) { .. never executes ... }
func TestNumberTruth(t *testing.T) {

	// Each test-case is expected to return `true` to pass.
	inputs := []string{
		`if ( 1 ) { return true; } return false;`,
		`if ( 1.0 ) { return true; } return false;`,
		`if ( -1 ) { return false ; } return true;`,
		`if ( -1.1 ) { return false ; } return true;`,
		`return 0 ? false : true ;`,
		`return 1 ? true : false ;`,
		`return -1 ? false : true ;`,
		`return -1.0 ? false : true ;`,
	}

	for _, tst := range inputs {

		obj := New(tst)

		p := obj.Prepare()
		if p != nil {
			t.Fatalf("Failed to compile: %s", tst)
		}

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst, err.Error())
		}

		if ret != true {
			t.Fatalf("Found unexpected result running script: %s - %v", tst, ret)
		}
	}
}

func TestMutators(t *testing.T) {

	inputs := []string{
		`1 *= 3;`,
		`"steve" += "kemp"`,
		`true -= false`,
		`3.4 /= 7`,

		// Using this same broken approach to testing
		// compilation results of various statements
		// assign:
		`a = 3 += 3;`,
		`if ( 3 += 7 ) { return true; } `,
		`if ( true ) { return 4+=3; } `,
		`if ( true ) { return 4 ; } else { return 4+=3; } `,
		`function foo() { 3 += 4; return true; } `,
		`foreach x in  [ 3, 4, 4+=4 ] { return 127 ; }`,

		// BUG:TODO: This should fail?
		// `foreach x in 1..3 { return true *= 3; }`,

		`return 4 += 3;`,
		`(4 += 34) + ( 4 /= 4)`,
		`a = [ 3, 4, 4+=4 ];`,
		`(4 + 34) * ( 4 /= 4)`,
		`√(4+=3);`,
		`while( 4 += 3 ) { return true; } `,
		`while( true ) { 4 += 3; }`,
		`3+= 3 ? true : false;`,
		`true ? 3+= 3 : false;`,
		`true ? false : 3+= 3;`,
		`"steve"[3+= 3];`,
		`print( 1, 2, 3, 4+= 3 );`,

		// TODO: index expression.
		//       `3+=1[3];`,
		// TODO: foreach (ast.Body),
		//       `foreach x,y in [ 3, 4, 5 ] { return 3+=2; }`,
	}

	for _, tst := range inputs {

		obj := New(tst)

		err := obj.Prepare()
		if err == nil {
			t.Fatalf("Expected error compiling test, got none: %s", tst)
		}
		if !strings.Contains(err.Error(), "must be an identifier") {
			t.Fatalf("Got an error on bogus types, but the wrong one: %s\n", err.Error())
		}
	}
}
