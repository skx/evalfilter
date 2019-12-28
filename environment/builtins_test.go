package environment

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

// Test float-conversion.
func TestFloat(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result object.Object
	}

	tests := []TestCase{
		{Input: &object.String{Value: "π"}, Result: &object.Null{}},
		{Input: &object.String{Value: "Steve"}, Result: &object.Null{}},
		{Input: &object.Integer{Value: 3}, Result: &object.Float{Value: 3}},
		{Input: &object.String{Value: "3.21"}, Result: &object.Float{Value: 3.21}},
		{Input: &object.Boolean{Value: true}, Result: &object.Null{}},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnFloat(args)

		if x.Type() != test.Result.Type() {
			t.Errorf("Invalid type result for '%s', got %s, expected %s", test.Input, x.Type(), test.Result.Type())
		}

		switch x.(type) {
		case *object.Float:
			if x.(*object.Float).Value != test.Result.(*object.Float).Value {
				t.Errorf("invalid float result")
			}
		case *object.Null:
		default:
			t.Errorf("unknown type")
		}
	}

	// ensure that zero arguments are handled
	var tmp []object.Object
	out := fnFloat(tmp)
	if out.Type() != "NULL" {
		t.Errorf("Invalid result for no args:%s", out.Type())
	}
}

// Test integer-conversion.
func TestInt(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result object.Object
	}

	tests := []TestCase{
		{Input: &object.String{Value: "π"}, Result: &object.Null{}},
		{Input: &object.String{Value: "Steve"}, Result: &object.Null{}},
		{Input: &object.Integer{Value: 3}, Result: &object.Integer{Value: 3}},
		{Input: &object.String{Value: "3"}, Result: &object.Integer{Value: 3}},
		{Input: &object.Boolean{Value: true}, Result: &object.Null{}},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnInt(args)

		if x.Type() != test.Result.Type() {
			t.Errorf("Invalid type result for return")
		}

		switch x.(type) {
		case *object.Integer:
			if x.(*object.Integer).Value != test.Result.(*object.Integer).Value {
				t.Errorf("Invalid integer result")
			}
		case *object.Null:
		default:
			t.Errorf("unknown type")
		}
	}

	// ensure that zero arguments are handled
	var tmp []object.Object
	out := fnInt(tmp)
	if out.Type() != "NULL" {
		t.Errorf("Invalid result for no args:%s", out.Type())
	}
}

// Test string-conversion.
func TestString(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result object.Object
	}

	tests := []TestCase{
		{Input: &object.String{Value: "π"}, Result: &object.String{Value: "π"}},
		{Input: &object.String{Value: "Steve"}, Result: &object.String{Value: "Steve"}},
		{Input: &object.Integer{Value: 3}, Result: &object.String{Value: "3"}},
		{Input: &object.Boolean{Value: true}, Result: &object.String{Value: "true"}},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnString(args)

		if x.Type() != test.Result.Type() {
			t.Errorf("Invalid type result for return")
		}

		switch x.(type) {
		case *object.String:
			if x.(*object.String).Value != test.Result.(*object.String).Value {
				t.Errorf("Invalid string result")
			}
		case *object.Null:
		default:
			t.Errorf("unknown type")
		}
	}

	// ensure that zero arguments are handled
	var tmp []object.Object
	out := fnString(tmp)
	if out.Type() != "NULL" {
		t.Errorf("Invalid result for no args:%s", out.Type())
	}
}

