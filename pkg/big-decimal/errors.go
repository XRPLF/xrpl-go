package bigdecimal

import (
	"errors"
	"fmt"
)

// ErrInvalidZeroValue indicates the value string represents zero or is invalid zero.
var ErrInvalidZeroValue = errors.New("value cannot be zero")

// ErrInvalidCharacter is returned when a string contains disallowed characters.
type ErrInvalidCharacter struct {
	Allowed string
}

// Error implements the error interface for ErrInvalidCharacter
func (e ErrInvalidCharacter) Error() string {
	return fmt.Sprintf("value contains invalid character: sonly the following are allowed: %s", e.Allowed)
}
