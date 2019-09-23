// This example demonstrates setting a variable in the host-application,
// which is then accessed inside the filter-script.

package main

import (
	"fmt"
	"time"

	"github.com/skx/evalfilter"
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
print( "The time is ", $time, "\n" );

// Simple of accessing the variable in a conditional-
if ( $time == "steve" ) {
   print( "\tThat is a surprise..\n");
} else {
   print("\tYay!\n");
}

// We're done
return false;
`)

	//
	// Loop forever.
	//
	for {

		//
		// Set the `time` variable to the current time.
		//
		eval.SetVariable("time", fmt.Sprintf("%v", time.Now().Unix()))

		//
		// Run the script.
		//
		ret, err := eval.Run(nil)
		if err != nil {
			fmt.Printf("Failed to run script: %s\n", err.Error())
			return
		}
		fmt.Printf("Script gave result %v\n", ret)

		// Sleep for a second, and repeat
		time.Sleep(1 * time.Second)

	}
}
