// Package object contains our core-definitions for objects.
//
// Our scripting language supports several different object-types:
//
// * Integer number.
// * Floating-point number.
// * String
// * Boolean values (true, or false).
// * Null
package object

// Type describes the type of an object.
type Type string

// pre-defined constant Type
const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	STRING_OBJ       = "STRING"
)

// Object is the interface that all of our various object-types must implmenet.
type Object interface {

	// Type returns the type of this object.
	Type() Type

	// Inspect returns a string-representation of the given object.
	Inspect() string
}
