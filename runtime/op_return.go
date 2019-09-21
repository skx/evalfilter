// This file contains the implementation for the return operation

package runtime

import "github.com/skx/evalfilter/environment"

// ReturnOperation holds state for the `return` operation
type ReturnOperation struct {
	// Value holds the value which will be returned.
	Value bool
}

// Run handles the return operation.
func (r *ReturnOperation) Run(env *environment.Environment, obj interface{}) (bool, bool, error) {
	return true, r.Value, nil
}
