// Mostly table-driven tests for exercising the different OpCode handlers
// individually.
//
// Similarly testing the optimizer.

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

// TestCase is the structure for describing our test-cases.
//
// Most of the tests in this file run a series of bytecode programs, and
// then validate that a particular return value was received.
//
// We've created `RunTestCases` to handle this common setup/execution
// of the tests.  This is the structure that it uses.
//
// The return value is either:
//
//    a) the topmost stack-item returned, or
//    b) an error-string returned in the case of an (expected) error.
//
// If the `optimized` field is filled out then the bytecode will be
// compared with that, after execution.
//
type TestCase struct {

	// The (bytecode) program we're going to run.
	program code.Instructions

	// When optimizations are enabled then this is the
	// output we expect to receive, once the execution
	// is complete.
	optimized code.Instructions

	// User-defined functions which are made available.
	functions map[string]environment.UserFunction

	// The expected result of the program.
	result string

	// Is there an error expected?  If so the result
	// must be contained in the error-message, otherwise
	// we'll compare the result with the output-value.
	error bool
}

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

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(0),
				byte(code.OpPush),
				byte(0),
				byte(0),
				byte(code.OpDiv),
				byte(code.OpReturn),
			},
			result: "division by zero",
			error:  true,

			// Optimizer makes no changes, because
			// of the division-by-zero error
			optimized: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(0),
				byte(code.OpPush),
				byte(0),
				byte(0),
				byte(code.OpDiv),
				byte(code.OpReturn),
			},
		},
	}

	RunTestCases(tests, []object.Object{}, t)
}

func TestEmptyProgram(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{},
			result:  "program is empty",
			error:   true,
		},
	}

	RunTestCases(tests, []object.Object{}, t)
}

func TestInit(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpReturn),
			},
			result: "true",
			error:  false,
		},
	}

	RunTestCases(tests, []object.Object{}, t)
}

func TestMissingReturn(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpFalse),
			},
			result: "null",
			error:  false,
		},
	}

	RunTestCases(tests, []object.Object{}, t)
}

func TestOpArray(t *testing.T) {

	tests := []TestCase{
		// Build an array with zero args.
		{
			program: code.Instructions{
				byte(code.OpArray),  // 0x00
				byte(0),             // 0x01
				byte(0),             // 0x02
				byte(code.OpReturn), // 0x03
			},
			result: "[]",
			error:  false,
		},

		// stack underflow
		{
			program: code.Instructions{
				byte(code.OpArray),  // 0x00
				byte(0),             // 0x01
				byte(1),             // 0x02
				byte(code.OpReturn), // 0x03
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// [true]
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpArray),  // 0x01
				byte(0),             // 0x02
				byte(1),             // 0x03
				byte(code.OpReturn), // 0x04
			},
			result: "[true]",
			error:  false,
		},

		// [true, true, true, true]
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpTrue),   // 0x01
				byte(code.OpTrue),   // 0x02
				byte(code.OpTrue),   // 0x03
				byte(code.OpArray),  // 0x04
				byte(0),             // 0x05
				byte(4),             // 0x06
				byte(code.OpReturn), // 0x07
			},
			result: "[true, true, true, true]",
			error:  false,
		},
	}

	RunTestCases(tests, []object.Object{}, t)
}

func TestOpBang(t *testing.T) {

	tests := []TestCase{

		// !true -> false
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpBang),   // 0x01
				byte(code.OpReturn), // 0x03
			},
			result: "false",
			error:  false},

		// !false -> true
		{
			program: code.Instructions{
				byte(code.OpFalse),  // 0x00
				byte(code.OpBang),   // 0x01
				byte(code.OpReturn), // 0x03
			},
			result: "true",
			error:  false,
		},

		// !2 -> false
		{
			program: code.Instructions{
				byte(code.OpPush),   // 0x00
				byte(0),             // 0x01
				byte(2),             // 0x02
				byte(code.OpBang),   // 0x03
				byte(code.OpReturn), // 0x04
			},
			result: "false",
			error:  false,
		},

		// !(null) -> true
		{
			program: code.Instructions{
				byte(code.OpLookup), // 0x00
				byte(0),             // 0x01
				byte(0),             // 0x02
				byte(code.OpBang),   // 0x03
				byte(code.OpReturn), // 0x04
			},
			result: "true",
			error:  false,
		},

		// !true -> false
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpBang),   // 0x01
				byte(code.OpReturn), // 0x03
			},
			result: "false",
			error:  false,
		},

		// Test empty stack
		{
			program: code.Instructions{
				byte(code.OpBang),
			},
			result: "Pop from an empty stack",
			error:  true,
		},
	}

	constants := []object.Object{
		&object.String{Value: "Steve"},
	}

	RunTestCases(tests, constants, t)
}

