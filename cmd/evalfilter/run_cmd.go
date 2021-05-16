package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/skx/evalfilter/v2"
	"github.com/skx/evalfilter/v2/object"
)

// Structure for our options and state.
type runCmd struct {

	// Show execution as it happens
	debug bool

	// Disable the bytecode optimizer
	raw bool

	// The user may specify a JSON file.
	jsonFile string

	// Maximum execution duration for the script.
	timeout time.Duration
}

// Info returns the name of this subcommand.
func (r *runCmd) Info() (string, string) {
	return "run", `Run a script file, against a JSON object.

This sub-command allows executing the specified evalfilter-script,
optionally you may specify a JSON object to run the script against.

Example:

  $ evalfilter run script.in
  $ evalfilter run -json /path/to/obj.json script.in

`
}

// Arguments adds per-command args to the object.
func (r *runCmd) Arguments(f *flag.FlagSet) {
	f.StringVar(&r.jsonFile, "json", "", "Run the script with the object contained within the specified JSON file as input.")
	f.BoolVar(&r.raw, "no-optimizer", false, "Disable the bytecode optimizer.")
	f.BoolVar(&r.debug, "debug", false, "Show instructions and the stack at ever step.")
	f.DurationVar(&r.timeout, "timeout", 0, "Specify the maximum execution time to allow for the script(s).")
}

// Run the given script.
func (r *runCmd) Run(file string) {

	//
	// The thing the script will run against.
	//
	obj := make(map[string]interface{})

	//
	// If we have a JSON file then populate our object.
	//
	if r.jsonFile != "" {

		//
		// Read the file contents.
		//
		dat, err := ioutil.ReadFile(r.jsonFile)
		if err != nil {
			fmt.Printf("Error reading file %s - %s\n", r.jsonFile, err.Error())
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
	// Read the script contents.
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
	// If we've been given a timeout period then set it here.
	//
	if r.timeout != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()
		eval.SetContext(ctx)
	}

	//
	// Flags to pass to the preparation function.
	//
	var flags []byte
	if r.raw {
		flags = append(flags, evalfilter.NoOptimize)
	}

	//
	// If we're to debug then set the appropriate variable
	//
	// NOTE: This must be done before `prepare` is invoked.
	//
	if r.debug {
		eval.SetVariable("DEBUG", &object.Boolean{Value: true})
	}

	//
	// Prepare
	//
	err = eval.Prepare(flags)
	if err != nil {
		fmt.Printf("Error compiling:%s\n", err.Error())
		return
	}

	//
	// Run the script.
	//
	ret, err := eval.Execute(obj)
	if err != nil {
		fmt.Printf("Failed to run script: %s\n", err.Error())
		return
	}

	//
	// Show the actual, literal, return-value, as well as the
	// truthiness of the result.
	//
	fmt.Printf("Script gave result type:%s value:%s - which is '%t'.\n",
		ret.Type(), ret.Inspect(), ret.True())

	// Now show as JSON, if we can.
	helper, ok := ret.(object.JSONAble)
	if ok {

		j, err := helper.JSON()
		if err != nil {
			fmt.Printf("Error converting result to JSON\n")
			return
		}

		fmt.Printf("JSON Result:\n\t%s\n", j)
	}

}

// Execute is invoked if the user specifies `run` as the subcommand.
func (r *runCmd) Execute(args []string) int {

	//
	// For each file we've been passed; run it.
	//
	for _, file := range args {
		r.Run(file)
	}

	return 0

}