// Test string length
func TestLen(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result int
	}

	tests := []TestCase{
		{Input: &object.String{Value: "π"}, Result: 1},
		{Input: &object.String{Value: "Steve"}, Result: 5},
		{Input: &object.Integer{Value: 1}, Result: 1},
		{Input: &object.Float{Value: 3.2}, Result: 3},
		{Input: &object.Boolean{Value: true}, Result: 4},
		{Input: &object.Boolean{Value: false}, Result: 5},

		// Arrays
		{Input: &object.Array{Elements: []object.Object{
			&object.String{Value: "steve"},
			&object.Integer{Value: 1}}},
			Result: 2},
		{Input: &object.Array{Elements: []object.Object{
			&object.String{Value: "steve"}}},
			Result: 1},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnLen(args)
		if int(x.(*object.Integer).Value) != test.Result {
			t.Errorf("Invalid length for %s", test.Input)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnLen(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
}

// Test lower-casing strings
func TestLower(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result string
	}

	tests := []TestCase{
		{Input: &object.String{Value: "STEVE"}, Result: "steve"},
		{Input: &object.String{Value: "Π"}, Result: "π"},
		{Input: &object.Integer{Value: 1}, Result: "1"},
		{Input: &object.Float{Value: 3.2}, Result: "3.2"},
		{Input: &object.Boolean{Value: true}, Result: "true"},
		{Input: &object.Null{}, Result: "null"},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnLower(args)
		if x.(*object.String).Value != test.Result {
			t.Errorf("Invalid result for %t", test.Input)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnLower(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
}

// Test regexp-matching
func TestMatch(t *testing.T) {

	type TestCase struct {
		String string
		Regexp string
		Result bool
	}

	tests := []TestCase{
		// All tests are doubled to test the cache-handle
		{String: "Steve", Regexp: "^Steve$", Result: true},
		{String: "Steve", Regexp: "^Steve$", Result: true},

		{String: "Steve", Regexp: "(?i)^steve$", Result: true},
		{String: "Steve", Regexp: "(?i)^steve$", Result: true},

		{String: "Steve", Regexp: "^steve$", Result: false},
		{String: "Steve", Regexp: "^steve$", Result: false},

		// invalid regexp
		{String: "Steve", Regexp: "+", Result: false},
		{String: "Steve", Regexp: "+", Result: false},
	}

	for _, test := range tests {

		var args []object.Object

		args = append(args, &object.String{Value: test.String})
		args = append(args, &object.String{Value: test.Regexp})

		res := fnMatch(args)

		if res.(*object.Boolean).Value != test.Result {
			t.Errorf("Invalid result for %s =~ /%s/", test.String, test.Regexp)
		}

	}

	// Calling the function with != 2 arguments should return false
	var args []object.Object
	out := fnMatch(args)
	if out.(*object.Boolean).Value != false {
		t.Errorf("no arguments returns a weird result")
	}

}

// Test trimming strings
func TestTrim(t *testing.T) {

	type TestCase struct {
		Input  string
		Result string
	}

	tests := []TestCase{
		{Input: "   Steve", Result: "Steve"},
		{Input: "Steve    ", Result: "Steve"},
		{Input: "   Steve    ", Result: "Steve"},
		{Input: " π   ", Result: "π"},
		{Input: "   ", Result: ""},
		{Input: "    π    π   ", Result: "π    π"},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, &object.String{Value: test.Input})

		x := fnTrim(args)
		if x.(*object.String).Value != test.Result {
			t.Errorf("Invalid result for %s", test.Input)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnTrim(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
}

// Test types
func TestType(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result string
	}

	tests := []TestCase{
		{Input: &object.String{Value: "Steve"}, Result: "string"},
		{Input: &object.Integer{Value: 1}, Result: "integer"},
		{Input: &object.Float{Value: 3.2}, Result: "float"},
		{Input: &object.Boolean{Value: true}, Result: "boolean"},
		{Input: &object.Null{}, Result: "null"},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnType(args)
		if x.(*object.String).Value != test.Result {
			t.Errorf("Invalid result for %v got %s", test.Input, x.(*object.String).Value)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnType(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

}

// Test upper-casing strings
func TestUpper(t *testing.T) {

	type TestCase struct {
		Input  object.Object
		Result string
	}

	tests := []TestCase{
		{Input: &object.String{Value: "Steve"}, Result: "STEVE"},
		{Input: &object.String{Value: "π"}, Result: "Π"},
		{Input: &object.Integer{Value: 1}, Result: "1"},
		{Input: &object.Float{Value: 3.2}, Result: "3.2"},
		{Input: &object.Boolean{Value: true}, Result: "TRUE"},
		{Input: &object.Null{}, Result: "NULL"},
	}

	// For each test
	for _, test := range tests {

		var args []object.Object
		args = append(args, test.Input)

		x := fnUpper(args)
		if x.(*object.String).Value != test.Result {
			t.Errorf("Invalid result for %s", test.Input)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnUpper(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
}

// NOP-test
func TestPrint(t *testing.T) {
	var args []object.Object
	fnPrint(args)

	args = append(args, &object.String{Value: ""})
	fnPrint(args)
}
