package object

import "testing"

// TODO: Iteration
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
