package definitions

import "github.com/ugorji/go/codec"

// FieldInstance is a struct that represents a field instance.
type FieldInstance struct {
	FieldName string
	*FieldInfo
	FieldHeader *FieldHeader
	Ordinal     int32
}

// FieldInfo is a struct that represents the field info.
type FieldInfo struct {
	Nth            int32
	IsVLEncoded    bool
	IsSerialized   bool
	IsSigningField bool
	Type           string
}

// FieldHeader is a struct that represents the field header.
type FieldHeader struct {
	TypeCode  int32
	FieldCode int32
}

// CreateFieldHeader creates a new field header.
func (d *Definitions) CreateFieldHeader(tc, fc int32) FieldHeader {
	return FieldHeader{
		TypeCode:  tc,
		FieldCode: fc,
	}
}

type fieldInstanceMap map[string]*FieldInstance

// CodecEncodeSelf implements the codec.SelfEncoder interface.
func (fi *fieldInstanceMap) CodecEncodeSelf(_ *codec.Encoder) {}

// CodecDecodeSelf implements the codec.SelfDecoder interface.
func (fi *fieldInstanceMap) CodecDecodeSelf(d *codec.Decoder) {
	var x [][]any
	d.MustDecode(&x)
	y := convertToFieldInstanceMap(x)
	*fi = y
}
