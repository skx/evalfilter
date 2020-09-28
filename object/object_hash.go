package object

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// HashKey is the structure used for hash-keys
type HashKey struct {
	// Type holds the type of the object.
	Type Type

	// Value holds the actual hash-key value.
	Value uint64
}

// HashPair is a structure which is used to store hash-entries
type HashPair struct {
	// Key holds our hash-key key.
	Key Object

	// Value holds our hash-key value.
	Value Object
}

// ByName implements sort.Interface for []HashPair based on the Key field.
type ByName []HashPair

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Key.Inspect() < a[j].Key.Inspect() }

// Hash wrap map[HashKey]HashPair and implements Object interface.
type Hash struct {
	// Pairs holds the key/value pairs of the hash we wrap
	Pairs map[HashKey]HashPair

	// offset holds our iteration-offset.
	offset int
}

// Type returns the type of this object.
func (h *Hash) Type() Type {
	return HASH
}

// Entries returns the sorted list of entries we maintain
//
// We need this to guarantee a stable order when iterating, or
// exporting to a string.
func (h *Hash) Entries() []HashPair {

	// Maintain an array of values
	var entries []HashPair

	for _, pair := range h.Pairs {
		entries = append(entries, pair)
	}

	// Sort them
	sort.Sort(ByName(entries))

	for i, n := range entries {
		fmt.Printf("%d: %v\n", i, n)
	}
	return entries
}

// Inspect returns a string-representation of the given object.
func (h *Hash) Inspect() string {

	// Get the list of entries, sorted by key-name.
	entries := h.Entries()

	// Now output
	var out bytes.Buffer

	pairs := make([]string, 0)
	for _, entry := range entries {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			entry.Key.Inspect(), entry.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// True returns whether this object wraps a true-like value.
//
// Used when this object is the conditional in a comparison, etc.
func (h *Hash) True() bool {
	return (len(h.Pairs) != 0)
}

// Reset implements the Iterable interface, and allows the contents
// of the array to be reset to allow re-iteration.
func (h *Hash) Reset() {
	h.offset = 0
}

// Next implements the Iterable interface, and allows the contents
// of our array to be iterated over.
func (h *Hash) Next() (Object, Object, bool) {

	// Get the list of entries, sorted by key-name.
	entries := h.Entries()

	// Now pick the next entry
	if h.offset < len(entries) {
		idx := 0

		for _, pair := range entries {
			if h.offset == idx {
				h.offset++
				return pair.Value, pair.Key, true
			}
			idx++
		}
	}

	return nil, &Integer{Value: 0}, false
}

// ToInterface converts this object to a go-interface, which will allow
// it to be used naturally in our sprintf/printf primitives.
//
// It might also be helpful for embedded users.
func (h *Hash) ToInterface() interface{} {
	return "<HASH>"
}

// Ensure this object implements the expected interfaces.
var _ Iterable = &Hash{}
