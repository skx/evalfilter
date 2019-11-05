// Package vm implements a simple virtual machine.
//
// We're constructed with a set of opcodes, and we process those forever,
// until we hit a `return` statement which terminates our run.
package vm

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/skx/evalfilter/v2/code"
	"github.com/skx/evalfilter/v2/object"
	"github.com/skx/evalfilter/v2/stack"
)

// True is our global "true" object.
var True = &object.Boolean{Value: true}

// False is our global "false" object.
var False = &object.Boolean{Value: false}

// Null is our global "false" object.
var Null = &object.Null{}

// VM is the structure which holds our state.
type VM struct {

	// constants holds constants in the program source
	// for example the script "1 + 3;" contains two
	// numeric constants "1" and "3".
	constants []object.Object

	// bytecode contains the actual instructions we'll execute.
	bytecode code.Instructions

	// stack stores our stack.
	//
	// We're a stack-based virtual machine so this is used for many
	// many of our internal facilities.
	stack *stack.Stack

	// environment holds the environment, which will allow variables
	// and functions to be get/set.
	environment *object.Environment

	// fields contains the contents of all the fields in the object
	// or map we're created with.
	//
	// These are discovered dynamically at run-time, and parsed only
	// once.
	fields map[string]object.Object
}

// New constructs a new virtual machine.
func New(constants []object.Object, bytecode code.Instructions, env *object.Environment) *VM {

	return &VM{
		constants:   constants,
		environment: env,
		bytecode:    bytecode,
		stack:       stack.NewStack(),
	}
}

// Run launches our virtual machine, intepreting the bytecode we were
// constructed with.
//
// We return when we hit a return-operation, or we hit the end of
// the supplied bytecode - the latter is an error.
func (vm *VM) Run(obj interface{}) (object.Object, error) {

	// Sanity-check
	if len(vm.bytecode) < 1 {
		return nil, fmt.Errorf("bytecode is empty.  Did you forget to call evalfilter.Prepare?")
	}

	//
	// Make an empty map to store field/map contents.
	//
	vm.fields = make(map[string]object.Object)

	//
	// Instruction pointer and length.
	//
	ip := 0
	ln := len(vm.bytecode)

	//
	// Loop over all the bytecode.
	//
	for ip < ln {

		op := code.Opcode(vm.bytecode[ip])

		switch op {

		// store constant
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.bytecode[ip+1:])
			ip += 3

			vm.stack.Push(vm.constants[constIndex])

			// Lookup variable/field, by name
		case code.OpLookup:
			constIndex := code.ReadUint16(vm.bytecode[ip+1:])
			ip += 3

			// Get the name.
			name := vm.constants[constIndex].Inspect()

			val := vm.lookup(obj, name)
			vm.stack.Push(val)

			// Set a variable by name
		case code.OpSet:
			var name object.Object
			var val object.Object
			var err error
			name, err = vm.stack.Pop()
			if err != nil {
				return nil, err
			}
			val, err = vm.stack.Pop()
			if err != nil {
				return nil, err
			}

			vm.environment.Set(name.Inspect(), val)
			ip++

			// maths & comparisons
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod, code.OpPower, code.OpLess, code.OpLessEqual, code.OpGreater, code.OpGreaterEqual, code.OpEqual, code.OpNotEqual, code.OpMatches, code.OpNotMatches, code.OpAnd, code.OpOr:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return nil, err
			}
			ip++

			// !true -> false
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return nil, err
			}
			ip++

			// -1
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return nil, err
			}
			ip++

			// square root
		case code.OpRoot:
			err := vm.executeSquareRoot()
			if err != nil {
				return nil, err
			}
			ip++

			// Boolean literal
		case code.OpTrue:
			vm.stack.Push(True)
			ip++

			// Boolean literal
		case code.OpFalse:
			vm.stack.Push(False)
			ip++

			// return from script
		case code.OpReturn:
			result, err := vm.stack.Pop()
			return result, err

			// flow-control: unconditional jump
		case code.OpJump:
			pos := int(code.ReadUint16(vm.bytecode[ip+1:]))
			ip = pos

			// flow-control: jump if stack contains non-true
		case code.OpJumpIfFalse:
			pos := int(code.ReadUint16(vm.bytecode[ip+1:]))
			ip += 3

			condition, err := vm.stack.Pop()
			if err != nil {
				return nil, err
			}
			if !vm.isTruthy(condition) {
				ip = pos
			}

			// function-call.  This is messy.
		case code.OpCall:

			// OpCall will contain the number of arguments
			// the function is being invoked with.
			args := code.ReadUint16(vm.bytecode[ip+1:])

			// skip the "Call NN,NN"
			ip += 3

			// get the name of the function from the stack.
			fName, err := vm.stack.Pop()
			if err != nil {
				return nil, err
			}

			// construct an array with the arguments we were
			// passed.
			var fArgs []object.Object
			for args > 0 {
				a, err := vm.stack.Pop()
				if err != nil {
					return nil, err
				}
				fArgs = append(fArgs, a)
				args--
			}

			for left, right := 0, len(fArgs)-1; left < right; left, right = left+1, right-1 {
				fArgs[left], fArgs[right] = fArgs[right], fArgs[left]
			}

			// Get the function we're to invoke
			fn, ok := vm.environment.GetFunction(fName.Inspect())
			if !ok {
				return nil, fmt.Errorf("the function %s does not exist", fName.Inspect())
			}

			// Cast the function
			out := fn.(func(args []object.Object) object.Object)

			// Call it
			ret := out(fArgs)

			// Store the result back on our stack.
			vm.stack.Push(ret)
		default:
			fmt.Printf("Unknown opcode: %v", op)
		}
	}

	//
	// If we get here we've hit the end of the bytecode, and we
	// didn't encounter a return-instruction.
	//
	// That means the script is malformed..
	//
	return nil, fmt.Errorf("missing return at the end of the script")
}

