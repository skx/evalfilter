package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/google/subcommands"
	"github.com/skx/evalfilter"
)

//
// The options set by our command-line flags.  A json file
//
type runCmd struct {

	// The user may specify a JSON file.
	jsonFile string
}

//
// Glue
//
func (*runCmd) Name() string     { return "run" }
func (*runCmd) Synopsis() string { return "Run a script file, against a JSON object." }
func (*runCmd) Usage() string {
	return `run -json=x.json script1 script2 .. [scriptN]:
  Run the given file(s), using the object in the JSON-file as input.
`
}

//
// Flag setup
//
func (p *runCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.jsonFile, "json", "", "The JSON file, containing the object to test the script with.")

}

//
// Run the given script.
//
func (p *runCmd) Run(file string) {

	obj := make(map[string]interface{})
	//
	// If we have a JSON file then populate our object.
	//
	if p.jsonFile != "" {

		//
		// Read the file contents.
		//
		dat, err := ioutil.ReadFile(p.jsonFile)
		if err != nil {
			fmt.Printf("Error reading file %s - %s\n", p.jsonFile, err.Error())
			return
		}

		//
		// Parse the JSON
		//
		err = json.Unmarshal(dat, &obj)
		if err != nil {
			fmt.Printf("Error parsing JSON %s\n", err.Error())
			return
		}
	}

	//
	// Read the file contents.
	//
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading file %s - %s\n", file, err.Error())
		return
	}

	//
	// Create the evaluator.
	//
	eval := evalfilter.New(string(dat))

	//
	// Run the script.
	//
	ret, err := eval.Run(obj)
	if err != nil {
		fmt.Printf("Failed to run script: %s\n", err.Error())
		return
	}

	fmt.Printf("Script gave result %v\n", ret)
}

//
// Entry-point.
//
func (p *runCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	//
	// For each file we've been passed; run it.
	//
	for _, file := range f.Args() {
		p.Run(file)
	}

	return subcommands.ExitSuccess

}
