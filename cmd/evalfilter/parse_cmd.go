package main

import (
	"fmt"
	"io/ioutil"

	"github.com/skx/evalfilter/v2/lexer"
	"github.com/skx/evalfilter/v2/parser"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type parseCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info returns the name of this subcommand.
func (p *parseCmd) Info() (string, string) {
	return "parse", `Show the parser output for a given script.

This sub-command allows you to see how a given input-script is
parsed, which can be useful if you're receiving syntax-errors.

Example:

  $ evalfilter parse script.in
`
}

// Parse parses the given file, and dumps the AST which resulted from it.
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
	program, err := parse.Parse()

	//
	// Where there any errors produced by the parser?
	//
	// If so report that.
	//
	if err != nil {
		fmt.Printf("Error parsing script: %s\n", err.Error())
		return
	}

	//
	// Print the parsed program.
	//
	fmt.Printf("%s\n", program.String())
}

// Execute is invoked if the user specifies `lex` as the subcommand.
func (p *parseCmd) Execute(args []string) int {

	//
	// For each file we've been passed.
	//
	for _, file := range args {
		p.Parse(file)
	}

	return 0

}
