[![GoDoc](https://godoc.org/github.com/skx/evalfilter?status.svg)](http://godoc.org/github.com/skx/evalfilter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/evalfilter)](https://goreportcard.com/report/github.com/skx/evalfilter)
[![license](https://img.shields.io/github/license/skx/evalfilter.svg)](https://github.com/skx/evalfilter/blob/master/LICENSE)

* [Sample Use](#sample-use)
* [Scripting Facilities](#scripting-facilities)
* [Function Invocation](#function-invocation)
   * [Built-In Functions](#built-in-functions)
* [Variables](#variables)
* [Alternatives](#alternatives)
* [Github Setup](#github-setup)



# eval-filter

The evalfilter package provides an embeddable evaluation-engine, which allows simple logic which might otherwise be hardwired into your golang application to be delegated to (user-written) script(s).

There is no shortage of embeddable languages which are available to the golang world, but this library is intended to be simpler; the ideal use case is defining rules which are applied to run tests against objects.

To give a feel for the way it works you may consult the simple example in the file [example_test.go](example_test.go), which filters a list of people by their age.



## Sample Use

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
    //
    //   return false;
    // or
    //   return true;
    //
    // Your host application may decide to do something interesting
    // when it receives a `true` result, and nothing when it sees `false`.
    //

    //
    // If we have a message from Steve it is "interesting"!
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
  * String contains:
    * "`if ( Content ~= "needle" )`"
  * Does not contain:
    * "`if ( Content !~ "some text we dont want" )`"
* You can also add new primitives to the engine.
  * By implementing them in your golang host application.
* Your host-application can set variables which are accessible to the user-script, with a `$`-prefix.
  * `if ( $time == "Steve" ) { print "You set the variable 'time' to 'Steve'\n"; }`
* Finally there is a `print` primitive to allow you to see what is happening, if you need to.

You'll note that you're referring to structure-fields by name, they are found dynamically via reflection.

`if` conditions can be nested as the following sample shows, and also support an `else` clause.


     if ( Count > 10 ) {
         print "Count is > 10\n";

         if ( Count > 50 ) {
              print "The count is super-big!\n";
         } else {
              print "The count is somewhat high!\n";
         }
     }

## Function Invocation

In addition to operating upon the fields of an object/structure literally you can also call functions with them.

For example you might have a list of people, which you wish to filter by the length of their names:

    // People have "name" + "age" attributes
    type Person struct {
      Name string
      Age  int
    }
    people := []Person{
        {"Bob", 31},
        {"John", 42},
        {"Michael", 17},
        {"Jenny", 26},
    }

You can filter the list based upon the length of their name via a filter-script like this:

    // Example filter - we only care about people with "long" names.
    if ( len(Name) > 4 ) { return true ; }

    // Since we return false the caller will know to ignore people here.
    return false;

This example is contained in [example_function_test.go](example_function_test.go) if you wish to see the complete code.


### Built-In Functions

The following functions are built-in, and available by default:

* `len(field | string)`
  * Returns the length of the given string, or the contents of the given field.
* `trim(field | string)`
  * Returns the given string, or the contents of the given field, with leading/trailing whitespace removed.


## Variables

Your host application can register variables which are accessible to your scripting environment via the `SetVariable` method.  The variables can have their values updated at any time before the call to `Eval` is made.

* **NOTE**: Variables are accessed with a `$`-prefix inside the users' script

For example the following example sets the contents of the variable `time`, and then outputs it.  Every second the output will change, because the value has been updated:

    eval := evalfilter.New(`print "The time is ", $time, "\n";
                            return false;`)

    for {

        // Set the variable `$time` to be the seconds past the epoch.
        eval.SetVariable("time", fmt.Sprintf("%v", time.Now().Unix()))

        // Run the script.
        ret, err := eval.Run(nil)

        // If there are errors - abort
        if err != nil {
            panic(err)
        }

        // Show the result
        fmt.Printf("Script gave result %v\n", ret)

        // Update every second.

        time.Sleep(1 * time.Second)
    }


## Alternatives

If this solution doesn't quite fit your needs you might investigate:

* https://github.com/Knetic/govaluate
* https://github.com/PaesslerAG/gval/
* https://github.com/antonmedv/expr

## Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.


Steve
--
