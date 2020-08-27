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
