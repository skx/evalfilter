// This file contains the implementation for the `if` operation.

package runtime

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skx/evalfilter/environment"
)

// IfOperation holds state for the `if` operation
type IfOperation struct {
	// Left argument
	Left Argument

	// Right argument - note that this might be missing.
	Right Argument

	// The comparison operation to be applied to the two arguments.
	//
	// If only a single argument is passed to the `if` operation
	// then the left-most argument will be tested for "truthiness".
	Op string

	// Operations to be carried out if the statement matches.
	True []Operation

	// Operations to be carried out if the statement does not.
	False []Operation
}

// Run executes an if statement.
func (i *IfOperation) Run(env *environment.Environment, obj interface{}) (bool, bool, error) {

	// Run the if-statement.
	res, err := i.doesMatch(env, obj)

	// Was there an error?
	if err != nil {
		return false, false, fmt.Errorf("failed to run if-test %s", err)
	}

	//
	// No error - and we got a match.
	//
	if res {

		//
		// The test matches so we should now handle
		// all the things that are in the `true`
		// list.
		//
		for _, t := range i.True {

			//
			// Process each operation.
			//
			// If this was a return statement then we return
			//
			ret, val, err := t.Run(env, obj)
			if ret {
				return ret, val, err
			}

		}
	} else {

		//
		// The test did not so we should now handle
		// all the things that are in the `false` list.
		//
		for _, t := range i.False {

			//
			// Process each operation.
			//
			// If this was a return statement then we return
			//
			ret, val, err := t.Run(env, obj)
			if ret {
				return ret, val, err
			}

		}

	}

	return false, false, nil
}

// doesMatch runs the actual comparison for the if-statement.
func (i *IfOperation) doesMatch(env *environment.Environment, obj interface{}) (bool, error) {

	//
	// Expand the left & right sides of the conditional
	//
	lVal := i.Left.Value(env, obj)

	//
	// Single argument form?
	//

	if i.Op == "" {

		//
		// Is the result true/false?
		//
		if i.truthy(lVal) {
			return true, nil
		}

		return false, nil
	}

	rVal := i.Right.Value(env, obj)

	//
	// Convert to strings, in case they're needed for the early
	// operations.
	//
	lStr := fmt.Sprintf("%v", lVal)
	rStr := fmt.Sprintf("%v", rVal)

	//
	// Basic operations
	//

	// Equality - string and number.
	if i.Op == "==" {
		return (lStr == rStr), nil
	}

	// Inequality - string and number.
	if i.Op == "!=" {
		return (lStr != rStr), nil
	}

	// String-contains
	if i.Op == "~=" {
		return strings.Contains(lStr, rStr), nil
	}

	// String does not contain
	if i.Op == "!~" {
		return !strings.Contains(lStr, rStr), nil
	}

	//
	// All remaining operations are numeric, so we need to convert
	// the values into numbers.
	//
	// Call them `a` and `b`.
	//
	var a float64
	var b float64
	var err error

	//
	// Convert
	//
	a, err = i.toNumberArg(lVal)
	if err != nil {
		return false, err
	}
	b, err = i.toNumberArg(rVal)
	if err != nil {
		return false, err
	}

	//
	// Now operate.
	//
	if i.Op == ">" {
		return (a > b), nil
	}
	if i.Op == ">=" {
		return (a >= b), nil
	}
	if i.Op == "<" {
		return (a < b), nil
	}
	if i.Op == "<=" {
		return (a <= b), nil
	}

	//
	// Invalid operator?
	//
	return false, fmt.Errorf("unknown operator %v", i.Op)
}

// toNumberArg tries to convert the given interface to a float64 value.
func (i *IfOperation) toNumberArg(value interface{}) (float64, error) {

	// string?
	_, ok := value.(string)
	if ok {
		a, _ := strconv.ParseFloat(value.(string), 32)
		return a, nil
	}

	// int
	_, ok = value.(int)
	if ok {
		return (float64(value.(int))), nil
	}

	// float?
	_, ok = value.(int64)
	if ok {
		return (float64(value.(int64))), nil
	}

	return 0, fmt.Errorf("failed to convert %v to number", value)
}

// truthy tests if the given object is "truthy", or true-like:
//
// * nil is false.
// * `false` is false
// * An empty string is false.
// * A numeric value of 0 is false.
//
// Everything else is true.
func (i *IfOperation) truthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return (val.(bool))
	case string:
		return (val.(string) != "")
	case int:
		return (val.(int) != 0)
	case int8:
		return (val.(int8) != 0)
	case int16:
		return (val.(int16) != 0)
	case int32:
		return (val.(int32) != 0)
	case int64:
		return (val.(int64) != 0)
	case uint:
		return (val.(uint) != 0)
	case uint8:
		return (val.(uint8) != 0)
	case uint16:
		return (val.(uint16) != 0)
	case uint32:
		return (val.(uint32) != 0)
	case uint64:
		return (val.(uint64) != 0)
	case float32:
		return (val.(float32) != 0)
	case float64:
		return (val.(float64) != 0)
	case nil:
		return false
	default:
		fmt.Printf("unexpected type %T", v)
	}

	return false
}
