// Package object contains the golang-implementation of our evalfilter
// object-types.
//
// Our scripting language supports several different object-types:
//
// * Array.
// * Boolean value.
// * Floating-point number.
// * Integer number.
// * Null
// * String value.
//
// To allow these objects to be used interchanagably there is a simple
// interface which all object-types must implement, which is simple to
// satisfy.
package object

// Type describes the type of an object.
type Type string

// pre-defined object types.
const (
	ARRAY   = "ARRAY"
	BOOLEAN = "BOOLEAN"
	ERROR   = "ERROR"
	FLOAT   = "FLOAT"
	INTEGER = "INTEGER"
	NULL    = "NULL"
	STRING  = "STRING"
)

// Object is the interface that all of our various object-types must implement.
//
// This is a deliberately minimal interface, which should allow the easy
// addition of new types in the future.
type Object interface {

	// Inspect returns a string-representation of the given object.
	Inspect() string

	// Type returns the type of this object.
	Type() Type

	// True returns whether this object is "true".
	//
	// This is used when an object is used in an `if` expression,
	// for example, or with the logical `&&` and `||` operations.
	True() bool
}
