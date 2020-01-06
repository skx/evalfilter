# state

This example demonstrates that by keeping the same `evalfilter` instance
between script-calls you can maintain state.

## Usage

Compile via `go build .`, then launch the program.   It will execute the
script embedded within itself 20 times, and you'll see the result change:

```
go build . && ./state
0 -> false
1 -> false
2 -> false
3 -> false
4 -> false
5 -> false
6 -> false
7 -> false
8 -> false
9 -> false
10 -> true
11 -> true
12 -> true
13 -> true
14 -> true
15 -> true
16 -> true
17 -> true
18 -> true
19 -> true
```
