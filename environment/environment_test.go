package environment

import (
	"testing"

	"github.com/skx/evalfilter/v2/object"
)

func TestFunctions(t *testing.T) {

	env := New()

	names := []string{"print", "trim", "type", "int", "float"}

	// Functions will exist.
	for _, fn := range names {
		_, ok := env.GetFunction(fn)
		if !ok {
			t.Errorf("Failed to get function %s", fn)
		}
	}

	// Functions will not exist when modified
	for _, fn := range names {
		_, ok := env.GetFunction(fn + fn)
		if ok {
			t.Errorf("Found function %s which should not exist", fn)
		}
	}
}

func TestGetSet(t *testing.T) {

	env := New()

	env.Set("foo", &object.String{Value: "bar"})

	out, ok := env.Get("foo")
	if !ok {
		t.Errorf("Failed to lookup known-value.")
	}
	if out.Type() != "STRING" {
		t.Errorf("known-value had wrong type")
	}
	if out.(*object.String).Value != "bar" {
		t.Errorf("known-value had wrong value")
	}
}

func TestMissingSet(t *testing.T) {

	env := New()

	_, ok := env.Get("foo")
	if ok {
		t.Errorf("lookup of a missing value worked, bogus.")
	}
}
