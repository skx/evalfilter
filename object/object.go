// Package object contains our core-definitions for objects.
//
// Our language supports several different object-types:
//
// * Integer number.
// * Floating-point number.
// * String.
// * Booleans.
// * Null
package object

// Type describes the type of an object.
type Type string

// pre-defined constant Type
const (
	INTEGER = "INTEGER"
	FLOAT   = "FLOAT"
	BOOLEAN = "BOOLEAN"
	NULL    = "NULL"
	ERROR   = "ERROR"
	STRING  = "STRING"
)

// Object is the interface that all of our various object-types must implement.
type Object interface {

	// Type returns the type of this object.
	Type() Type

	// Inspect returns a string-representation of the given object.
	Inspect() string
}
