package definitions

import (
	"errors"
	"fmt"
)

// definitions

// ErrUnableToCastFieldInfo is returned when the field info cannot be cast.
var ErrUnableToCastFieldInfo = errors.New("unable to cast to field info")

// Dynamic errors

// NotFoundError is an error that occurs when a value is not found.
type NotFoundError struct {
	Instance string
	Input    string
}

// Error implements the error interface.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%v %v not found", e.Instance, e.Input)
}

// NotFoundErrorInt is an error that occurs when a value is not found.
type NotFoundErrorInt struct {
	Instance string
	Input    int32
}

// Error implements the error interface.
func (e *NotFoundErrorInt) Error() string {
	return fmt.Sprintf("%v %v not found", e.Instance, e.Input)
}

// NotFoundErrorFieldHeader is an error that occurs when a value is not found.
type NotFoundErrorFieldHeader struct {
	Instance string
	Input    FieldHeader
}

// Error implements the error interface.
func (e *NotFoundErrorFieldHeader) Error() string {
	return fmt.Sprintf("%v %v not found", e.Instance, e.Input)
}