func TestOpCall(t *testing.T) {

	functions := make(map[string]environment.UserFunction)

	// test returns true
	functions["test"] = environment.UserFunction{
		Bytecode: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpReturn), // 0x01
		},
	}

	// bang returns the inverse of the input value
	functions["bang"] = environment.UserFunction{
		Bytecode: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(4), // "input"
			byte(code.OpLocal),
			byte(code.OpLookup),
			byte(0),
			byte(4), // "input"
			byte(code.OpBang),
			byte(code.OpReturn),
		},
		Arguments: []string{"input"},
	}

	// error tries to pop from an empty stack.
	functions["error"] = environment.UserFunction{
		Bytecode: code.Instructions{
			byte(code.OpLocal),
			byte(code.OpReturn),
		},
	}

	tests := []TestCase{

		// empty stack
		{program: code.Instructions{
			byte(code.OpLocal), // 0x00
		}, result: "Pop from an empty stack", error: true},

		// empty stack
		{program: code.Instructions{
			byte(code.OpCall), // 0x00
			byte(0),           // 0x01
			byte(0),           // 0x02 -> val
		}, result: "Pop from an empty stack", error: true},

		// call: len(steve) -> but missing the string
		{
			program: code.Instructions{
				byte(code.OpConstant), // len
				byte(0),
				byte(1),
				byte(code.OpCall),
				byte(0),
				byte(1),
				byte(code.OpReturn),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

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

		// call: user-defined function "Steve" - doesn't exist.
		{
			program: code.Instructions{
				byte(code.OpConstant), // "Steve"
				byte(0),
				byte(0),
				byte(code.OpCall),
				byte(0),
				byte(0),

				byte(code.OpReturn),
			},
			functions: functions,
			result:    "the function Steve does not exist",
			error:     true,
		},
		// call: test()
		{
			program: code.Instructions{
				byte(code.OpConstant), // "test"
				byte(0),
				byte(2),
				byte(code.OpCall),
				byte(0),
				byte(0),

				byte(code.OpReturn),
			},
			functions: functions,
			result:    "true",
			error:     false,
			optimized: code.Instructions{
				byte(code.OpConstant), // "test"
				byte(0),
				byte(2),
				byte(code.OpCall),
				byte(0),
				byte(0),

				byte(code.OpReturn),
			},
		},
		// call: bang()
		{
			program: code.Instructions{
				byte(code.OpConstant), // "bang"
				byte(0),
				byte(3),
				byte(code.OpCall),
				byte(0),
				byte(0),
				byte(code.OpReturn),
			},
			functions: functions,
			result:    "mismatch in argument-counts",
			error:     true,
		},
		// call: error()
		{
			program: code.Instructions{
				byte(code.OpConstant), // "error"
				byte(0),
				byte(5),
				byte(code.OpCall),
				byte(0),
				byte(0),
				byte(code.OpReturn),
			},
			functions: functions,
			result:    "Pop from an empty stack",
			error:     true,
		},
		// call: bang(true)
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpConstant), // "bang"
				byte(0),
				byte(3),
				byte(code.OpCall),
				byte(0),
				byte(1),
				byte(code.OpReturn),
			},
			functions: functions,
			result:    "false",
			error:     false,
		},
		// call: bang(false)
		{
			program: code.Instructions{
				byte(code.OpFalse),
				byte(code.OpConstant), // "bang"
				byte(0),
				byte(3),
				byte(code.OpCall),
				byte(0),
				byte(1),
				byte(code.OpReturn),
			},
			functions: functions,
			result:    "true",
			error:     false,
		},
	}

	// Constants
	constants := []object.Object{&object.String{Value: "Steve"},
		&object.String{Value: "len"},   // global function
		&object.String{Value: "test"},  // user defined fun
		&object.String{Value: "bang"},  // user defined fun
		&object.String{Value: "input"}, // input param to bang()
		&object.String{Value: "error"}, // user defined fun
	}

	RunTestCases(tests, constants, t)
}

