package account

import "errors"

// ErrNoAccountID is returned when no account ID is specified in a request.
var ErrNoAccountID = errors.New("no account ID specified")
