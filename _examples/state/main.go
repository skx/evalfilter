// This example demonstrates that we can keep state persistent
// across runs.
//

package main

import (
	"fmt"

	"github.com/skx/evalfilter"
	"github.com/skx/evalfilter/object"
)

//
// Entry-point
//
func main() {

	//
	// Parse the given script
	//
	eval := evalfilter.New(`

//
// This script will be called and will return 'false' the
// first ten times it is executed.
//
// After that it will return true.
//
// The reason this works is because we use the same evalfilter
// instance each time we launch it, so variables set persist
// between runs.
//

if ( ! count ) {
  count = 0;
} else {
  count = count + 1;
}

//
// Set a variable which we'll fetch back from our host application,
// via 'eval.GetVariable()'.
//
loop = count;

//
// If we've been invoked this many times we will return 'true'.
//
if ( count >= 10 )  {
  return true;
}

//
// Otherwise we're done.
//
return false;
`)

	//
	// Loop 20 times
	//
	i := 0
	for i < 20 {

		//
		// Run the script.
		//
		ret, err := eval.Run(nil)
		if err != nil {
			fmt.Printf("Failed to run script: %s\n", err.Error())
			return
		}

		//
		// Get the value of the variable `loop` which has
		// been set by the script.
		//
		loop := eval.GetVariable("loop")

		fmt.Printf("%d -> %v [loop:%d]\n", i, ret, loop.(*object.Integer).Value)

		i++
	}
}
