// The vm package implements a simple virtual machine.
//
// We're constructed with a set of opcodes, and we process those forever,
// until we hit a `return` statement which terminates our run.
package vm

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/skx/evalfilter/code"
	"github.com/skx/evalfilter/object"
)

const StackSize = 2048

var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

// VM is the structure which holds our state.
type VM struct {

	// constants holds constants in the program source
	// for example the script "1 + 3;"  refers to two
	// numeric constants "1" and "3".
	constants []object.Object

	// bytecode contains the actual instructions we'll execute
	bytecode []byte

	// stack stores our stack.
	// We're a stack-based virtual machine so this is used for many
	// many of our internal facilities.
	stack []object.Object

	// environment holds the environment, which will allow variables
	// and functions to be get/set.
	environment *object.Environment

	// sp contains our stack-pointer
	sp int
}

// New constructs a new virtual machine.
func New(constants []object.Object, bytecode []byte, env *object.Environment) *VM {

	return &VM{
		constants:   constants,
		environment: env,
		bytecode:    bytecode,
		stack:       make([]object.Object, StackSize),
		sp:          0,
	}
}

// Run launches our virtual machine, intepreting the bytecode we were
// constructed with.
//
// We return when we hit a return-operation, or we hit the end of
// the supplied bytecode - the latter is an error.
func (vm *VM) Run(obj interface{}) (object.Object, error) {

	ip := 0
	ln := len(vm.bytecode)

	for ip < ln {

		op := code.Opcode(vm.bytecode[ip])

		switch op {

		// store constant
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.bytecode[ip+1:])
			ip += 3

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return nil, err
			}

			// Lookup variable/field, by name
		case code.OpLookup:
			constIndex := code.ReadUint16(vm.bytecode[ip+1:])
			ip += 3

			// Get the name.
			name := vm.constants[constIndex].Inspect()

			val := vm.lookup(obj, name)
			vm.push(val)

			// Set a variable by name
		case code.OpSet:
			name := vm.pop()
			val := vm.pop()

			vm.environment.Set(name.Inspect(), val)
			ip += 1

			// maths & comparisions
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod, code.OpPower, code.OpLess, code.OpLessEqual, code.OpGreater, code.OpGreaterEqual, code.OpEqual, code.OpNotEqual, code.OpMatches, code.OpNotMatches, code.OpAnd, code.OpOr:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return nil, err
			}
			ip += 1

			// !true -> false
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return nil, err
			}
			ip += 1

			// -1
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return nil, err
			}
			ip += 1

			// square root
		case code.OpRoot:
			err := vm.executeSquareRoot()
			if err != nil {
				return nil, err
			}
			ip += 1

			// Boolean literal
		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return nil, err
			}
			ip += 1

			// Boolean literal
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return nil, err
			}
			ip += 1

			// return from script
		case code.OpReturn:
			result := vm.pop()
			return result, nil

			// flow-control: unconditional jump
		case code.OpJump:
			pos := int(code.ReadUint16(vm.bytecode[ip+1:]))
			ip = pos

			// flow-control: jump if stack contains non-true
		case code.OpJumpIfFalse:
			pos := int(code.ReadUint16(vm.bytecode[ip+1:]))
			ip += 3

			condition := vm.pop()
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
			fName := vm.pop()

			// construct an array with the arguments we were
			// passed.
			var fArgs []object.Object
			for args > 0 {
				fArgs = append(fArgs, vm.pop())
				args--
			}

			for left, right := 0, len(fArgs)-1; left < right; left, right = left+1, right-1 {
				fArgs[left], fArgs[right] = fArgs[right], fArgs[left]
			}

			// Get the function we're to invoke
			fn := vm.environment.GetFunction(fName.Inspect())
			if fn == nil {
				return nil, fmt.Errorf("the function %s does not exist", fName.Inspect())
			}

			// Cast the function
			out := fn.(func(args []object.Object) object.Object)

			// Call it
			ret := out(fArgs)

			// Store the result back on our stack.
			vm.push(ret)
		default:
			fmt.Printf("Unknown opcode: %v", op)
		}
	}

	return nil, fmt.Errorf("Unhandled return!")
}

// Push a value upon our stack.
func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// pop a value from our stack.
func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

// Execute an operation against two arguments, i.e "foo == bar", "2 + 3", etc.
//
// This is a crazy-big function, because we have to cope with different operand
// types and operators.
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

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
			vm.push(False)
			return nil

		}
		r := vm.objectToNativeBoolean(right)
		if r {
			vm.push(True)
		} else {
			vm.push(False)
		}
		return nil
	case op == code.OpOr:
		// if left is true skip right
		l := vm.objectToNativeBoolean(left)
		if l {
			vm.push(True)
			return nil
		}
		r := vm.objectToNativeBoolean(right)
		if r {
			vm.push(True)
		} else {
			vm.push(False)
		}
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return vm.evalBooleanInfixExpression(op, left, right)
	case left.Type() != right.Type():
		return fmt.Errorf("type mismatch: %s %v %s",
			left.Type(), op, right.Type())
	default:
		return fmt.Errorf("unknown operator: %s %v %s",
			left.Type(), op, right.Type())
	}

	return nil
}

