// This file contains the code which walks the AST, which was created by
// the parser, and generates our bytecode-program, along with appropriate
// constants.

package evalfilter

import (
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/skx/evalfilter/v2/ast"
	"github.com/skx/evalfilter/v2/code"
	"github.com/skx/evalfilter/v2/environment"
	"github.com/skx/evalfilter/v2/object"
)

// compile is core-code for converting the AST into a series of bytecodes.
func (e *Eval) compile(node ast.Node) error {

	switch node := node.(type) {

	case *ast.Program:
		for _, s := range node.Statements {
			err := e.compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := e.compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.BooleanLiteral:
		if node.Value {
			e.emit(code.OpTrue)
		} else {
			e.emit(code.OpFalse)
		}

	case *ast.FloatLiteral:
		str := &object.Float{Value: node.Value}
		e.emit(code.OpConstant, e.addConstant(str))

	case *ast.IntegerLiteral:

		// Get the value of the literal
		v := node.Value

		// If this is an integer between 0 & 65535 we
		// can push it naturally.
		if v%1 == 0 && v >= 0 && v <= 65534 {
			e.emit(code.OpPush, int(v))
		} else {

			//
			// Otherwise we emit it as a constant
			// to our pool.
			//
			integer := &object.Integer{Value: node.Value}
			e.emit(code.OpConstant, e.addConstant(integer))
		}

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		e.emit(code.OpConstant, e.addConstant(str))

	case *ast.RegexpLiteral:

		// The regexp body
		val := node.Value

		// The regexp flags
		if node.Flags != "" {

			// Which we pretend were part of the body
			// because that is what Golang expects.
			val = "(?" + node.Flags + ")" + val
		}

		// The value + flags
		reg := &object.Regexp{Value: val}
		e.emit(code.OpConstant, e.addConstant(reg))

	case *ast.ArrayLiteral:
		for _, el := range node.Elements {
			err := e.compile(el)
			if err != nil {
				return err
			}
		}
		e.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteral:
		keys := []ast.Expression{}

		// get the keys
		for k := range node.Pairs {
			keys = append(keys, k)
		}

		// sort them
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		// for each key + value compile them
		for _, k := range keys {
			err := e.compile(k)
			if err != nil {
				return err
			}
			err = e.compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}

		// Now the number of key+values we've saved
		e.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.ReturnStatement:
		err := e.compile(node.ReturnValue)
		if err != nil {
			return err
		}
		e.emit(code.OpReturn)

	case *ast.ExpressionStatement:
		err := e.compile(node.Expression)
		if err != nil {
			return err
		}

	case *ast.InfixExpression:
		err := e.compile(node.Left)
		if err != nil {
			return err
		}

		err = e.compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {

		// mutators
		// special-handling here:
		//    foo += 3;
		//    -> foo
		//    -> 3
		//    OpAdd
		//    OpSet foo
		//
		case "+=", "-=", "*=", "/=":

			l, ok := node.Left.(*ast.Identifier)
			if !ok {
				return fmt.Errorf("left-most operand for %s must be an identifier", node.Operator)
			}
			if node.Operator == "+=" {
				e.emit(code.OpAdd)
			}
			if node.Operator == "-=" {
				e.emit(code.OpSub)
			}
			if node.Operator == "*=" {
				e.emit(code.OpMul)
			}
			if node.Operator == "/=" {
				e.emit(code.OpDiv)
			}
			str := &object.String{Value: l.Token.Literal}

			e.emit(code.OpConstant, e.addConstant(str))

			// And make it work.
			e.emit(code.OpSet)

		// maths
		case "+":
			e.emit(code.OpAdd)
		case "-":
			e.emit(code.OpSub)
		case "*":
			e.emit(code.OpMul)
		case "/":
			e.emit(code.OpDiv)
		case "%":
			e.emit(code.OpMod)
		case "**":
			e.emit(code.OpPower)

			// comparisons
		case "<":
			e.emit(code.OpLess)
		case "<=":
			e.emit(code.OpLessEqual)
		case ">":
			e.emit(code.OpGreater)
		case ">=":
			e.emit(code.OpGreaterEqual)
		case "==":
			e.emit(code.OpEqual)
		case "!=":
			e.emit(code.OpNotEqual)

			// special matches - regexp and array membership
		case "~=":
			e.emit(code.OpMatches)
		case "!~":
			e.emit(code.OpNotMatches)
		case "in":
			e.emit(code.OpArrayIn)

			// misc
		case "..":
			e.emit(code.OpRange)

			// logical operators
		case "&&":
			e.emit(code.OpAnd)
		case "||":
			e.emit(code.OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		err := e.compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			e.emit(code.OpBang)
		case "-":
			e.emit(code.OpMinus)
		case "√":
			e.emit(code.OpSquareRoot)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PostfixExpression:

		if node.Operator == "++" {
			name := &object.String{Value: node.Token.Literal}
			e.emit(code.OpInc, e.addConstant(name))
		} else if node.Operator == "--" {
			name := &object.String{Value: node.Token.Literal}
			e.emit(code.OpDec, e.addConstant(name))
		} else {
			return fmt.Errorf("unknown postfix operator %s", node.Operator)
		}

	case *ast.LocalVariable:

		// get the name and declare it as local.
		e.emit(code.OpConstant, e.addConstant(&object.String{Value: node.Token.Literal}))
		e.emit(code.OpLocal)

	case *ast.ForeachStatement:

		// Put the array on the stack
		err := e.compile(node.Value)
		if err != nil {
			return err
		}

		// Reset the iteration state of the
		// object - in case it has been iterated
		// over in the post.
		e.emit(code.OpIterationReset)

		// Now we're at the start of our loop,
		// we'll jump back to this point each
		// time round.
		start := len(e.instructions)

		// Store the name of the index-variable.
		str := &object.String{Value: node.Index}
		e.emit(code.OpConstant, e.addConstant(str))

		// Set the name of the body-variable
		str = &object.String{Value: node.Ident}
		e.emit(code.OpConstant, e.addConstant(str))

		// Get the next piece of the iterable,
		// or push False if that fails..
		e.emit(code.OpIterationNext)

		// jump end
		end := e.emit(code.OpJumpIfFalse, 9999)

		// Output the body
		err = e.compile(node.Body)
		if err != nil {
			return nil
		}

		// repeat
		e.emit(code.OpJump, start)

		// back-patch
		e.changeOperand(end, len(e.instructions))

		return nil

	case *ast.FunctionDefinition:

		//
		// Hack: Reset the instructions.
		//
		// What we're doing here is ensuring that
		// we start compiling each function-body as
		// a new set of bytecode.
		//
		// Because things like `if` and our `iterators`
		// have offsets in the generated bytecode we're
		// going to end up with a chunk of bytecode
		// for each function that starts from offset
		// ZERO.
		//
		// So:
		//    blah ..
		//    blah ..
		//    function foo() { ... }
		//    blah ..
		//    blah ..
		//
		// Will _ALWAYS_ result in a new set of bytecode
		// for the function that has an instruction pointer
		// starting at offset ZERO.  Regardless of the length
		// of any preceding bytecode that has already been
		// generated.
		//
		// This is hacky, but it is also safe, because we're
		// single-threaded.  We CANNOT compile N-function
		// definitions at the same time.  We'll only do so
		// sequentially, and nothing else will mess with
		// vm.instructions behind our back.
		//
		before := e.instructions
		e.instructions = code.Instructions{}

		// Compile the body of the function
		err := e.compile(node.Body)
		if err != nil {

			// reset our instructions if we
			// have an error.
			//
			// This is not required as errors
			// will cause termination of our
			// compiler-function but it feels
			// like a neat thing to do.
			e.instructions = before
			return err
		}

		//
		// Ensure that every function will return something.
		//
		// We're doing this because we'll be executing the
		// compiled functions in (essentially) a child-VM.
		//
		// Our VM will terminate execution when it hits a
		// return-statement - so this guarantees that will
		// happen even in the case of a function like:
		//
		//    function alive() { printf("We're alive now\n" ); }
		//
		// Without an explicit return there is .. no return
		// value, and no clean termination.  Instead we'd walk
		// off the end of our bytecode array.
		//
		if len(e.instructions) == 0 ||
			code.Opcode(e.instructions[len(e.instructions)-1]) != code.OpReturn {
			e.emit(code.OpVoid)
			e.emit(code.OpReturn)
		}

		// Save the bytecode away, remember we generated
		// in our "internal" instruction space, which we
		// swapped out for safety.
		x := environment.UserFunction{Bytecode: e.instructions}

		// Copy the function-arguments.
		for _, nm := range node.Parameters {
			x.Arguments = append(x.Arguments, nm.Value)
		}

		// And save this function-reference by name.
		e.functions[node.Token.Literal] = x

		// Now we can restore our bytecode to what it was
		// before we started to deal with the body.
		e.instructions = before

	case *ast.IfExpression:

		// Compile the expression.
		err := e.compile(node.Condition)
		if err != nil {
			return err
		}

		//
		//  Assume the following input:
		//
		//    if ( blah ) {
		//       // A
		//    }
		//    else {
		//       // B
		//    }
		//    // C
		//
		// We've now compiled `blah`, which is the expression
		// above.
		//
		// So now we generate an `OpJumpIfFalse` to handle the case
		// where the if statement is not true. (If the `blah` condition
		// was true we just continue running it ..)
		//
		// Then the jump we're generating here will jump to either
		// B - if there is an else-block - or C if there is not.
		//
		jumpNotTruthyPos := e.emit(code.OpJumpIfFalse, 9999)

		//
		// Compile the code in block A
		//
		err = e.compile(node.Consequence)
		if err != nil {
			return err
		}

		//
		// Here we're calculating the length END of A.
		//
		// Because if the expression was false we want to
		// jump to the START of B.
		//
		afterConsequencePos := len(e.instructions)
		e.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		//
		// If we don't have an `else` block then we're done.
		//
		// If we do then the end of the A-block needs to jump
		// to C - to skip over the else-branch.
		//
		// If there is no else block then we're all good, we only
		// needed to jump over the first block if the condition
		// was not true - and we've already handled that case.
		//
		if node.Alternative != nil {

			//
			// Add a jump to the end of A - which will
			// take us to C.
			//
			// Emit an `OpJump` with a bogus value
			jumpPos := e.emit(code.OpJump, 9999)

			//
			// We're jumping to the wrong place here,
			// so we have to cope with the updated target
			//
			// (We're in the wrong place because we just
			// added a JUMP at the end of A)
			//
			afterConsequencePos = len(e.instructions)
			e.changeOperand(jumpNotTruthyPos, afterConsequencePos)

			//
			// Compile the block
			//
			err := e.compile(node.Alternative)
			if err != nil {
				return err
			}

			//
			// Now we change the offset to be C, which
			// is the end of B.
			//
			afterAlternativePos := len(e.instructions)
			e.changeOperand(jumpPos, afterAlternativePos)
		}

		//
		// Hopefully that is clear.
		//
		// We end up with a simple case where there is no else-clause:
		//
		//   if ..
		//     JUMP IF NOT B:
		//     body
		//     body
		// B:
		//
		// And when there are both we have a pair of jumps:
		//
		//   if ..
		//     JUMP IF NOT B:
		//     body
		//     body
		//     JUMP C:
		//
		//  B: // else clause
		//     body
		//     body
		//     // fall-through
		//  C:
		//

	case *ast.TernaryExpression:

		//
		// We'll have
		//
		//   x = COND ? bar : baz
		//
		// We'll emit
		//
		//   if ! COND  jmp BAZ
		//      bar
		//      jmp END
		//  BAZ:
		//      baz
		//  END:
		//

		//
		// Compile COND
		//
		err := e.compile(node.Condition)
		if err != nil {
			return err
		}

		//
		// Jump to BAZ if this fails - placeholder
		//
		jumpNotTruthyPos := e.emit(code.OpJumpIfFalse, 9999)

		//
		// Compile the bar-code
		//
		err = e.compile(node.IfTrue)
		if err != nil {
			return err
		}

		//
		// Jump to the end of this statement - placeholder
		//
		jumpEnd := e.emit(code.OpJump, 9999)

		//
		// Now we're at the start of the baz-handler
		// we can update our `jump false` to come to this
		// offset.
		//
		e.changeOperand(jumpNotTruthyPos, len(e.instructions))

		//
		// Compile the baz-code
		//
		err = e.compile(node.IfFalse)
		if err != nil {
			return err
		}

		//
		// And now we know our ending position.
		//
		// Update our placeholder.
		//
		e.changeOperand(jumpEnd, len(e.instructions))

	case *ast.SwitchExpression:

		//
		// So a switch statement will look like this:
		//
		//   switch( foo ) {
		//     case "one" {
		//         one_code;
		//         ..
		//     }
		//     case "two" {
		//         two_code;
		//         ..
		//     }
		//     default {
		//         fail_code;
		//
		//   }
		//
		// We want to compile that into:
		//
		//   if ! foo eq one
		//     jmp two
		//       one_code
		//       jmp END
		//  two:
		//   if ! foo eq two
		//     jmp three
		//       two_code
		//       jmp END
		//  three:
		//  default:
		//     fail_code
		//  END:
		//
		//
		// NOTE: This means that multiple cases cannot match.
		//
		//       This is because at the end of each block
		//       we add a jump to the position AFTER the default
		//       block.
		//
		//       TLDR: We run either ONE block, or the default.
		//             We cannot run multiple matches.
		//       is valid:
		//
		//   switch (foo ) {
		//      // The first case wins.
		//     case 1 { printf("ONE\n"); }
		//     case 1 { printf("ONE - again\n"); }
		//   }
		//
		//
		//

		//
		//
		patches := []int{}

		// We have to assemble each choice
		for _, opt := range node.Choices {

			// skipping the default-case, which we'll
			// handle later.
			if opt.Default {
				continue
			}

			// Look at any expression we've got in this case.
			for _, val := range opt.Expr {

				// OK so we have an expression.
				//
				// Emit "test Value Expression"
				// jump if false to end

				// compile the thing we're testing
				err := e.compile(node.Value)
				if err != nil {
					return err
				}

				// now compile the express
				err = e.compile(val)
				if err != nil {
					return err
				}

				// the comparison
				e.emit(code.OpCase)

				// Now the jump over the block to run
				// if it matches - for the case when it
				// actually DOESN'T.
				pos := e.emit(code.OpJumpIfFalse, 9999)

				// finally the block
				err = e.compile(opt.Block)
				if err != nil {
					return err
				}

				// And jump to after the default-block
				//
				end := e.emit(code.OpJump, 9999)
				patches = append(patches, end)

				// now we know the end of the block.
				e.changeOperand(pos, len(e.instructions))
			}

		}

		//
		// Now the default-block
		//
		for _, opt := range node.Choices {
			if !opt.Default {
				continue
			}

			// Compile the block
			err := e.compile(opt.Block)
			if err != nil {
				return err
			}
		}

		// And now we're after the default - if there wasn't on,
		// or after the last choice if here wasn't.
		for _, offset := range patches {
			e.changeOperand(offset, len(e.instructions))
		}

		// Finally add a NOP
		// Because our "jmp END" will jump to an instruction which
		// doesn't exist otherwise
		e.emit(code.OpTrue)
		e.emit(code.OpPop)

	case *ast.WhileStatement:

		//
		// Record our starting position
		//
		cur := len(e.instructions)

		//
		// Compile the condition.
		//
		err := e.compile(node.Condition)
		if err != nil {
			return err
		}

		//
		//  Assume the following input:
		//
		//    // A
		//    while ( cond ) {
		//       // B
		//       statement(s)
		//       // b2 -> jump to A to retest the condition
		//    }
		//    // C
		//
		// We've now compiled `cond`, which is the expression
		// above.
		//
		// If the condition is false we jump to C, skipping the
		// body.
		//
		// If the condition is true we fall through, and at
		// B2 we jump back to A
		//

		//
		// So now we generate an `OpJumpIfFalse` to handle the case
		// where the condition is not true.
		//
		// This will jump to C, the position after the body.
		//
		jumpNotTruthyPos := e.emit(code.OpJumpIfFalse, 9999)

		//
		// Compile the code in the body
		//
		err = e.compile(node.Body)
		if err != nil {
			return err
		}

		//
		// Append the b2 jump to retry the loop
		//
		e.emit(code.OpJump, cur)

		//
		// Change the jump to skip the block if the condition
		// was false.
		//
		e.changeOperand(jumpNotTruthyPos, len(e.instructions))

	case *ast.AssignStatement:

		// Get the value
		err := e.compile(node.Value)
		if err != nil {
			return err
		}

		// Store the name
		str := &object.String{Value: node.Name.String()}
		e.emit(code.OpConstant, e.addConstant(str))

		// And make it work.
		e.emit(code.OpSet)

	case *ast.Identifier:
		str := &object.String{Value: node.Value}
		e.emit(code.OpLookup, e.addConstant(str))

	case *ast.CallExpression:

		//
		// call to print(1) will have the stack setup as:
		//
		//  1
		//  print
		//  call 1
		//
		// call to print( "steve", "kemp" ) will have:
		//
		//  "steve"
		//  "kemp"
		//  "print"
		//  call 2
		//
		// i.e. We store the arguments on the stack and
		// emit `OpCall NN` where NN is the number of arguments
		// to pop and invoke the function with.
		//
		args := len(node.Arguments)
		for _, a := range node.Arguments {

			err := e.compile(a)
			if err != nil {
				return err
			}
		}

		// call - has the string on the stack
		str := &object.String{Value: node.Function.String()}
		e.emit(code.OpConstant, e.addConstant(str))

		// then a call instruction with the number of args.
		e.emit(code.OpCall, args)

	case *ast.IndexExpression:
		err := e.compile(node.Left)
		if err != nil {
			return err
		}

		err = e.compile(node.Index)
		if err != nil {
			return err
		}

		e.emit(code.OpIndex)

	default:
		return fmt.Errorf("unknown node type %T %v", node, node)
	}
	return nil
}

// addConstant adds a constant to the pool
func (e *Eval) addConstant(obj object.Object) int {

	//
	// Look to see if the constant is present already
	//
	for i, c := range e.constants {

		//
		// If the existing constant has the same
		// type and value - then return the offset.
		//
		if c.Type() == obj.Type() &&
			c.Inspect() == obj.Inspect() {
			return i
		}
	}

	//
	// Otherwise this is a distinct constant and should
	// be added.
	//
	e.constants = append(e.constants, obj)
	return len(e.constants) - 1
}

// emit generates a bytecode operation, and adds it to our program-array.
func (e *Eval) emit(op code.Opcode, operands ...int) int {

	ins := make([]byte, 1)
	ins[0] = byte(op)

	if len(operands) == 1 {

		// Make a buffer for the arg
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(operands[0]))

		// append
		ins = append(ins, b...)
	}

	posNewInstruction := len(e.instructions)
	e.instructions = append(e.instructions, ins...)

	return posNewInstruction
}

// changeOperand is designed to patch the operand of
// an instruction.
//
// This function is used to rewrite the target of our jump
// instructions in the handling of `if`, `while` and our
// ternary expressions.
func (e *Eval) changeOperand(opPos int, operand int) {

	//
	// We're pointed at the instruction,
	// so our offset will be something like
	//
	//        OpBlah
	// opPos: OpJump target
	//
	// In terms of our bytecode that translates
	// to our e.instructions containing something
	// like this:
	//
	//  ..
	//  e.instructions[opPos]   = OpJump
	//  e.instructions[opPos+1] = arg1
	//  e.instructions[opPos+2] = arg2
	//  ..

	//
	// So we ignore the opcode, which doesn't
	// change, and just update the argument-bytes
	// in-place.
	//

	// Make a buffer for the arg, which we can
	// use to split it into two bytes.
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(operand))

	// replace the argument in-place
	e.instructions[opPos+1] = b[0]
	e.instructions[opPos+2] = b[1]
}
