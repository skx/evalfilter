// builtins.go contains our in-built functions.

package environment

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/skx/evalfilter/v2/object"
)

// regCache is a cache of compiled regular expression objects.
// These may persist between runs because a regular expression object
// is essentially constant.
var regCache map[string]*regexp.Regexp

// init ensures that our regexp cache is populated
func init() {
	regCache = make(map[string]*regexp.Regexp)
}

// fnBetween is the implementation of our between function.
func fnBetween(args []object.Object) object.Object {

	// We expect three items "the value", and the lower/upper bounds.
	if len(args) != 3 {
		return &object.Null{}
	}

	// All arguments must be numbers
	for _, obj := range args {
		if obj.Type() != object.FLOAT && obj.Type() != object.INTEGER {
			return &object.Null{}
		}
	}

	// Get the values
	val := args[0]
	min := args[1]
	max := args[2]

	// val < min?
	lower := fnMin([]object.Object{val, min})
	if lower == val {

		if val.Inspect() != min.Inspect() {
			return &object.Boolean{Value: false}
		}
	}

	// val > max
	upper := fnMax([]object.Object{val, max})
	if upper == val {
		if val.Inspect() != max.Inspect() {
			return &object.Boolean{Value: false}
		}
	}

	return &object.Boolean{Value: true}
}

// fnFloat is the implementation of the `float` function.
//
// It converts an object to a float, if it can.
//
// On failure it returns Null
func fnFloat(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Stringify
	str := args[0].Inspect()

	i, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return &object.Null{}
	}

	return &object.Float{Value: i}
}

// fnGetenv is the implementation of the `getenv` function.
func fnGetenv(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Stringify
	str := args[0].Inspect()

	// Fetch & return
	return &object.String{Value: os.Getenv(str)}
}

// fnInt is the implementation of the `int` function.
//
// It converts an object to an integer, if it can.
//
// On failure it returns Null
func fnInt(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Stringify
	str := args[0].Inspect()

	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return &object.Null{}
	}

	return &object.Integer{Value: i}
}


// Join the given array with a string.
func fnJoin(args []object.Object) object.Object {

	// We expect two arguments
	if len(args) != 2 {
		return &object.Null{}
	}

	// The first argument must be an array
	if args[0].Type() != object.ARRAY {
		return &object.Null{}
	}
	if args[1].Type() != object.STRING {
		return &object.Null{}
	}

	// Do the join
	out := ""
	len := len(args[0].(*object.Array).Elements)

	for i, entry := range(args[0].(*object.Array).Elements) {
		out += entry.Inspect()
		if i != len-1 {
			out += args[1].(*object.String).Value
	        }
	}

	return &object.String{Value: out}
}

// Get the (sorted) keys from the specified hash.
func fnKeys(args []object.Object) object.Object {

	// We expect a single argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// The argument must be a hash
	if args[0].Type() != object.HASH {
		return &object.Null{}
	}

	// The object we're working with
	hash := args[0].(*object.Hash)
	entries := hash.Entries()

	// Create a new array for the results.
	array := make([]object.Object, len(entries))

	// Now copy the keys into it.
	for i, ent := range entries {
		array[i] = ent.Key
	}

	// Return the array.
	return &object.Array{Elements: array}
}

// fnLen is the implementation of our `len` function.
//
// Interestingly this function doesn't just count the length of string
// objects, instead we cast all objects to strings and allow their lengths
// to be calculated.
//
// The obvious exception is the handling of arrays and hashes.  The length of
// an array is the number of elements which it contains.  The length of a
// hash is the number of key-value pairs present.
//
// So `len(false)` is 5, len(3) is 1, and `len(0.123)` is 5, and arrays
// work as expected: len([]) is zero, and len(["steve", "kemp"]) is two.
//
func fnLen(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// array is handled differently
	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.Hash:
		return &object.Integer{Value: int64(len(arg.Pairs))}
	}

	// Stringify
	str := args[0].Inspect()
	sum := utf8.RuneCountInString(str)

	// return
	return &object.Integer{Value: int64(sum)}
}

