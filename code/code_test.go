package code

import (
	"strings"
	"testing"
)

func TestOpcodes(t *testing.T) {

	var i Opcode

	for i <= OpFinal {

        // Stringify and check it looks sane
		x := String(i)
		if !strings.HasPrefix(x, "Op") {
			t.Fatalf("opcode doesn't have a good prefix:%s", x)
		}

        // Opcode length
        if ( i < OpCodeSingleArg ) {
            if ( 3 != Length(i) ) {
                t.Fatalf("Invalid length of opcode %s", x)
            }
        } else {
            if ( 1 != Length(i) ) {
                t.Fatalf("Invalid length of opcode %s", x)
            }

        }

		i++
	}
}