func TestOpCase(t *testing.T) {

	tests := []TestCase{
		// stack underflow
		{
			program: code.Instructions{
				byte(code.OpCase), // 0x00
			},
			result: "Pop from an empty stack",
			error:  true,
		},
		// stack underflow
		{
			program: code.Instructions{
				byte(code.OpTrue), // 0x00
				byte(code.OpCase), // 0x01
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// switch("steve") { case "steve" { return true } }

		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02 "steve"
				byte(code.OpConstant), // 0x03
				byte(0),               // 0x04
				byte(0),               // 0x05 "steve"
				byte(code.OpCase),     // 0x06
				byte(code.OpReturn),   // 0x07
			},
			result: "true",
			error:  false,
		},

		// switch("steve") { case /steve/ { return true } }

		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02 "steve"
				byte(code.OpConstant), // 0x03
				byte(0),               // 0x04
				byte(1),               // 0x05 /steve/
				byte(code.OpCase),     // 0x06
				byte(code.OpReturn),   // 0x07
			},
			result: "true",
			error:  false,
		},

		// switch("steve") { case /steve/ { return true }  return false;}

		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(2),               // 0x02 "foo"
				byte(code.OpConstant), // 0x03
				byte(0),               // 0x04
				byte(1),               // 0x05 /steve/
				byte(code.OpCase),     // 0x06
				byte(code.OpReturn),   // 0x07
			},
			result: "false",
			error:  false,
		},

		// switch("steve") { case FALSE { return true }  return false;}

		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02 "steve"
				byte(code.OpConstant), // 0x03
				byte(0),               // 0x04
				byte(3),               // 0x05 FALSE
				byte(code.OpCase),     // 0x06
				byte(code.OpReturn),   // 0x07
			},
			result: "false",
			error:  false,
		},
	}

	constants := []object.Object{
		&object.String{Value: "steve"},
		&object.Regexp{Value: "steve"},
		&object.String{Value: "foo"},
		&object.Boolean{Value: false},
	}

	RunTestCases(tests, constants, t)
}

func TestOpConstant(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02 -> val
				byte(code.OpReturn),
			},
			result: "Steve",
			error:  false,
		},
		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(1),               // 0x02 -> val
				byte(code.OpReturn),
			},
			result: "access to constant which doesn't exist",
			error:  true,
		},
	}

	// Constants
	constants := []object.Object{&object.String{Value: "Steve"}}

	RunTestCases(tests, constants, t)
}

func TestOpDec(t *testing.T) {

	tests := []TestCase{

		// dec on a string
		{
			program: code.Instructions{
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
			},
			result: "doesn't implement the Decrement() interface",
			error:  true,
		},

		// dec on a number
		{
			program: code.Instructions{
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
			},
			result: "8",
			error:  false,
		},

		// dec on a number - but no OpLookup
		{
			program: code.Instructions{
				byte(code.OpPush),     // 0x00
				byte(0),               // 0x01
				byte(9),               // 0x02 -> 9
				byte(code.OpConstant), // 0x03
				byte(0),               // 0x04
				byte(0),               // 0x05 -> key
				byte(code.OpSet),      // 0x06

				// now "key = 9", decrease it
				byte(code.OpDec),
				byte(0),
				byte(0),
				byte(code.OpReturn),

				// That OpLookup causes this.
			},
			result: "Pop from an empty stack",
			error:  true,
		},
		{
			program: code.Instructions{
				byte(code.OpDec), // 0x00
				byte(0),          // 0x01
				byte(2),          // 0x02
				byte(code.OpReturn),
			},
			result: "access to constant which doesn't exist",
			error:  true,
		},
	}

	// Constants
	constants := []object.Object{&object.String{Value: "key"},
		&object.String{Value: "val"}}

	RunTestCases(tests, constants, t)
}

