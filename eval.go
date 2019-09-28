// Package evalfilter allows running a user-supplied script against an object.
package evalfilter

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/skx/evalfilter/ast"
	"github.com/skx/evalfilter/lexer"
	"github.com/skx/evalfilter/object"
	"github.com/skx/evalfilter/parser"
)

// pre-defined objects; Null, True and False
var (
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
)

// Eval is our public-facing structure which stores our state.
type Eval struct {
	// Input script
	Script string

	// Parser
	Parser *parser.Parser

	// Parsed program
	Program ast.Node

	// Environment
	Environment *object.Environment

	// User-supplied functions
	Functions map[string]interface{}

	// Object we operate upon
	Object interface{}
}

// New creates a new instance of the evaluator.
func New(script string) *Eval {

	//
	// Create our object.
	//
	e := &Eval{
		Environment: object.NewEnvironment(),
		Functions:   make(map[string]interface{}),
		Parser:      parser.New(lexer.New(script)),
	}

	//
	// Add our default functions.
	//
	e.AddFunction("len", fnLen)
	e.AddFunction("match", fnMatch)
	e.AddFunction("print", fnPrint)
	e.AddFunction("trim", fnTrim)
	e.AddFunction("type", fnType)

	//
	// Parse the program we were given.
	//
	e.Program = e.Parser.ParseProgram()

	//
	// Return it.
	//
	return e
}

// Run takes the program which was passed in the constructor, and
// executes it.  The supplied object will be used for performing
// dynamic field-lookups, etc.
func (e *Eval) Run(obj interface{}) (bool, error) {

	//
	// Where there any errors produced by the parser?
	//
	if len(e.Parser.Errors()) > 0 {
		return false, fmt.Errorf("\nErrors parsing script:\n" +
			strings.Join(e.Parser.Errors(), "\n"))
	}

	//
	// Store the object we're working against.
	//
	e.Object = obj

	//
	// Evaluate the program, recursively.
	//
	out := e.EvalIt(e.Program, e.Environment)
	if out == nil {
		return false, nil
	}

	//
	// Ok we ran, and returned a value, convert that
	// to a boolean value.
	//
	return e.isTruthy(out), nil
}

// AddFunction adds a function to our runtime.
//
// Once a function has been added it may be used by the filter script.
func (e *Eval) AddFunction(name string, fun interface{}) {
	e.Environment.Set(name, &object.String{Value: name})
	e.Functions[name] = fun
}

// SetVariable adds, or updates, a variable which will be available
// to the filter script.
func (e *Eval) SetVariable(name string, value object.Object) {
	e.Environment.Set(name, value)
}

// EvalIt is our core function for evaluating nodes.
func (e *Eval) EvalIt(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	//Statements
	case *ast.Program:
		return e.evalProgram(node, env)
	case *ast.ExpressionStatement:
		return e.EvalIt(node.Expression, env)

	//Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.Boolean:
		return e.nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := e.EvalIt(node.Right, env)
		if e.isError(right) {
			return right
		}
		return e.evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := e.EvalIt(node.Left, env)
		if e.isError(left) {
			return left
		}
		right := e.EvalIt(node.Right, env)
		if e.isError(right) {
			return right
		}
		res := e.evalInfixExpression(node.Operator, left, right)
		if e.isError(res) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", res.Inspect())
		}
		return (res)

	case *ast.BlockStatement:
		return e.evalBlockStatement(node, env)

	// `if`, `return`, assignment
	case *ast.IfExpression:
		return e.evalIfExpression(node, env)
	case *ast.ReturnStatement:
		val := e.EvalIt(node.ReturnValue, env)
		if e.isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	// identifiers (names / field-members)
	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	// function-calls
	case *ast.CallExpression:
		function := e.EvalIt(node.Function, env)
		if e.isError(function) {
			return function
		}
		args := e.evalExpression(node.Arguments, env)
		if len(args) == 1 && e.isError(args[0]) {
			return args[0]
		}
		res := e.applyFunction(env, function, args)
		if e.isError(res) {
			fmt.Fprintf(os.Stderr, "Error calling `%s` : %s\n", node.Function, res.Inspect())
		}
		return res
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.AssignStatement:
		return e.evalAssignStatement(node, env)
	default:
		fmt.Fprintf(os.Stderr, "Unknown AST-type: %t %v\n", node, node)
	}
	return nil
}

// eval block statement
func (e *Eval) evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = e.EvalIt(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

// for performance, using single instance of boolean
func (e *Eval) nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// eval prefix expression
func (e *Eval) evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return e.evalBangOperatorExpression(right)
	case "-":
		return e.evalMinusPrefixOperatorExpression(right)
	default:
		return e.newError("unknown operator: %s%s", operator, right.Type())
	}
}

