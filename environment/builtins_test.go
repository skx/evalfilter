package environment

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

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
	if out.(*object.String).Value != "null" {
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
}
