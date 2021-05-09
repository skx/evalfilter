package environment

import (
	"os"
	"testing"
	"time"

	"github.com/skx/evalfilter/v2/object"
)

// Test between functions
func TestBetween(t *testing.T) {

	type TestCase struct {
		v   object.Object
		min object.Object
		max object.Object
		res bool
	}

	tests := []TestCase{

		// 0 in 0-10 -> OK
		{
			v:   &object.Integer{Value: 0},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: true,
		},

		// -4 in 0-10 -> NO
		{
			v:   &object.Integer{Value: -5},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: false,
		},

		// 1 in 0-10 -> OK
		{
			v:   &object.Integer{Value: 1},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: true,
		},

		// 10 in 0-10 -> OK
		{
			v:   &object.Integer{Value: 10},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: true,
		},

		// 11 in 0-10 -> False
		{
			v:   &object.Integer{Value: 11},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: false,
		},

		// -10 in 0-10 -> NO
		{
			v:   &object.Float{Value: -10},
			min: &object.Integer{Value: 0},
			max: &object.Integer{Value: 10},
			res: false,
		},

		// 0 in 0-10 -> OK
		{
			v:   &object.Integer{Value: 0},
			min: &object.Float{Value: 0},
			max: &object.Float{Value: 10},
			res: true,
		},

		// 1 in 0-10 -> OK
		{
			v:   &object.Integer{Value: 1},
			min: &object.Float{Value: 0},
			max: &object.Float{Value: 10},
			res: true,
		},

		// 10 in 0-10
		{
			v:   &object.Integer{Value: 10},
			min: &object.Float{Value: 0},
			max: &object.Float{Value: 10},
			res: true,
		},
	}

	// For each test
	for _, test := range tests {

		args := []object.Object{test.v, test.min, test.max}

		ret := fnBetween(args)

		if ret.(*object.Boolean).Value != test.res {

			t.Errorf("Invalid result for between( %s, %s, %s) - got %s", test.v.Inspect(), test.min.Inspect(), test.max.Inspect(), ret.Inspect())
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnBetween(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// one arg is bogus
	args = []object.Object{&object.Integer{Value: 1}}
	out = fnBetween(args)
	if out.Type() != object.NULL {
		t.Errorf("one argument returned a weird result")
	}

	// two args are bogus
	args = []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 1}}
	out = fnBetween(args)
	if out.Type() != object.NULL {
		t.Errorf("two arguments returned a weird result")
	}

	// three string args are bogus
	args = []object.Object{&object.String{Value: "Moi!"},
		&object.String{Value: "Moi!"},
		&object.String{Value: "Moi!"},
	}
	out = fnBetween(args)
	if out.Type() != object.NULL {
		t.Errorf("bad type returned a weird result")
	}

	// four are too!
	args = []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 1}, &object.Integer{Value: 1}, &object.Integer{Value: 1}}
	out = fnBetween(args)
	if out.Type() != object.NULL {
		t.Errorf("four arguments returned a weird result")
	}
}

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

