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

// TestScopeLookup tests that scoping works as I expect
//
//  1.  Original Scope
//  2.  Child Scope where there is "Steve"
//  3.  Second Child Scope whether there is "Zuul"
//  4.  4th scope, empty
//
func TestScopeLookup(t *testing.T) {

	// Create a scope.  Empty
	env := New()

	// Add a scope
	env.AddScope()
	env.SetLocal("Steve", &object.String{Value: "Kemp"})

	// Add another Scope
	env.AddScope()
	env.SetLocal("Zuul", &object.String{Value: "There is only"})

	// Add a final Scope
	env.AddScope()

	// Test lookups work
	for _, name := range []string{"Steve", "Zuul"} {
		_, ok := env.Get(name)
		if !ok {
			t.Errorf("Failed to lookup known-value: %s", name)
		}
	}

	// Set the Zuul, as a local
	env.SetLocal("Zuul", &object.String{Value: "Yes"})
	get, _ := env.Get("Zuul")
	if get.Inspect() != "Yes" {
		t.Errorf("Getting value failed")
	}

	// Set it as a global, which should still work
	env.Set("Zuul", &object.String{Value: "luuZ"})
	get, _ = env.Get("Zuul")
	if get.Inspect() != "luuZ" {
		t.Errorf("Getting value failed")
	}

	//
	// Drop the final scope
	//
	// If `Set` didn't walk up the scopes properly then it
	// will be gone
	//
	err := env.RemoveScope()
	if err != nil {
		t.Fatalf("Error removing scope")
	}
	get, _ = env.Get("Zuul")
	if get.Inspect() != "luuZ" {
		t.Errorf("Getting value failed")
	}

	// Update the second-layer
	env.Set("Steve", &object.String{Value: "evetS"})
	get, _ = env.Get("Steve")
	if get.Inspect() != "evetS" {
		t.Errorf("Getting value failed")
	}

	//
	// Drop the Z-scope.  Now we have:
	//  1.  Empty
	//  2.  Steve
	//
	err = env.RemoveScope()
	if err != nil {
		t.Fatalf("Error removing scope")
	}

	//
	// Drop the Steve-scope, now we have one empty
	//
	err = env.RemoveScope()
	if err != nil {
		t.Fatalf("Error removing scope")
	}

	//
	// All our variables are gone
	//
	for _, name := range []string{"Steve", "Zuul"} {
		res, ok := env.Get(name)
		if ok {
			t.Errorf("Got a result, but we've dropped the scopes! %v", res)
		}
	}

	//
	// Now we've added/removed the same number of scopes
	//
	err = env.RemoveScope()
	if err == nil {
		t.Fatalf("Should have seen an error removing invalid scope")
	}

}