func TestOpHash(t *testing.T) {

	tests := []TestCase{
		// Build a hash with zero args.
		{
			program: code.Instructions{
				byte(code.OpHash),   // 0x00
				byte(0),             // 0x01
				byte(0),             // 0x02
				byte(code.OpReturn), // 0x03
			},
			result: "{}",
			error:  false,
		},

		// stack underflow
		{
			program: code.Instructions{
				byte(code.OpHash),   // 0x00
				byte(0),             // 0x01
				byte(1),             // 0x02
				byte(code.OpReturn), // 0x03
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// stack underflow
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpHash),   // 0x01
				byte(0),             // 0x02
				byte(1),             // 0x03
				byte(code.OpReturn), // 0x04
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// {"foo":"bar"}
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0), // "foo"
				byte(code.OpConstant),
				byte(0),
				byte(1), // "bar"
				byte(code.OpHash),
				byte(0),
				byte(2),
				byte(code.OpReturn),
			},
			result: "{foo: bar}",
			error:  false,
		},

		// return len({"foo":"bar"})
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0), // "foo"
				byte(code.OpConstant),
				byte(0),
				byte(1), // "bar"
				byte(code.OpHash),
				byte(0),
				byte(2),
				byte(code.OpConstant),
				byte(0),
				byte(2), // "len"
				byte(code.OpCall),
				byte(0),
				byte(1),
				byte(code.OpReturn),
			},
			result: "1",
			error:  false,
		},

		// {true: bar} - > invalid key
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpConstant),
				byte(0),
				byte(1), // "bar"
				byte(code.OpHash),
				byte(0),
				byte(2),
				byte(code.OpReturn),
			},
			result: "unusable as hash key",
			error:  true,
		},
	}

	constants := []object.Object{
		&object.String{Value: "foo"},
		&object.String{Value: "bar"},
		&object.String{Value: "len"},
		&object.Boolean{Value: true},
	}

	RunTestCases(tests, constants, t)
}

func TestOpInc(t *testing.T) {

	tests := []TestCase{

		// inc on a string
		{
			program: code.Instructions{
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
			},
			result: "doesn't implement the Increment() interface",
			error:  true,
		},

		// inc on a number
		{
			program: code.Instructions{
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
			},
			result: "10",
			error:  false,
		},

		// inc on a number - with no OpLookup
		{
			program: code.Instructions{
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
			},
			result: "Pop from an empty stack",
			error:  true,
		},
		{
			program: code.Instructions{
				byte(code.OpInc), // 0x00
				byte(0),          // 0x01
				byte(2),          // 0x02
				byte(code.OpReturn),
			},
			result: "access to constant which doesn't exist",
			error:  true,
		},
	}

	// Constants
	constants := []object.Object{&object.String{Value: "key"},
		&object.String{Value: "val"}}

	RunTestCases(tests, constants, t)
}

func TestOpIndex(t *testing.T) {

	tests := []TestCase{

		// empty stack
		{
			program: code.Instructions{
				byte(code.OpIndex),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// stack is too small.
		{
			program: code.Instructions{
				byte(code.OpFalse),
				byte(code.OpIndex),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// 1[1] -> "type error"
		{
			program: code.Instructions{
				byte(code.OpPush), // 0x00
				byte(0),           // 0x01
				byte(0),           // 0x02
				byte(code.OpPush), // 0x03
				byte(0),
				byte(0),
				byte(code.OpIndex),
				byte(code.OpReturn),
			},
			result: "the index operator can only be applied to arrays, hashes, and strings,",
			error:  true,
		},

		// "steve"["steve"] -> "type error"
		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02
				byte(code.OpConstant), // 0x03
				byte(0),
				byte(0),
				byte(code.OpIndex),
				byte(code.OpReturn),
			},
			result: "must be given an integer",
			error:  true,
		},

		// "steve"[1] -> "t"
		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02
				byte(code.OpPush),     // 0x03
				byte(0),
				byte(1),
				byte(code.OpIndex),
				byte(code.OpReturn),
			},
			result: "t",
			error:  false,
		},

		// "steve"[2570] -> "NULL"
		{
			program: code.Instructions{
				byte(code.OpConstant), // 0x00
				byte(0),               // 0x01
				byte(0),               // 0x02
				byte(code.OpPush),     // 0x03
				byte(10),
				byte(10),
				byte(code.OpIndex),
				byte(code.OpReturn),
			},
			result: "null",
			error:  false,
		},

		// create array: index[1]
		{
			program: code.Instructions{
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
			},
			result: "2",
			error:  false,
		},

		// create array: array[255]
		{
			program: code.Instructions{
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
			},
			result: "null",
			error:  false,
		},
	}

	// Constants
	constants := []object.Object{&object.String{Value: "Steve"}}

	RunTestCases(tests, constants, t)
}

