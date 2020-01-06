// This example demonstrates setting a variable in the host-application,
//
// which is then accessed inside the filter-script.
//
// The host application will also retrieve a variable, to prove that works.
//

package main

import (
	"fmt"
	"time"

	"github.com/skx/evalfilter/v2"
	"github.com/skx/evalfilter/v2/object"
)

//
// Entry-point
//
func main() {

	//
	// Parse the given script
	//
	eval := evalfilter.New(`

// Show the time.
print( "The variable the script received was ", var, "\n" );

// Simple of accessing the variable in a conditional-
if ( var >= 20 ) {
   print( "\tWe've run a long time..\n");
} else {
   print("\tWe'll keep going until we hit 20 iterations.\n");
}

// Set a variable.
new = var * 4;

// We're done
return false;
`)

	//
	// Prepare the script
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile script: %s\n", err.Error())
		return
	}

	//
	// Counter variable we'll pass to the script.
	//
	count := 0

	//
	// Loop forever.
	//
	for {

		//
		// Set the `time` variable to the current time.
		//
		eval.SetVariable("var", &object.Integer{Value: int64(count)})

		//
		// Run the script.
		//
		ret, err := eval.Run(nil)
		if err != nil {
			fmt.Printf("Failed to run script: %s\n", err.Error())
			return
		}
		fmt.Printf("\tScript gave result %v\n", ret)

		// Get the variable which the script set
		set := eval.GetVariable("new")
		if set.Type() == object.INTEGER {
			fmt.Printf("\tThe script set the variable 'new' to %d\n", set.(*object.Integer).Value)
		} else {
			fmt.Printf("\tRetrieved variable of surprising value: %s\n", set.Inspect())
		}

		// Sleep for a second, and repeat
		time.Sleep(1 * time.Second)

		// Bump
		count++

		fmt.Printf("\n\n")
	}
}
