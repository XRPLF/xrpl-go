package types

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/Peersyst/xrpl-go/binary-codec/definitions"
	"github.com/Peersyst/xrpl-go/binary-codec/types/interfaces"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

var (
	// ErrUnsupportedPermissionType is returned when the JSON type is unsupported for PermissionValue.
	ErrUnsupportedPermissionType = errors.New("unsupported JSON type for PermissionValue")
	// ErrPermissionValueOutOfRange is returned when the permission value cannot be coerced to a uint32 in the [0, 4294967295] range.
	ErrPermissionValueOutOfRange = errors.New("permission value out of uint32 range")
)

// PermissionValue represents a 32-bit unsigned integer permission value.
type PermissionValue struct{}

// FromJSON converts a JSON value into a serialized byte slice representing a 32-bit unsigned integer permission value.
// If the input value is a string, it's assumed to be a permission name, and the method will
// attempt to convert it into a corresponding permission value. If the conversion fails, an error is returned.
func (p *PermissionValue) FromJSON(value any) ([]byte, error) {
	if s, ok := value.(string); ok {
		pv, err := definitions.Get().GetDelegatablePermissionValueByName(s)
		if err != nil {
			return nil, err
		}
		value = pv
	}

	ui32, ok := typecheck.ToUint32(value)
	if !ok {
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint64, float32, float64, json.Number:
			return nil, ErrPermissionValueOutOfRange
		default:
			return nil, ErrUnsupportedPermissionType
		}
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, ui32)
	return buf, nil
}

// ToJSON takes a BinaryParser and optional parameters, and converts the serialized byte data
// back into a JSON value. If a permission name is found for the value, it returns the name;
// otherwise, it returns the numeric value. If the parsing fails, an error is returned.
func (p *PermissionValue) ToJSON(parser interfaces.BinaryParser, _ ...int) (any, error) {
	b, err := parser.ReadBytes(4)
	if err != nil {
		return nil, err
	}

	permissionValue := binary.BigEndian.Uint32(b)

	// #nosec G115
	if name, err := definitions.Get().GetDelegatablePermissionNameByValue(int32(permissionValue)); err == nil {
		return name, nil
	}

	return permissionValue, nil
}
