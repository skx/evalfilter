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
// * Regular-expression objects.
//
// To allow these objects to be used interchanagably each kind of object
// must implement the same simple interface.
//
// There are additional interfaces for adding support for more advanced
// operations - such as iteration, incrementing, and decrementing.
package object

// Type describes the type of an object.
type Type string

// pre-defined object types.
const (
	ARRAY   = "ARRAY"
	BOOLEAN = "BOOLEAN"
	FLOAT   = "FLOAT"
	HASH    = "HASH"
	INTEGER = "INTEGER"
	NULL    = "NULL"
	REGEXP  = "REGEXP"
	STRING  = "STRING"
	VOID    = "VOID"
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

	// ToInterface converts the given object to a "native" golang value,
	// which is required to ensure that we can use the object in our
	// `sprintf` or `printf` primitives.
	ToInterface() interface{}
}

// Increment is an interface that some objects might wish to support.
//
// If this interface is implemented then it will be possible to use the
// the postfix `++` operator upon objects of that type, without that
// a run-time error will be generated.
type Increment interface {

	// Increase raises the object's value by one.
	Increase()
}

// Decrement is an interface that some objects might wish to support.
//
// If this interface is implemented then it will be possible to use the
// the postfix `--` operator upon objects of that type, without that
// a run-time error will be generated.
type Decrement interface {

	// Decrease lowers the object's value by one.
	Decrease()
}

// Iterable is an interface that some objects might wish to support.
//
// If this interface is implemented then it will be possible to
// use the `foreach` function to iterate over the object.  If
// the interface is not implemented then a run-time error will
// be generated instead.
type Iterable interface {

	// Reset the state of any previous iteration.
	Reset()

	// Get the next "thing" from the object being iterated
	// over.
	//
	// The return values are the item which is to be returned
	// next, the index of that object, and finally a boolean
	// to say whether the function succeeded.
	//
	// If the boolean value returned is false then that
	// means the iteration has completed and no further
	// items are available.
	Next() (Object, Object, bool)
}

// Hashable type can be hashed
type Hashable interface {

	// HashKey returns a hash key for the given object.
	HashKey() HashKey
}
