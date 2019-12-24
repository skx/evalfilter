# Fuzz Testing

Fuzz-testing involves creating random input, and running the program to test with that, to see what happens.

The intention is that most of the random inputs will be invalid, so you'll be able to test your error-handling and see where you failed to consider specific input things.


## Usage

The [fuzz/](fuzz/) directory contains a simple fuzz-handler which is designed to be used by [go-fuzz](https://github.com/dvyukov/go-fuzz).

To use it you must first install the tools:

```
go get github.com/dvyukov/go-fuzz/go-fuzz
go get github.com/dvyukov/go-fuzz/go-fuzz-build
```

Once you've got the tooling installed you can now build the fuzz-package:

```
go-fuzz-build github.com/skx/evalfilter/v2/fuzz
```

Finally you can now launch the fuzzer, like so:

```
go-fuzz -procs=1 -bin=fuzz-fuzz.zip -workdir=workdir
```


## Results

As the fuzzer runs it will regularly output a status-line showing how long it has been running for, how many "crashers" (i.e. bugs, or error-conditions which were not handled) it has found, and similar metrics.

Sample output might look like this:

```
2019/12/24 12:59:27 workers: 2, corpus: 1425 (49m37s ago), crashers: 0, restarts: 1/10000, execs: 598384404 (11382/sec), cover: 1513, uptime: 14h36m
2019/12/24 12:59:30 workers: 2, corpus: 1425 (49m40s ago), crashers: 0, restarts: 1/10000, execs: 598416850 (11381/sec), cover: 1513, uptime: 14h36m
2019/12/24 12:59:33 workers: 2, corpus: 1425 (49m43s ago), crashers: 0, restarts: 1/10000, execs: 598449839 (11381/sec), cover: 1513, uptime: 14h36m
2019/12/24 12:59:36 workers: 2, corpus: 1425 (49m46s ago), crashers: 0, restarts: 1/10000, execs: 598487533 (11382/sec), cover: 1513, uptime: 14h36m
```

The last line shows us:

* There are two workers running.
* The fuzzer has been running for 14 hours and 36 minutes.
* Zero crashing-cases have been found.
  * If you see `crasers: 1`, or higher, then you have something to examine.
* 598487533 executions have taken place
* We're averaging the execution of a test-case 11382 times a sec.


## Output

You can kill (Ctrl-c) the fuzzer and restart it, and it will keep going from where it left off, because all state is stored in `./workdir`.

If you find crashing-input please report a bug.  The cases that caused a crash will be found int `workdir/crashers`.