// fnLower is the implementation of our `lower` function.
//
// Much like the `len` function here we cast to a string before
// we lower-case.
func fnLower(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Stringify and lower-case
	arg := fmt.Sprintf("%v", args[0].Inspect())
	arg = strings.ToLower(arg)

	// Return
	return &object.String{Value: arg}
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

// fnMax is the implementation of our `max` function.
func fnMax(args []object.Object) object.Object {

	// We expect two arguments
	if len(args) != 2 {
		return &object.Null{}
	}

	// Create an array.  Yeah.
	elements := make([]object.Object, 2)
	elements[0] = args[0]
	elements[1] = args[1]

	// Construct an actual array.
	arr := &object.Array{Elements: elements}

	// sort it
	out := fnSort([]object.Object{arr})

	// max
	return (out.(*object.Array).Elements[1])

}

// fnMin is the implementation of our `min` function.
func fnMin(args []object.Object) object.Object {

	// We expect two arguments
	if len(args) != 2 {
		return &object.Null{}
	}

	// Create an array.  Yeah.
	elements := make([]object.Object, 2)
	elements[0] = args[0]
	elements[1] = args[1]

	// Construct an actual array.
	arr := &object.Array{Elements: elements}

	// sort it
	out := fnSort([]object.Object{arr})

	// max
	return (out.(*object.Array).Elements[0])

}

// fnNow is the implementation of our `now` function.
func fnNow(args []object.Object) object.Object {

	// Handle timezones, by reading $TZ, and if not set
	// defaulting to UTC.
	env := os.Getenv("TZ")
	if env == "" {
		env = "UTC"
	}

	now := time.Now()

	// Ensure we set that timezone.
	loc, err := time.LoadLocation(env)
	if err == nil {
		now = now.In(loc)
	}

	return &object.Integer{Value: now.Unix()}
}

// fnSplit is the implementation of our `split` primitive.
func fnSplit(args []object.Object) object.Object {

	// We expect two arguments
	if len(args) != 2 {
		return &object.Null{}
	}

	// String to split
	input := args[0]

	// String to split by
	split := args[1]

	// Typecheck
	if input.Type() != object.STRING ||
		split.Type() != object.STRING {
		return &object.Null{}
	}

	// Perform the split
	pieces := strings.Split(input.(*object.String).Value,
		split.(*object.String).Value)

	// Convert the results into an array of string-objects
	elements := make([]object.Object, len(pieces))
	for i, e := range pieces {
		elements[i] = &object.String{Value: e}
	}

	// Now return that as an array.
	return (&object.Array{Elements: elements})
}

// fnString is the implementation of our `string` function.
func fnString(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	str := args[0].Inspect()
	return &object.String{Value: str}
}

// fnTrim is the implementation of our `trim` function.
func fnTrim(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	arg := args[0]
	val := strings.TrimSpace(arg.Inspect())

	return &object.String{Value: val}
}

// fnType is the implementation of our `type` function.
func fnType(args []object.Object) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Get the arg
	arg := args[0]

	// Get the type - lower-case
	val := string(arg.Type())
	val = strings.ToLower(val)

	// Return
	return &object.String{Value: val}
}

// fnPanic throws an error
func fnPanic(args []object.Object) (out object.Object) {

	out = &object.Void{}
	if len(args) == 1 {
		panic(args[0].Inspect())
	}

	panic("panic!")
}

// fnPrint is the implementation of our `print` function.
func fnPrint(args []object.Object) object.Object {
	for _, e := range args {
		fmt.Printf("%s", e.Inspect())
	}
	return &object.Void{}
}

// fnPrintf is the implementation of our `printf` function.
func fnPrintf(args []object.Object) object.Object {

	// Convert to the formatted version, via our `sprintf`
	// function.
	out := fnSprintf(args)

	// If that returned a string then we can print it
	if out.Type() == object.STRING {
		fmt.Print(out.(*object.String).Value)

	}

	return &object.Void{}
}

// fnSort implements our `sort` function
func fnSort(args []object.Object) object.Object {

	// We expect either one or two arguments
	//    sort([array], bool)
	if len(args) != 1 && len(args) != 2 {
		return &object.Null{}
	}

	// Type-check the first argument
	if args[0].Type() != object.ARRAY {
		return &object.Null{}
	}

	// Default to not lower-casing items
	lower := false

	// Second (optional) argument controls case-sensitivity.
	if len(args) == 2 {

		// Type-check second argument
		if args[1].Type() != object.BOOLEAN {
			return &object.Null{}
		}

		// Copy value.
		lower = args[1].(*object.Boolean).Value
	}

	// defer to our helper method
	return (sortHelper(args, lower, false))
}

// fnReverse implements our `reverse` function
func fnReverse(args []object.Object) object.Object {

	// We expect either one or two arguments
	//    reverse([array], bool)
	if len(args) != 1 && len(args) != 2 {
		return &object.Null{}
	}

	// Type-check the first argument
	if args[0].Type() != object.ARRAY {
		return &object.Null{}
	}

	// Default to not lower-casing items
	lower := false

	// Second (optional) argument controls case-sensitivity.
	if len(args) == 2 {

		// Type-check second argument
		if args[1].Type() != object.BOOLEAN {
			return &object.Null{}
		}

		// Copy value.
		lower = args[1].(*object.Boolean).Value
	}

	// Defer to our helper method.
	return (sortHelper(args, lower, true))
}

