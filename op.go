// Operation is the abstract interface any of our operations
// must implement.

package evalfilter

// Operation is the abstract interface all operations much implement.
type Operation interface {

	// Run runs the operation.
	//
	// Return values:
	//   return - If true we're returning
	//
	//   value  - The value we terminate with
	//
	//   error  - An error occurred
	//
	Run(self *Evaluator, obj interface{}) (bool, bool, error)
}
