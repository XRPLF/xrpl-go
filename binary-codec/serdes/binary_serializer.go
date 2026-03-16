package serdes

import (
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/interfaces"
)

// ErrLengthPrefixTooLong is returned when the length of the value exceeds 918744 bytes.
var ErrLengthPrefixTooLong = errors.New("length of value must not exceed 918744 bytes of data")

// BinarySerializer serializes field instances and their values into binary format.
type BinarySerializer struct {
	sink         []byte
	fieldIDCodec interfaces.FieldIDCodec
}

// NewBinarySerializer creates a BinarySerializer with the provided FieldIDCodec.
func NewBinarySerializer(fieldIDCodec interfaces.FieldIDCodec) *BinarySerializer {
	return &BinarySerializer{
		fieldIDCodec: fieldIDCodec,
	}
}

func (s *BinarySerializer) put(v []byte) {
	s.sink = append(s.sink, v...)
}

// GetSink returns the serialized binary data accumulated in the serializer.
func (s *BinarySerializer) GetSink() []byte {
	return s.sink
}

// WriteFieldAndValue serializes a field instance and its value into binary format and writes it to the sink.
func (s *BinarySerializer) WriteFieldAndValue(fi definitions.FieldInstance, value []byte) error {
	h, err := s.fieldIDCodec.Encode(fi.FieldName)
	if err != nil {
		return err
	}

	s.put(h)

	if fi.IsVLEncoded {
		vl, err := encodeVariableLength(len(value))
		if err != nil {
			return err
		}
		s.put(vl)
	}

	s.put(value)

	if fi.Type == "STObject" {
		s.put([]byte{0xE1})
	}
	return nil
}

func encodeVariableLength(length int) ([]byte, error) {
	if length <= 192 {
		return []byte{byte(length)}, nil
	}
	if length < 12480 {
		length -= 193
		b1 := byte((length >> 8) + 193)
		b2 := byte((length & 0xFF))
		return []byte{b1, b2}, nil
	}
	if length <= 918744 {
		length -= 12481
		b1 := byte((length >> 16) + 241)
		b2 := byte((length >> 8) & 0xFF)
		b3 := byte(length & 0xFF)
		return []byte{b1, b2, b3}, nil
	}
	return nil, ErrLengthPrefixTooLong
}
