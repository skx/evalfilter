// This example demonstrates working with `time.Time`.

package main

import (
	"fmt"
	"time"

	"github.com/skx/evalfilter/v2"
)

//
// Entry-point
//
func main() {

	//
	// We'll use an instance of this object to run our script against.
	//
	// Here the important thing is the `At` field.
	//
	type Input struct {
		Message string
		At      time.Time
	}

	//
	// Create an instance of this object.
	//
	in := &Input{Message: "This is a test",
		At: time.Now()}

	//
	// This is the script we'll run.
	//
	input := `

if ( hour(At) < 07 || hour(At) >= 14) {
   print( "This is out of hours ..\n" );
}

print( "The message is '", Message, "'\n" );
print( "The time was '", At, "'\n" );
print( "In a more humane format that is ",
       hour( At ), ":", minute( At), ":", seconds(At) ,
       "\n" );
print( "The message was sent upon ", weekday(At), "\n" );
print( "The date was " , day(At), "/", month(At), "/", year(At), "\n");

// No real decision here.
return true;
`

	//
	// Create an evaluator, with the given script.
	//
	eval := evalfilter.New(input)

	//
	// Prepare the script
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile script: %s\n", err.Error())
		return
	}

	//
	// Run our script, using the instance of the structure.
	//
	ret, err := eval.Run(in)
	if err != nil {
		fmt.Printf("Failed to run script:%s\n", err.Error())
		return
	}

	//
	// Show the output.
	//
	if ret {
		fmt.Printf("Script gave result %v\n\n", ret)
	}

	//
	// All done
	//
}