// sortHelper is a helper function which allows sorting/reversing an array of items.
//
// TODO: See if we can avoid juggling this sort with a temporary object.
// TODO: It might be we can sort.Slice on array.Elements,  but when I tried
// TODO: that first I got "panic Swapper on nil"
func sortHelper(args []object.Object, lowerCase bool, doReverse bool) object.Object {

	// We convert the input-array we're sorting into an
	// array of structures which we can sort natively.
	type Temp struct {

		// key is the (stringified) copy of the
		// object-members contents
		key string

		// index is the ORIGINAL index of the item
		// in the input array.
		index int
	}

	//
	// Make a copy of the keys + indexes
	//
	items := make([]Temp, len(args[0].(*object.Array).Elements))
	for i := range items {
		items[i].key = args[0].(*object.Array).Elements[i].Inspect()
		items[i].index = i
	}

	// Sort the temporary structure we have been given.
	//
	// Here we handle "sort vs. reverse".
	//
	// We also handle the optional case-insensitivity.
	sort.Slice(items, func(i, j int) bool {

		a := items[i].key
		b := items[j].key

		if lowerCase {
			a = strings.ToLower(a)
			b = strings.ToLower(b)
		}

		if doReverse {
			return b < a
		}
		return a < b
	})

	// Now we've sorted our result - populate an array
	// with the values which have been sorted.
	//
	// Here we copy items from the original array so
	// the types are the same.
	//
	// e.g. "sort(["Steve", 3])" works as expected with
	// regard to the items keeping their types.
	//
	out := make([]object.Object, len(items))
	for i, e := range items {
		out[i] = args[0].(*object.Array).Elements[e.index]
	}

	// All done.
	return &object.Array{Elements: out}
}

// fnSprintf is the implementation of our `sprintf` function.
func fnSprintf(args []object.Object) object.Object {

	// We expect 1+ arguments
	if len(args) < 1 {
		return &object.Null{}
	}

	// Type-check
	if args[0].Type() != object.STRING {
		return &object.Null{}
	}

	// Get the format-string.
	fs := args[0].(*object.String).Value

	// Convert the arguments to something go's sprintf
	// code will understand.
	argLen := len(args)
	fmtArgs := make([]interface{}, argLen-1)

	// Here we convert and assign.
	for i, v := range args[1:] {
		fmtArgs[i] = v.ToInterface()
	}

	// Call the helper
	out := fmt.Sprintf(fs, fmtArgs...)

	// And now return the value.
	return &object.String{Value: out}
}

// fnUpper is the implementation of our `upper` function.
//
// Again we stringify our arguments here so `upper(true)` is
// the string `TRUE`.
func fnUpper(args []object.Object) object.Object {
	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// Stringify and upper-case
	arg := fmt.Sprintf("%v", args[0].Inspect())
	arg = strings.ToUpper(arg)

	// Return
	return &object.String{Value: arg}
}

// getTimeField handles returning a time-related field from an object
// which is assumed to contain a time in the Unix Epoch format.
func getTimeField(args []object.Object, val string) object.Object {

	// We expect one argument
	if len(args) != 1 {
		return &object.Null{}
	}

	// It must be an integer
	if args[0].Type() != object.INTEGER {
		return &object.Null{}
	}

	// Convert that to a time
	ts := time.Unix(args[0].(*object.Integer).Value, 0)

	// Handle timezones, by reading $TZ, and if not set
	// defaulting to UTC.
	env := os.Getenv("TZ")
	if env == "" {
		env = "UTC"
	}

	// Ensure we set that timezone.
	loc, err := time.LoadLocation(env)
	if err == nil {
		ts = ts.In(loc)
	}

	// Now get the fields
	hr, min, sec := ts.Clock()
	year, month, day := ts.Date()

	// And return the one we should
	switch val {
	case "hour":
		return &object.Integer{Value: int64(hr)}
	case "minute":
		return &object.Integer{Value: int64(min)}
	case "seconds":
		return &object.Integer{Value: int64(sec)}
	case "day":
		return &object.Integer{Value: int64(day)}
	case "month":
		return &object.Integer{Value: int64(month)}
	case "year":
		return &object.Integer{Value: int64(year)}
	case "weekday":
		return &object.String{Value: ts.Weekday().String()}
	}

	// Unknown field: can't happen?
	return &object.Null{}
}

// fnHour returns the hour of the given time-object.
func fnHour(args []object.Object) object.Object {
	return getTimeField(args, "hour")
}

// fnMinute returns the minute of the given time-object.
func fnMinute(args []object.Object) object.Object {
	return getTimeField(args, "minute")
}

// fnSeconds returns the seconds of the given time-object.
func fnSeconds(args []object.Object) object.Object {
	return getTimeField(args, "seconds")
}

// fnDay returns the day of the given time-object.
func fnDay(args []object.Object) object.Object {
	return getTimeField(args, "day")
}

// fnMonth returns the month of the given time-object.
func fnMonth(args []object.Object) object.Object {
	return getTimeField(args, "month")
}

// fnYear returns the year of the given time-object.
func fnYear(args []object.Object) object.Object {
	return getTimeField(args, "year")
}

// fnWeekday returns the name of the day in the given time-object.
func fnWeekday(args []object.Object) object.Object {
	return getTimeField(args, "weekday")
}
