# variable

This example demonstrates setting a variable in the host-application, and
accessing that from inside the filter script.

## Usage

Compile the script with `go build .`.

Then launch it to see that the value `time`, as accessed by the script
keeps changing each time the script is evaluated:

```
$ ./variable
The time is 1569264364
	Yay!
Script gave result false
The time is 1569264365
	Yay!
Script gave result false
The time is 1569264366
	Yay!
Script gave result false
^C

```
