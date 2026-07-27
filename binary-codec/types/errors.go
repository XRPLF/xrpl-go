//revive:disable:var-naming
package types

import "errors"

var (
	errNotValidJSON         = errors.New("not a valid json")
	errDecodeClassicAddress = errors.New("unable to decode classic address")
	errReadBytes            = errors.New("read bytes error")
	// ErrDuplicateXAddressTag is returned when an X-address contains a tag and the transaction also defines the matching tag field.
	ErrDuplicateXAddressTag = errors.New("duplicate X-address tag")
	// ErrAccountIDTagNotAllowed is returned when an AccountID-typed field receives an X-address that carries an embedded tag.
	// Only top-level Account and Destination can carry an embedded tag (promoted to SourceTag/DestinationTag).
	ErrAccountIDTagNotAllowed = errors.New("AccountID field cannot have an associated tag")
	// ErrUInt8OutOfRange is returned when a value is outside the uint8 range (0-255).
	ErrUInt8OutOfRange = errors.New("value out of uint8 range (0-255)")
	// ErrUInt16OutOfRange is returned when a value is outside the uint16 range (0-65535).
	ErrUInt16OutOfRange = errors.New("value out of uint16 range (0-65535)")
	// ErrUInt32OutOfRange is returned when a value is outside the uint32 range (0-4294967295).
	ErrUInt32OutOfRange = errors.New("value out of uint32 range (0-4294967295)")
	// ErrInt32OutOfRange is returned when a value is outside the int32 range (-2147483648 to 2147483647).
	ErrInt32OutOfRange = errors.New("value out of int32 range (-2147483648 to 2147483647)")
	// ErrInvalidStringNumber is returned when a value contains only legal XRPL String Number characters
	// but does not match the grammar (e.g. "00.1", ".5", "1.", "1e", "", "-", "+1"). Distinct from
	// bigdecimal.ErrInvalidCharacter, which signals an out-of-set character.
	ErrInvalidStringNumber = errors.New("value is not a valid XRPL String Number")
)
