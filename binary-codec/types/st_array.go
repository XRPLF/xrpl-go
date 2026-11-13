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

// FromJSON converts a JSON value, which must be a slice of JSON objects, to the
// serialized form of the STArray, including the array end marker.
// It returns an error if the JSON value is not a slice.
func (t *STArray) FromJSON(json any) ([]byte, error) {
	return arrayFromJSON(json, 0)
}

func arrayFromJSON(json any, depth int) ([]byte, error) {
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
		b, err := st.fromJSON(v, nil, depth)
		if err != nil {
			return nil, err
		}
		sink = append(sink, b...)
	}
	sink = append(sink, ArrayEndMarker)

	return sink, nil
}

// ToJSON reads serialized STObjects from the BinaryParser until the array end marker
// or the end of the data, and returns them as a slice of JSON objects.
func (t *STArray) ToJSON(p interfaces.BinaryParser, _ ...int) (any, error) {
	return arrayToJSON(p, 0)
}

func arrayToJSON(p interfaces.BinaryParser, depth int) (any, error) {
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
		if depth+1 > maxNestingDepth {
			return nil, ErrMaxNestingDepth
		}

		res, err := objectToJSON(p, depth+1)
		if err != nil {
			return nil, err
		}

		stObj := make(map[string]any)
		stObj[fi.FieldName] = res
		value = append(value, stObj)
	}
	return value, nil
}
