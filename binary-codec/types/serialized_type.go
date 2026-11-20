//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
)

// DecodeContext describes the field being read. Its zero value decodes a root
// object or a standalone fixed-width value.
type DecodeContext struct {
	// FieldName selects field-specific JSON representations, such as UInt64 amounts.
	FieldName string
	// Length is the byte length from the variable-length prefix. A pointer to
	// zero represents an empty field. A nil pointer means no prefix was read.
	Length *int
	depth  int
}

// Depth is the nesting level of this field, starting at zero for the root.
func (c DecodeContext) Depth() int { return c.depth }

// SerializedType converts between a JSON value and its binary representation.
type SerializedType interface {
	FromJSON(json any) ([]byte, error)
	ToJSON(parser interfaces.BinaryParser, ctx DecodeContext) (any, error)
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
	case "Int32":
		return &Int32{}
	case "UInt64":
		return &UInt64{}
	case "Hash128":
		return NewHash128()
	case "Hash160":
		return NewHash160()
	case "Hash192":
		return NewHash192()
	case "Hash256":
		return NewHash256()
	case "AccountID":
		return &AccountID{}
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
	case "Number":
		return &Number{}
	}
	return nil
}
