package token

import (
	"strings"
	"testing"
)

// Test looking up values succeeds, then fails
func TestLookup(t *testing.T) {

	for key, val := range keywords {

		// Obviously this will pass.
		if LookupIdentifier(key) != val {
			t.Errorf("Lookup of %s failed", key)
		}

		// Once the keywords are uppercase they'll no longer
		// match - so we find them as identifiers.
		if LookupIdentifier(strings.ToUpper(key)) != IDENT {
			t.Errorf("Lookup of %s failed", key)
		}
	}
}

// TestPosition doesn't really test anything :/
func TestPosition(t *testing.T) {
	x := &Token{}
	if !strings.Contains(x.Position(), ", column") {
		t.Fatalf("failed to get position")
	}
}
