package environment

import (
	"os"
	"testing"
	"time"

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

// TestTime performs *minimal* invocation of time-fields
func TestTime(t *testing.T) {

	//
	// Call all the functions with no arguments
	//
	var args []object.Object
	var out object.Object

	out = fnHour(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnMinute(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnSeconds(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnDay(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnMonth(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnYear(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnWeekday(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	//
	// Now create a known-time
	//
	args = append(args, &object.Integer{Value: 195315316})

	//
	// And check the values.
	//
	if fnDay(args).(*object.Integer).Value != 10 {
		t.Errorf("Failed to get the correct date")
	}
	if fnMonth(args).(*object.Integer).Value != 3 {
		t.Errorf("Failed to get the correct date")
	}
	if fnYear(args).(*object.Integer).Value != 1976 {
		t.Errorf("Failed to get the correct date")
	}
	if fnWeekday(args).(*object.String).Value != "Wednesday" {
		t.Errorf("Failed to get the correct date")
	}

	if fnHour(args).(*object.Integer).Value != 14 {
		t.Errorf("Failed to get the correct time")
	}
	if fnMinute(args).(*object.Integer).Value != 15 {
		t.Errorf("Failed to get the correct time")
	}
	if fnSeconds(args).(*object.Integer).Value != 16 {
		t.Errorf("Failed to get the correct time")
	}

	//
	// Test bogus field-name to the internal-helper.
	//
	if getTimeField(args, "bogus").Type() != object.NULL {
		t.Errorf("unexpected value passing bogus argument")
	}

	//
	// Finally check with a bogus-type to one of the methods
	//
	var bogus []object.Object
	bogus = append(bogus, &object.String{Value: "not int"})

	if fnSeconds(bogus).Type() != object.NULL {
		t.Errorf("unexpected value passing bogus argument")
	}

}

// Test `now`
func TestNow(t *testing.T) {

	// Handle timezones, by reading $TZ, and if not set
	// defaulting to UTC.
	env := os.Getenv("TZ")
	if env == "" {
		env = "UTC"
	}

	now := time.Now()

	// Ensure we set that timezone.
	loc, err := time.LoadLocation(env)
	if err == nil {
		now = now.In(loc)
	}

	// Call the function
	var empty []object.Object
	out := fnNow(empty)

	// type-check
	if out.Type() != object.INTEGER {
		t.Errorf("output of `now` was not an integer")
	}

	// get the value
	val := out.(*object.Integer).Value

	// diff
	diff := val - now.Unix()
	if diff < 0 {
		diff *= -1
	}

	if diff > 2 {
		t.Errorf("getting current time differed from expected value by more than two seconds.  weird")
	}
}

// Test formatting strings
func TestSprintf(t *testing.T) {

	type TestCase struct {
		Input  []object.Object
		Result string
	}

	tests := []TestCase{
		{Input: []object.Object{
			&object.String{Value: "%s %s"},
			&object.String{Value: "steve"},
			&object.String{Value: "kemp"}},
			Result: "steve kemp"},

		{Input: []object.Object{
			&object.String{Value: "%d"},
			&object.Integer{Value: 12}},
			Result: "12"},

		{Input: []object.Object{
			&object.String{Value: "%f %d"},
			&object.Float{Value: 3.222219},
			&object.Integer{Value: -3}},
			Result: "3.222219 -3"},

		{Input: []object.Object{
			&object.String{Value: "%t %t %t"},
			&object.Boolean{Value: true},
			&object.Boolean{Value: false},
			&object.Boolean{Value: true}},
			Result: "true false true"},

		{Input: []object.Object{&object.String{Value: "%%"}},
			Result: "%"},

		{Input: []object.Object{&object.String{Value: "no arguments"}},
			Result: "no arguments"},
	}

	// For each test
	for i, test := range tests {

		var args []object.Object
		args = append(args, test.Input...)

		x := fnSprintf(args)
		if x.(*object.String).Value != test.Result {
			t.Errorf("Invalid result for test %d, got %s", i, x)
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnSprintf(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// Calling the an argument of non-string
	args = append(args, &object.Integer{Value: 32})
	out = fnSprintf(args)
	if out.Type() != object.NULL {
		t.Errorf("non-string argument returns a weird result")
	}
}

// Test printing formatting strings
func TestPrintf(t *testing.T) {

	type TestCase struct {
		Input  []object.Object
		Result bool
	}

	tests := []TestCase{
		{Input: []object.Object{
			&object.String{Value: "%s %s"},
			&object.String{Value: "steve"},
			&object.String{Value: "kemp"}},
			Result: true},

		{Input: []object.Object{
			&object.String{Value: "%d"},
			&object.Integer{Value: 12}},
			Result: true},

		{Input: []object.Object{
			&object.String{Value: "%f %d"},
			&object.Float{Value: 3.222219},
			&object.Integer{Value: -3}},
			Result: true},

		{Input: []object.Object{
			&object.String{Value: "%t %t %t"},
			&object.Boolean{Value: true},
			&object.Boolean{Value: false},
			&object.Boolean{Value: true}},
			Result: true},

		{Input: []object.Object{&object.String{Value: "%%"}},
			Result: true},

		{Input: []object.Object{&object.String{Value: "no arguments"}},
			Result: true},

		// no arg
		{Input: []object.Object{},
			Result: false},

		// bad type
		{Input: []object.Object{&object.Boolean{Value: false}},
			Result: false},
	}

	// For each test
	for i, test := range tests {

		var args []object.Object
		args = append(args, test.Input...)

		x := fnPrintf(args)
		if x.(*object.Boolean).Value != test.Result {
			t.Errorf("Invalid result for test %d, got %s", i, x)
		}
	}
}
