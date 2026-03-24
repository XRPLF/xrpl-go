// Package types contains data structures for account query types.
// revive:disable:var-naming
package types

import "errors"

// ErrNoAccountID is returned when no account ID is specified.
var ErrNoAccountID = errors.New("no account ID specified")
