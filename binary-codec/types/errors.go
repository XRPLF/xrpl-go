//revive:disable:var-naming
package types

import "errors"

var (
	errNotValidJSON         = errors.New("not a valid json")
	errDecodeClassicAddress = errors.New("unable to decode classic address")
	errReadBytes            = errors.New("read bytes error")
	// ErrUInt8OutOfRange is returned when a value is outside the uint8 range (0-255).
	ErrUInt8OutOfRange = errors.New("value out of uint8 range (0-255)")
	// ErrUInt16OutOfRange is returned when a value is outside the uint16 range (0-65535).
	ErrUInt16OutOfRange = errors.New("value out of uint16 range (0-65535)")
	// ErrUInt32OutOfRange is returned when a value is outside the uint32 range (0-4294967295).
	ErrUInt32OutOfRange = errors.New("value out of uint32 range (0-4294967295)")
	// ErrUInt64OutOfRange is returned when a value is outside the uint64 range.
	ErrUInt64OutOfRange = errors.New("value out of uint64 range")
)