// inspectObject discovers the names/values of all structure fields, or
// map contents.  We do this before we begin running our bytecode.
//
// We call this the first time a lookup is attempted against our object
// which means once called all values should be present.
func (vm *VM) inspectObject(obj interface{}) {

	//
	// If the reference is nil we have nothing to walk.
	//
	if obj == nil {
		return
	}

	//
	// Get the value
	//
	val := reflect.Indirect(reflect.ValueOf(obj))

	//
	// Is this a map?
	//
	if val.Kind() == reflect.Map {

		//
		// Get all keys
		//
		for _, key := range val.MapKeys() {
			name := key.Interface().(string)
			field := val.MapIndex(key).Elem()

			var ret object.Object
			ret = Null

			switch field.Kind() {
			case reflect.Int, reflect.Int64:
				ret = &object.Integer{Value: field.Int()}
			case reflect.Float32, reflect.Float64:
				ret = &object.Float{Value: field.Float()}
			case reflect.String:
				ret = &object.String{Value: field.String()}
			case reflect.Bool:
				ret = &object.Boolean{Value: field.Bool()}
			}

			vm.fields[name] = ret
		}
		return
	}

	//
	// OK this is an object
	//
	for i := 0; i < val.NumField(); i++ {

		field := val.Field(i)
		typeField := val.Type().Field(i)
		name := typeField.Name

		var ret object.Object
		ret = Null

		switch field.Kind() {
		case reflect.Int, reflect.Int64:
			ret = &object.Integer{Value: field.Int()}
		case reflect.Float32, reflect.Float64:
			ret = &object.Float{Value: field.Float()}
		case reflect.String:
			ret = &object.String{Value: field.String()}
		case reflect.Bool:
			ret = &object.Boolean{Value: field.Bool()}
		}

		vm.fields[name] = ret
	}
}

