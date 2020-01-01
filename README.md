[![GoDoc](https://godoc.org/github.com/skx/evalfilter?status.svg)](http://godoc.org/github.com/skx/evalfilter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/evalfilter)](https://goreportcard.com/report/github.com/skx/evalfilter)
[![license](https://img.shields.io/github/license/skx/evalfilter.svg)](https://github.com/skx/evalfilter/blob/master/LICENSE)

* [eval-filter](#eval-filter)
  * [Implementation](#implementation)
  * [Scripting Facilities](#scripting-facilities)
  * [Use Cases](#use-cases)
* [Sample Usage](#sample-usage)
  * [Built-In Functions](#built-in-functions)
  * [Variables](#variables)
* [Standalone Use](#standalone-use)
* [Benchmarking](#benchmarking)
* [Fuzz Testing](#fuzz-testing)
* [API Stability](#api-stability)
* [Github Setup](#github-setup)


# eval-filter

The evalfilter package provides an embeddable evaluation-engine, which allows simple logic which might otherwise be hardwired into your golang application to be delegated to (user-written) script(s).

There is no shortage of embeddable languages which are available to the golang world, this library is intended to be something that is:

* Simple to embed.
* Simple to use, as there are only three methods you need to call:
  * [New](https://godoc.org/github.com/skx/evalfilter#New)
  * [Prepare](https://godoc.org/github.com/skx/evalfilter#Eval.Prepare)
  * Then either [Execute(object)](https://godoc.org/github.com/skx/evalfilter#Eval.Execute) or [Run(object)](https://godoc.org/github.com/skx/evalfilter#Eval.Run) depending upon what kind of return value you would like..
* Simple to understand.
* As fast as it can be, without being too magical.

The scripting language is C-like, and is generally intended to allow you to _filter_ objects, which means you might call the same script upon multiple objects, and the script will return either `true` or `false` as appropriate to denote whether some action might be taken by your application against that particular object.

It _is_ possible for you to handle arbitrary return-values from the script(s) you execute, and indeed the script itself could call back into your application to carry out tasks, via the addition of new primitives implemented and exported by your host application, which would make the return value almost irrelevant.

The ideal use-case is that your application receives objects of some kind, perhaps as a result of incoming webhook submissions, network events, or similar, and you wish to decide how to handle those objects in a flexible fashion.



## Implementation

In terms of implementation the script to be executed is split into [tokens](token/token.go) by the [lexer](lexer/lexer.go), then those tokens are [parsed](parser/parser.go) into an abstract-syntax-tree.   Once the AST exists it is walked by the [compiler](compiler.go) and a series of [bytecode instructions](code/code.go) operations are generated.

Once the bytecode has been generated it can be reused multiple times, there is no state which needs to be maintained, which makes actually executing the script (i.e. running the bytecode) a fast process.

At execution-time the bytecode which was generated is interpreted by a simple [virtual machine](vm/vm.go).  The virtual machine is fairly naive implementation of a [stack-based](stack/stack.go) virtual machine, with some runtime support to provide the [builtin-functions](environment/builtins.go), as well as supporting the addition of host-specific functions.

The bytecode itself is documented briefly in [BYTECODE.md](BYTECODE.md), but it is not something you should need to understand to use the library - although it might be useful for debugging issues.


## Scripting Facilities

The scripting-language this package presents supports the basic types you'd expect:

* Arrays.
* Floating-point numbers.
* Integers.
* Strings.
* Time / Date values.
  * i.e. We can use reflection to handle `time.Time` values in any structure/map we're operating upon.

The types are supported both in the language itself, and in the reflection-layer which is used to allow the script access to fields in the Golang object/map you supply to it.

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
* There are series of built-in primitives which can be used by your scripts, and you can export your own host-specified functions easily.
  * For example the `print` function to generate output from your script is just a simple function implemented in Golang and exported to the environment.



## Use Cases

The backstory behind this project is explained in [this blog-post](https://blog.steve.fi/a_slack_hack.html), but in brief I wanted to react to incoming Slack message:

* I wanted to implement a simple "on-call notifier".
   * When messages were posted to Slack channels I wanted to _sometimes_ trigger a phone-call to the on-call engineer.
   * Of course not _all_ Slack-messages were worth waking up an engineer for..

The expectation was that non-developers might want to change the matching of messages to update the messages which were deemed worthy of waking up the on-call engineer.  They shouldn't need to worry about rebuilding the on-call application, nor should they need to understand Go.  So the logic was moved into a script and this evaluation engine was born.

This is a pretty good use-case for an evaluation engine, solving a real problem, and not requiring a large degree of technical knowledge to update.

As noted the application was pretty simple, each time a Slack message was received it would be decoded into a simple structure:

    type Message struct {
        Author  string
        Channel string
        Message string
        Sent    time.Time
    }

Once the message was decoded a simple script could then be executed against that particular object to decide whether to initiate a phone-call:

    //
    // You can see that comments are prefixed with "//".
    //
    // In my case if this script hit `return true;` a phone call would be initiated.
    //
    // If the script terminated with `return false;` I would do nothing.

    //
    // If this is within office hours we'll assume somebody is around who
    // can handle the issue, so there is no need to raise a call.
    //
    if ( hour(Sent) >= 9 || hour(Sent) <= 17 ) {

        //
        // Of course we need to exclude the weekend.  Nobody works
        // weekends, if they can help it!
        //
        if ( day(Sent) != "Saturday" && day(Sent) != "Sunday" ) {
           return false;
        }
    }

    //
    // A service crashed with a panic?
    //
    // If so raise the engineer.
    //
    if ( Message ~=  /panic/i ) { return true; }


    //
    // At this point we decide the message is not important, so
    // we ignore it.
    //
    // In a real-life implementation we'd probably work the other
    // way round.  Default to triggering the call unless we knew
    // it was a low-priority/irrelevant message.
    //
    return false;

You'll notice that we test fields such as `Sent` and `Message` here which come from the object we were given.  That works due to the magic of reflection.  Similarly we called a number of built-in functions related to time/date.  These functions understand the golang `time.Time` type, from which the `Sent` value was read via reflection.

(All `time.Time` values are converted to seconds-past the Unix Epoch, but you can retrieve all the appropriate fields via `hour()`, `minute()`, `day()`, `year()`, `weekday()`, etc, as you would expect.  Using them literally will return the Epoch value.)


# Sample Usage

To give you a quick feel for how things look you could consult these two simple examples:

* [example_test.go](example_test.go).
  * This filters a list of people by their age.
* [example_function_test.go](example_function_test.go).
  * This exports a function from the golang-host application to the script.
  * The new function is then used to filter a list of people.

Additional examples of using the library to embed scripting support into simple host applications are available beneath the [_examples/](_examples/) directory.

There is also a standalone driver located in [cmd/evalfilter](cmd/evalfilter) which allows you to examine bytecode, tokens, and run scripts - this is discussed [later](#standalone-use) in this README file.



## Built-In Functions

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
* `now()`
  * Returns the current time.


## Variables

Your host application can also register variables which are accessible to your scripting environment via the `SetVariable` method.  The variables can have their values updated at any time before the call to `Eval` is made.

Similarly you can _retrieve_ values which have been set within scripts, via `GetVariable`.

You can see an example of this in [_examples/variable/](_examples/variable/)



# Standalone Use

If you wish to experiment with script-syntax you can install the standalone driver:

```
go get github.com/skx/evalfilter/v2/cmd/evalfilter

```

This driver, contained within the repository at [cmd/evalfilter](cmd/evalfilter) has a number of sub-commands to allow you to experiment with the scripting environment:

* Output a dissassembly of [bytecode](BYTECODE.md).
* Run a script.
  * Optionally with a JSON object as input.


# Benchmarking

The scripting language should be fast enough for most purposes; it will certainly cope well with running simple scripts for every incoming HTTP-request, for example.  If you wish to test the speed there are some local benchmarks available.

You can run the benchmarks as follows:

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


# Fuzz Testing

Fuzz-testing is basically magic - you run your program with random input, which stress-tests it and frequently exposes corner-cases you've not considered.

This project has been fuzz-tested repeatedly, and [FUZZING.md](FUZZING.md) contains notes on how you can carry out testing of your own.


## API Stability

The API will remain as-is for given major release number, so far we've had we've had two major releases:

* 1.x.x
  * The initial implementation which parsed script into an AST then walked it.
* 2.x.x
  * The updated design which parses the given script into an AST, then generates bytecode to execute when the script is actually run.

The second release was implemented to perform a significant speedup for the case where the same script might be reused multiple times.



# Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.



Steve
--
