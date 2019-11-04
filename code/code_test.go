package code

import (
	"strings"
	"testing"
)

func TestOpcodes(t *testing.T) {

	var i Opcode

	for i <= OpFinal {
		x := String(i)
		if !strings.HasPrefix(x, "Op") {
			t.Fatalf("opcode doesn't have a good prefix:%s", x)
		}

		i++
	}
}
