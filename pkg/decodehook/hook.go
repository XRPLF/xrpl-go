// Package decodehook provides shared decode hooks for mapstructure.
package decodehook

import (
	"encoding/json"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

var jsonUnmarshalerType = reflect.TypeFor[json.Unmarshaler]()

// JSON returns a mapstructure decode hook that delegates to json.Unmarshaler
// for any target type that implements it. The source value is re-encoded to JSON
// and then passed to the target's UnmarshalJSON method, which lets individual
// types define their own flexible decoding logic (e.g. accepting both strings
// and numbers for a field that is modelled as a string in Go).
func JSON() mapstructure.DecodeHookFuncValue {
	return func(from reflect.Value, to reflect.Value) (any, error) {
		toType := to.Type()
		if toType.Kind() != reflect.Pointer {
			toType = reflect.PointerTo(toType)
		}
		if !toType.Implements(jsonUnmarshalerType) {
			return from.Interface(), nil
		}

		// Re-encode the source (a map[string]any or similar) to JSON bytes.
		jsonBytes, err := json.Marshal(from.Interface())
		if err != nil {
			return nil, err
		}

		// Allocate a new value of the concrete (non-pointer) target type and
		// unmarshal into it via the custom UnmarshalJSON.
		target := reflect.New(to.Type())
		if err := json.Unmarshal(jsonBytes, target.Interface()); err != nil {
			return nil, err
		}
		return target.Elem().Interface(), nil
	}
}
