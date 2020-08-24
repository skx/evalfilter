package environment

import (
	"github.com/skx/evalfilter/v2/code"
)

// UserFunction holds details about a single user-defined function.
//
// A user-defined function
type UserFunction struct {

	// The names of the arguments which are passed to the
	// function, if any.
	Arguments []string

	// The function will be compiled into a set of bytecode
	// instructions which will be stored here.
	Bytecode code.Instructions
}
