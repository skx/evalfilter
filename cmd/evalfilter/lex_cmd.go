package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/evalfilter/v2/lexer"
)

//
// The options set by our command-line flags: None
//
type lexCmd struct {
}

//
// Glue
//
func (*lexCmd) Name() string     { return "lex" }
func (*lexCmd) Synopsis() string { return "Show our lexer output." }
func (*lexCmd) Usage() string {
	return `lexer file1 file2 .. [fileN]:
  Show the output from our lexer
`
}

//
// Flag setup
//
func (p *lexCmd) SetFlags(f *flag.FlagSet) {
}

func (p *lexCmd) Lex(file string) {
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
	l := lexer.New(string(dat))

	//
	// Dump the tokens.
	//
	for {
		tok := l.NextToken()
		fmt.Printf("%v\n", tok)
		if tok.Type == "EOF" {
			break
		}
	}
}

//
// Entry-point.
//
func (p *lexCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we've been passed.
	//
	for _, file := range f.Args() {
		p.Lex(file)
	}

	return subcommands.ExitSuccess

}
