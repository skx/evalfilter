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
}

func TestArrayLength(t *testing.T) {

	// Content of our array.
	content := []Object{&String{Value: "foo"},
		&String{Value: "bar"},
		&String{Value: "baz"}}

	// Create it
	arr := &Array{Elements: content}

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
}

func TestArrayIteration(t *testing.T) {

	// Content of our array.
	content := []Object{&String{Value: "foo"},
		&String{Value: "bar"},
		&String{Value: "baz"}}

	//
	// We implement our foreach behaviour via an
	// iteration interface.  Test that.
	//
	vals := []string{"foo", "bar", "baz"}

	// Create it
	arr := &Array{Elements: content}

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

	tb := &Boolean{Value: true}
	fb := &Boolean{Value: false}

	// Inspect
	if tb.Inspect() != "true" {
		t.Fatalf("Invalid value!")
	}
	if fb.Inspect() != "false" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if tb.Type() != BOOLEAN {
		t.Fatalf("Wrong type")
	}
	if fb.Type() != BOOLEAN {
		t.Fatalf("Wrong type")
	}

	// True
	if !tb.True() {
		t.Fatalf("Truth test on boolean failed")
	}
	if fb.True() {
		t.Fatalf("Truth test on boolean failed")
	}

	tX := tb.ToInterface()
	if !tX.(bool) {
		t.Fatalf("interface usage failed")
	}
	fX := fb.ToInterface()
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

	// Hash checks - two identical values should hash
	// in the same way
	a := Float{Value: 3.1}
	b := Float{Value: 3.1}
	c := Float{Value: 33.1}

	aH := a.HashKey()
	bH := b.HashKey()
	cH := c.HashKey()

	if aH != bH {
		t.Fatalf("two identical values should have the same hash")
	}
	if aH == cH {
		t.Fatalf("two different values should have different hashes")
	}
}

// TestHash tests our hash object in a basic way
func TestHash(t *testing.T) {
	tmp := &Hash{}

	if tmp.True() {
		t.Fatalf("empty hash should be false")
	}

	if tmp.Type() != HASH {
		t.Fatalf("hash has the wrong type")
	}

	x := tmp.ToInterface()
	if x.(string) != "<HASH>" {
		t.Fatalf("interface usage failed")
	}

	// Create some values for the hash
	a := HashPair{Key: &String{Value: "Name"}, Value: &String{Value: "Steve"}}
	aK := &String{Value: "Name"}
	b := HashPair{Key: &String{Value: "Country"}, Value: &String{Value: "Finland"}}
	bK := &String{Value: "Country"}

	tmp.Pairs = make(map[HashKey]HashPair)
	tmp.Pairs[aK.HashKey()] = a
	tmp.Pairs[bK.HashKey()] = b

	// Now we have some entries
	if !tmp.True() {
		t.Fatalf("populated hash should be true")
	}

	if tmp.Inspect() != "{Country: Finland, Name: Steve}" {
		t.Fatalf("Got %s for hash", tmp.Inspect())
	}

	// Reset the iteration
	tmp.Reset()

	// Get the next-value from the array, via the
	// iterator.
	v1, k1, more1 := tmp.Next()

	if !more1 {
		t.Fatalf("we expect more iterations")
	}

	if k1.Inspect() != "Country" {
		t.Fatalf("wrong key, got %s", k1.Inspect())
	}
	if v1.Inspect() != "Finland" {
		t.Fatalf("wrong key")
	}

	// Get the next-value from the array, via the
	// iterator.
	v2, k2, more2 := tmp.Next()
	if !more2 {
		t.Fatalf("we expect more iterations")
	}
	if k2.Inspect() != "Name" {
		t.Fatalf("wrong key, got %s", k2.Inspect())
	}
	if v2.Inspect() != "Steve" {
		t.Fatalf("wrong key")
	}

	_, _, more3 := tmp.Next()
	if more3 {
		t.Fatalf("iteration should be over now")
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

	// Hash checks - two identical values should hash
	// in the same way
	a := Integer{Value: 31}
	b := Integer{Value: 31}
	c := Integer{Value: 33}

	aH := a.HashKey()
	bH := b.HashKey()
	cH := c.HashKey()

	if aH != bH {
		t.Fatalf("two identical values should have the same hash")
	}
	if aH == cH {
		t.Fatalf("two different values should have different hashes")
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

// TestRegexp tests our Regexp-object in a basic way.
func TestRegexp(t *testing.T) {

	tmp := &Regexp{Value: "Steve"}
	nul := &Regexp{Value: ""}

	// Inspect
	if tmp.Inspect() != "Steve" {
		t.Fatalf("Invalid value!")
	}

	// Type
	if tmp.Type() != REGEXP {
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
}

func TestStringIteration(t *testing.T) {

	//
	// We implement our foreach behaviour via an
	// iteration interface.  Test that.
	//
	txt := "Steve狐犬"
	tmp := &String{Value: txt}

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

func TestStringHash(t *testing.T) {
	// Hash checks - two identical values should hash
	// in the same way
	a := String{Value: "Steve"}
	b := String{Value: "Steve"}
	c := String{Value: "steve"}

	aH := a.HashKey()
	bH := b.HashKey()
	cH := c.HashKey()

	if aH != bH {
		t.Fatalf("two identical values should have the same hash")
	}
	if aH == cH {
		t.Fatalf("two different values should have different hashes")
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
