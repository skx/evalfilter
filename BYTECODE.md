# Bytecode

When the evalfilter is executed a user-supplied script is lexed, parsed, and transformed into a series of bytecode operations.

Although we don't expect users to care about the implementation details here are some brief notes.


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

To view the bytecode:

```
./evalfilter bytecode ./simple.txt
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

## Overview

We're a stack-based ..

## Opcodes

Brief list ..
