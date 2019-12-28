package evalfilter

import (
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
		{Input: `if ( Count <= 3 ) { print( "", ""); return false; }`, Result: false},
		{Input: `if ( len("steve") <= 3 ) { return false; } else { return true; }`, Result: true},
		{Input: `if ( len("π") == 1) { return true; } else { return false; }`, Result: true},
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
		{Input: `if ( Greeting ~= "World" ) { return true; } return false;`, Result: true},
		{Input: `if ( Greeting ~= "Moi" ) { return true; } return false;`, Result: false},
		{Input: `if ( Greeting !~ "Cake" ) { return true; } return false;`, Result: true},
		{Input: `if ( Greeting ~= "Cake" ) { return true; } return false;`, Result: false},
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
		{Input: `if ( True() ) { return true; } return false;`, Result: true},
		{Input: `if ( True() == false ) { return true; } return false;`, Result: false},
		{Input: `if ( True() != false ) { return true; } return false;`, Result: true},
		{Input: `if ( ! True() ) { return true; } else { return false; }`, Result: false},
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

		// int OP int
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
		{Input: `return( "Steve" in "Steve" );`, Error: true},
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