func TestGetenv(t *testing.T) {

	val := "FOO1234"

	// Should be empty
	x := fnGetenv([]object.Object{&object.String{Value: val}})
	if x.Type() != "STRING" {
		t.Errorf("Invalid type result for return")
	}

	// Cast to string
	ret := x.(*object.String).Value
	if ret != "" {
		t.Errorf("Got an unexpected environmental variable")
	}

	// Now try again, after setting the value
	os.Setenv(val, val)

	x = fnGetenv([]object.Object{&object.String{Value: val}})
	if x.Type() != "STRING" {
		t.Errorf("Invalid type result for return")
	}

	// Cast to string
	ret = x.(*object.String).Value
	if ret != val {
		t.Errorf("Got an unexpected environmental variable")
	}

	// Reset
	os.Setenv(val, "")

	// ensure that zero arguments are handled
	var tmp []object.Object
	out := fnGetenv(tmp)
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

// Test keys-primitive
func TestKeys(t *testing.T) {

	// One argument is required
	var args []object.Object
	out := fnKeys(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// The argument must be a hash
	args = append(args, &object.String{Value: "bogus type"})
	out = fnKeys(args)
	if out.Type() != object.NULL {
		t.Errorf("bad argument-type returns a weird result")
	}

	// OK now we have a hash.  Empty one is easy
	tmp := &object.Hash{}
	out = fnKeys([]object.Object{tmp})
	if out.Type() != object.ARRAY {
		t.Errorf("keys(hash) returned non-arry")
	}

	// Populate the hash with a pair of keys
	a := object.HashPair{Key: &object.String{Value: "Name"}, Value: &object.String{Value: "Steve"}}
	aK := &object.String{Value: "Name"}
	b := object.HashPair{Key: &object.String{Value: "Country"}, Value: &object.String{Value: "Finland"}}
	bK := &object.String{Value: "Country"}

	tmp.Pairs = make(map[object.HashKey]object.HashPair)
	tmp.Pairs[aK.HashKey()] = a
	tmp.Pairs[bK.HashKey()] = b

	// Get the keys now
	out = fnKeys([]object.Object{tmp})
	if out.Type() != object.ARRAY {
		t.Errorf("keys(hash) returned non-arry")
	}

	// Array should have two elements.
	if len(out.(*object.Array).Elements) != 2 {
		t.Errorf("wrong number of keys")
	}

	// Key names will be stored in the array.
	//
	// Since keys are sorted we can assume:
	//    key1 = Country
	//    key2 = Name
	//
	values := out.(*object.Array).Elements
	if values[0].Type() != object.STRING {
		t.Errorf("not right type")
	}
	if values[0].(*object.String).Value != "Country" {
		t.Errorf("wrong key-name - 'Country' got %s", values[0].(*object.String).Value)
	}
	if values[1].Type() != object.STRING {
		t.Errorf("not right type")
	}
	if values[1].(*object.String).Value != "Name" {
		t.Errorf("wrong key-name - expected 'Name' got %s", values[1].(*object.String).Value)
	}
}

// Test split-primitive
func TestSplit(t *testing.T) {

	// Two arguments are required
	var args []object.Object
	out := fnSplit(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// 1 arg
	args = append(args, &object.Null{})
	out = fnSplit(args)
	if out.Type() != object.NULL {
		t.Errorf("one argument returns a weird result")
	}

	// 2 args - but wrong types
	args = append(args, &object.Null{})
	out = fnSplit(args)
	if out.Type() != object.NULL {
		t.Errorf("one argument returns a weird result")
	}

	// Valid
	args = []object.Object{&object.String{Value: "Steve\nKemp"},
		&object.String{Value: "\n"}}

	out = fnSplit(args)
	if out.Type() != object.ARRAY {
		t.Errorf("didn't get an array back with two string-args")
	}
	if (len(out.(*object.Array).Elements)) != 2 {
		t.Errorf("return value was incorrect")
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

	// Hash entries
	a := object.HashPair{Key: &object.String{Value: "Name"}, Value: &object.String{Value: "Steve"}}
	aK := &object.String{Value: "Name"}
	b := object.HashPair{Key: &object.String{Value: "Country"}, Value: &object.String{Value: "Finland"}}
	bK := &object.String{Value: "Country"}

	tmpOne := make(map[object.HashKey]object.HashPair)
	tmpTwo := make(map[object.HashKey]object.HashPair)

	tmpOne[aK.HashKey()] = a
	tmpTwo[aK.HashKey()] = a
	tmpTwo[bK.HashKey()] = b

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

		// Hashes
		{Input: &object.Hash{}, Result: 0},
		{Input: &object.Hash{Pairs: tmpOne}, Result: 1},
		{Input: &object.Hash{Pairs: tmpTwo}, Result: 2},
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

// Test minimum/maximum number
func TestMinMax(t *testing.T) {

	type TestCase struct {
		a   object.Object
		b   object.Object
		op  string
		res object.Object
	}

	tests := []TestCase{

		// two ints
		{a: &object.Integer{Value: 1}, b: &object.Integer{Value: 10}, res: &object.Integer{Value: 1}, op: "min"},
		{a: &object.Integer{Value: 1}, b: &object.Integer{Value: 10}, res: &object.Integer{Value: 10}, op: "max"},

		// Two floats
		{a: &object.Float{Value: 1}, b: &object.Float{Value: 10}, res: &object.Float{Value: 1}, op: "min"},
		{a: &object.Float{Value: 1}, b: &object.Float{Value: 10}, res: &object.Float{Value: 10}, op: "max"},

		// int + float
		{a: &object.Integer{Value: 1}, b: &object.Float{Value: 1.3}, res: &object.Integer{Value: 1}, op: "min"},
		{a: &object.Integer{Value: 1}, b: &object.Float{Value: 1.3}, res: &object.Float{Value: 1.3}, op: "max"},

		// float + int
		{a: &object.Float{Value: 1.3}, b: &object.Integer{Value: 1}, res: &object.Float{Value: 1.3}, op: "max"},
		{a: &object.Float{Value: -91.3}, b: &object.Integer{Value: 2}, res: &object.Integer{Value: 2}, op: "max"},
	}
	// For each test
	for _, test := range tests {

		args := []object.Object{test.a, test.b}
		var ret object.Object

		if test.op == "min" {
			ret = fnMin(args)
		}
		if test.op == "max" {
			ret = fnMax(args)
		}

		if ret.Inspect() != test.res.Inspect() {
			t.Errorf("Invalid result for %s(%s, %s) (got %s wanted %s)", test.op, test.a.Inspect(), test.b.Inspect(), ret.Inspect(), test.res.Inspect())
		}
	}

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnMin(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}
	out = fnMax(args)
	if out.Type() != object.NULL {
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
}

func TestTimeKnown(t *testing.T) {

	var args []object.Object

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
		Input []object.Object
	}

	tests := []TestCase{
		{Input: []object.Object{
			&object.String{Value: "%s %s"},
			&object.String{Value: "steve"},
			&object.String{Value: "kemp"}},
		},

		{Input: []object.Object{
			&object.String{Value: "%d"},
			&object.Integer{Value: 12}},
		},

		{Input: []object.Object{
			&object.String{Value: "%f %d"},
			&object.Float{Value: 3.222219},
			&object.Integer{Value: -3}},
		},

		{Input: []object.Object{
			&object.String{Value: "%t %t %t"},
			&object.Boolean{Value: true},
			&object.Boolean{Value: false},
			&object.Boolean{Value: true}},
		},

		{Input: []object.Object{&object.String{Value: "%%"}}},

		{Input: []object.Object{&object.String{Value: "no arguments"}}},

		// no arg
		{Input: []object.Object{}},

		// bad type
		{Input: []object.Object{&object.Boolean{Value: false}}},
	}

	// For each test
	for i, test := range tests {

		var args []object.Object
		args = append(args, test.Input...)

		x := fnPrintf(args)
		if x.Type() != object.VOID {
			t.Errorf("Invalid return type for test %d, got %s", i, x)
		}

	}
}

// Test sorting works
func TestSort(t *testing.T) {

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnSort(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// Calling the an initial argument which isn't an array
	args = append(args, &object.Integer{Value: 32})
	out = fnSort(args)
	if out.Type() != object.NULL {
		t.Errorf("non-string argument returns a weird result")
	}

	// Now calling with a first argument which is an array,
	// but a second argument which isn't.
	args = []object.Object{&object.Array{Elements: []object.Object{
		&object.String{Value: "steve"}}},
		&object.Integer{Value: 32}}
	out = fnSort(args)
	if out.Type() != object.NULL {
		t.Errorf("non-string argument returns a weird result")
	}

	//
	// OK now we test sorting works
	//
	var items []object.Object
	items = append(items, &object.String{Value: "Cake"})
	items = append(items, &object.String{Value: "Apples"})
	items = append(items, &object.String{Value: "and"})

	var in []object.Object
	in = append(in, &object.Array{Elements: items})

	//
	// Expected result: "Apples", "Cake", "and"
	//
	out = fnSort(in)
	first := out.(*object.Array).Elements[0]
	if first.(*object.String).Value != "Apples" {
		t.Errorf("post-sort the result was wrong")
	}

	//
	// Case-insensitive - via the second-argument of `true`
	//
	// Expected result: "and", "Apples", "Cake".
	//
	in = append(in, &object.Boolean{Value: true})
	out = fnSort(in)
	first = out.(*object.Array).Elements[0]
	if first.(*object.String).Value != "and" {
		t.Errorf("post-sort the result was wrong")
	}

	//
	// Final test of case
	//
	cased := []object.Object{
		&object.Array{Elements: []object.Object{&object.String{Value: "x"},
			&object.String{Value: "A"}},
		},
		&object.Boolean{Value: true}}

	out = fnSort(cased)
	first = out.(*object.Array).Elements[0]
	if first.(*object.String).Value != "A" {
		t.Errorf("post-sort the result was wrong")
	}
}

// Test sorting works in reverse
func TestReverseSort(t *testing.T) {

	// Calling the function with no-arguments should return null
	var args []object.Object
	out := fnReverse(args)
	if out.Type() != object.NULL {
		t.Errorf("no arguments returns a weird result")
	}

	// Calling the an initial argument which isn't an array
	args = append(args, &object.Integer{Value: 32})
	out = fnReverse(args)
	if out.Type() != object.NULL {
		t.Errorf("non-string argument returns a weird result")
	}

	// Now calling with a first argument which is an array,
	// but a second argument which isn't.
	args = []object.Object{&object.Array{Elements: []object.Object{
		&object.String{Value: "steve"}}},
		&object.Integer{Value: 32}}
	out = fnReverse(args)
	if out.Type() != object.NULL {
		t.Errorf("non-string argument returns a weird result")
	}

	//
	// OK now we test sorting works
	//
	var items []object.Object
	items = append(items, &object.String{Value: "Cake"})
	items = append(items, &object.String{Value: "Apples"})
	items = append(items, &object.String{Value: "and"})

	var in []object.Object
	in = append(in, &object.Array{Elements: items})

	//
	// Expected result: "and", "Cake", "Apples"
	//
	out = fnReverse(in)
	first := out.(*object.Array).Elements[0]
	if first.(*object.String).Value != "and" {
		t.Errorf("post-sort the result was wrong")
	}
	last := out.(*object.Array).Elements[2]
	if last.(*object.String).Value != "Apples" {
		t.Errorf("post-sort the result was wrong")
	}

	//
	// Case-insensitive - via the second-argument of `true`
	//
	// Expected result: "Cake", "Apples", "and"
	//
	in = append(in, &object.Boolean{Value: true})
	out = fnReverse(in)
	first = out.(*object.Array).Elements[0]
	if first.(*object.String).Value != "Cake" {
		t.Errorf("post-sort the result was wrong")
	}
	last = out.(*object.Array).Elements[2]
	if last.(*object.String).Value != "and" {
		t.Errorf("post-sort the result was wrong")
	}

}
