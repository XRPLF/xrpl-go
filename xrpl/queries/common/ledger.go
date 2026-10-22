// Package common provides shared types for XRPL ledger specifiers and parsing utilities.
//
//revive:disable:var-naming
package common

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// LedgerSpecifier defines an interface for types that can represent a ledger identifier as a string.
type LedgerSpecifier interface {
	Ledger() string
}

// UnmarshalLedgerSpecifier parses JSON data into a LedgerSpecifier, handling both string and numeric forms.
func UnmarshalLedgerSpecifier(data []byte) (LedgerSpecifier, error) {
	if len(data) == 0 {
		return nil, nil
	}
	switch data[0] {
	case '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return nil, err
		}
		switch s {
		case Current.Ledger():
			return Current, nil
		case Validated.Ledger():
			return Validated, nil
		case Closed.Ledger():
			return Closed, nil
		}
		return nil, fmt.Errorf("%w: %q", ErrInvalidLedgerSpecifier, s)
	default:
		var i LedgerIndex
		if err := json.Unmarshal(data, &i); err != nil {
			return nil, err
		}
		return i, nil
	}
}

// LedgerIndex represents a numeric index of a ledger.
type LedgerIndex uint32

// Ledger returns the LedgerIndex as a decimal string.
func (l LedgerIndex) Ledger() string {
	return strconv.FormatUint(uint64(l), 10)
}

// Uint32 returns the LedgerIndex as a uint32.
func (l LedgerIndex) Uint32() uint32 {
	return uint32(l)
}

// Int returns the LedgerIndex as an int.
func (l LedgerIndex) Int() int {
	return int(l)
}

// LedgerTitle represents a named ledger specifier like "current", "validated", or "closed".
type LedgerTitle string

// Named ledger specifiers.
const (
	// Current is the current open ledger.
	Current LedgerTitle = "current"
	// Validated is the most recently validated ledger.
	Validated LedgerTitle = "validated"
	// Closed is the most recently closed ledger.
	Closed LedgerTitle = "closed"
)

// Ledger returns the LedgerTitle as a string.
func (l LedgerTitle) Ledger() string {
	return string(l)
}

// LedgerHash represents the hash of a ledger.
type LedgerHash string
