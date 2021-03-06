// Package code contains definitions of the bytecode instructions
// our compiler emits, and our virtual machine executes.
package code

import "fmt"

// Opcode is a type-alias.
type Opcode byte

// Instructions is a type alias.
type Instructions []byte

// Opcodes we support
const (

	// Push the value of one of our constant objects onto the stack.
	//
	// The 16-bit argument is the offset of the constant.
	OpConstant Opcode = iota

	// Unconditionally jump to the specified offset.
	//
	// 16-bit argument is the offset to jump to.
	OpJump

	// Pop a value from the stack, and if the value is false then jump
	// to the specified offset.
	//
	// 16-bit argument is the offset to jump to.
	OpJumpIfFalse

	// Call one of our built-in functions.
	//
	// Pop the name from the the stack, then use the 16-bit argument
	// as the number of additional items to pop off the stack.  (i.e
	// the number of arguments to pass to the function.)
	//
	// Once complete push the result of the call back to the stack.
	OpCall

	// Load a variable by name.
	// 16-bit offset to the name to lookup
	//
	// TODO: This could be a single-byte operation, we could
	// pop the name from the stack ..
	OpLookup

	// Push an integer upon the stack
	OpPush

	// Store a literal array.
	OpArray

	// Store a literal hash.
	OpHash

	// NOP
	OpNop

	// NOP - But one that the optimizer won't remove
	OpPlaceholder

	// Set a variable by name
	OpSet

	// Setup a local variable, by name
	OpLocal

	// Push a TRUE value onto the stack.
	OpTrue

	// Push a FALSE value onto the stack.
	OpFalse

	// Push a VOID value onto the stack.
	OpVoid

	// Run a case-comparison
	OpCase

	// Pop two values from the stack, add them, and push the result.
	OpAdd

	// Pop two values from the stack, subtract them, and push the result.
	OpSub

	// Pop two values from the stack, multiply them, and push the result.
	OpMul

	// Pop two values from the stack, divide them, and push the result.
	OpDiv

	// Pop two values from the stack, run a modulus op, push the result.
	OpMod

	// Pop two values from the stack, raise to the power, push the result.
	OpPower

	// Increment the value of the given object
	OpInc

	// Decrement the value of the given object
	OpDec

	// Use  the top value from the stack as a the return value
	// and cease execution.
	OpReturn

	// Pop a value from the the stack, invert, push back.
	OpMinus

	// Pop a value from the the stack, negate, push back.
	OpBang

	// Pop a value from the stack, and push back the square root of it.
	OpSquareRoot

	// Pop two values and push TRUE if the first is less than the second,
	// otherwise push FALSE.
	OpLess

	// Pop two values and push TRUE if the first is less than, or equal
	// to the second, otherwise push FALSE.
	OpLessEqual

	// Pop two values and push TRUE if the first is greater than the second,
	// otherwise push FALSE
	OpGreater

	// Pop two values and push TRUE if the first is greater than, or equal
	// to the second.  Otherwise push TRUE.
	OpGreaterEqual

	// Pop two values from the stack.  If equal push TRUE, else push FALSE.
	OpEqual

	// Pop two values from the stack.  If unequal push TRUE, else push FALSE.
	OpNotEqual

	// Pop two values from the stack, if the first matches the regexp
	// in the second push TRUE, else push FALSE.
	OpMatches

	// Pop two values from the stack, if the first does not match the
	// regexp in the second push TRUE, else push FALSE.
	OpNotMatches

	// Pop two values from the stack.  If both are TRUE push TRUE,
	// otherwise push FALSE.
	OpAnd

	// Pop two values from the stack.  If either is TRUE push TRUE,
	// otherwise push FALSE.
	OpOr

	// String / Array index operation.
	OpIndex

	// Pop two values from the the stack, if the first value is
	// contained in the second-argument (which must be an array),
	// push TRUE, else push FALSE
	OpArrayIn

	// OpIterationReset resets the state of a given object,
	// such that the same object can be iterated upon multiple
	// times.
	OpIterationReset

	// OpIterationNext is used for walking over items in an array.
	//
	// It is a horrible opcode because the interpreter shares
	// knowledge on the back-end with the fake-code generated
	// on the front-end.
	//
	// Assuming an array is on the stack (!):
	//
	//  1. We pop the array FROM the stack.
	//
	//  2. We push the array back, after bumping the count-field.
	//
	//  3. We then push the next item.
	//
	//  4. We then push TRUE - which lets our OpJumpIfFalse work.
	//
	// UNLESS we're at the end, in which case we do:
	//
	//  1.  We pop the value FROM the stack
	//
	//  2.  We push FALSE - breaking out of out body.
	//
	OpIterationNext

	// Given two integer values produce an array holding
	// items between them.
	OpRange
)

// OpCodeNames allows mapping opcodes to their names.
var OpCodeNames = [...]string{
	OpAdd:            "OpAdd",
	OpAnd:            "OpAnd",
	OpArray:          "OpArray",
	OpArrayIn:        "OpArrayIn",
	OpBang:           "OpBang",
	OpCall:           "OpCall",
	OpCase:           "OpCase",
	OpConstant:       "OpConstant",
	OpDec:            "OpDec",
	OpDiv:            "OpDiv",
	OpEqual:          "OpEqual",
	OpFalse:          "OpFalse",
	OpGreater:        "OpGreater",
	OpGreaterEqual:   "OpGreaterEqual",
	OpHash:           "OpHash",
	OpInc:            "OpInc",
	OpIndex:          "OpIndex",
	OpIterationNext:  "OpIterationNext",
	OpIterationReset: "OpIterationReset",
	OpJump:           "OpJump",
	OpJumpIfFalse:    "OpJumpIfFalse",
	OpLess:           "OpLess",
	OpLessEqual:      "OpLessEqual",
	OpLocal:          "OpLocal",
	OpLookup:         "OpLookup",
	OpMatches:        "OpMatches",
	OpMinus:          "OpMinus",
	OpMod:            "OpMod",
	OpMul:            "OpMul",
	OpNop:            "OpNop",
	OpNotEqual:       "OpNotEqual",
	OpNotMatches:     "OpNotMatches",
	OpOr:             "OpOr",
	OpPlaceholder:    "OpPlaceholder",
	OpPower:          "OpPower",
	OpPush:           "OpPush",
	OpRange:          "OpRange",
	OpReturn:         "OpReturn",
	OpSet:            "OpSet",
	OpSquareRoot:     "OpSquareRoot",
	OpSub:            "OpSub",
	OpTrue:           "OpTrue",
	OpVoid:           "OpVoid",
}

// Length returns the length of the given opcode, including any optional
// argument.
//
// Opcodes default to being a single byte, but some require a mandatory
// argument which is currently limited to a single 16-bit / 2-byte
// value.
//
// This means our instructions are either a single 8-bit byte, or
// three such bytes.
//
// This function returns the appropriate length for a given opcode.
func Length(op Opcode) int {

	switch op {
	case OpArray:
		return 3
	case OpHash:
		return 3
	case OpCall:
		return 3
	case OpConstant:
		return 3
	case OpDec:
		return 3
	case OpJump, OpJumpIfFalse:
		return 3
	case OpInc:
		return 3
	case OpLookup:
		return 3
	case OpPush:
		return 3
	}

	return 1
}

// String converts the given opcode to a string.
//
// This is used by our bytecode disassembler/dumper.
func String(op Opcode) string {

	// Sanity-check
	if int(op) >= len(OpCodeNames) {
		fmt.Printf("Warning: Invalid opcode 0x%02X\n", op)
		return "OpUnknown"
	}

	return OpCodeNames[op]
}
