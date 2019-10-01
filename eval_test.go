package evalfilter

import (
	"testing"

	"github.com/skx/evalfilter/object"
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
		{Input: `if ( len(trim() ) == 0 ) { return true; } return false;`, Result: true},
		{Input: `if ( Count == 3 ) { return true; } return false;`, Result: false},
		{Input: `if ( Count != 1 ) { return true; }`, Result: true},
		{Input: `if ( Count != 12.4 ) { return false; } return true;`, Result: true},
		{Input: `if ( "Steve" == "Steve" ) { return true; }`, Result: true},
		{Input: `if ( "Steve" == "Kemp" ) { return true; } return false;`, Result: false},
	}

	for _, tst := range tests {

		obj := New(tst.Input)

		ret, err := obj.Run(object)
		if err != nil {
			t.Fatalf("Found unexpected error running test %s\n", err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script")
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
	}

	for _, tst := range tests {

		obj := New(tst.Input)

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

// TestObjectToBool tests our Object -> boolean results work as expected.
func TestObjectToBool(t *testing.T) {

	type Test struct {
		Object object.Object
		Result bool
	}
	var x object.Object

	tests := []Test{
		// Bool
		{Object: &object.Boolean{Value: true}, Result: true},
		{Object: &object.Boolean{Value: false}, Result: false},

		// Int
		{Object: &object.Integer{Value: 10}, Result: true},
		{Object: &object.Integer{Value: 0}, Result: false},

		// Float
		{Object: &object.Float{Value: 1.30}, Result: true},
		{Object: &object.Float{Value: 0.0}, Result: false},

		// String
		{Object: &object.String{Value: "Steve"}, Result: true},
		{Object: &object.String{Value: ""}, Result: false},

		// Misc
		{Object: &object.ReturnValue{Value: &object.Boolean{Value: true}}, Result: true},
		{Object: &object.ReturnValue{Value: &object.Boolean{Value: false}}, Result: false},
		{Object: &object.ReturnValue{Value: &object.String{Value: ""}}, Result: false},
		{Object: &object.ReturnValue{Value: &object.String{Value: "not-empty"}}, Result: true},
		{Object: &object.Null{}, Result: false},
		{Object: x, Result: true},
	}

	for _, tst := range tests {

		obj := New("")

		if obj.objectToNativeBoolean(tst.Object) != tst.Result {
			t.Fatalf("Found unexpected error running test '%s'\n", tst.Object.Inspect())
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

		ret, err := obj.Run(nil)
		if err != nil {
			t.Fatalf("Found unexpected error running test '%s' - %s\n", tst.Input, err.Error())
		}

		if ret != tst.Result {
			t.Fatalf("Found unexpected result running script: %s", tst.Input)
		}
	}
}