func TestOpIterationNext(t *testing.T) {

	tests := []TestCase{

		// OpIterationNext requires three stack entries: give it none
		{
			program: code.Instructions{
				byte(code.OpIterationNext),
			},
			error:  true,
			result: "Pop from an empty stack",
		},

		// OpIterationNext requires three stack entries: give it one
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpIterationNext),
			},
			error:  true,
			result: "Pop from an empty stack",
		},

		// OpIterationNext requires three stack entries: give it two
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpTrue),
				byte(code.OpIterationNext),
			},
			error:  true,
			result: "Pop from an empty stack",
		},

		// OpIterationNext requires three stack entries: give it three,
		// but not an iterable thing.
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpTrue),
				byte(code.OpTrue),
				byte(code.OpIterationNext),
			},
			error:  true,
			result: "object doesn't implement the Iterable interface",
		},
		// iterate over characters in a string
		{
			program: code.Instructions{
				byte(code.OpConstant),       // 0x00
				byte(0),                     // 0x01
				byte(0),                     // 0x02 -> "Steve"
				byte(code.OpIterationReset), // 0x03
				byte(code.OpConstant),       // 0x04 XXXX:
				byte(0),                     // 0x05
				byte(1),                     // 0x06 -> "i"
				byte(code.OpConstant),       // 0x07
				byte(0),                     // 0x08
				byte(2),                     // 0x09 -> "c"
				byte(code.OpIterationNext),  // 0x0A
				byte(code.OpJumpIfFalse),    // 0x0B
				byte(0),                     // 0x0C
				byte(32),                    // 0x0D -> YYYY
				byte(code.OpConstant),       // 0x0E
				byte(0),                     // 0x0F
				byte(3),                     // 0x10 -> "%d: %s\n"
				byte(code.OpLookup),         // 0x11
				byte(0),                     // 0x12
				byte(1),                     // 0x13 -> "lookup: i"
				byte(code.OpLookup),         // 0x14
				byte(0),                     // 0x15
				byte(2),                     // 0x16 -> "lookup: c"
				byte(code.OpConstant),       // 0x17
				byte(0),                     // 0x18
				byte(4),                     // 0x19 -> "printf"
				byte(code.OpCall),           // 0x1A
				byte(0),                     // 0x1B
				byte(3),                     // 0x1C -> call printf with 3 args
				byte(code.OpJump),           // 0x1D
				byte(0),                     // 0x1E
				byte(4),                     // 0x1F -> XXXX
				byte(code.OpTrue),           // 0x20 YYYYY:
				byte(code.OpReturn),         // 0x21
			},
			result: "true",
			error:  false,
		},
	}

	// Constants
	constants := []object.Object{
		&object.String{Value: "Steve"},    // 0x00
		&object.String{Value: "i"},        // 0x01
		&object.String{Value: "c"},        // 0x02
		&object.String{Value: "%d: %s\n"}, // 0x03
		&object.String{Value: "printf"},   // 0x04
	}

	RunTestCases(tests, constants, t)

}

func TestOpIterationReset(t *testing.T) {

	tests := []TestCase{

		// empty stack
		{
			program: code.Instructions{
				byte(code.OpIterationReset),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// something that cannot be iterated over
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpIterationReset),
			},
			result: "object doesn't implement the Iterable interface",
			error:  true,
		},

		// something that can be iterated over
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpIterationReset),
				byte(code.OpReturn),
			},
			result: "Steve",
			error:  false,
		},
	}

	constants := []object.Object{&object.String{Value: "Steve"}}

	RunTestCases(tests, constants, t)
}

