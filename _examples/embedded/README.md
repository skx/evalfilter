# Embedded Examples

This directory contains some example programs which embed the evalfilter scripting language / evaluation engine.


## [passwd](passwd/)

This loops over the entries in your system `/etc/passwd` file, running a script against each entry.


## [state](state/)

This example shows how you can maintain state while running the same script multiple times with the same intepreter.


## [time](time/)

This example demonstrates how golang's `time.Time` values are correctly retrieved using reflection.


## [variable](variable/)

This example demonstrates setting a variable via golang code, and accessing that within the scripting environment.

This is a distinct process different from passing a structure/object/map to the `Run` or `Execute` method of the evaluation engine.
