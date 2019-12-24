// +build gofuzz

package fuzz

import "github.com/skx/evalfilter/v2"

func Fuzz(data []byte) int {

	//
	// Create the helper
	//
	eval := evalfilter.New(string(data))

	//
	// Parse the program.
	//
	// Create the bytecode.
	//
	// Optimize it.
	//
	eval.Prepare()

	return 0
}
