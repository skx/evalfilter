// This example reads the contents of the file `/etc/passwd`, then
// processes the entries found within it.

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"github.com/skx/evalfilter"
	"github.com/skx/evalfilter/environment"
	"github.com/skx/evalfilter/runtime"
)

//
// Entry-point
//
func main() {

	//
	// List of login-names / user-names we found.
	//
	var Users []string

	//
	// We require a single argument, which is the name of
	// a script to run.
	//
	if len(os.Args[1:]) != 1 {
		fmt.Printf("Usage: passwd script.file\n")
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
	// Add a custom-function, for demonstration purposes.
	//
	eval.AddFunction("dump",
		func(env *environment.Environment, obj interface{}, args []runtime.Argument) interface{} {

			//
			// Show the arguments we received.
			//
			for i, arg := range args {

				fmt.Printf("\tArg %d - %v\n", i, arg.Value(env, obj))
			}
			return 0
		})

	//
	// Process /etc/passwd.
	//
	file, err := os.Open("/etc/passwd")
	if err != nil {
		fmt.Printf("Failed to open /etc/passwd - %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	//
	// Create a reader.
	//
	reader := bufio.NewReader(file)

	//
	// For each line ..
	//
	for {
		//
		// Read a line.
		//
		line, err := reader.ReadString('\n')

		//
		// Error?
		//
		if err != nil {
			if err == io.EOF {
				break
			} else {

				fmt.Printf("Error processing line from passwd file %s\n", err)
				os.Exit(1)
			}
		}

		// get the username and description
		lineSlice := strings.FieldsFunc(line, func(divide rune) bool {
			return divide == ':'
		})

		if len(lineSlice) > 0 {
			Users = append(Users, lineSlice[0])
		}

		// End of file?  Break
	}

	//
	// now we have a list of user-names.
	//
	// Process each one via the user-specified script
	//
	for _, name := range Users {

		//
		// Lookup the details of the user.
		//
		usr, err := user.Lookup(name)
		if err != nil {
			panic(err)
		}

		//
		// Run our script
		//
		ret, err := eval.Run(usr)
		if err != nil {
			fmt.Printf("Failed to run script:%s\n", err.Error())
			return
		}

		//
		// Show the script.
		//
		fmt.Printf("User %s gave result %v\n\n", name, ret)
	}

	//
	// All done
	//
}
