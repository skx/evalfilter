// This example reads the contents of the file `/etc/passwd`, then
// processes the entries found within it.

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/skx/evalfilter"
)

//
// ObjectInstance is what we'll run the script against.
//
type ObjectInstance struct {
	Name string
	Age  int
}

//
// Entry-point
//
func main() {

	v := &ObjectInstance{Name: "Steve Kemp", Age: 100}

	//
	// We require a single argument, which is the name of
	// a script to parse.
	//
	if len(os.Args[1:]) != 1 {
		fmt.Printf("Usage: %s script.file\n", os.Args[0])
		return
	}

	//
	// Load the contents of the given file.
	//
	fmt.Printf("Loading %s\n", os.Args[1])
	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to load %s %s\n", os.Args[1], err)
		return
	}

	//
	// Create an evaluator, with the script inside it.
	//
	eval := evalfilter.New(string(content))

	//
	// Run our script
	//
	ret, err := eval.Run(v)
	if err != nil {
		fmt.Printf("Failed to run script:%s\n", err.Error())
		return
	}

	//
	// Show the result
	//
	if ret {
		fmt.Printf("Script result %v\n\n", ret)
	}

	//
	// All done
	//
}
