// Operation is the abstract interface any of our operations
// must implement.

package runtime

import "github.com/skx/evalfilter/environment"

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
	Run(env *environment.Environment, obj interface{}) (bool, bool, error)
}
