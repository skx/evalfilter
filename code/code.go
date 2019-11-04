// Package code contains definitions of the bytecode instruction.
//
// The instructions are used in two different ways, first of all the
// compiler will generate them as it walks the AST which resulted from
// parsing the users' program.  Secondly the virtual machine itself
// will intepret those instructions.
package code

import "encoding/binary"

// Instructions is a type alias.
type Instructions []byte

// Opcode is a type-alias.
type Opcode byte

// Opcodes we support
const (

	// load a constant object; number, string, bool.
	// 16-bit argument is the offset of the constant
	OpConstant Opcode = iota

	// Jump for IF implementation
	// 16-bit argument is the offset to jump to
	OpJump
	OpJumpIfFalse

	// Function call.
	// 16-bit offset is the number of arguments used
	OpCall

	// Load a variable by name
	// 16-bit offset to the name to lookup
	OpLookup

	//
	// Everything higher than this opcode has no arguments.
	//
	// This is a fake opcode.
	//
	OpCodeSingleArg

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

	// Last opcode - NOP
	OpFinal
)

// ReadUint16 Return a 16-bit number from the stream.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// String converts the given opcode to a string.
// This is useful for diagnostics.
func String(op Opcode) string {

	switch op {
	case OpConstant:
		return "OpConstant"
	case OpJump:
		return "OpJump"
	case OpJumpIfFalse:
		return "OpJumpIfFalse"
	case OpCall:
		return "OpCall"
	case OpLookup:
		return "OpLookup"
	case OpCodeSingleArg:
		return "OpCodeSingleArg"
	case OpSet:
		return "OpSet"
	case OpTrue:
		return "OpTrue"
	case OpFalse:
		return "OpFalse"
	case OpAdd:
		return "OpAdd"
	case OpSub:
		return "OpSub"
	case OpMul:
		return "OpMul"
	case OpDiv:
		return "OpDiv"
	case OpMod:
		return "OpMod"
	case OpPower:
		return "OpPower"
	case OpReturn:
		return "OpReturn"
	case OpMinus:
		return "OpMinus"
	case OpBang:
		return "OpBang"
	case OpRoot:
		return "OpRoot"
	case OpLess:
		return "OpLess"
	case OpLessEqual:
		return "OpLessEqual"
	case OpGreater:
		return "OpGreater"
	case OpGreaterEqual:
		return "OpGreaterEqual"
	case OpEqual:
		return "OpEqual"
	case OpNotEqual:
		return "OpNotEqual"
	case OpMatches:
		return "OpMatches"
	case OpNotMatches:
		return "OpNotMatches"
	case OpAnd:
		return "OpAnd"
	case OpOr:
		return "OpOr"
	default:
		return "OpUnknown"
	}

}
