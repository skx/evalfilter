package evalfilter

import "fmt"

// Example is a function which will filter a list of people, to return
// only those members who are above a particular age, via the use of
// a simple script.
func Example() {

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

// Example filter - we only care about people over 30.
if ( Age > 30 ) { return true ; }

// Since we return false the caller will know to ignore people here.
return false;
`

	//
	// Create the evaluator
	//
	eval := New(script)

	//
	// Process each person.
	//
	for _, entry := range people {

		//
		// Call the filter
		//
		res, err := eval.Run(entry)

		//
		// Error-detection is important.
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
	// {Bob 31}
	// {John 42}
}
