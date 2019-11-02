// Package code contains the instructions that we use.
//
// The instructions are used in two different ways, first of all the
// compiler will generate them as it walks the AST which resulted from
// parsing the users' program.  Secondly the virtual machine itself
// will intepret instructions.
package code

import (
	"encoding/binary"
)

// Instructions is a type alias.
type Instructions []byte

// Opcode is a type-alias.
type Opcode byte

// Opcodes we support
const (

	// load a constant object; number, string, bool.
	OpConstant Opcode = iota

	// Load a variable by name
	OpLookup

	// Set a variable by name
	OpSet

	// Load true/false value.
	OpTrue
	OpFalse

	// Maths operations
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPower

	// Return from script.
	OpReturn

	// Prefix operations.
	OpMinus
	OpBang
	OpRoot

	// Comparison operations
	OpLess
	OpLessEqual
	OpGreater
	OpGreaterEqual

	// Equality
	OpEqual
	OpNotEqual

	// Regexp (string) match
	OpMatches
	OpNotMatches

	// Logical
	OpAnd
	OpOr

	// Jump for IF implementation
	OpJump
	OpJumpIfFalse

	// Function call
	OpCall
)

// Definition holds an opcode definition.
type Definition struct {
	// Name is the name of the opcode.
	Name string

	// OperandWidths holds the number of arguments it has.
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{

	// Lookup a constant
	OpConstant: {"OpConstant", []int{2}},

	// Lookup a variable/field, by name
	OpLookup: {"OpLookup", []int{2}},

	// Set a variable, by name
	OpSet: {"OpSet", []int{}},

	// Store a boolean value
	OpTrue:  {"OpTrue", []int{}},
	OpFalse: {"OpFalse", []int{}},

	// Maths
	OpAdd:   {"OpAdd", []int{}},
	OpSub:   {"OpSub", []int{}},
	OpMul:   {"OpMul", []int{}},
	OpDiv:   {"OpDiv", []int{}},
	OpMod:   {"OpMod", []int{}},
	OpPower: {"OpPower", []int{}},

	// Return from script
	OpReturn: {"OpReturn", []int{}},

	// Prefix-operations
	OpMinus: {"OpMinus", []int{}},
	OpRoot:  {"OpRoot", []int{}},
	OpBang:  {"OpBang", []int{}},

	// Comparisons
	OpLess:         {"OpLess", []int{}},
	OpLessEqual:    {"OpLessEqual", []int{}},
	OpGreater:      {"OpGreater", []int{}},
	OpGreaterEqual: {"OpGreaterEqual", []int{}},
	OpEqual:        {"OpEqual", []int{}},
	OpNotEqual:     {"OpNotEqual", []int{}},
	OpMatches:      {"OpMatches", []int{}},
	OpNotMatches:   {"OpNotMatches", []int{}},

	// Logical
	OpAnd: {"OpAnd", []int{}},
	OpOr:  {"OpOr", []int{}},

	// For IF
	OpJump:        {"OpJump", []int{2}},
	OpJumpIfFalse: {"OpJumpIfFalse", []int{2}},

	OpCall: {"OpCall", []int{2}},
}

// Make generates an opcode.
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadUint16 Return a 16-bit number from the stream.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
