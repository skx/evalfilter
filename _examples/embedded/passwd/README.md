# passwd

This example demonstrates running a user-supplied script, along with a custom function.

## Usage

Compile via `go build .`, then launch the program with one of the script-files as the single argument:

```
./passwd and.txt
Loading and.txt
**********************************************************************
User nobody gave result true
```

That showed that that the script returned `true` for the user `nobody`,
matching the condition:

    if ( Username == "nobody" && Uid > "1000" ) {

## Notes

The second test there looks odd, but that is because `Uid` is a string-field,
rather than a numeric one.
