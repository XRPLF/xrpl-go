package common

import "errors"

// ErrInvalidLedgerSpecifier is returned when a ledger specifier string is not one of the
// named ledger titles (current, validated, closed).
var ErrInvalidLedgerSpecifier = errors.New("invalid ledger specifier")
