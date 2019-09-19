// This file contains the implementation for the print operation

package evalfilter

import "fmt"

// PrintOperation holds state for the `print` operation.
type PrintOperation struct {
	// Values are the various values to be printed.
	Values []Argument
}

// Run runs the print operation.
func (p *PrintOperation) Run(e *Evaluator, obj interface{}) (bool, bool, error) {
	for _, val := range p.Values {
		fmt.Printf("%v", val.Value(e, obj))
	}
	return false, false, nil
}
