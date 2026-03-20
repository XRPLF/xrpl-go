// Package serdes provides utilities to parse and serialize XRPL binary data fields.
package serdes

import (
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/serdes/interfaces"
)

var (
	// Static errors

	// ErrParserOutOfBound is returned when the parser is out of bounds.
	ErrParserOutOfBound = errors.New("parser out of bounds")
	// ErrInvalidTypecode is returned when the typecode is invalid.
	ErrInvalidTypecode = errors.New("invalid typecode")
	// ErrInvalidFieldcode is returned when the fieldcode is invalid.
	ErrInvalidFieldcode = errors.New("invalid fieldcode")
)

// BinaryParser parses binary-encoded XRPL data into field instances based on definitions.
type BinaryParser struct {
	data        []byte
	definitions interfaces.Definitions
}

// NewBinaryParser returns a new BinaryParser initialized with the given data.
func NewBinaryParser(d []byte, definitions interfaces.Definitions) *BinaryParser {
	return &BinaryParser{
		data:        d,
		definitions: definitions,
	}
}

// ReadField reads the next field in the data.
// It reads the field's header, fetches the field's name based on its header,
// and then gets the FieldInstance for that field name.
func (p *BinaryParser) ReadField() (*definitions.FieldInstance, error) {
	fh, err := p.readFieldHeader()
	if err != nil {
		return nil, err
	}
	fn, err := p.definitions.GetFieldNameByFieldHeader(*fh)
	if err != nil {
		return nil, err
	}
	f, err := p.definitions.GetFieldInstanceByFieldName(fn)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// readFieldHeader reads the header of the next field in the data.
func (p *BinaryParser) readFieldHeader() (*definitions.FieldHeader, error) {
	// Read the first byte of the field header
	typeCode, err := p.ReadByte()
	if err != nil {
		return nil, err
	}

	// The field code is the last 4 bits of the first byte
	fieldCode := typeCode & 15
	typeCode >>= 4

	// Read the type code if it's not in the first byte
	if typeCode == 0 {
		typeCode, err = p.ReadByte()
		if err != nil {
			return nil, err
		}
		if typeCode == 0 || typeCode < 16 {
			return nil, ErrInvalidTypecode
		}
	}

	// Read the field code if it's not in the first byte
	if fieldCode == 0 {
		fieldCode, err = p.ReadByte()
		if err != nil {
			return nil, err
		}
		if fieldCode == 0 || fieldCode < 16 {
			return nil, ErrInvalidFieldcode
		}
	}

	// Return the field header
	return &definitions.FieldHeader{
		TypeCode:  int32(typeCode),
		FieldCode: int32(fieldCode),
	}, nil
}

// ReadByte reads the next byte in the data.
// It returns an error if no more data is available.
func (p *BinaryParser) ReadByte() (byte, error) {
	if len(p.data) < 1 {
		return 0, ErrParserOutOfBound
	}
	b := p.data[0]
	p.data = p.data[1:]
	return b, nil
}

// Peek returns the next byte in the data without advancing the read cursor.
// It returns an error if no more data is available.
func (p *BinaryParser) Peek() (byte, error) {
	if len(p.data) < 1 {
		return 0, ErrParserOutOfBound
	}
	return p.data[0], nil
}

// ReadBytes reads the next n bytes in the data.
// It returns an error if fewer than n bytes are available.
func (p *BinaryParser) ReadBytes(n int) ([]byte, error) {
	var bytes []byte
	for range n {
		b, err := p.ReadByte()
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, b)
	}
	return bytes, nil
}

// HasMore returns true if there is more data to read, and false otherwise.
func (p *BinaryParser) HasMore() bool {
	return len(p.data) != 0
}

// ReadVariableLength reads a variable-length field from the binary data
// and returns the length as an integer. The length is determined by
// 1 to 3 bytes length prefix according to XRPL documentation.
func (p *BinaryParser) ReadVariableLength() (int, error) {
	b1, err := p.ReadByte()
	if err != nil {
		return 0, err
	}
	if b1 > 192 && b1 < 241 {
		b2, err := p.ReadByte()
		if err != nil {
			return 0, err
		}
		return 193 + ((int(b1) - 193) * 256) + int(b2), nil
	} else if b1 > 240 && b1 < 255 {
		b2, err := p.ReadByte()
		if err != nil {
			return 0, err
		}
		b3, err := p.ReadByte()
		if err != nil {
			return 0, err
		}
		return 12481 + ((int(b1) - 241) * 65536) + (int(b2) * 256) + int(b3), nil
	}
	return int(b1), nil
}
