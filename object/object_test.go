package object

import (
	"fmt"
	"testing"
)

// TestArray tests our Array object a little
func TestArray(t *testing.T) {

	// Content of our array.
	content := []Object{&String{Value: "foo"},
		&String{Value: "bar"},
		&String{Value: "baz"}}

	// Create it
	arr := &Array{Elements: content}
	empty := &Array{}

	// Inspect
	if arr.Inspect() != "[foo, bar, baz]" {
		t.Fatalf("Unexpected content:%s\n", arr.Inspect())
	}

	// Type
	if arr.Type() != ARRAY {
		t.Fatalf("Invalid type")
	}

	// True
	if empty.True() {
		t.Fatalf("Empty array should be false!")
	}
	if !arr.True() {
		t.Fatalf("Non-empty array should be true!")
	}

	//
	// We implement our foreach behaviour via an
	// iteration interface.  Test that.
	//
	vals := []string{"foo", "bar", "baz"}

	// Interface
	arrInt := arr.ToInterface().([]interface{})

	// Ensure we got the right count
	if len(arrInt) != 3 {
		t.Fatalf("Length of array is wrong")
	}

	// Ensure each value matches expectations
	for i, entry := range arrInt {
		if vals[i] != entry.(string) {
			t.Fatalf("toInterface results not matched")
		}
	}

	//
	// No harm in repeating this test a few times
	//
	for range []int{0, 1, 2} {

		// Reset the iteration and count of loops.
		arr.Reset()
		count := 0

		// For each of the known array-values we expect
		for i := range vals {

			// Get the next-value from the array, via the
			// iterator.
			obj, offset, more := arr.Next()

			// Ensure the offset matches what we expect
			if int(offset.(*Integer).Value) != count {
				t.Fatalf("Iteration offset got messed up: %d != %d", offset, count)
			}

			// And that the value matches
			if obj.Inspect() != vals[i] {
				t.Fatalf("wrong value for offset %d %s != %s", i, vals[i], obj.Inspect())
			}

			// More?
			if i != len(vals) && !more {
				t.Fatalf("Expected more, but got none")
			}
			count++
		}

		// Now we've exhausted our iteration
		obj, offset, more := arr.Next()
		if more {
			t.Fatalf("We didn't expect more text, but found it")
		}
		if int(offset.(*Integer).Value) != 0 {
			t.Fatalf("At the end of the iteration we got a weird offset")
		}
		if obj != nil {
			t.Fatalf("At the end of the iteration we got a weird object")
		}
	}

}

// TestBoolean tests our Bool-object in a basic way.
func TestBool(t *testing.T) {

	true := &Boolean{Value: true}
	false := &Boolean{Value: false}

	// Inspect
	if true.Inspect() != "true" {
		t.Fatalf("Invalid value!")
	}
	if false.Inspect() != "false" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if true.Type() != BOOLEAN {
		t.Fatalf("Wrong type")
	}
	if false.Type() != BOOLEAN {
		t.Fatalf("Wrong type")
	}

	// True
	if !true.True() {
		t.Fatalf("Truth test on boolean failed")
	}
	if false.True() {
		t.Fatalf("Truth test on boolean failed")
	}

	tX := true.ToInterface()
	if !tX.(bool) {
		t.Fatalf("interface usage failed")
	}
	fX := false.ToInterface()
	if fX.(bool) {
		t.Fatalf("interface usage failed")
	}
}

// TestFloat tests our Float-object in a basic way.
func TestFloat(t *testing.T) {

	tmp := &Float{Value: 3.7}
	nul := &Float{Value: 0}

	// Inspect
	if tmp.Inspect() != "3.7" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if tmp.Type() != FLOAT {
		t.Fatalf("Wrong type")
	}

	// True
	if !tmp.True() {
		t.Fatalf("Non-zero float should be true")
	}
	if nul.True() {
		t.Fatalf("zero-value should be false")
	}

	x := tmp.ToInterface()
	if x.(float64) != 3.7 {
		t.Fatalf("interface usage failed")
	}

	// Our language implements "foo++" and "foo--"
	// via an interface.  Test those methods
	tmp.Increase()
	if tmp.Inspect() != "4.7" {
		t.Errorf("Increase() failed")
	}
	tmp.Decrease()
	tmp.Decrease()
	if tmp.Inspect() != "2.7" {
		t.Errorf("Decrease() failed")
	}
}

