package vm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/skx/evalfilter/v2/code"
	"github.com/skx/evalfilter/v2/environment"
	"github.com/skx/evalfilter/v2/object"
)

func TestBool(t *testing.T) {

	vm := New(nil, nil, nil, environment.New())
	tb := vm.nativeBoolToBooleanObject(true)
	fb := vm.nativeBoolToBooleanObject(false)

	if tb != True {
		t.Fatalf("bool mismatch")
	}
	if fb != False {
		t.Fatalf("bool mismatch")
	}
}

// Test that we can handle a timeout when running an endless program.
//
// [Special]
func TestContextTimeout(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run - endless loop
	bytecode := code.Instructions{
		byte(code.OpJump),
		byte(0),
		byte(0),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()

	// Timeout after a second
	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "timeout during execution") {
		t.Fatalf("got error, but not the expected one: %s", err.Error())
	}
}

// TestDivideByZero is testing a division by zero in the optimizer
// NOT at runtime.
//
// [Optimize]
func TestDivideByZero(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpDiv),
		byte(code.OpReturn),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Fatalf("got error, but not the expected one: %s", err.Error())
	}
}

func TestEmptyProgram(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "program is empty") {
		t.Fatalf("got error, but not the expected one: %s", err.Error())
	}
}

// [Special]
func TestInit(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(code.OpTrue),
		byte(code.OpReturn)}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	ctx := context.Background()

	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	_, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

}

// Special
func TestMissingReturn(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(code.OpTrue),
		byte(code.OpFalse),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "missing return") {
		t.Fatalf("got error, but not the expected one: %s", err.Error())
	}
}