// Execute an operation against two arguments, i.e "foo == bar", "2 + 3", etc.
//
// This is a crazy-big function, because we have to cope with different operand
// types and operators.
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	var left object.Object
	var right object.Object
	var err error

	right, err = vm.stack.Pop()
	if err != nil {
		return err
	}
	left, err = vm.stack.Pop()
	if err != nil {
		return err
	}

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.evalIntegerInfixExpression(op, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return vm.evalFloatInfixExpression(op, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.evalFloatIntegerInfixExpression(op, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return vm.evalIntegerFloatInfixExpression(op, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return vm.evalStringInfixExpression(op, left, right)
	case op == code.OpAnd:
		// if left is false skip right
		l := vm.objectToNativeBoolean(left)
		if !l {
			vm.stack.Push(False)
			return nil

		}
		r := vm.objectToNativeBoolean(right)
		if r {
			vm.stack.Push(True)
		} else {
			vm.stack.Push(False)
		}
		return nil
	case op == code.OpOr:
		// if left is true skip right
		l := vm.objectToNativeBoolean(left)
		if l {
			vm.stack.Push(True)
			return nil
		}
		r := vm.objectToNativeBoolean(right)
		if r {
			vm.stack.Push(True)
		} else {
			vm.stack.Push(False)
		}
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return vm.evalBooleanInfixExpression(op, left, right)
	case left.Type() != right.Type():
		return fmt.Errorf("type mismatch: %s %s %s",
			left.Type(), code.String(op), right.Type())
	default:
		return fmt.Errorf("unknown operator: %s %s %s",
			left.Type(), code.String(op), right.Type())
	}

	return nil
}

// integer OP integer
func (vm *VM) evalIntegerInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		vm.stack.Push(&object.Integer{Value: leftVal + rightVal})
	case code.OpSub:
		vm.stack.Push(&object.Integer{Value: leftVal - rightVal})
	case code.OpMul:
		vm.stack.Push(&object.Integer{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.stack.Push(&object.Integer{Value: leftVal / rightVal})
	case code.OpMod:
		vm.stack.Push(&object.Integer{Value: leftVal % rightVal})
	case code.OpPower:
		vm.stack.Push(&object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))})
	case code.OpLess:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %s %s", left.Type(), code.String(op), right.Type()))
	}

	return nil
}

// float OP float
func (vm *VM) evalFloatInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch op {
	case code.OpAdd:
		vm.stack.Push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.stack.Push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.stack.Push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.stack.Push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.stack.Push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.stack.Push(&object.Float{Value: math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %s %s", left.Type(), code.String(op), right.Type()))
	}

	return nil
}

// float OP int
func (vm *VM) evalFloatIntegerInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Float).Value
	rightVal := float64(right.(*object.Integer).Value)

	switch op {
	case code.OpAdd:
		vm.stack.Push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.stack.Push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.stack.Push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.stack.Push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.stack.Push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.stack.Push(&object.Float{Value: math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %s %s", left.Type(), code.String(op), right.Type()))
	}

	return nil
}

// int OP float
func (vm *VM) evalIntegerFloatInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value

	switch op {
	case code.OpAdd:
		vm.stack.Push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.stack.Push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.stack.Push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.stack.Push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.stack.Push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.stack.Push(&object.Float{Value: math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %s %s", left.Type(), code.String(op), right.Type()))
	}

	return nil
}

// string OP string
func (vm *VM) evalStringInfixExpression(op code.Opcode, left object.Object, right object.Object) error {
	l := left.(*object.String)
	r := right.(*object.String)

	switch op {
	case code.OpEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value == r.Value))
	case code.OpNotEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value != r.Value))
	case code.OpGreaterEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value >= r.Value))
	case code.OpGreater:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value > r.Value))
	case code.OpLessEqual:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value <= r.Value))
	case code.OpLess:
		vm.stack.Push(vm.nativeBoolToBooleanObject(l.Value < r.Value))
	case code.OpMatches:
		vm.stack.Push(vm.nativeBoolToBooleanObject(strings.Contains(l.Value, r.Value)))
	case code.OpNotMatches:
		vm.stack.Push(vm.nativeBoolToBooleanObject(!strings.Contains(l.Value, r.Value)))
	case code.OpAdd:
		vm.stack.Push(&object.String{Value: l.Value + r.Value})
	default:
		return (fmt.Errorf("unknown operator: %s %s %s", left.Type(), code.String(op), right.Type()))
	}

	return nil
}

