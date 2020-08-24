package main

import (
	"fmt"
	"io/ioutil"

	"github.com/skx/evalfilter/v2/lexer"
	"github.com/skx/subcommands"
)

// Structure for our options and state.
type lexCmd struct {

	// We embed the NoFlags option, because we accept no command-line flags.
	subcommands.NoFlags
}

// Info returns the name of this subcommand.
func (l *lexCmd) Info() (string, string) {
	return "lex", `Show the lexer output for a given script.

This sub-command allows you to see how a given input-script is
split into tokens by our lexer.

Example:

  $ evalfilter run script.in
`
}

// Lex actually lexes the specified file, and shows the tokens that
// were produced.
func (l *lexCmd) Lex(file string) {
	//

	// Read the file contents.
	//
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s - %s\n", file, err.Error())
		return
	}

	//
	// Create a lexer object with those contents.
	//
	lex := lexer.New(string(dat))

	//
	// Dump the tokens.
	//
	for {
		tok := lex.NextToken()
		fmt.Printf("%v\n", tok)
		if tok.Type == "EOF" || tok.Type == "ILLEGAL" {
			break
		}
	}
}

// Execute is invoked if the user specifies `lex` as the subcommand.
func (l *lexCmd) Execute(args []string) int {

	//
	// For each file we've been passed.
	//
	for _, file := range args {

		l.Lex(file)
	}

	return 0

}
