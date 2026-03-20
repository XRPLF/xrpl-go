package transactions

import "errors"

// ErrNoTxBlob is returned when no TxBlob is defined in the SubmitRequest.
var ErrNoTxBlob = errors.New("no TxBlob defined")
