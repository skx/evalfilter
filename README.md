[![GoDoc](https://godoc.org/github.com/skx/evalfilter?status.svg)](http://godoc.org/github.com/skx/evalfilter)
[![Go Report Card](https://goreportcard.com/badge/github.com/skx/evalfilter)](https://goreportcard.com/report/github.com/skx/evalfilter)
[![license](https://img.shields.io/github/license/skx/evalfilter.svg)](https://github.com/skx/evalfilter/blob/master/LICENSE)


# eval-filter

The evalfilter package provides a basic embeddable evaluation-engine, which allows simple logic which might otherwise be hardwired into your golang application to be delegated to (user-written) script(s).

There is no shortage of embeddable languages which are available to the golang world, but this library is intended to be simpler.  The ideal use case is defining rules which are applied test specific objects.

In short don't think of this as a scripting-language, but instead a simple way of applying a set of rules to an object, or a filtering a large collection of objects via a user-defined fashion.

You may view a quick and simple example in the file [example_test.go](example_test.go), which filters a list of people by their age.


## Sample Use

You might have a chat-bot which listens to incoming messages and does something interesting when specific messages are seen.  You don't necessarily need to have a full-scripting language, you just need to be write a snippet of script which returns `true` if the given message is interesting.

Upon the receipt of each incoming message you call the filter, if the message is interesting it will return `true`.  If the return-value is `false` you know the message is uninteresting and you do nothing.

Assume you have a structure describing incoming messages:

    type Message struct {
        Author  string
        Channel string
        Message string
        Sent    time.Time
    }

Now you have an instance of that message:

    var msg Message

You want to decide if this message is interesting, so you might invoke the evaluator with like so:

    eval, er := evaluator.New( `script goes here ...` )
    out, err := eval.Run( msg )

Assuming no error the `out` value will contain the return-result of your script which will be a `boolean`, because these scripts are _filters_.



## Scripting

The scripting language itself is where things get interesting, because you can access members of the structure passed as you would expect:

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
    // Your host application uses this script as a filter, so that
    // any message which return `true` will be processed further.
    //

    //
    // If we have messages from Steve they're "interesting".
    //
    if ( Author == "Steve" ) { return true; }

    //
    // We should listen to our parents.
    //
    if ( Author == "YourParent" ) { return true; }

    //
    // OK the message is uninteresting, and will be discarded, or
    // otherwise ignored.
    //
    return false;

You'll notice that we don't define the _object_ here, because it is implied that the script operates upon a single instance of a particular structure, whatever that might be.   That means `Author` is implicitly the author-field of the message object, which the `Run` method was invoked with.

(i.e. We access "`Author`", rather than `msg.Author` because we're opering upon a single object - the name of that object is redundent.)


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

You'll note that you're referring to structure-fields by name, they are found dynamically via reflection.


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

You can implement your own functions in your application, which can be
called by scripts - see [example_function_test.go](example_function_test.go) for an example of doing that.


## Built-In Functions

The following functions are built-in, and available by default:

* `len( field | string)`
  * Returns the length of the given string, or the contents of the given field.
* `trim( field | string)`
  * Returns the given string, or the contents of the given field, with leading/trailing whitespace removed.



Steve
--