func TestOpJumpIfFalse(t *testing.T) {

	tests := []TestCase{

		// empty stack
		{
			program: code.Instructions{
				byte(code.OpJumpIfFalse),
				byte(0),
				byte(0),
			},
			result: "Pop from an empty stack",
			error:  true},

		// true
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpJumpIfFalse),
				byte(0),
				byte(0),
				byte(code.OpTrue),
				byte(code.OpReturn),
			},
			result: "true",
			error:  false,
		},

		// false
		{
			program: code.Instructions{
				byte(code.OpFalse),       // 0x00
				byte(code.OpJumpIfFalse), // 0x01
				byte(0),                  // 0x02
				byte(6),                  // 0x03
				byte(code.OpTrue),        // 0x04
				byte(code.OpReturn),      // 0x05
				byte(code.OpFalse),       // 0x06
				byte(code.OpReturn),      // 0x07
			},
			result: "false",
			error:  false,
		},
		// false: out of bounds
		{
			program: code.Instructions{
				byte(code.OpTrue), // 0x00
				byte(code.OpBang),
				byte(code.OpJumpIfFalse), // 0x01
				byte(0),                  // 0x02
				byte(11),                 // 0x03
				byte(code.OpTrue),        // 0x04
				byte(code.OpReturn),      // 0x05
				byte(code.OpFalse),       // 0x06
				byte(code.OpReturn),      // 0x07
			},
			result: "instruction pointer is out of bounds",
			error:  true,
		},
		// OpJump
		{
			program: code.Instructions{
				byte(code.OpJump),   // 0x01
				byte(0),             // 0x02
				byte(11),            // 0x03
				byte(code.OpReturn), // 0x07
			},
			result: "instruction pointer is out of bounds",
			error:  true,
		},
	}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

// This is mostly a test of our reflection.
//
// We use both "Object" and "Map" for more complete testing.
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

		// lookup bogus
		{program: code.Instructions{
			byte(code.OpLookup),
			byte(0),
			byte(6),
			byte(code.OpReturn),
		}, result: "access to constant which doesn't exist", error: true},
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

	m := make(map[string]interface{})
	m["Name"] = "Steve Kemp"
	m["Array"] = []string{"Bart", "Lisa", "Maggie"}
	m["Time"] = time.Unix(1598613755, 0)
	m["True"] = true
	m["Int"] = 17
	m["Float"] = 3.2

	// These are here to give us coverage on the
	// introspection/reflection - they are not actually used.
	m["MiscInt"] = []int{2, 1, 2}
	m["MiscInt32"] = []int32{2, 1, 2}
	m["MiscInt64"] = []int64{2, 1, 2}
	m["MiscFloat32"] = []float32{2.3, 1.1, 2.3}
	m["MiscFloat64"] = []float64{2.3, 1.1, 2.3}
	m["MiscBool"] = []bool{true, false}
	m["MiscTime"] = []time.Time{time.Now()}

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

		// Run - with the map
		out, err = vm.Run(m)

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

	tests := []TestCase{

		// minus -> empty stack
		{
			program: code.Instructions{
				byte(code.OpMinus),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// minus 10 -> -10
		{
			program: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(10),
				byte(code.OpMinus),
				byte(code.OpReturn),
			},
			result: "-10",
			error:  false,
		},

		// minus 3.1 -> -3.1
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpMinus),
				byte(code.OpReturn),
			},
			result: "-3.1",
			error:  false,
		},

		// minus "string" -> type error
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(1),
				byte(code.OpMinus),
				byte(code.OpReturn),
			},
			result: "unsupported type for negation",
			error:  true,
		},
	}

	// Constants
	constants := []object.Object{&object.Float{Value: 3.1},
		&object.String{Value: "not a number"},
	}

	RunTestCases(tests, constants, t)
}

func TestOpRange(t *testing.T) {

	tests := []TestCase{
		// empty stack
		{
			program: code.Instructions{
				byte(code.OpRange),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// too small stack
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpRange),
			},
			result: "Pop from an empty stack",
			error:  true,
		},

		// range( string, int )
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpPush),
				byte(1),
				byte(1),
				byte(code.OpRange),
			},
			result: "the range must be an integer",
			error:  true,
		},

		// range(int, string)
		{
			program: code.Instructions{
				byte(code.OpPush),
				byte(1),
				byte(1),
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpRange),
			},
			result: "the range must be an integer",
			error:  true,
		},

		// range(string, string)
		{
			program: code.Instructions{
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpConstant),
				byte(0),
				byte(0),
				byte(code.OpRange),
			},
			result: "the range must be an integer",
			error:  true,
		},

		// range(1, 10)
		{
			program: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(1),
				byte(code.OpPush),
				byte(0),
				byte(10),
				byte(code.OpRange),
				byte(code.OpReturn),
			},
			result: "[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]",
			error:  false,
		},

		// range(10,1) -> error
		{
			program: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(10),
				byte(code.OpPush),
				byte(0),
				byte(1),
				byte(code.OpRange),
				byte(code.OpReturn),
			},
			result: "the start of a range must be smaller than the end",
			error:  true,
		},
	}

	constants := []object.Object{&object.String{Value: "hello!"}}

	RunTestCases(tests, constants, t)
}

