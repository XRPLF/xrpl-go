package types

import (
	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// SerializedType is an interface representing any type that can be serialized
// and deserialized to and from JSON.
// The FromJson method takes a JSON value and converts it to a byte slice.
// The ToJson method takes a BinaryParser and optional parameters, and converts
// the serialized byte data back to a JSON value.
type SerializedType interface {
	FromJSON(json any) ([]byte, error)
	ToJSON(parser interfaces.BinaryParser, opts ...int) (any, error)
}

// GetSerializedType is a function that returns the correct SerializedType instance
// based on the string parameter.
// It creates a new instance of the type described by the parameter, allowing
// the appropriate methods of that type to be called.
// If the input string does not match a known type, the function returns nil.
func GetSerializedType(t string) SerializedType {
	switch t {
	case "UInt8":
		return &UInt8{}
	case "UInt16":
		return &UInt16{}
	case "UInt32":
		return &UInt32{}
	case "UInt64":
		return &UInt64{}
	case "Hash128":
		return NewHash128()
	case "Hash160":
		return NewHash160()
	case "Hash256":
		return NewHash256()
	case "AccountID":
		return &AccountID{}
	case "Number":
		return &Number{}
	case "Amount":
		return &Amount{}
	case "Vector256":
		return &Vector256{}
	case "Blob":
		return &Blob{}
	case "STObject":
		return NewSTObject(serdes.NewBinarySerializer(serdes.NewFieldIDCodec(definitions.Get())))
	case "STArray":
		return &STArray{}
	case "PathSet":
		return &PathSet{}
	case "XChainBridge":
		return &XChainBridge{}
	case "Issue":
		return &Issue{}
	case "Currency":
		return &Currency{}
	}
	return nil
}