// integer OP integer
func (vm *VM) evalIntegerInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case code.OpAdd:
		vm.push(&object.Integer{Value: leftVal + rightVal})
	case code.OpSub:
		vm.push(&object.Integer{Value: leftVal - rightVal})
	case code.OpMul:
		vm.push(&object.Integer{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.push(&object.Integer{Value: leftVal / rightVal})
	case code.OpMod:
		vm.push(&object.Integer{Value: leftVal % rightVal})
	case code.OpPower:
		vm.push(&object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))})
	case code.OpLess:
		vm.push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %v %s", left.Type(), op, right.Type()))
	}

	return nil
}

// float OP float
func (vm *VM) evalFloatInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value

	switch op {
	case code.OpAdd:
		vm.push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.push(&object.Float{Value: math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %v %s", left.Type(), op, right.Type()))
	}

	return nil
}

// float OP int
func (vm *VM) evalFloatIntegerInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Float).Value
	rightVal := float64(right.(*object.Integer).Value)

	switch op {
	case code.OpAdd:
		vm.push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.push(&object.Float{math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %v %s", left.Type(), op, right.Type()))
	}

	return nil
}

// int OP float
func (vm *VM) evalIntegerFloatInfixExpression(op code.Opcode, left, right object.Object) error {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value

	switch op {
	case code.OpAdd:
		vm.push(&object.Float{Value: leftVal + rightVal})
	case code.OpSub:
		vm.push(&object.Float{Value: leftVal - rightVal})
	case code.OpMul:
		vm.push(&object.Float{Value: leftVal * rightVal})
	case code.OpDiv:
		vm.push(&object.Float{Value: leftVal / rightVal})
	case code.OpMod:
		vm.push(&object.Float{Value: float64(int(leftVal) % int(rightVal))})
	case code.OpPower:
		vm.push(&object.Float{Value: math.Pow(leftVal, rightVal)})
	case code.OpLess:
		vm.push(vm.nativeBoolToBooleanObject(leftVal < rightVal))
	case code.OpLessEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal <= rightVal))
	case code.OpGreater:
		vm.push(vm.nativeBoolToBooleanObject(leftVal > rightVal))
	case code.OpGreaterEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal >= rightVal))
	case code.OpEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal == rightVal))
	case code.OpNotEqual:
		vm.push(vm.nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return (fmt.Errorf("unknown operator: %s %v %s", left.Type(), op, right.Type()))
	}

	return nil
}

// string OP string
func (vm *VM) evalStringInfixExpression(op code.Opcode, left object.Object, right object.Object) error {
	l := left.(*object.String)
	r := right.(*object.String)

	switch op {
	case code.OpEqual:
		vm.push(vm.nativeBoolToBooleanObject(l.Value == r.Value))
	case code.OpNotEqual:
		vm.push(vm.nativeBoolToBooleanObject(l.Value != r.Value))
	case code.OpGreaterEqual:
		vm.push(vm.nativeBoolToBooleanObject(l.Value >= r.Value))
	case code.OpGreater:
		vm.push(vm.nativeBoolToBooleanObject(l.Value > r.Value))
	case code.OpLessEqual:
		vm.push(vm.nativeBoolToBooleanObject(l.Value <= r.Value))
	case code.OpLess:
		vm.push(vm.nativeBoolToBooleanObject(l.Value < r.Value))
	case code.OpMatches:
		vm.push(vm.nativeBoolToBooleanObject(strings.Contains(l.Value, r.Value)))
	case code.OpNotMatches:
		vm.push(vm.nativeBoolToBooleanObject(!strings.Contains(l.Value, r.Value)))
	case code.OpAdd:
		vm.push(&object.String{Value: l.Value + r.Value})
	default:
		return (fmt.Errorf("unknown operator: %s %v %s", left.Type(), op, right.Type()))
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
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

// Allow negative numbers.
func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()
	var res object.Object

	switch obj := operand.(type) {
	case *object.Integer:
		res = &object.Integer{Value: -obj.Value}
	case *object.Float:
		res = &object.Float{Value: -obj.Value}
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	return vm.push(res)
}

// The square root operation is just too cute :).
func (vm *VM) executeSquareRoot() error {
	operand := vm.pop()
	var res object.Object

	switch obj := operand.(type) {
	case *object.Integer:
		res = &object.Float{Value: math.Sqrt(float64(obj.Value))}
	case *object.Float:
		res = &object.Float{Value: math.Sqrt(obj.Value)}
	default:
		return fmt.Errorf("unsupported type for square-root: %s", operand.Type())
	}

	return vm.push(res)
}

// convert an object to a native (go) boolean
func (vm *VM) objectToNativeBoolean(o object.Object) bool {
	if r, ok := o.(*object.ReturnValue); ok {
		o = r.Value
	}
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
	// Look for this as a variable before
	// looking for field values.
	//
	if val, ok := vm.environment.Get(name); ok {
		return val
	}

	if obj != nil {

		ref := reflect.ValueOf(obj)

		//
		// The field we find by reflection.
		//
		var field reflect.Value

		//
		// Are we dealing with a map?
		//
		if ref.Kind() == reflect.Map {
			for _, key := range ref.MapKeys() {
				if key.Interface() == name {

					field = ref.MapIndex(key).Elem()
					break
				}
			}
		} else {

			field = reflect.Indirect(ref).FieldByName(name)
		}

		switch field.Kind() {
		case reflect.Int, reflect.Int64:
			return &object.Integer{Value: field.Int()}
		case reflect.Float32, reflect.Float64:
			return &object.Float{Value: field.Float()}
		case reflect.String:
			return &object.String{Value: field.String()}
		case reflect.Bool:
			return &object.Boolean{Value: field.Bool()}
		}
	}

	// Not found
	return Null

}