func TestOpArray(t *testing.T) {

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// Build an array with zero args.
		{program: code.Instructions{
			byte(code.OpArray),  // 0x00
			byte(0),             // 0x01
			byte(0),             // 0x02
			byte(code.OpReturn), // 0x03
		}, result: "[]", error: false},

		// stack underflow
		{program: code.Instructions{
			byte(code.OpArray),  // 0x00
			byte(0),             // 0x01
			byte(1),             // 0x02
			byte(code.OpReturn), // 0x03
		}, result: "Pop from an empty stack", error: true},

		// [true]
		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpArray),  // 0x01
			byte(0),             // 0x02
			byte(1),             // 0x03
			byte(code.OpReturn), // 0x04
		}, result: "[true]", error: false},

		// [true, true, true, true]
		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpTrue),   // 0x01
			byte(code.OpTrue),   // 0x02
			byte(code.OpTrue),   // 0x03
			byte(code.OpArray),  // 0x04
			byte(0),             // 0x05
			byte(4),             // 0x06
			byte(code.OpReturn), // 0x07
		}, result: "[true, true, true, true]", error: false},
	}

	for _, test := range tests {

		// One constants
		constants := []object.Object{&object.String{Value: "Steve"}}

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

func TestOpBang(t *testing.T) {

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// !true -> false
		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "false", error: false},

		// !false -> true
		{program: code.Instructions{
			byte(code.OpFalse),  // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "true", error: false},

		// !2 -> false
		{program: code.Instructions{
			byte(code.OpPush),   // 0x00
			byte(0),             // 0x01
			byte(2),             // 0x02
			byte(code.OpBang),   // 0x03
			byte(code.OpReturn), // 0x04
		}, result: "false", error: false},

		// !(null) -> true
		{program: code.Instructions{
			byte(code.OpLookup), // 0x00
			byte(0),             // 0x01
			byte(0),             // 0x02
			byte(code.OpBang),   // 0x03
			byte(code.OpReturn), // 0x04
		}, result: "true", error: false},

		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "false", error: false},

		// Test empty stack
		{program: code.Instructions{
			byte(code.OpBang),
		}, result: "Pop from an empty stack", error: true},
	}

	for _, test := range tests {

		// One constants
		constants := []object.Object{&object.String{Value: "Steve"}}

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

// TODO: User-Defined Function(s)
func TestOpCall(t *testing.T) {

	// Constants
	constants := []object.Object{&object.String{Value: "Steve"},
		&object.String{Value: "len"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// empty stack
		{program: code.Instructions{
			byte(code.OpCall), // 0x00
			byte(0),           // 0x01
			byte(0),           // 0x02 -> val
		}, result: "Pop from an empty stack", error: true},

		// call: len(steve) -> but missing the string
		{program: code.Instructions{
			byte(code.OpConstant), // len
			byte(0),
			byte(1),
			byte(code.OpCall),
			byte(0),
			byte(1),
			byte(code.OpReturn),
		}, result: "Pop from an empty stack", error: true},

		// call: len(steve)
		{program: code.Instructions{
			byte(code.OpConstant), // steve
			byte(0),
			byte(0),
			byte(code.OpConstant), // len
			byte(0),
			byte(1),
			byte(code.OpCall),
			byte(0),
			byte(1),
			byte(code.OpReturn),
		}, result: "5", error: false},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}

}

// [Special]
func TestOpConstant(t *testing.T) {
	// One constant
	constants := []object.Object{&object.String{Value: "Steve"}}

	// The program we run:
	bytecode := code.Instructions{
		byte(code.OpConstant), // 0x00
		byte(0),               // 0x01
		byte(0),               // 0x02
		byte(code.OpReturn),   // 0x03
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	// Default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	out, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Result should be our constant.
	if out.Inspect() != "Steve" {
		t.Errorf("program has wrong result: %v", out)
	}
}

func TestOpDec(t *testing.T) {

	// Constants
	constants := []object.Object{&object.String{Value: "key"},
		&object.String{Value: "val"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// dec on a string
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(1),               // 0x02 -> val
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = steve"
			byte(code.OpDec),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "doesn't implement the Decrement() interface", error: true},

		// dec on a number
		{program: code.Instructions{
			byte(code.OpPush),     // 0x00
			byte(0),               // 0x01
			byte(9),               // 0x02 -> 9
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = 9", decrease  it

			// BUG?  Here we MUST set OpLookup otherwise
			// we get a stack-mismatch.
			byte(code.OpLookup),
			byte(0),
			byte(0),

			byte(code.OpDec),
			byte(0),
			byte(0),

			// Get the value
			byte(code.OpLookup),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "8", error: false},

		// dec on a number - but no OpLookup
		{program: code.Instructions{
			byte(code.OpPush),     // 0x00
			byte(0),               // 0x01
			byte(9),               // 0x02 -> 9
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = 9", decrease it

			// BUG?  Here we MUST set OpLookup otherwise
			// we get a stack-mismatch.
			byte(code.OpDec),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "Pop from an empty stack", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}

}

func TestOpInc(t *testing.T) {

	// Constants
	constants := []object.Object{&object.String{Value: "key"},
		&object.String{Value: "val"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// inc on a string
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(1),               // 0x02 -> val
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = steve"
			byte(code.OpInc),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "doesn't implement the Increment() interface", error: true},

		// inc on a number
		{program: code.Instructions{
			byte(code.OpPush),     // 0x00
			byte(0),               // 0x01
			byte(9),               // 0x02 -> 9
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = 9", increase it

			// BUG?  Here we MUST set OpLookup otherwise
			// we get a stack-mismatch.
			byte(code.OpLookup),
			byte(0),
			byte(0),

			byte(code.OpInc),
			byte(0),
			byte(0),

			// Get the value
			byte(code.OpLookup),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "10", error: false},

		// inc on a number - with no OpLookup
		{program: code.Instructions{
			byte(code.OpPush),     // 0x00
			byte(0),               // 0x01
			byte(9),               // 0x02 -> 9
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> key
			byte(code.OpSet),      // 0x06

			// now "key = 9", increase it
			byte(code.OpInc),
			byte(0),
			byte(0),

			// Get the value
			byte(code.OpLookup),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "Pop from an empty stack", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}

}

func TestOpIndex(t *testing.T) {
	// Constants
	constants := []object.Object{&object.String{Value: "Steve"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// empty stack
		{program: code.Instructions{
			byte(code.OpIndex),
		}, result: "Pop from an empty stack", error: true},

		// stack is too small.
		{program: code.Instructions{
			byte(code.OpFalse),
			byte(code.OpIndex),
		}, result: "Pop from an empty stack", error: true},

		// 1[1] -> "type error"
		{program: code.Instructions{
			byte(code.OpPush), // 0x00
			byte(0),           // 0x01
			byte(0),           // 0x02
			byte(code.OpPush), // 0x03
			byte(0),
			byte(0),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "the index operator can only be applied to string", error: true},

		// "steve"["steve"] -> "type error"
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(0),               // 0x02
			byte(code.OpConstant), // 0x03
			byte(0),
			byte(0),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "must be given an integer", error: true},

		// "steve"[1] -> "t"
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(0),               // 0x02
			byte(code.OpPush),     // 0x03
			byte(0),
			byte(1),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "t", error: false},

		// "steve"[2570] -> "NULL"
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(0),               // 0x02
			byte(code.OpPush),     // 0x03
			byte(10),
			byte(10),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "null", error: false},

		// create array: index[1]
		{program: code.Instructions{
			// create [1,2,3,...8,9,10]
			byte(code.OpPush),
			byte(0),
			byte(1),
			byte(code.OpPush),
			byte(0),
			byte(10),
			byte(code.OpRange),

			// index
			byte(code.OpPush),
			byte(0),
			byte(1),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "2", error: false},

		// create array: array[255]
		{program: code.Instructions{
			// create [1,2,3,...8,9,10]
			byte(code.OpPush),
			byte(0),
			byte(1),
			byte(code.OpPush),
			byte(0),
			byte(10),
			byte(code.OpRange),

			// index
			byte(code.OpPush),
			byte(0),
			byte(255),
			byte(code.OpIndex),
			byte(code.OpReturn),
		}, result: "null", error: false},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

func TestOpIterationNext(t *testing.T) {
}

func TestOpIterationReset(t *testing.T) {
	constants := []object.Object{&object.String{Value: "Steve"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// empty stack
		{program: code.Instructions{
			byte(code.OpIterationReset),
		}, result: "Pop from an empty stack", error: true},

		// something that cannot be iterated over
		{program: code.Instructions{
			byte(code.OpTrue),
			byte(code.OpIterationReset),
		}, result: "object doesn't implement the Iterable interface", error: true},
		// something that can be iterated over
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpIterationReset),
			byte(code.OpReturn),
		}, result: "Steve", error: false},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

func TestOpJumpIfFalse(t *testing.T) {
	constants := []object.Object{}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// empty stack
		{program: code.Instructions{
			byte(code.OpJumpIfFalse),
			byte(0),
			byte(0),
		}, result: "Pop from an empty stack", error: true},

		// true
		{program: code.Instructions{
			byte(code.OpTrue),
			byte(code.OpJumpIfFalse),
			byte(0),
			byte(0),
			byte(code.OpTrue),
			byte(code.OpReturn),
		}, result: "true", error: false},

		// false
		{program: code.Instructions{
			byte(code.OpFalse),       // 0x00
			byte(code.OpJumpIfFalse), // 0x01
			byte(0),                  // 0x02
			byte(6),                  // 0x03
			byte(code.OpTrue),        // 0x04
			byte(code.OpReturn),      // 0x05
			byte(code.OpFalse),       // 0x06
			byte(code.OpReturn),      // 0x07
		}, result: "false", error: false},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

// This is mostly a test of our reflection.
//
// [Special]
func TestOpLookup(t *testing.T) {

	// Constants: field names we lookup in our struct
	constants := []object.Object{
		&object.String{Value: "Name"},
		&object.String{Value: "Array"},
		&object.String{Value: "Time"},
		&object.String{Value: "True"},
		&object.String{Value: "Int"},
		&object.String{Value: "Float"},
	}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// lookup name
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(0),
			byte(code.OpReturn),
		}, result: "Steve Kemp", error: false},

		// lookup array
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(1),
			byte(code.OpReturn),
		}, result: "[Bart, Lisa, Maggie]", error: false},

		// lookup array
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(2),
			byte(code.OpReturn),
		}, result: "1598613755", error: false},

		// lookup bool
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(3),
			byte(code.OpReturn),
		}, result: "true", error: false},

		// lookup int
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(4),
			byte(code.OpReturn),
		}, result: "17", error: false},

		// lookup float
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(5),
			byte(code.OpReturn),
		}, result: "3.2", error: false},
	}

	// The object we pass to the engine
	type Input struct {
		Name  string
		Array []string
		Time  time.Time
		True  bool
		Int   int64
		Float float64
	}

	// The instance of that object.
	in := Input{Name: "Steve Kemp",
		Array: []string{"Bart", "Lisa", "Maggie"},
		Time:  time.Unix(1598613755, 0),
		True:  true,
		Int:   17,
		Float: 3.2}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run - with the structure
		out, err := vm.Run(in)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

func TestOpMinus(t *testing.T) {
	// Constants
	constants := []object.Object{&object.Float{Value: 3.1},
		&object.String{Value: "not a number"},
	}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// minus -> empty stack
		{program: code.Instructions{
			byte(code.OpMinus),
		}, result: "Pop from an empty stack", error: true},

		// minus 10 -> -10
		{program: code.Instructions{
			byte(code.OpPush),
			byte(0),
			byte(10),
			byte(code.OpMinus),
			byte(code.OpReturn),
		}, result: "-10", error: false},

		// minus 3.1 -> -3.1
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpMinus),
			byte(code.OpReturn),
		}, result: "-3.1", error: false},

		// minus "string" -> type error
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(1),
			byte(code.OpMinus),
			byte(code.OpReturn),
		}, result: "unsupported type for negation", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

