# variable

This example demonstrates setting a variable in the host-application, and
accessing that from inside the filter script.

## Usage

Compile the script with `go build .`.

Then launch it to see that the value `var`, as accessed by the script
keeps changing each time the script is evaluated:

```
$ ./variable
The variable we received was 0
	We'll keep going until we hit 20 iterations.
	Script gave result false
	The script set new=0


The variable we received was 1
	We'll keep going until we hit 20 iterations.
	Script gave result false
	The script set new=4

..
^C

```

Similarly you'll see that the value the script set was retrieved by the host
application, and also changes each iteration.
