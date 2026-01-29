//revive:disable:var-naming
package types

import (
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

const (
	// ArrayEndMarker denotes the end of an STArray serialized sequence.
	ArrayEndMarker = 0xF1
	// ObjectEndMarker denotes the end of an STObject within an STArray.
	ObjectEndMarker = 0xE1
)

// STArray represents an array of STObject instances.
type STArray struct{}

// ErrNotSTObjectInSTArray is returned when a non-STObject value is found in an STArray.
var ErrNotSTObjectInSTArray = errors.New("not STObject in STArray. Array fields must be STObjects")

// FromJSON is a method that takes a JSON value (which should be a slice of JSON objects),
// and converts it to a byte slice, representing the serialized form of the STArray.
// It loops through the JSON slice, and for each element, calls the FromJSON method
// of an STObject, appending the resulting byte slice to a "sink" slice.
// The method returns an error if the JSON value is not a slice.
func (t *STArray) FromJSON(json any) ([]byte, error) {
	switch v := json.(type) {
	case []any:
		json = v
	case []map[string]any:
		json = make([]any, len(v))
		for i, m := range v {
			json.([]any)[i] = m
		}
	default:
		return nil, ErrNotSTObjectInSTArray
	}

	var sink []byte
	for _, v := range json.([]any) {
		st := NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))
		b, err := st.FromJSON(v)
		if err != nil {
			return nil, err
		}
		sink = append(sink, b...)
	}
	sink = append(sink, ArrayEndMarker)

	return sink, nil
}

// ToJSON is a method that takes a BinaryParser and optional parameters, and converts
// the serialized byte data back to a JSON value.
// The method loops until the BinaryParser has no more data, and for each loop,
// it calls the ToJSON method of an STObject, appending the resulting JSON value to a "value" slice.
func (t *STArray) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	// Initialize as empty slice (not nil) so empty arrays marshal to [] not null
	value := make([]any, 0)

	for p.HasMore() {
		fi, err := p.ReadField()
		if err != nil {
			return nil, err
		}

		// Check for ArrayEndMarker FIRST (handles empty arrays)
		if fi.FieldName == "ArrayEndMarker" {
			break
		}

		// All array elements must be STObjects
		if fi.Type != "STObject" {
			return nil, ErrNotSTObjectInSTArray
		}

		st := GetSerializedType(fi.Type)
		res, err := st.ToJSON(p)
		if err != nil {
			return nil, err
		}

		stObj := make(map[string]any)
		stObj[fi.FieldName] = res
		value = append(value, stObj)
	}
	return value, nil
}
