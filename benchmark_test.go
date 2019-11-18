package evalfilter

import (
	"fmt"
	"testing"
)

// Benchmark_evalfilter_complex_map - This is a complex test against a map.
//
// This is designed to see if there is any speed difference in using a
// map vs a structure.  See `Benchmark_evalfilter_complex_obj` for
// the alternative.
func Benchmark_evalfilter_complex_map(b *testing.B) {

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

// Benchmark_evalfilter_complex_obj - This is a complex test against an object
//
// This is designed to see if there is any speed difference in using a
// map vs a structure.  See `Benchmark_evalfilter_complex_map` for
// the alternative.
func Benchmark_evalfilter_complex_obj(b *testing.B) {

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
	type Input struct {
		Country string
		Origin  string
		Adults  int
		Value   int
	}
	o := &Input{Country: "RU", Origin: "MOW", Adults: 1, Value: 99}

	var ret bool

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ret, err = eval.Run(o)
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
	eval := New(`return ( Name == "Steve" && Name == "Steve" );`)

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
	type Input struct {
		Name string
	}
	obj := Input{Name: "Steve"}

	var ret bool

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ret, err = eval.Run(obj)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !ret {
		b.Fail()
	}
}

// Benchmark_evalfilter_trivial - This is a trivial test that uses no fields.
func Benchmark_evalfilter_trivial(b *testing.B) {

	//
	// Prepare the script
	//
	eval := New(`if ( 1 + 2 * 3 == 7 ) { return true; } return false;`)

	//
	// Ensure this compiled properly.
	//
	err := eval.Prepare()
	if err != nil {
		fmt.Printf("Failed to compile: %s\n", err.Error())
		return
	}

	var ret bool

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ret, err = eval.Run(nil)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !ret {
		b.Fail()
	}
}
