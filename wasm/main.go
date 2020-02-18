// Simple example script which will work as a WASM binary.

package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/skx/evalfilter/v2"
	"github.com/skx/evalfilter/v2/object"
)

// Replace the value of the given field with the specified text.
func out(i js.Value, val string) {
	js.Global().Get("document").Call("getElementById", i.String()).Set("value", val)
}

// Append text to the current output of the given field.
func append(i js.Value, val string) {

	cur := js.Global().Get("document").Call("getElementById", i.String()).Get("value").String()
	cur += val

	js.Global().Get("document").Call("getElementById", i.String()).Set("value", cur)
}

// run takes the script in 0 and outputs the result to 1
func run(this js.Value, i []js.Value) interface{} {

	// empty the output
	out(i[1], "")

	// Get the input
	in := js.Global().Get("document").Call("getElementById", i[0].String()).Get("value").String()

	// create the environment.
	eval := evalfilter.New(string(in))

	// prepare the script
	err := eval.Prepare()
	if err != nil {
		out(i[1], "Error compiling:"+err.Error())
		return nil
	}

	// ensure that print works
	eval.AddFunction("print",
		func(args []object.Object) object.Object {
			for _, e := range args {
				append(i[1], fmt.Sprintf("%s", e.Inspect()))
			}
			return &object.Void{}
		})

	// call the script
	ret, err := eval.Execute(nil)
	if err != nil {
		out(i[1], "Error running:"+err.Error())
		return nil
	}

	// Show the text
	txt := fmt.Sprintf("Script result was '%s' (type %s) which is '%t'.\n",
		ret.Inspect(), strings.ToLower(fmt.Sprintf("%s", ret.Type())), ret.True())
	append(i[1], txt)
	return nil
}

func registerCallbacks() {
	js.Global().Set("run", js.FuncOf(run))
}

func main() {
	c := make(chan struct{}, 0)

	println("WASM Go Initialized")
	// register functions
	registerCallbacks()
	<-c
}
