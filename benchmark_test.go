package evalfilter

import (
	"fmt"
	"testing"
)

// Benchmark_evalfilter_complex - This is a complex test against a map.
func Benchmark_evalfilter_complex(b *testing.B) {

	//
	// Prepare the script
	//
	eval := New(`if ( (Origin == "MOW" || Country == "RU") && (Value >= 100 || Adults == 1) ) { return true; }  else { return false; }`)

	//
	// Ensure this compiled properly.
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile: %s\n", err.Error())
		return
	}

	//
	// Create the object we'll test against.
	//
	params := make(map[string]interface{})
	params["Origin"] = "MOW"
	params["Country"] = "RU"
	params["Adults"] = 1
	params["Value"] = 99

	var ret bool

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ret, err = eval.Run(params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !ret {
		b.Fail()
	}
}

// Benchmark_evalfilter_simple - This is a very simple test which should
// be pretty quick.
func Benchmark_evalfilter_simple(b *testing.B) {

	//
	// Prepare the script
	//
	eval := New(`return ( Name == "Steve" );`)

	//
	// Ensure this compiled properly.
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile: %s\n", err.Error())
		return
	}

	//
	// Create the object we'll test against.
	//
	params := make(map[string]interface{})
	params["Name"] = "Steve"

	var ret bool

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ret, err = eval.Run(params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !ret {
		b.Fail()
	}
}
