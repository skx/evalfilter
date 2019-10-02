//
// Entry-point to the CLI service.
//

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/google/subcommands"
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

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&lexCmd{}, "")
	subcommands.Register(&parseCmd{}, "")
	subcommands.Register(&runCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
