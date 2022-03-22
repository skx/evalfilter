//go:build go1.18
// +build go1.18

package evalfilter

import (
	"strings"
	"testing"
)

// FuzzEvaluator runs the fuzz-testing against our evaluation engine
func FuzzEvaluator(f *testing.F) {

	// empty + whitespace
	f.Add([]byte(nil))
	f.Add([]byte(""))
	f.Add([]byte(`\n\r\t`))
	f.Add([]byte("`/bin/ls`"))

	// hash
	f.Add([]byte(`
a = { "Name": "Steve",
      "Age": 2020 - 1976,
      "Location": "Helsinki", }
`))

	// iteration
	f.Add([]byte(`
printf("Iterating over the hash via the output of keys():\n");
foreach name in k {
  printf("\tKEY:%s Value:'%s'\n", name, string(a[name]))
}`))

	// loop
	f.Add([]byte(`
i = 3;

print( "Starting value: ", i, "\n");

while( i < 10 ) {
   print("In loop, value: ", i, "\n");
   i++
}
`))

	// Sorting
	f.Add([]byte(`
foreach index,entry in ["foo", "bar","baz"] {
   printf("\t%d:%s\n", index, entry );
}`))

	// Known errors we might see
	known := []string{
		"no prefix parse function",
		"invalid character for identifier",
		"unexpected end of file reached",
		"unterminated string",
		"unterminated regular expression",
		"expected next token",
		"illegal regexp flag",
		"token to",
		"as integer",
		"as float",
		"must be an identifier",
		"nested ternary",
		"must be ident",
		"incomplete block",
		"unterminated switch statement",
		"expected case|default",
		"unterminated function parameters",
	}

	// switch
	f.Add([]byte(`
function test( name ) {

  switch( name ) {
    case "Ste" + "ve" {
	printf("I know you %s - expression-match!\n", name );
    }
    case "Steven" {
	printf("I know you %s - literal-match!\n", name );
    }
    case /^steve$/ {
        printf("I know you %s - regexp-match!\n", name );
    }
    default {
	printf("I don't know who you are %s\n", name );
    }
  }
}

test("me")
test("Steve")
test("steve")
test("Steven")
test(2/2)
test(false)
`))

	f.Fuzz(func(t *testing.T, input []byte) {

		// Create the helper
		eval := New(string(input))

		//
		// Parse the program.
		//
		// Create the bytecode.
		//
		// Optimize it.
		//
		err := eval.Prepare()

		// Found an error
		if err != nil {

			falsePositive := false

			// is it a known one?
			for _, expected := range known {
				if strings.Contains(err.Error(), expected) {
					falsePositive = true
				}
			}

			if !falsePositive {
				t.Fatalf("error running input %s -> %s", input, err.Error())
			}

		}
	})
}
