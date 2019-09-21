// This file contains the implementation for the print operation

package runtime

import (
	"fmt"

	"github.com/skx/evalfilter/environment"
)

// PrintOperation holds state for the `print` operation.
type PrintOperation struct {
	// Values are the various values to be printed.
	Values []Argument
}

// Run runs the print operation.
func (p *PrintOperation) Run(env *environment.Environment, obj interface{}) (bool, bool, error) {
	for _, val := range p.Values {
		fmt.Printf("%v", val.Value(env, obj))
	}
	return false, false, nil
}