func TestOpSet(t *testing.T) {

	tests := []TestCase{
		// Set "Steve" -> "Kemp" then lookup the result
		{
			program: code.Instructions{
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
			},
			result: "Kemp",
			error:  false,
		},

		// Empty stack
		{
			program: code.Instructions{
				byte(code.OpSet),    // 0x00
				byte(code.OpReturn), // 0x01,
			}, result: "Pop from an empty stack",
			error: true,
		},

		// Only one stack entry
		{
			program: code.Instructions{
				byte(code.OpTrue),   // 0x00
				byte(code.OpSet),    // 0x01
				byte(code.OpReturn), // 0x02,
			},
			result: "Pop from an empty stack",
			error:  true,
		},
	}

	// A pair of constants
	constants := []object.Object{&object.String{Value: "Steve"},
		&object.String{Value: "Kemp"},
	}

	RunTestCases(tests, constants, t)
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

func TestOpVoid(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpVoid),   // 0x00
				byte(code.OpReturn), // 0x01
			},
			result: "void",
			error:  false,
		},
	}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

// Test constant-comparisons are removed.
func TestOptimizerConstants(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
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
			},
			result: "false",
			error:  false,
			optimized: code.Instructions{
				byte(code.OpJump),
				byte(0),
				byte(3),
				byte(code.OpTrue),  // 0 == 0
				byte(code.OpFalse), // 1 == 0
				byte(code.OpTrue),  // 1 != 0
				byte(code.OpFalse), // 0 != 0
				byte(code.OpReturn)},
		},
	}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

func TestOptimizerEnabled(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpReturn),
				byte(code.OpTrue),
				byte(code.OpReturn),
			},
			result: "true", error: false,
			optimized: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpReturn),
			}}}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

// Test constant-jumps are removed.
// [Special]
// [Optimize]
func TestOptimizerJumps(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
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
			},
			result: "true", error: false,
			optimized: code.Instructions{
				byte(code.OpTrue),
				byte(code.OpReturn),
			}}}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

// Test constant-maths expressions are replaced with their results.
func TestOptimizerMaths(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
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
				byte(code.OpPush),
				byte(0),
				byte(7),
				byte(code.OpReturn)},
			result: "7",
			error:  false,
			optimized: code.Instructions{
				byte(code.OpPush),
				byte(0),
				byte(7),
				byte(code.OpReturn),
			}}}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

// Test OpNops are removed.
func TestOptimizerNops(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(code.OpNop),
				byte(code.OpFalse),
				byte(code.OpNop),
				byte(code.OpReturn)},
			result: "false",
			error:  false,
			optimized: code.Instructions{
				byte(code.OpFalse),
				byte(code.OpReturn),
			}}}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
}

func TestUnknownOpcode(t *testing.T) {

	tests := []TestCase{
		{
			program: code.Instructions{
				byte(200),
			},
			result: "unhandled opcode",
			error:  true,
		},
	}

	constants := []object.Object{}

	RunTestCases(tests, constants, t)
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

// Given an array of test-cases run them.
//
// These run with zero optimizations.
func RunTestCases(tests []TestCase, objects []object.Object, t *testing.T) {

	for _, test := range tests {

		funs := test.functions

		// No functions
		if len(funs) == 0 {
			funs = make(map[string]environment.UserFunction)
		}

		// Default environment
		env := environment.New()

		if len(test.optimized) > 0 {
			env.Set("OPTIMIZE", &object.Boolean{Value: true})
			env.Set("DEBUG", &object.Boolean{Value: true})
		}

		// Create
		vm := New(objects, test.program, funs, env)

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

		if len(test.optimized) > 0 {

			if len(test.optimized) != len(vm.bytecode) {
				t.Fatalf("optimized bytecode had wrong size: %d != %d", len(test.optimized), len(vm.bytecode))
			}
			for i, op := range test.optimized {
				if vm.bytecode[i] != op {
					t.Fatalf("index %d opcode was %s not %s", i,
						code.String(code.Opcode(vm.bytecode[i])), code.String(code.Opcode(op)))
				}
			}
		}
	}

}
