package reflection

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

// TestBasic ensures our basic stuff is working
func TestBasic(t *testing.T) {

	type foo struct {
		Name   string
		Age    int
		OK     bool
		Number float32
	}

	obj := foo{Name: "Steve Kemp", Age: 42, Number: 3.25, OK: true}

	r := New(obj)

	//
	// For each value we'll test we got something
	//

	// Name: String
	nm, err := r.Get("Name")
	if err != nil {
		t.Fatalf("error getting name:%s", err.Error())
	}
	if nm.Type() != object.STRING {
		t.Fatalf("wrong type")
	}
	if nm.Inspect() != "Steve Kemp" {
		t.Fatalf("name is wrong")
	}

	// Age: Integer
	age, err := r.Get("Age")
	if err != nil {
		t.Fatalf("error getting age")
	}
	if age.Type() != object.INTEGER {
		t.Fatalf("wrong type, got: %s - %s", age.Type(), age.Inspect())
	}
	if age.Inspect() != "42" {
		t.Fatalf("age is wrong")
	}

	// Number: Integer
	n, err := r.Get("Number")
	if err != nil {
		t.Fatalf("error getting value")
	}
	if n.Type() != object.FLOAT {
		t.Fatalf("wrong type")
	}
	if n.Inspect() != "3.25" {
		t.Fatalf("number is wrong")
	}

	// OK: bool
	b, err := r.Get("OK")
	if err != nil {
		t.Fatalf("error getting value")
	}
	if b.Type() != object.BOOLEAN {
		t.Fatalf("wrong type")
	}
	if b.Inspect() != "true" {
		t.Fatalf("bool is wrong")
	}
}

// TestNested tries to test a nested structure.
func TestNested(t *testing.T) {

	type bar struct {
		Value string
	}
	type foo struct {
		Forename string
		Surname  bar
	}

	obj := foo{Forename: "Steve", Surname: bar{Value: "Kemp"}}

	r := New(obj)

	// Name: String
	sur, err := r.Get("Surname.Value")
	if err != nil {
		t.Fatalf("error getting value :%s", err.Error())
	}
	if sur.Type() != object.STRING {
		t.Fatalf("wrong type")
	}
	if sur.Inspect() != "Kemp" {
		t.Fatalf("wrong value")
	}
}

func TestArray(t *testing.T) {

	type Parent struct {
		Name     string
		Children []string
	}

	var dad Parent
	dad.Name = "Homer"
	dad.Children = []string{"Bart", "Lisa", "Maggie"}

	r := New(dad)

	// Children: String Array
	c, err := r.Get("Children")
	if err != nil {
		t.Fatalf("error getting value :%s", err.Error())
	}
	if c.Type() != object.ARRAY {
		t.Fatalf("wrong type, got %s", c.Type())
	}
}
