package evalfilter

import (
	"fmt"
)

// ExampleUserFunction demonstrates that defining functions inside
// our scripting language works roughly as you would expect.
//
// There are issues with recursion, but basic/naive operation works
// correctly.
func ExampleUserFunction() {

	//
	// We'll run this script, which defines a function and uses
	// it.
	//
	// Recursion is not 100% supported, so this is a non-recursive
	// implementation of the fibonacci algorithm.
	//
	script := `
//
// non-recursive fibonacci sequence
//
function fibonacci(n) {
 if(n <= 1){
  return n;
 }

 fibo = 1;
 fiboPrev = 1;

 x = 2;
 for( x < n ) {
  temp = fibo;
  fibo = fibo + fiboPrev;
  fiboPrev = temp;
  x++;
 }
 return fibo;
}

// Invoke the function with a bunch of numbers.
foreach i in 1..20 {
   printf("N:%d F:%d\n", i, fibonacci(i));
}

return true;
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
	// N:1 F:1
	// N:2 F:1
	// N:3 F:2
	// N:4 F:3
	// N:5 F:5
	// N:6 F:8
	// N:7 F:13
	// N:8 F:21
	// N:9 F:34
	// N:10 F:55
	// N:11 F:89
	// N:12 F:144
	// N:13 F:233
	// N:14 F:377
	// N:15 F:610
	// N:16 F:987
	// N:17 F:1597
	// N:18 F:2584
	// N:19 F:4181
	// N:20 F:6765
	// true
}
