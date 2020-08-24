//
// Entry-point to the CLI service.
//

package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/skx/subcommands"
)

//
// Setup our sub-commands and use them.
//
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic at the disco: \n" + string(debug.Stack()))
		}
	}()

	subcommands.Register(&lexCmd{})
	subcommands.Register(&bytecodeCmd{})
	subcommands.Register(&parseCmd{})
	subcommands.Register(&runCmd{})

	os.Exit(subcommands.Execute())
}