// bool OP bool
func (vm *VM) evalBooleanInfixExpression(op code.Opcode, left object.Object, right object.Object) error {
	// convert the bools to strings.
	l := &object.String{Value: left.Inspect()}
	r := &object.String{Value: right.Inspect()}

	return (vm.evalStringInfixExpression(op, l, r))
}

// Implement !.
func (vm *VM) executeBangOperator() error {
	operand, err := vm.stack.Pop()
	if err != nil {
		return err
	}

	switch operand {
	case True:
		vm.stack.Push(False)
	case False:
		vm.stack.Push(True)
	case Null:
		vm.stack.Push(True)
	default:
		vm.stack.Push(False)
	}
	return nil
}

// Allow negative numbers.
func (vm *VM) executeMinusOperator() error {
	operand, err := vm.stack.Pop()
	if err != nil {
		return err
	}
	var res object.Object

	switch obj := operand.(type) {
	case *object.Integer:
		res = &object.Integer{Value: -obj.Value}
	case *object.Float:
		res = &object.Float{Value: -obj.Value}
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	vm.stack.Push(res)
	return nil
}

// The square root operation is just too cute :).
func (vm *VM) executeSquareRoot() error {
	operand, err := vm.stack.Pop()
	if err != nil {
		return err
	}
	var res object.Object

	switch obj := operand.(type) {
	case *object.Integer:
		res = &object.Float{Value: math.Sqrt(float64(obj.Value))}
	case *object.Float:
		res = &object.Float{Value: math.Sqrt(obj.Value)}
	default:
		return fmt.Errorf("unsupported type for square-root: %s", operand.Type())
	}

	vm.stack.Push(res)
	return nil
}

// convert an object to a native (go) boolean
func (vm *VM) objectToNativeBoolean(o object.Object) bool {

	switch obj := o.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.String:
		return obj.Value != ""
	case *object.Null:
		return false
	case *object.Integer:
		if obj.Value == 0 {
			return false
		}
		return true
	case *object.Float:
		if obj.Value == 0.0 {
			return false
		}
		return true
	default:
		return true
	}
}

// convert a native (go) boolean to an Object
func (vm *VM) nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

// Does the given object contain a "true-like" value?
func (vm *VM) isTruthy(obj object.Object) bool {

	//
	// Is this a boolean object?
	//
	// If so look for the stored value.
	//
	switch tmp := obj.(type) {
	case *object.Boolean:
		return tmp.Value
	case *object.String:
		return (tmp.Value != "")
	case *object.Null:
		return false
	case *object.Integer:
		return (tmp.Value != 0)
	case *object.Float:
		return (tmp.Value != 0.0)
	}

	//
	// If not we return based on our constants.
	//
	switch obj {
	case Null:
		return false
	case True:
		return true
	case False:
		return false
	default:
		return true
	}
}

// lookup the name of the given field/map-member.
func (vm *VM) lookup(obj interface{}, name string) object.Object {

	//
	// Remove legacy "$" prefix, if present.
	//
	name = strings.TrimPrefix(name, "$")

	//
	// Look for this as a variable first, they take precedence.
	//
	if val, ok := vm.environment.Get(name); ok {
		return val
	}

	//
	// Now we assume this is a reference to a map-key, or
	// object member.
	//
	// If we've not discovered them then do so now
	//
	if len(vm.fields) == 0 {
		vm.inspectObject(obj)
	}

	//
	// Then perform the lookup
	//
	if cached, found := vm.fields[name]; found {
		return cached
	}

	//
	// If it was not found it is an unknown/unset value.
	//
	return Null
}
