[![GoDoc](https://godoc.org/github.com/skx/evalfilter?status.svg)](http://godoc.org/github.com/skx/evalfilter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/evalfilter)](https://goreportcard.com/report/github.com/skx/evalfilter)
[![license](https://img.shields.io/github/license/skx/evalfilter.svg)](https://github.com/skx/evalfilter/blob/master/LICENSE)

* [eval-filter](#eval-filter)
  * [Overview](#overview)
  * [Implementation](#implementation)
    * [Bytecode](#bytecode)
  * [Use Cases](#use-cases)
  * [Sample Usage](#sample-usage)
  * [API Stability](#api-stability)
  * [Scripting Facilities](#scripting-facilities)
	 * [Built-In Functions](#built-in-functions)
     * [Variables](#variables)
  * [Standalone Use](#standalone-use)
     * [Debugging via standalone use](#debugging-via-standalone-use)
  * [Benchmarking](#benchmarking)
  * [Fuzz Testing](#fuzz-testing)
  * [Github Setup](#github-setup)


# eval-filter

The evalfilter package provides an embeddable evaluation-engine, which allows simple logic which might otherwise be hardwired into your golang application to be delegated to (user-written) script(s).

There is no shortage of embeddable languages which are available to the golang world, this library is intended to be something that is:

* Simple to embed.
* Simple to use.
  * There are only three methods, call `New`, `Prepare`, then `Run(object)`.
* Simple to understand.
* As fast as it can be, without being too magical.



## Overview

The `evalfilter` library provides the means to embed a small scripting engine in your golang application (which is known as the host application).

The scripting language is C-like, and is generally intended to allow you to _filter_ objects, with the general expectation that a script will return `true` or `false` allowing you to decide what to do after running it.

The ideal use-case is that your application receives objects of some kind, perhaps as a result of incoming webhook submissions, network events, or similar, and you wish to decide how to handle those objects in a flexible fashion.



## Implementation

In terms of implementation the script to be executed is split into [tokens](token/token.go) by the [lexer](lexer/lexer.go), then those tokens are [parsed](parser/parser.go) into an abstract-syntax-tree.   The AST is then processed, and from it a series of [bytecode](code/code.go) operations are generated.  The bytecode runs through a simple optimizer-stage and then the compiler is done.

Once the bytecode has been generated it can be reused multiple times, there is no state which needs to be maintained.  This makes actually executing the script (i.e. running the bytecode) a fast process.

At execution-time the bytecode which was generated is interpreted by a simple [virtual machine](vm/vm.go) in the `Run` method.  As this is a stack-based virtual machine, rather than a register-based one, we have a simple [stack](stack/stack.go) implementation, along with some runtime support to provide the [builtin-functions](environment/builtins.go).



### Bytecode

The bytecode is not exposed externally, but it is documented in [BYTECODE.md](BYTECODE.md).


## Use Cases

The backstory behind this project is explained in [this blog-post](https://blog.steve.fi/a_slack_hack.html), but in brief I wanted to read incoming Slack messages and react to specific ones to carry out an action.

* In brief I wanted to implement a simple "on-call notifier".
* When messages were posted to Slack channels I wanted to _sometimes_ trigger a phone-call to the on-call engineer who was nominated to handle events/problems/issues that evening.
* Of course not _all_ Slack-messages were worth waking up an engineer for..

The expectation was that non-developers might want to change the matching of messages, without having to know how to rebuild the application, or understand Go.  So the logic was moved into a script and this evaluation engine was born.

This is a pretty good use-case for an evaluation engine, solving a real problem, and not requiring a large degree of technical knowledge to update.

As noted the application was pretty simple, logically:

* Create an instance of the `evalfilter`.
* Load the user's script, which will let messages be matched.
* For each incoming message run the users' script against it.
  * If it returns `true` you know you should trigger the on-call notification.
  * Otherwise ignore the message.

To make this more concrete we'll pretend we have the following structure to describe incoming messages:

    type Message struct {
        Author  string
        Channel string
        Message string
        Sent    time.Time
    }

The user could now write the following script to decide whether to initiate a notification:

    //
    // You can see that comments are prefixed with "//".
    //
    // This script is invoked by your Golang application as a filter,
    // the intent is that the user's script will terminate with either:
    //   return false;
    // or
    //   return true;
    //

    //
    // If we have a message from Steve it is interesting!
    //
    // Here `return true` means to initiate the phone-call.
    //
    if ( Author == "Steve" ) { return true; }

    //
    // A bug is being discussed?  Awesome.  That's worth waking
    // somebody for.
    //
    if ( Message ~=  /panic/i ) { return true; }

    //
    // If this is outside office hours we'll raise a phone-call.
    //
    if ( hour(Sent) <= 7 || hour(Sent) >= 19) { return true; }

    //
    // At this point we decide the message is not important, so
    // we ignore it.
    //
    // In a real-life implementation we'd probably work the other
    // way round.  Default to triggering the call unless we knew
    // it was a low-priority/irrelevant message.
    //
    return false;

You'll notice that we test fields such as `Message` here, which come from the object we were given.  That works due to the magic of reflection.  Similarly we managed to call the built-in function `hour` to get the hour of the `Sent` field which was a golang `time.Time` value, again this works due to the magic of reflection.

(All `time.Time` values are converted to seconds-past the Unix Epoch, but you can retrieve all the appropriate fields via `hour()`, `minute()`, `day()`, `year()`, `weekday()`, etc, as you would expect.)



## Sample Usage

To give you a quick feel for how things look you could consult these two simple examples:

* [example_test.go](example_test.go).
  * This filters a list of people by their age.
* [example_function_test.go](example_function_test.go).
  * This exports a function from the golang-host application to the script.
  * The new function is then used to filter a list of people.

Additional examples are available beneath the [_examples/](_examples/) directory, and there is a standalone driver located in [cmd/evalfilter](cmd/evalfilter) which allows you to examine bytecode, tokens, and run scripts.



## API Stability

The API will remain as-is for given major release number, so far we've had we've had two major releases:

* 1.x.x
  * The initial implementation which parsed script into an AST then walked it.
* 2.x.x
  * The updated design which parses the given script into an AST, then generates bytecode to execute when the script is actually run.

The second release was implemented to perform a significant speedup for the case where the same script might be reused multiple times.


## Scripting Facilities

The engine supports the basic types you'd expect:

* Arrays
* Floating-point numbers
* Integers
* Strings
* Time / Date values
  * i.e. We can use reflection to handle `time.Time` values in any structure/map we're operating upon.


These types are supported both in the language itself, and in the reflection-layer which is used to allow the script access to fields in the Golang object/map you supply to it.

Again as you'd expect the facilities are pretty normal/expected:

* Perform comparisons of strings and numbers:
  * equality:
    * "`if ( Message == "test" ) { return true; }`"
  * inequality:
    * "`if ( Count != 3 ) { return true; }`"
  * size (`<`, `<=`, `>`, `>=`):
    * "`if ( Count >= 10 ) { return false; }`"
    * "`if ( Hour >= 8 && Hour <= 17 ) { return false; }`"
  * String matching against a regular expression:
    * "`if ( Content ~= /needle/ )`"
    * "`if ( Content ~= /needle/i )`"
      * With case insensitivity
  * Does not match a regular expression:
    * "`if ( Content !~ /some text we don't want/ )`"
  * Test if an array contains a value:
    * "`return ( Name in [ "Alice", "Bob", "Chris" ] );`"
* You can also easily add new primitives to the engine.
  * By implementing them in your golang host application.
  * Your host-application can also set variables which are accessible to the user-script.
* Finally there is a `print` primitive to allow you to see what is happening, if you need to.
  * This is just one of the built-in functions, but perhaps the most useful.



### Built-In Functions

As we noted earlier you can export functions from your host-application and make them available to the scripting environment, as demonstrated in the [example_function_test.go](example_function_test.go) sample, but of course there are some built-in functions which are always available:

* `float(value)`
  * Tries to convert the value to a floating-point number, returns Null on failure.
  * e.g. `float("3.13")`.
* `int(value)`
  * Tries to convert the value to an integer, returns Null on failure.
  * e.g. `int("3")`.
* `len(field | value)`
  * Returns the length of the given value, or the contents of the given field.
  * For arrays it returns the number of elements, as you'd expect.
* `lower(field | value)`
  * Return the lower-case version of the given input.
* `string( )`
  * Converts a value to a string.  e.g. "`string(3/3.4)`".
* `trim(field | string)`
  * Returns the given string, or the contents of the given field, with leading/trailing whitespace removed.
* `type(field | value)`
  * Returns the type of the given field, as a string.
    * For example `string`, `integer`, `float`, `array`, `boolean`, or `null`.
* `upper(field | value)`
  * Return the upper-case version of the given input.
* `hour(field|value)`, `minute(field:value)`, `seconds(field:value`
  * Allow converting a time to HH:MM:SS.
* `day(field|value)`, `month(field:value)`, `year(field:value`
  * Allow converting a time to DD/MM/YYYY.
* `weekday(field|value)`
  * Allow converting a time to "Saturday", "Sunday", etc.


## Variables

Your host application can also register variables which are accessible to your scripting environment via the `SetVariable` method.  The variables can have their values updated at any time before the call to `Eval` is made.

Similarly you can _retrieve_ values which have been set within scripts, via `GetVariable`.

You can see an example of this in [_examples/variable/](_examples/variable/)


## Standalone Use

If you wish to experiment with script-syntax you can install the standalone driver:

```
go get github.com/skx/evalfilter/cmd/evalfilter

```

The driver has a number of sub-commands to allow you to test a script, for example viewing the parse-tree, the bytecode, or even running a script against a JSON object.

For example in the [cmd/evalfilter](cmd/evalfilter) directory you might run:

     ./evalfilter run -json on-call.json on-call.script

This will test a script against a JSON object, allowing you to experiment with changing either.


### Debugging via standalone use

Using the standalone driver is very useful to debug execution of scripts,
for example the `-debug` and `-no-optimizer` flags will change the way
that the script is run.

Consider this example:

    print( "Hello, World\n" );
    return true;

You can trace how it is executed via:

```
$ evalfilter run -debug ./example.in

	Stack: []
0000	OpConstant	0000

	Stack: [Hello, World\n]
0003	OpConstant	0001

	Stack: [Hello, World\n, print]
0006	OpCall	0001
Hello, World

..
Script gave result true
```

Here you're show the state of the stack and every opcode which is executed, along with the arguments.  This is perhaps more useful when coupled with seeing the raw bytecode disassembly:

```
$ evalfilter bytecode ./example.in
Bytecode:
  000000	    OpConstant	0	// load constant: "Hello, World\n"
  000003	    OpConstant	1	// load constant: "print"
  000006	        OpCall	1	// call function with 1 arg(s)
  000009	        OpTrue
  000010	      OpReturn


Constants:
  000000 Type:STRING Value:"Hello, World\n"
  000001 Type:STRING Value:"print"
```


## Benchmarking

If you wish to run a local benchmark you should be able to do so as follows:

```
go test -test.bench=evalfilter_ -benchtime=10s -run=^t
goos: linux
goarch: amd64
pkg: github.com/skx/evalfilter/v2
Benchmark_evalfilter_complex_map-4   	 4426123	      2721 ns/op
Benchmark_evalfilter_complex_obj-4   	 7657472	      1561 ns/op
Benchmark_evalfilter_simple-4        	15309301	       818 ns/op
Benchmark_evalfilter_trivial-4       	100000000	       105 ns/op
PASS
ok  	github.com/skx/evalfilter/v2	52.258s
```

The examples there are not particularly representative, but they will give you
an idea of the general speed.  In the real world the speed of the evaluation
engine is unlikely to be a significant bottleneck.

One interesting thing that shows up clearly is that working with a `struct` is significantly faster than working with a `map`.  I can only assume that the reflection overhead is shorter there, but I don't know why.


## Fuzz Testing

Fuzz-testing is basically magic - you run your program with random input, which stress-tests it and frequently exposes corner-cases you've not considered.

This project has been fuzz-tested repeatedly, and [FUZZING.md](FUZZING.md) contains notes on how you can carry out testing of your own.


## Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.



Steve
--
