[![GoDoc](https://godoc.org/github.com/skx/evalfilter?status.svg)](http://godoc.org/github.com/skx/evalfilter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/evalfilter)](https://goreportcard.com/report/github.com/skx/evalfilter)
[![license](https://img.shields.io/github/license/skx/evalfilter.svg)](https://github.com/skx/evalfilter/blob/master/LICENSE)

* [eval-filter](#eval-filter)
  * [API Stability](#api-stability)
  * [Sample Usecase](#sample-usecase)
  * [Scripting Facilities](#scripting-facilities)
  * [Function Invocation](#function-invocation)
     * [Built-In Functions](#built-in-functions)
  * [Variables](#variables)
  * [Standalone Use](#standalone-use)
  * [Benchmarking](#benchmarking)
  * [Github Setup](#github-setup)



# eval-filter

The evalfilter package provides an embeddable evaluation-engine, which allows simple logic which might otherwise be hardwired into your golang application to be delegated to (user-written) script(s).

There is no shortage of embeddable languages which are available to the golang world, this library is intended to be less complex, allowing only simple tests to be made against structures/objects.  That said the flexibility present means that if you don't need full scripting support this might just be sufficient for your needs.

* The backstory behind this project is explained in [this blog-post](https://blog.steve.fi/a_slack_hack.html)

To give you a quick feel for how things look you could consult:

* [example_test.go](example_test.go).
  * This filters a list of people by their age.
* [example_function_test.go](example_function_test.go).
  * This exports a function from the golang-host application to the script.
  * Then uses that to filter a list of people.
* Some other simple examples are available beneath the [_examples/](_examples/) directory.


## API Stability

The API will remain as-is for given major release number, so far we've had we've had two major releases:

* 1.x.x
  * The initial implementation which parsed script into an AST then walked it.
* 2.x.x
  * The updated design which parses the given script into an AST, then generates bytecode to execute when the script is actually run.

The second release was implemented to perform a significant speedup for the case where the same script might be reused multiple times.


## Sample Usecase

You might have a chat-bot which listens to incoming messages and runs "something interesting" when specific messages are seen.  You don't necessarily need to have a full-scripting language, you just need to allow a user to specify whether the interesting-action should occur, on a per-message basis.

* Create an instance of the `evalfilter`.
* Load the user's script.
* For each incoming message run the script against it.
  * If it returns `true` you know you should carry out your interesting activity.
  * Otherwise you will not.

Assume you have a structure describing your incoming messages which looks something like this:

    type Message struct {
        Author  string
        Channel string
        Message string
        Sent    time.Time
    }

The user could now write following script to let you know that the incoming message was interesting:

    //
    // You can see that comments are prefixed with "//".
    //
    // This script is invoked by your Golang application as a filter,
    // the intent is that the user's script will terminate with either:
    //   return false;
    // or
    //   return true;
    //
    // Your host application will then carry out the interesting operation
    // when it receives a `true` result.
    //

    //
    // If we have a message from Steve it is interesting!
    //
    if ( Author == "Steve" ) { return true; }

    //
    // A bug is being discussed?  Awesome.
    //
    if ( Message ~=  "panic" ) { return true; }

    //
    // OK the message is uninteresting, and will be discarded, or
    // otherwise ignored.
    //
    return false;

You'll notice that we don't define the _object_ here, because it is implied that the script operates upon a single instance of a particular structure, whatever that might be.   That means `Author` is implicitly the author-field of the message object, which the `Run` method was invoked with.



## Scripting Facilities

The engine supports scripts which:

* Perform comparisons of strings and numbers:
  * equality:
    * "`if ( Message == "test" ) { return true; }`"
  * inequality:
    * "`if ( Count != 3 ) { return true; }`"
  * size (`<`, `<=`, `>`, `>=`):
    * "`if ( Count >= 10 ) { return false; }`"
    * "`if ( Hour >= 8 && Hour <= 17 ) { return false; }`"
  * String contains:
    * "`if ( Content ~= "needle" )`"
  * Does not contain:
    * "`if ( Content !~ "some text we dont want" )`"
* You can also add new primitives to the engine.
  * By implementing them in your golang host application.
  * Your host-application can also set variables which are accessible to the user-script.
* Finally there is a `print` primitive to allow you to see what is happening, if you need to.
  * This is just one of the built-in functions, but perhaps the most useful.

You'll note that you're referring to structure-fields by name, they are found dynamically via reflection.  The  `if` conditions can be nested, and also support an `else` clause.



## Function Invocation

In addition to operating upon the fields of an object/structure literally you can also call functions with them.

For example you might have a list of people, which you wish to filter by the length of their names:

    // People have "name" + "age" attributes
    type Person struct {
      Name string
      Age  int
    }

    // Now here is a list of people-objects.
    people := []Person{
        {"Bob", 31},
        {"John", 42},
        {"Michael", 17},
        {"Jenny", 26},
    }

You can filter the list based upon the length of their name via a script such as this:

    // Example filter - we only care about people with "long" names.
    if ( len(Name) > 4 ) { return true; } else { return false; }

This example is contained in [example_function_test.go](example_function_test.go) if you wish to see the complete code.


### Built-In Functions

The following functions are built-in and available by default:

* `len(field | value)`
  * Returns the length of the given value, or the contents of the given field.
* `lower(field | value)`
  * Return the lower-case version of the given input.
* `match(field | str, regexp)`
  * Returns `true` if the specified string matches the supplied regular expression.
  * You can make this case-insensitive using `(?i)`, for example:
    * `if ( match( "Steve" , "(?i)^steve$" ) ) { ... `
* `trim(field | string)`
  * Returns the given string, or the contents of the given field, with leading/trailing whitespace removed.
* `type(field | value)`
  * Returns the type of the given field, as a string.
    * For example `string`, `integer`, `float`, `boolean`, or `null`.
* `upper(field | value)`
  * Return the upper-case version of the given input.


## Variables

Your host application can register variables which are accessible to your scripting environment via the `SetVariable` method.  The variables can have their values updated at any time before the call to `Eval` is made.

For example the following example sets the contents of the variable `time`, and then outputs it.  Every second the output will change, because the value has been updated:

    eval := evalfilter.New(`
                print("The time is ", time, "\n");
                return false;
            `)

    eval.Prepare()

    for {

        // Set the variable `time` to be the seconds past the epoch.
        eval.SetVariable("time", &object.Integer{Value: time.Now().Unix()})

        // Run the script.
        eval.Run(nil)

        // Update every second.
        time.Sleep(1 * time.Second)
    }

This example is available, with error-checking, in [_examples/variable/](_examples/variable/).


## Standalone Use

If you wish to experiment with script-syntax you can install the standalone driver:

```
go get github.com/skx/evalfilter/cmd/evalfilter

```

The driver has a number of sub-commands to allow you to test a script, for example viewing the parse-tree, the byecode, or even running a script against a JSON object.

For example in the [cmd/evalfilter](cmd/evalfilter) directory you might run:

     ./evalfilter run -json on-call.json on-call.script

This will test a script against a JSON object, allowing you to experiment with changing either.  This driver an also be used to reproduce any problems identified via [fuzz-testing](FUZZING.md).


## Benchmarking

If you wish to run a local benchmark you should be able to do so as follows:

```
$ go test -test.bench=evalfilter -benchtime=10s -run=^t
goos: linux
goarch: amd64
pkg: github.com/skx/evalfilter
Benchmark_evalfilter_complex-4   	 5000000	      3895 ns/op
Benchmark_evalfilter_simple-4    	500000000	        25.6 ns/op
PASS
ok  	github.com/skx/evalfilter	38.934s
```

Neither example there is completely representative, but it will give you
an idea of the speed.  In the majority of cases the speed of the evaluation
engine will be acceptible.


## Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.


Steve
--
