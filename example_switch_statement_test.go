package evalfilter

import (
	"fmt"
)

// ExampleSwitchFunction demonstrates how the case/switch statement
// in our scripting language works.
//
func ExampleSwitchFunction() {

	//
	// We'll run this script, which defines a function and uses
	// it.
	//
	//
	script := `

function test( name ) {

  switch( name ) {
    case "Ste" + "ve" {
	printf("\tI know you - expression-match!\n" );
    }
    case "Steven" {
	printf("\tI know you - literal-match!\n" );
    }
    case /steve kemp/i {
        printf("\tI know you - regexp-match!\n");
    }
    default {
	printf("\tDefault: I don't know who you are\n" );
    }
  }
}

printf("Running: Steve\n");
test( "Steve" );

printf("Running: Steven\n");
test( "Steven" );

printf("Running: steve kemp\n");
test( "steve kemp" );

printf("Running: bob\n");
test( "Bob" );

printf("Testing is all over\n")
`

	//
	// Create the evaluator
	//
	eval := New(script)

	//
	// Prepare the evaluator.
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile the code:%s\n", err.Error())
		return
	}

	//
	// Call the filter - since we're testing
	// the user-defined function we don't
	// care about passing a real-object to it.
	//
	res, err := eval.Run(nil)

	//
	// Error-detection is important (!)
	//
	if err != nil {
		panic(err)
	}

	//
	// Show the output of the call.
	//
	if res {
		fmt.Printf("%v\n", res)
	}

	// Output:
	// Running: Steve
	//	I know you - expression-match!
	//Running: Steven
	//	I know you - literal-match!
	//Running: steve kemp
	//	I know you - regexp-match!
	//Running: bob
	//	Default: I don't know who you are
	//Testing is all over
}