func TestOpRange(t *testing.T) {
	// Empty constants
	constants := []object.Object{&object.String{Value: "hello!"}}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{
		// empty stack
		{program: code.Instructions{
			byte(code.OpRange),
		},
			result: "Pop from an empty stack", error: true},

		// too small stack
		{program: code.Instructions{
			byte(code.OpTrue),
			byte(code.OpRange),
		},
			result: "Pop from an empty stack", error: true},

		// range( string, int )
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpPush),
			byte(1),
			byte(1),
			byte(code.OpRange),
		},
			result: "the range must be an integer", error: true},

		// range(int, string)
		{program: code.Instructions{
			byte(code.OpPush),
			byte(1),
			byte(1),
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpRange),
		},
			result: "the range must be an integer", error: true},

		// range(string, string)
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpRange),
		},
			result: "the range must be an integer", error: true},

		// range(1, 10)
		{program: code.Instructions{
			byte(code.OpPush),
			byte(0),
			byte(1),
			byte(code.OpPush),
			byte(0),
			byte(10),
			byte(code.OpRange),
			byte(code.OpReturn),
		},
			result: "[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]", error: false},

		// range(10,1) -> error
		{program: code.Instructions{
			byte(code.OpPush),
			byte(0),
			byte(10),
			byte(code.OpPush),
			byte(0),
			byte(1),
			byte(code.OpRange),
			byte(code.OpReturn),
		},
			result: "the start of a range must be smaller than the end", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out.Inspect())
			}
		}
	}
}

