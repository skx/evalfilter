package code

import (
	"strings"
	"testing"
)

func TestOpcodes(t *testing.T) {

	var i Opcode

	//
	// For each OpCode byte-value and name
	//
	for k, v := range OpCodeNames {

		// Stringify and check it looks sane
		x := String(Opcode(k))
		if !strings.HasPrefix(x, "Op") {
			t.Fatalf("opcode doesn't have a good prefix:%s", x)
		}
		if x != v {
			t.Fatalf("Stringifying didn't result in a good value: %s != %s", x, v)
		}

		// Lengths here should only
		l := Length(Opcode(k))

		switch l {
		case 1:
			// nop
		case 3:
			c := Opcode(k)
			if c != OpArray &&
				c != OpCall &&
				c != OpConstant &&
				c != OpJump &&
				c != OpJumpIfFalse &&
				c != OpLookup &&
				c != OpInc &&
				c != OpDec &&
				c != OpPush {

				t.Errorf("found opcode which requires an argument %s", x)
			}
		default:
			t.Errorf("unexpected opcode length %d %s", l, x)
		}
		i++
	}
}

// TestInvalidCode ensures that an unknown code is identified as such.
func TestInvalidCode(t *testing.T) {

	name := String(Opcode(244))
	if name != "OpUnknown" {
		t.Fatalf("unknown opcodes returned something unexpected:%s", name)
	}
}