func (e *Eval) evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func (e *Eval) evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return e.newError("unknown operator: -%s", right.Type())
	}
}

func (e *Eval) evalInfixExpression(operator string, left, right object.Object) object.Object {

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalFloatIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return e.evalIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(operator, left, right)
	case operator == "&&":
		return e.nativeBoolToBooleanObject(e.objectToNativeBoolean(left) && e.objectToNativeBoolean(right))
	case operator == "||":
		return e.nativeBoolToBooleanObject(e.objectToNativeBoolean(left) || e.objectToNativeBoolean(right))
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return e.evalBooleanInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return e.newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

// boolean operations
func (e *Eval) evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	// convert the bools to strings.
	l := &object.String{Value: string(left.Inspect())}
	r := &object.String{Value: string(right.Inspect())}

	switch operator {
	case "<":
		return e.evalStringInfixExpression(operator, l, r)
	case "==":
		return e.evalStringInfixExpression(operator, l, r)
	case "!=":
		return e.evalStringInfixExpression(operator, l, r)
	case "<=":
		return e.evalStringInfixExpression(operator, l, r)
	case ">":
		return e.evalStringInfixExpression(operator, l, r)
	case ">=":
		return e.evalStringInfixExpression(operator, l, r)
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Eval) evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return e.nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return e.nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return e.nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return e.nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return e.nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}
func (e *Eval) evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return e.nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return e.nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return e.nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return e.nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return e.nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Eval) evalFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := float64(right.(*object.Integer).Value)
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return e.nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return e.nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return e.nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return e.nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return e.nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Eval) evalIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return e.nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return e.nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return e.nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return e.nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return e.nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return e.newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func (e *Eval) evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	l := left.(*object.String)
	r := right.(*object.String)

	switch operator {
	case "==":
		return e.nativeBoolToBooleanObject(l.Value == r.Value)
	case "!=":
		return e.nativeBoolToBooleanObject(l.Value != r.Value)
	case ">=":
		return e.nativeBoolToBooleanObject(l.Value >= r.Value)
	case ">":
		return e.nativeBoolToBooleanObject(l.Value > r.Value)
	case "<=":
		return e.nativeBoolToBooleanObject(l.Value <= r.Value)
	case "<":
		return e.nativeBoolToBooleanObject(l.Value < r.Value)
	case "~=":
		return e.nativeBoolToBooleanObject(strings.Contains(l.Value, r.Value))
	case "!~":
		return e.nativeBoolToBooleanObject(!strings.Contains(l.Value, r.Value))
	case "+":
		return &object.String{Value: l.Value + r.Value}
	case "+=":
		return &object.String{Value: l.Value + r.Value}
	}

	return e.newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

func (e *Eval) evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := e.EvalIt(ie.Condition, env)
	if e.isError(condition) {
		return condition
	}
	if e.isTruthy(condition) {
		return e.EvalIt(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return e.EvalIt(ie.Alternative, env)
	} else {
		return NULL
	}
}

func (e *Eval) evalAssignStatement(a *ast.AssignStatement, env *object.Environment) (val object.Object) {
	evaluated := e.EvalIt(a.Value, env)
	if e.isError(evaluated) {
		return evaluated
	}
	env.Set(a.Name.String(), evaluated)
	return evaluated
}

func (e *Eval) isTruthy(obj object.Object) bool {

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
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func (e *Eval) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = e.EvalIt(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func (e *Eval) newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func (e *Eval) isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func (e *Eval) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {

	//
	// Can we get a value from the environment?
	//
	// If so return it.
	//

	//
	// Note that for legacy purposes we might have a leading `$`
	// on our variable-names.
	//
	// If so remove it.
	//
	name := node.Value
	name = strings.TrimPrefix(name, "$")

	if val, ok := env.Get(name); ok {
		return val
	}

	//
	// Otherwise we need to perform a structure lookup.
	//
	if e.Object != nil {
		ref := reflect.ValueOf(e.Object)
		field := reflect.Indirect(ref).FieldByName(name)

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
	return NULL
}

func (e *Eval) evalExpression(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, exp := range exps {
		evaluated := e.EvalIt(exp, env)
		if e.isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

//
// Here is where we call a user-supplied function.
//
func (e *Eval) applyFunction(env *object.Environment, fn object.Object, args []object.Object) object.Object {

	// Get the function
	res, ok := e.Functions[fn.Inspect()]
	if !ok {
		fmt.Printf("Failed to find function \n")
		return (&object.String{Value: "Function not found " + fn.Inspect()})
	}

	// Cast it into the correct type, and then invoke it.
	out := res.(func(args []object.Object) object.Object)
	ret := (out(args))

	return ret
}

func (e *Eval) objectToNativeBoolean(o object.Object) bool {
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
