# Fuzz Testing

Fuzz-testing involves creating random input, and running the program to test with that, to see what happens.

The intention is that most of the random inputs will be invalid, so you'll be able to test your error-handling and see where you failed to consider specific input things.

The 1.18 release of the golang compiler/toolset has integrated support for fuzz-testing.


## Usage

If you're running golang 1.18beta1 or higher you can run the fuzz-testing against the evaluator like so:

```
go test -fuzztime=300s -parallel=1 -fuzz=FuzzEvaluator -v
```


## Results

As the fuzzer runs it will regularly output a status-line showing how long it has been running for, how many "crashers" (i.e. bugs, or error-conditions which were not handled) it has found, and similar metrics.

Sample output might look like this:

```
fuzz: elapsed: 0s, gathering baseline coverage: 0/5 completed
fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 1 workers
fuzz: elapsed: 3s, execs: 28632 (9541/sec), new interesting: 99 (total: 104)
fuzz: elapsed: 6s, execs: 60575 (10651/sec), new interesting: 160 (total: 165)
fuzz: elapsed: 9s, execs: 86143 (8522/sec), new interesting: 195 (total: 200)
```

The last line shows us:

* The fuzzer has been running for 9 seconds.
* Zero crashing-cases have been found.
* 86143 executions have taken place
* We're averaging the execution of a test-case 8522 times a sec.
* We've found/evolved 200 interesting test-cases