func TestOpSet(t *testing.T) {
	// A pair of constants
	constants := []object.Object{&object.String{Value: "Steve"},
		&object.String{Value: "Kemp"},
	}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{
		// Set "Steve" -> "Kemp" then lookup the result
		{program: code.Instructions{
			byte(code.OpConstant), // 0x00
			byte(0),               // 0x01
			byte(1),               // 0x02 -> Steve
			byte(code.OpConstant), // 0x03
			byte(0),               // 0x04
			byte(0),               // 0x05 -> Kemp
			byte(code.OpSet),      // 0x06
			byte(code.OpLookup),
			byte(0),
			byte(0),
			byte(code.OpReturn), // 0x08,
		}, result: "Kemp", error: false},

		// Empty stack
		{program: code.Instructions{
			byte(code.OpSet),    // 0x00
			byte(code.OpReturn), // 0x01,
		}, result: "Pop from an empty stack", error: true},

		// Only one stack entry
		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpSet),    // 0x01
			byte(code.OpReturn), // 0x02,
		}, result: "Pop from an empty stack", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result, expected %s: %s", test.result, out)
			}
		}
	}
}

// [Special]
// [Optimize]
func TestOpSquareRoot(t *testing.T) {

	// A pair of constants
	constants := []object.Object{&object.Integer{Value: 9},
		&object.Float{Value: 16.0},
	}

	type TestCase struct {
		program code.Instructions
		result  string
		error   bool
	}

	tests := []TestCase{

		// root(9) -> 3
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(0),
			byte(code.OpSquareRoot),
			byte(code.OpReturn),
		}, result: "3", error: false},

		// root(16.0) -> 4
		{program: code.Instructions{
			byte(code.OpConstant),
			byte(0),
			byte(1),
			byte(code.OpSquareRoot),
			byte(code.OpReturn),
		}, result: "4", error: false},

		// root(false) -> error
		{program: code.Instructions{
			byte(code.OpFalse),
			byte(code.OpSquareRoot),
			byte(code.OpReturn),
		}, result: "unsupported type for square-root", error: true},

		// root() -> error
		{program: code.Instructions{
			byte(code.OpSquareRoot),
			byte(code.OpReturn),
		}, result: "Pop from an empty stack", error: true},
	}

	for _, test := range tests {

		// No functions
		functions := make(map[string]environment.UserFunction)

		// Default environment
		env := environment.New()
		env.Set("DEBUG", &object.Boolean{Value: true})
		env.Set("OPTIMIZE", &object.Boolean{Value: true})

		// Default context
		ctx := context.Background()

		// Create
		vm := New(constants, test.program, functions, env)
		vm.SetContext(ctx)

		// Run
		out, err := vm.Run(nil)

		if test.error {
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("Error '%s' didn't contain '%s'", err.Error(), test.result)
			}
		} else {
			if err != nil {
				t.Fatalf("expected no error - found:%s\n", err.Error())
			}

			// Result should be our constant.
			if out.Inspect() != test.result {
				t.Errorf("program has wrong result: %v", out)
			}
		}
	}
}

