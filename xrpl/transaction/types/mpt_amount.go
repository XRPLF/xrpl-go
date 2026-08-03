package types

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// MaxMPTAmount is the largest MPT amount currently supported by the protocol.
const MaxMPTAmount MPTAmount = 1<<63 - 1

// MPTAmount represents a non-negative number of MPT units. XRPL JSON encodes
// the value as a quoted base-10 integer to preserve full integer precision.
type MPTAmount uint64

// Uint64 returns the amount as a uint64.
func (m MPTAmount) Uint64() uint64 {
	return uint64(m)
}

// String returns the base-10 representation of the amount.
func (m MPTAmount) String() string {
	return strconv.FormatUint(uint64(m), 10)
}

// IsZero reports whether the amount is zero.
func (m MPTAmount) IsZero() bool {
	return m == 0
}

// IsValid reports whether the amount is within the protocol MPT range.
func (m MPTAmount) IsValid() bool {
	return m <= MaxMPTAmount
}

// Flatten returns the quoted-decimal JSON value expected by the binary codec.
func (m MPTAmount) Flatten() any {
	return m.String()
}

// MarshalJSON serializes the amount as a quoted base-10 integer.
func (m MPTAmount) MarshalJSON() ([]byte, error) {
	if !m.IsValid() {
		return nil, ErrInvalidMPTAmount
	}
	return json.Marshal(m.String())
}

// UnmarshalJSON parses a quoted base-10 integer in the protocol MPT range.
func (m *MPTAmount) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%w: expected a quoted base-10 integer", ErrInvalidMPTAmount)
	}
	return m.UnmarshalText([]byte(value))
}

// UnmarshalText parses a base-10 integer in the protocol MPT range.
func (m *MPTAmount) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidMPTAmount
	}
	for _, digit := range data {
		if digit < '0' || digit > '9' {
			return ErrInvalidMPTAmount
		}
	}

	value, err := strconv.ParseUint(string(data), 10, 63)
	if err != nil {
		return ErrInvalidMPTAmount
	}
	*m = MPTAmount(value)
	return nil
}
