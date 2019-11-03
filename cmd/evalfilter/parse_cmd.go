package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/subcommands"
	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/parser"
)

//
// The options set by our command-line flags: None
//
type parseCmd struct {
}

//
// Glue
//
func (*parseCmd) Name() string     { return "parse" }
func (*parseCmd) Synopsis() string { return "Show our parser output." }
func (*parseCmd) Usage() string {
	return `parse file1 file2 .. [fileN]:
  Show the output from our parser
`
}

//
// Flag setup
//
func (p *parseCmd) SetFlags(f *flag.FlagSet) {
}

// Parse parses the given file, and dumps the AST which
// resulted from it.
func (p *parseCmd) Parse(file string) {

	//
	// Read the file contents.
	//
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s - %s\n", file, err.Error())
		return
	}

	//
	// Create the helper
	//
	parse := parser.New(lexer.New(string(dat)))

	//
	// Parse the program
	//
	program := parse.ParseProgram()

	//
	// Where there any errors produced by the parser?
	//
	// If so report that.
	//
	if len(parse.Errors()) > 0 {
		fmt.Printf("\nErrors parsing script:\n" +
			strings.Join(parse.Errors(), "\n"))
		return
	}

	//
	// Print the parsed program.
	//
	fmt.Printf("%s\n", program.String())
}

//
// Entry-point.
//
func (p *parseCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we've been passed.
	//
	for _, file := range f.Args() {
		p.Parse(file)
	}

	return subcommands.ExitSuccess

}