// [Special]
func TestOpVoid(t *testing.T) {
	// No constants
	constants := []object.Object{&object.String{Value: "Steve"}}

	// The program we run:
	bytecode := code.Instructions{
		byte(code.OpVoid),   // 0x00
		byte(code.OpReturn), // 0x01
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	// Default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	out, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Result should be void
	if out.Inspect() != "void" {
		t.Errorf("program has wrong result: %v", out)
	}
}

// Test constant-comparisons are removed.
// [Optimize]
func TestOptimizerConstants(t *testing.T) {
	// No constants
	constants := []object.Object{}

	// The program we run:
	bytecode := code.Instructions{
		// We add a jump here - to kill the deadcode
		// eliminator
		byte(code.OpJump), // 0x00
		byte(0),           // 0x01
		byte(3),           // 0x02

		// if ( 0 == 0 ) ..
		byte(code.OpPush), // 0x03
		byte(0),
		byte(0),
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpEqual),

		// if ( 1 == 0 )
		byte(code.OpPush),
		byte(0),
		byte(1),
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpEqual),

		// if ( 1 != 0 ) ..
		byte(code.OpPush),
		byte(0),
		byte(1),
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpNotEqual),

		// if ( 0 != 0 ) ..
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpPush),
		byte(0),
		byte(0),
		byte(code.OpNotEqual),

		// done
		byte(code.OpReturn),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// Default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	out, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Result should be false - as the second comparison will
	// be "0 != 0" which is false.
	if out.Inspect() != "false" {
		t.Errorf("optimized program has wrong result: %v", out)
	}

	// Check the bytecode was optimized
	after := len(vm.bytecode)

	if after == len(bytecode) {
		t.Fatalf("optimizer made no change to our bytecode")
	}
	expectedSize := 8
	if after != expectedSize {
		t.Fatalf("bytecode size was %d not %d", after, expectedSize)
	}

	// check we have the right output
	expected := code.Instructions{
		byte(code.OpJump),
		byte(0),
		byte(3),
		byte(code.OpTrue),  // 0 == 0
		byte(code.OpFalse), // 1 == 0
		byte(code.OpTrue),  // 1 != 0
		byte(code.OpFalse), // 0 != 0
		byte(code.OpReturn)}

	for i, op := range expected {
		if vm.bytecode[i] != op {
			t.Fatalf("index %d opcode was %s not %s", i,
				code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
		}
	}
}

