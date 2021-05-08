// Package reflection is the magic which lets us access members of
// objects and structures by name.
//
// Given an input object of the form:
//
//   foo{
//     bar {
//      baz: "Steve"
//     }
//   }
//
// We can access that as `foo.bar.baz`.
//
// This is all a bit horrid, because we've tried to unify the handling by
// first of all encoding the input-object to JSON, then decoding it again.
//
// The upshot is a working system though, for either objects or maps.
package reflection

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/skx/evalfilter/v2/object"
)

// Reflection holds our state
type Reflection struct {

	// input is the object we're passed in our constructor
	input interface{}

	// fields contains the contents of all the fields in the object
	// or map we're executing against.  We discover these via reflection
	// at run-time.
	//
	// Reflection is slow so the map here is used as a cache, avoiding
	// the need to reparse the same object multiple times.
	fields map[string]object.Object

	// Have we walked our object?
	processed bool
}

// New creates a new helper which will inspect the fields of the given
// structure, or object.
func New(input interface{}) *Reflection {

	r := &Reflection{
		input:     input,
		fields:    make(map[string]object.Object),
		processed: false,
	}

	return r
}

// Get looks up the value from the parsed structure, based upon the name of
// the field.
func (r *Reflection) Get(key string) (object.Object, error) {

	// Not processed?
	if !r.processed {

		// Then process.
		e := r.Process()
		if e != nil {
			return &object.Null{}, fmt.Errorf("error processing for key %s - %s", key, e.Error())
		}
	}

	// Lookup the key, if we found it return it.
	val, ok := r.fields[key]
	if ok {
		return val, nil
	}

	// Otherwise we have an error
	return &object.Null{}, fmt.Errorf("key not found '%s'", key)
}

// Process handles getting the fields of the input we were given, via
// reflection and storing the values in a local cache.
func (r *Reflection) Process() error {

	// Already processed?  Then avoid a repeat
	if r.processed {
		return nil
	} else {
		r.processed = true
	}

	// Yeah, horrid.
	data, err := json.Marshal(r.input)
	if err != nil {
		return err
	}

	// Now we have json.
	//
	// unmarshal
	var obj interface{}
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return err
	}

	//
	// Flatten and store
	//
	r.flatten(obj, "")

	return nil
}

func (r *Reflection) flatten(obj interface{}, prefix string) {
	switch obj := obj.(type) {
	case bool:
		r.fields[prefix] = &object.Boolean{Value: obj}
	case int:
		r.fields[prefix] = &object.Integer{Value: int64(obj)}
	case int32:
		r.fields[prefix] = &object.Integer{Value: int64(obj)}
	case int64:
		r.fields[prefix] = &object.Integer{Value: int64(obj)}
	case float32:
		// really an int?
		if float32(int32(obj)) == obj {
			r.fields[prefix] = &object.Integer{Value: int64(obj)}
		} else {
			r.fields[prefix] = &object.Float{Value: float64(obj)}
		}
	case time.Time:
		r.fields[prefix] = &object.Integer{Value: interface{}(obj).(time.Time).Unix()}
	case float64:
		// really an int?
		if float64(int64(obj)) == obj {
			r.fields[prefix] = &object.Integer{Value: int64(obj)}
		} else {
			r.fields[prefix] = &object.Float{Value: float64(obj)}
		}
	case string:
		r.fields[prefix] = &object.String{Value: obj}
	case []interface{}:
		var el []object.Object

		for _, v := range obj {

			var tmp object.Object

			switch v := v.(type) {
			case bool:
				tmp = &object.Boolean{Value: v}
			case int:
				tmp = &object.Integer{Value: int64(v)}
			case int32:
				tmp = &object.Integer{Value: int64(v)}
			case int64:
				tmp = &object.Integer{Value: int64(v)}
			case float32:
				// really an int?
				if float32(int32(v)) == v {
					tmp = &object.Integer{Value: int64(v)}
				} else {
					tmp = &object.Float{Value: float64(v)}
				}
			case float64:
				// really an int?
				if float64(int64(v)) == v {
					tmp = &object.Integer{Value: int64(v)}
				} else {
					tmp = &object.Float{Value: float64(v)}
				}
			case string:
				tmp = &object.String{Value: v}
			}

			el = append(el, tmp)

		}
		r.fields[prefix] = &object.Array{Elements: el}

	case map[string]interface{}:
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if len(prefix) > 0 {

				r.flatten(obj[k], fmt.Sprintf("%s.%s", prefix, k))
			} else {
				r.flatten(obj[k], fmt.Sprintf("%s", k))
			}
		}
	default:
		fmt.Printf("%s = null\n", prefix)
	}
}
