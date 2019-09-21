package evalfilter

import (
	"fmt"

	"github.com/skx/evalfilter/environment"
	"github.com/skx/evalfilter/runtime"
)

// ExampleCustomFunction demonstrates how you can add a custom function
// to your host-application, which is available to filter scripts.
//
// In this example we add the function `length`, which will calculate the
// length of strings, or the contents of fields, and make it available
// to the scripting-environment.
//
// We use this function to output only those users with names greater
// than four characters in length.
//
func ExampleCustomFunction() {

	//
	// This is the structure our script will operate upon.
	//
	type Person struct {
		Name string
		Age  int
	}

	//
	// Here is a list of people.
	//
	people := []Person{
		{"Bob", 31},
		{"John", 42},
		{"Michael", 17},
		{"Jenny", 26},
	}

	//
	// We'll run this script against each entry in the list
	//
	script := `
//
// Example filter - we only care about people with "long" names.
//
if ( length(Name) > 4 ) { return true ; }

// Since we return false the caller will know to ignore people here.
return false;
`

	//
	// Create the evaluator
	//
	eval := New(script)

	//
	// Helper function to calculate the length of a string.
	//
	// Note that we receive a variable number of arguments, for
	// simplicity we only calculate the length of the first.
	//
	// Also note that the function `len` does this job, and is
	// built-in and already available.
	//
	// This is just an example :)
	//
	eval.AddFunction("length",
		func(env *environment.Environment, obj interface{}, args []runtime.Argument) interface{} {

			//
			// Loop over the arguments
			//
			for _, arg := range args {

				//
				// Get the first argument.
				//
				val := arg.Value(env, obj)

				//
				// Now as a string.
				//
				str := fmt.Sprintf("%v", val)

				//
				// Return the length
				//
				return len(str)
			}
			return 0
		})

	//
	// Process each person.
	//
	for _, entry := range people {

		//
		// Call the filter
		//
		res, err := eval.Run(entry)

		//
		// Error-detection is important (!)
		//
		if err != nil {
			panic(err)
		}

		//
		// We only care about the people for whom the filter
		// returned `true`.
		//
		if res {
			fmt.Printf("%v\n", entry)
		}
	}

	// Output:
	// {Michael 17}
	// {Jenny 26}
}