// [Special]
// [Optimize]
func TestOptimizerEnabled(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(code.OpTrue),
		byte(code.OpReturn),
		byte(code.OpTrue),
		byte(code.OpReturn)}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// Default context
	ctx := context.Background()

	// Create

	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Check the bytecode was optimized
	after := len(vm.bytecode)

	if after == len(bytecode) {
		t.Fatalf("optimizer made no change to our bytecode")
	}
	expectedSize := 2
	if after != expectedSize {
		t.Fatalf("bytecode size was %d not %d", after, expectedSize)
	}

	// check we have the right output
	expected := code.Instructions{
		byte(code.OpTrue),
		byte(code.OpReturn)}

	for i, op := range expected {
		if vm.bytecode[i] != op {
			t.Fatalf("index %d opcode was %s not %s", i,
				code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
		}
	}
}

// Test constant-jumps are removed.
// [Special]
// [Optimize]
func TestOptimizerJumps(t *testing.T) {
	// No constants
	constants := []object.Object{}

	// The program we run:
	bytecode := code.Instructions{
		// should be removed:
		byte(code.OpTrue),        // 0x00
		byte(code.OpJumpIfFalse), // 0x01
		byte(33),                 // 0x02
		byte(44),                 // 0x03
		// should be removed.
		byte(code.OpFalse),       // 0x04
		byte(code.OpJumpIfFalse), // 0x05
		byte(0),                  // 0x06
		byte(9),                  // 0x07
		byte(code.OpFalse),       // 0x08
		byte(code.OpTrue),        // 0x09 -> JumpTarget
		byte(code.OpReturn),      // 0x0A
		byte(code.OpReturn),      // 0x0B
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// Default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	out, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Result should be true.
	if out.Inspect() != "true" {
		t.Errorf("optimized program has wrong result: %v", out)
	}

	// Check the bytecode was optimized
	after := len(vm.bytecode)

	if after == len(bytecode) {
		t.Fatalf("optimizer made no change to our bytecode")
	}
	expectedSize := 2
	if after != expectedSize {
		t.Fatalf("bytecode size was %d not %d", after, expectedSize)
	}

	// check we have the right output
	expected := code.Instructions{
		byte(code.OpTrue),
		byte(code.OpReturn)}

	for i, op := range expected {
		if vm.bytecode[i] != op {
			t.Fatalf("index %d opcode was %s not %s", i,
				code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
		}
	}
}

// Test constant-maths expressions are replaced with their results.
//
// [Special]
// [Optimize]
func TestOptimizerMaths(t *testing.T) {
	// No constants
	constants := []object.Object{}

	// The program we run:
	//    return 4 + 2 * 3 / 2 ;
	//  => 7
	bytecode := code.Instructions{
		byte(code.OpPush),
		byte(0),
		byte(4),

		byte(code.OpPush),
		byte(0),
		byte(2),

		byte(code.OpPush),
		byte(0),
		byte(3),

		byte(code.OpMul),

		byte(code.OpPush),
		byte(0),
		byte(2),

		byte(code.OpDiv),
		byte(code.OpAdd),

		// extra: - 0
		byte(code.OpPush),
		byte(0), byte(0),
		byte(code.OpSub),

		byte(code.OpReturn),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// Default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	out, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got :%s\n", err.Error())
	}

	// Result should be 7.
	if out.Inspect() != "7" {
		t.Errorf("optimized program has wrong result: %v", out)
	}

	// Check the bytecode was optimized
	after := len(vm.bytecode)

	if after == len(bytecode) {
		t.Fatalf("optimizer made no change to our bytecode")
	}
	expectedSize := 4
	if after != expectedSize {
		t.Fatalf("bytecode size was %d not %d", after, expectedSize)
	}

	// check we have the right output
	expected := code.Instructions{
		byte(code.OpPush),
		byte(0),
		byte(7),
		byte(code.OpReturn)}

	for i, op := range expected {
		if vm.bytecode[i] != op {
			t.Fatalf("index %d opcode was %s not %s", i,
				code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
		}
	}
}

// Test OpNops are removed.
// [Special]
// [Optimize]
func TestOptimizerNops(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(code.OpNop),
		byte(code.OpFalse),
		byte(code.OpNop),
		byte(code.OpReturn)}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Environment will enable the optimizer
	env := environment.New()
	env.Set("OPTIMIZE", &object.Boolean{Value: true})

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err != nil {
		t.Fatalf("expected no error, got: %s", err.Error())
	}

	// Check the bytecode was optimized
	after := len(vm.bytecode)

	if after == len(bytecode) {
		t.Fatalf("optimizer made no change to our bytecode")
	}
	expectedSize := 2
	if after != expectedSize {
		t.Fatalf("bytecode size was %d not %d", after, expectedSize)
	}

	// check we have the right output
	expected := code.Instructions{
		byte(code.OpFalse),
		byte(code.OpReturn)}

	for i, op := range expected {
		if vm.bytecode[i] != op {
			t.Fatalf("index %d opcode was %s not %s", i,
				code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
		}
	}
}

func TestUnknownOpcode(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run
	bytecode := code.Instructions{
		byte(200),
		byte(code.OpReturn),
	}

	// No functions
	functions := make(map[string]environment.UserFunction)

	// Default environment
	env := environment.New()

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	// Run
	_, err := vm.Run(nil)
	if err == nil {
		t.Fatalf("expected an error, got none")
	}
	if !strings.Contains(err.Error(), "unhandled opcode") {
		t.Fatalf("got error, but not the expected one: %s", err.Error())
	}
}

// [Special]
func TestWalkFunctionBytecode(t *testing.T) {

	// No constants
	constants := []object.Object{}

	// The program we run.
	bytecode := code.Instructions{
		byte(code.OpFalse),
		byte(code.OpReturn)}

	// A single function
	functions := make(map[string]environment.UserFunction)

	functions["steve"] = environment.UserFunction{
		Bytecode: bytecode,
	}

	// Environment will enable the optimizer
	env := environment.New()

	// default context
	ctx := context.Background()

	// Create
	vm := New(constants, bytecode, functions, env)
	vm.SetContext(ctx)

	err := vm.WalkFunctionBytecode("bob", func(offset int, opCode code.Opcode, opArg interface{}) (bool, error) {
		return false, nil
	})
	if err == nil {
		t.Fatalf("Expected error walking function that doesn't exist")
	}

	// Now the function that does exist
	count := 0
	err = vm.WalkFunctionBytecode("steve", func(offset int, opCode code.Opcode, opArg interface{}) (bool, error) {
		count++
		return false, nil
	})
	if err != nil {
		t.Fatalf("Unexpected error walking function that doesn't exist")
	}
	if count != 1 {
		t.Fatalf("callback didn't get invoked once")
	}

}
