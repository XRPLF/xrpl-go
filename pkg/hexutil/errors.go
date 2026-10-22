package hexutil

import "errors"

// ErrInvalidHexLength is returned when a hex string does not decode to the expected number of bytes.
var ErrInvalidHexLength = errors.New("invalid hex length")