// TestInt tests our Integer-object in a basic way.
func TestInt(t *testing.T) {

	tmp := &Integer{Value: 3}
	nul := &Integer{Value: 0}

	// Inspect
	if tmp.Inspect() != "3" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if tmp.Type() != INTEGER {
		t.Fatalf("Wrong type")
	}

	// True
	if !tmp.True() {
		t.Fatalf("Non-zero integer should be true")
	}
	if nul.True() {
		t.Fatalf("zero-value should be false")
	}

	// Interface
	x := tmp.ToInterface()
	if x.(int64) != 3 {
		t.Fatalf("interface usage failed")
	}

	// Our language implements "foo++" and "foo--"
	// via an interface.  Test those methods
	tmp.Increase()
	if tmp.Inspect() != "4" {
		t.Errorf("Increase() failed")
	}
	tmp.Decrease()
	tmp.Decrease()
	if tmp.Inspect() != "2" {
		t.Errorf("Decrease() failed")
	}
}

// TestNull tests our Null-object in a basic way.
func TestNull(t *testing.T) {

	v := &Null{}

	// Inspect
	if v.Inspect() != "null" {
		t.Fatalf("Invalid Inspect() value!")
	}

	// Type
	if v.Type() != NULL {
		t.Fatalf("Wrong type")
	}

	// True
	if v.True() {
		t.Fatalf("null object should never be True")
	}

	x := v.ToInterface()
	if x != nil {
		t.Fatalf("interface usage failed")
	}
}

// TestString tests our String-object in a basic way.
func TestString(t *testing.T) {

	tmp := &String{Value: "Steve"}
	nul := &String{Value: ""}

	// Inspect
	if tmp.Inspect() != "Steve" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if tmp.Type() != STRING {
		t.Fatalf("Wrong type")
	}

	// True
	if !tmp.True() {
		t.Fatalf("Non-empty string should be true")
	}
	if nul.True() {
		t.Fatalf("empty string should be false")
	}

	x := tmp.ToInterface()
	if x.(string) != "Steve" {
		t.Fatalf("interface usage failed")
	}

	//
	// We implement our foreach behaviour via an
	// iteration interface.  Test that.
	//
	txt := "Steve狐犬"
	tmp = &String{Value: txt}

	//
	// No harm in repeating this test a few times
	//
	for range []int{0, 1, 2} {

		// Reset the iteration and count of loops.
		tmp.Reset()
		count := 0

		// For each of the known string-characters we expect
		for i, c := range txt {

			// Get the next-value from the array, via the
			// iterator.
			obj, offset, more := tmp.Next()

			// Ensure the offset matches what we expect
			if int(offset.(*Integer).Value) != count {
				t.Fatalf("Iteration offset got messed up: %d != %d", offset, count)
			}

			if i != len(txt) && !more {
				t.Fatalf("Apparently there is no more")
			}
			if obj.Inspect() != fmt.Sprintf("%c", c) {
				t.Fatalf("wrong value for offset %d %c != %s", i, c, obj.Inspect())
			}

			count++
		}

		// Now we've exhausted our iteration
		obj, offset, more := tmp.Next()
		if more {
			t.Fatalf("We didn't expect more text, but found it")
		}
		if int(offset.(*Integer).Value) != 0 {
			t.Fatalf("At the end of the iteration we got a weird offset")
		}
		if obj != nil {
			t.Fatalf("At the end of the iteration we got a weird object")
		}
	}
}

// TestVoid tests our Void-object in a basic way.
func TestVoid(t *testing.T) {

	v := &Void{}

	// Inspect
	if v.Inspect() != "void" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if v.Type() != VOID {
		t.Fatalf("Wrong type")
	}

	// True
	if v.True() {
		t.Fatalf("Void object should never be True")
	}

	x := v.ToInterface()
	if x != nil {
		t.Fatalf("interface usage failed")
	}
}
