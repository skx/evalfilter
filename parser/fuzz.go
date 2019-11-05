// +build gofuzz

//
// This file is only used for fuzzing the parser,
// which will detect hangs, infinite loops & etc.
//

package parser

import "github.com/skx/evalfilter/v2/lexer"

// Fuzz is the function that our fuzzer-application uses.
// See `FUZZING.md` in our distribution for how to invoke it.
func Fuzz(data []byte) int {

	l := lexer.New(string(data))
	p := New(l)

	p.ParseProgram()
	return 1

}
