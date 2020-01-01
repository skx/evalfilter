# Bytecode

When the evalfilter package is executed a user-supplied script is lexed, parsed, and transformed into a series of bytecode operations, then these bytecode operations are executed by a simple stack-based virtual machine.

Although we don't expect users to care about the implementation details here are some brief notes.

The opcodes we're discussing are found in [code/code.go](code/code.go), and the interpreter for the virtual machine is contained in [vm/vm.go](vm/vm.go).


* [Bytecode](#bytecode)
  * [Examining Bytecode](#examining-bytecode)
* [Bytecode Overview](#bytecode-overview)
* [Mathematical Operations](#mathematical-operations)
* [Comparison Operations](#comparison-operations)
* [Control-Flow Operations](#control-flow-operations)
* [Misc Operations](#misc-operations)
* [Function Calls](#function-calls)
* [Program Walkthrough](#program-walkthrough)
* [Debugging](#debugging)
* [Optimization](#optimization)


## Examining Bytecode

You may use the `cmd/evalfilter` utility to view the bytecode representation of a program.

For example consider the following script:

```
if ( 1 == 0.5 * 2 ) {
  return true;
}

print( "This is weird\n" );

return false;
```

To view the bytecode you would run `evalfilter bytecode ./simple.txt`, which would produce a result similar to this:

```
Bytecode:
0000	        OpPush	   1	// Push 1 to stack
0003	    OpConstant	   0	// push constant onto stack: "0.5"
0006	        OpPush	   2	// Push 2 to stack
0009	         OpMul
0010	       OpEqual
0011	 OpJumpIfFalse	  16
0014	        OpTrue
0015	      OpReturn
0016	    OpConstant	   1	// push constant onto stack: "This is weird\n"
0019	    OpConstant	   2	// push constant onto stack: "print"
0022	        OpCall	   1	// call function with 1 arg(s)
0025	       OpFalse
0026	      OpReturn


Constant Pool:
0000 Type:FLOAT Value:"0.5"
0001 Type:STRING Value:"This is weird\n"
0002 Type:STRING Value:"print"
```


# Bytecode Overview

Our bytecode interpreter understands approximately 30 different instructions, which are broadly grouped into categories:

* Constant, and field operations.
* Mathematical operations.
* Comparison operations.
* Control-flow operations.
* Misc operations.

The virtual machine I've implemented needs two things to work:

* A list of instructions to execute.
* A list of constants.

For example the program "`print( 1.0 + 2.0 ); return true;`" contains __three__ constants:

* The name of the function `print`.
* The floating point value `1.0`.
* The floating point value `2.0`.
  * See [optimization](#optimization) for details of why this example uses floating-point numbers.

That program would be encoded like so:

```
Bytecode:
0000	    OpConstant	   0	// push constant onto stack: "1"
0003	    OpConstant	   1	// push constant onto stack: "2"
0006	         OpAdd
0007	    OpConstant	   2	// push constant onto stack: "print"
0010	        OpCall	   1	// call function with 1 arg(s)
0013	        OpTrue
0014	      OpReturn


Constant Pool:
0000 Type:FLOAT Value:"1"
0001 Type:FLOAT Value:"2"
0002 Type:STRING Value:"print"
```

To provide more details about the output:

* The value to the left of the instruction is the position in the code.
  * The control-flow instructions generate jumps to these indexes, so they're worth showing.
* The middle field is the instruction to be executed.
  * Some instructions include an argument, but most do not.
  * Some instructions contain helpful comments to the right.
* After the bytecode has been disassembled you'll see the list of constants.
  * Each of which is identified by numeric ID.

In this overview we're focusing upon the instruction `OpConstant`.  The `OpConstant` instruction has a single argument, which is the index of the constant to load and push on the stack.

When we start running the program the stack is empty.

* We run `OpConstant 0`
  * That loads the constant with ID `0` from the constant-area.
    * This is the floating-point number `1.0`.
  * The constant is then pushed upon the stack.
* We run `OpConstant 1`
  * That loads the constant with ID `0` from the constant-area.
    * This is the floating-point number `2.0`.
  * The constant is then pushed upon the stack.
* We then execute the `OpAdd` instruction.
  * This pops two values from the stack
    * i.e. The `2.0` we just added.
    * Then the `1.0` we added.
      * The stack is emptied in reverse.
  * The two values are added, producing a result of `3.0`.
  * Then the value is placed back upon the stack.


# Mathematical Operations

The mathematical-related opcodes all work in the same way, they each pop two operands from the stack, carry out the appropriate instruction, and then push the result back upon the stack.

We saw these described briefly earlier, but the full list of instructions is:

* `OpAdd`
  * Add two numbers.
* `OpSub`
  * Subtract a number from another
* `OpMul`
  * Multiply two numbers.
* `OpDiv`
  * Divide a number by another.
* `OpMod`
  * Calculate a modulus operation
* `OpPower`
  * Raise a number to the power of another.


# Comparison Operations

The comparison operations are very similar to the mathematical operations, and work in the same way:

* They pop two values from the stack.
* They run the comparison operation:
  * If the comparison succeeds they push `true` upon the stack.
  * Otherwise they push `false`.

Comparison operations include:

* `OpEqual`
  * This pushes `true` upon the stack if the two values to be compared are equal.
  * `false` otherwise.
* `OpLess`
  * This pushes `true` upon the stack if the first argument is less than the second.
  * `false` otherwise.

The full list is:

* `OpLess` / `<`
* `OpLessEqual` / `<=`
* `OpGreater` / `>`
* `OpGreaterEqual` / `>=`
* `OpEqual` / `==`
* `OpNotEqual` / `!=`
* `OpMatches` / `~=`
* `OpNotMatches` / `!~`
* `OpArrayIn` / `in`
  * This is an array-specific opcode which tests whether a value is contained within an array.


# Control-Flow Operations

There are two control-flow operations:

* `OpJump`
  * Which takes the offset within the bytecode to jump to.
  * Control flow immediately branches there.
* `OpJumpIfFalse`
  * A value is popped from the stack, if it is false then control moves to the offset specified as the argument.
  * Otherwise we proceed to the next instruction as expected.


# Misc Operations

There are some miscellaneous instructions:

* `OpBang`
  * Calculate negation.
* `OpMinus`
  * Calculate unary minus.
* `OpSquareRoot`
  * Calculate a square root.
* `OpTrue`
  * Pushes a `true` value to the stack.
* `OpFalse`
  * Pushes a `false` value to the stack.
* `OpReturn`
  * Pops a value off the stack and terminates processing.
    * The value taken from the stack is the return-code.
* `OpLookup`
  * Much like loading a constant by reference this loads the value from the structure field with the given name.
* `OpCall`
  * Pops the name of a function to call from the stack.
  * Called with an argument noting how many arguments to pass to the function, and pops that many arguments from the stack to use in the function-call.


# Function Calls

There are several built-in functions supplied with the interpreter, such as `len()`, `print()`, and similar.  Your host application which embeds the library can install more easily too.

The prototype of all functions is:

     func foo( args []object.Object ) object.Object { .. }

i.e. All functions take an array of objects, and return a single object.  The objects allow recieving or returning arrays, strings, numbers, booleans, and errors.

We've already seen how constants can be loaded from the constant area onto the stack, and that along with the `OpCall` instruction is all we need to support calling functions.

The `OpCall` instruction comes with a single operand, which is the number of arguments that should be supplied to the function, these arguments will be popped off the stack as you should have come to expect.


This is an excerpt from the program we saw at the top of this document, and shows a function being called:

```
  000019	OpConstant	3		// load constant: &{This is weird\n}
  000022	OpConstant	4		// load constant: &{print}
  000025	OpCall	1			// call function with 1 arguments
  000028	OpFalse
```

* The first operation loads the constant with ID 3 and pushes it onto the stack.
  * This is the string "`This is weird\n`", as the comment indicates.
* The second instruction loads the constant with ID 4 and pushes it onto the stack.
  * This is the string "`print`", which is the name of the function we're going to invoke.
* The third instruction is `OpCall 1` which means that the machine should call a function with one argument.

The end result of that is that the function call happens:

* `OpCall` pops the first value off the stack.
  * This is the function to invoke.
  * i.e. `print`.
* The argument to `OpCall` is the number of arguments to supply to that function.
  * There is one argument in this example.
  * So one value is popped off the stack.
    * This will be the string `This is weird\n`.
* Now that the arguments are handled the function is invoked.
* The return result from that call is then pushed onto the stack.


# Program Walkthrough

We already demonstrated a simple program earlier, with the following bytecode:

```
Bytecode:
  000000	OpConstant	0		// load constant: &{1}
  000003	OpConstant	1		// load constant: &{0.5}
  000006	OpConstant	2		// load constant: &{2}
  000009	OpMul
  000010	OpEqual
  000011	OpJumpIfFalse	19
  000014	OpTrue
  000015	OpReturn
  000016	OpJump	19
  000019	OpConstant	3		// load constant: &{This is weird\n}
  000022	OpConstant	4		// load constant: &{print}
  000025	OpCall	1			// call function with 1 arguments
  000028	OpFalse
  000029	OpReturn
```

Now we'll walk through what happens:

* We load the value of the constant with ID 0 and push to the stack.
  * The stack now looks like this:  [1]
* We load the value of the constant with ID 1 and push to the stack.
  * The stack now looks like this:  [1 0.5]
* We load the value of the constant with ID 2 and push to the stack.
  * The stack now looks like this:  [1 0.5 2]
* We come across the `OpMul` instruction, which pops two numbers from the stack, multiples them, and adds the result back.
  * The stack now looks like this:  [1 1]
    * The second value there is the result of `0.5 * 2`.
* We see the `Equal` instruction, which pops two items from the stack, and compares them.
  * `1` and `1` are equal so the result is `true`.
  * The stack now looks like this: [true]
* We see the `OpJumpIfFalse` instruction, which pops a value from the stack and jumps if that is false.
  * Since the value on the stack is `true` the jump is not taken.
  * The stack now looks like this: []
* We now see the `OpTrue` instruction.
  * This pushes the value `true` onto the stack.
  * The stack now looks like this: [true]
* We see the `OpReturn` instruction.
  * This pops a value from the stack, and terminates execution.
  * The stack now looks like this: []


# Debugging

Seeing the dump of bytecode, as shown above, is useful but it still can be
hard to reason about the run-time behaviour.

For this reason it is possible to output each opcode before it is executed,
as well as view the current state of the stack.

To show this debug-output simply invoke the `run` sub-command with the `-debug` flag:

     $ evalfilter run -debug ./path/to/script


# Optimization

The virtual machine performs some simple optimizations when it is constructed.

The optimizations are naive, but are designed to simplify the bytecode which is intepreted.  There are a few distinct steps which are taken, although precise details will vary over time.

* Mathematical operations which only refer to integers will be collapsed
  * i.e. The statement `if ( 1 + 2 == 3 ) { ...` will be converted to `if ( true ) { ..`
  * Because the condition is provably always true.

* Jump statements (i.e. the opcode instructions `OpJump` and `OpJumpIfFalse`) will be removed if appropriate.
  * In the case of a jump which is never taken `if ( false ) { ..` the code will be removed.
    * This code wouldn't be written by a user, but could be generated via the first optimization.

* If a program contains no jump operations, and a OpReturn instruction is encounted the program will be truncated.
  * For example the program `return true; print( "What?"); return false;` will be truncated to become `return true;` because nothing after that can execute.
