package vm

import (
	"context"
	"strings"
	"testing"

	"github.com/skx/evalfilter/v2/code"
	"github.com/skx/evalfilter/v2/environment"
	"github.com/skx/evalfilter/v2/object"
)

// TestDivideByZero is testing a division by zero in the optimizer
// NOT at runtime.
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

func TestOpBang(t *testing.T) {

	type TestCase struct {
		program code.Instructions
		result  string
	}

	tests := []TestCase{

		// !true -> false
		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "false"},

		// !false -> true
		{program: code.Instructions{
			byte(code.OpFalse),  // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "true"},

		// !2 -> false
		{program: code.Instructions{
			byte(code.OpPush),   // 0x00
			byte(0),             // 0x01
			byte(2),             // 0x02
			byte(code.OpBang),   // 0x03
			byte(code.OpReturn), // 0x04
		}, result: "false"},

		// !(null) -> true
		{program: code.Instructions{
			byte(code.OpLookup), // 0x00
			byte(0),             // 0x01
			byte(0),             // 0x02
			byte(code.OpBang),   // 0x03
			byte(code.OpReturn), // 0x04
		}, result: "true"},

		{program: code.Instructions{
			byte(code.OpTrue),   // 0x00
			byte(code.OpBang),   // 0x01
			byte(code.OpReturn), // 0x03
		}, result: "false"},

		// TODO: Test empty stack
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
		if err != nil {
			t.Fatalf("expected no error, got :%s\n", err.Error())
		}

		// Result
		if out.Inspect() != test.result {
			t.Errorf("program has wrong result: %v", out)
		}
	}
}

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

func TestOpSet(t *testing.T) {
	// A pair of constants
	constants := []object.Object{&object.String{Value: "Steve"},
		&object.String{Value: "Kemp"},
	}

	// The program we run:
	bytecode := code.Instructions{
		byte(code.OpConstant), // 0x00
		byte(0),               // 0x01
		byte(1),               // 0x02 -> Steve
		byte(code.OpConstant), // 0x03
		byte(0),               // 0x04
		byte(0),               // 0x05 -> Kemp
		byte(code.OpSet),      // 0x06
		byte(code.OpTrue),     // 0x07
		byte(code.OpReturn),   // 0x08
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
	if out.Inspect() != "true" {
		t.Errorf("program has wrong result: %v", out)
	}

	// The variable "Steve" should now exist
	s, res := vm.environment.Get("Steve")
	if !res {
		t.Fatalf("Failed to lookup variable:%v", s)
	}
	if s.Inspect() != "Kemp" {
		t.Fatalf("Expected variable has the wrong value")
	}
}

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
