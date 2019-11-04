// eval_functions.go contains our in-built functions.

package evalfilter

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/skx/evalfilter/object"
)

// regCache is a cache of compiled regular expression objects.
// These may persist between runs because a regular expression object
// is essentially constant.
var regCache map[string]*regexp.Regexp

// init ensures that our regexp cache is populated
func init() {
	regCache = make(map[string]*regexp.Regexp)
}

// fnLen is the implementation of our `len` function.
func fnLen(args []object.Object) object.Object {
	sum := 0

	for _, e := range args {
		switch e := e.(type) {
		case *object.String:
			sum += utf8.RuneCountInString(e.Value)
		}
	}
	return &object.Integer{Value: int64(sum)}
}

// fnLower is the implementation of our `lower` function.
func fnLower(args []object.Object) object.Object {

	out := ""

	// Join all input arguments
	for _, arg := range args {
		val := fmt.Sprintf("%v", arg.Inspect())

		out += strings.ToLower(val)
	}
	return &object.String{Value: out}
}

// fnMatch is the implementation of our regex `match` function.
func fnMatch(args []object.Object) object.Object {
	// We expect two arguments
	if len(args) != 2 {
		return &object.Boolean{Value: false}
	}

	str := args[0].Inspect()
	reg := args[1].Inspect()

	// Look for the compiled regular-expression object in our cache.
	r, ok := regCache[reg]
	if !ok {

		// OK it wasn't found, so compile it.
		var err error
		r, err = regexp.Compile(reg)

		// Ensure it compiled
		if err != nil {
			fmt.Printf("Invalid regular expression %s %s", reg, err.Error())
			return &object.Boolean{Value: false}
		}

		// store in the cache for next time
		regCache[reg] = r
	}

	// Split the input by newline.
	for _, s := range strings.Split(str, "\n") {

		// Strip leading-trailing whitespace
		s = strings.TrimSpace(s)

		// Test if it matched
		if r.MatchString(s) {
			return &object.Boolean{Value: true}
		}
	}
	return &object.Boolean{Value: false}
}

// fnTrim is the implementation of our `trim` function.
func fnTrim(args []object.Object) object.Object {
	str := ""
	for _, e := range args {
		str += fmt.Sprintf("%v", (e.Inspect()))
	}
	return &object.String{Value: strings.TrimSpace(str)}
}

// fnType is the implementation of our `type` function.
func fnType(args []object.Object) object.Object {
	for _, e := range args {
		return &object.String{Value: strings.ToLower(fmt.Sprintf("%v", e.Type()))}
	}

	return &object.String{Value: "null"}
}

// fnPrint is the implementation of our `print` function.
func fnPrint(args []object.Object) object.Object {
	for _, e := range args {
		fmt.Printf("%s", e.Inspect())
	}
	return &object.Integer{Value: 0}
}

// fnUpper is the implementation of our `upper` function.
func fnUpper(args []object.Object) object.Object {
	out := ""

	// Join all input arguments
	for _, arg := range args {
		val := fmt.Sprintf("%v", arg.Inspect())
		out += strings.ToUpper(val)
	}
	return &object.String{Value: out}

}
