# Bytecode

When the evalfilter package is executed a user-supplied script is lexed, parsed, and transformed into a series of bytecode operations.  These bytecode operations are executed by a simple stack-based virtual machine.

Although we don't expect users to care about the implementation details here are some brief notes.

The opcodes we're discussing are found in [code/code.go](code/code.go), and the virtual machine in [vm/vm.go](vm/vm.go).


* [Bytecode](#bytecode)
  * [Examining Bytecode](#examining-bytecode)
* [Bytecode Overview](#bytecode-overview)
  * [Constant / Field Operations](#constant--field-operations)
  * [Mathematical Operations](#mathematical-operations)
  * [Comparison Operations](#comparison-operations)
  * [Control-Flow Operations](#control-flow-operations)
  * [Misc Operations](#misc-operations)
* [Example Program](#example-program)

## Examining Bytecode

You may use `cmd/evalfilter` to view the bytecode representation of a program.

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


Constants:
  0 - &{1}
  1 - &{0.5}
  2 - &{2}
  3 - &{This is weird\n}
  4 - &{print}
```


# Bytecode Overview

Our bytecode understands approximately 30 different instructions, which are broadly grouped into categories:

* Constant / Field Operations
* Mathematical Operations
* Comparison Operations
* Control-Flow Operations


## Constant / Field Operations

The virtual machine I've implemented needs two things to work:

* A list of instructions to execute.
* A list of constants.

For example the program "`print( 1 + 2 ); return true;`" contains __three__ constants:

* The name of the function `print`.
* The integer value `1`.
* The integer value `2`.

That program would be encoded like so:

```
Bytecode:
  ...
  0000NN	OpConstant	0		// load constant: &{1}
  0000NN	OpConstant	1		// load constant: &{2}
  0000NN	OpAdd
...

Constants:
  0 - &{1}
  1 - &{2}
  2 - &{print}
```

This is the first time we've looked at our bytecode so there are several things to note:

* The value to the left of the instruction is the position in the code.
* The middle field is the instruction to be executed.
  * Some instructions contain a single argument, but most do not.
  * Some instructions contain helpful comments to the right.
* After the bytecode has been disassembled you'll see the pool of constants.
  * Each of which is identified by a unique number.

In this overview we're focusing upon the instruction `OpConstant`.  The `OpConstant` instruction has a single argument, which is the index of the constant to load and push on the stack.

When we start running the program the stack is empty.

* We run `OpConstant 0`
  * That loads the constant with ID `0` from the constant-area
    * This is the number `1`.
  * The constant is then pushed upon the stack.
* We run `OpConstant 1`
  * That loads the constant with ID `0` from the constant-area
    * This is the number `2`.
  * The constant is then pushed upon the stack.
* We then execute the `OpAdd` instruction.
  * This pops two values from the stack
    * i.e. The `2` we just added.
    * Then the `1` we added.
      * The stack is emptied in reverse
  * The two values are added, producing a result of `3`.
  * Then the value is placed back upon the stack.


## Mathematical Operations

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


## Comparison Operations

The comparison operations are very similar to the mathematical operations, and work in the same way:

* They pop two values from the stack.
* They run the comparision operation
  * If the comparison succeeds they push `true` upon the stack.
  * Otherwise they push `false`.

Comparision operations include:

* `OpEqual`
  * This pushes `true` upon the stack if the two values it is comparing are equal.
  * `false` otherwise.
* `OpLess`
  * This pushes `true` upon the stack if the first argumetn is less than the second.
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


## Control-Flow Operations

There are two control-flow operations:

* `OpJump`
  * Which takes the offset within the bytecode to jump to.
  * Control flow immediately branches there.
* `OpJumpIfFalse`
  * A value is poped from the stack, if it is false then control moves to the offset specified as the argument.
  * Otherwise we proceed to the next instruction as expected.


## Misc Operations

There are some miscellaneous instructions:

* `OpBang`
  * Calculate negation
* `OpMinus`
  * Calculate unary minus.
* `OpRoot`
  * Calculate a square root.
* `OpTrue`
 * Pushes a `true` value to the stack.
* `OpFalse`
 * Pushes a `false` value to the stack.
* `OpReturn`
 * Pops a value off the stack and terminates processing.
 * The value is the return-code.
* `OpLookup`
 * Much like loading a constant by reference this loads the value from the structure field with the given name.
* `OpCall`
 * Pops the name of a function to call from the stack.
 * Called with an argument noting how many arguments to pass to the function, and pops that many arguments from the stack to use in the function-call.


# Example Program

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
  * This pops a value from the stack, and terminates exection.
  * The stack now looks like this: []
