// Package types contains data structures for wallet operations and batch signing.
// revive:disable:var-naming
package types

import (
	"errors"
	"fmt"
)

var (
	// batch

	// ErrBatchSignableInvalid is returned when the batch signable is invalid.
	ErrBatchSignableInvalid = errors.New("batch signable is invalid")

	// fields

	// ErrAccountFieldIsNotAString is returned when the outer Batch account is missing or is not a string.
	ErrAccountFieldIsNotAString = errors.New("account field is not a string")
	// ErrSequenceFieldIsNotAnUint32 is returned when the outer Batch sequence is not a uint32.
	ErrSequenceFieldIsNotAnUint32 = errors.New("sequence field is not an uint32")
	// ErrTicketSequenceFieldIsNotAnUint32 is returned when the outer Batch ticket sequence is not a uint32.
	ErrTicketSequenceFieldIsNotAnUint32 = errors.New("ticket sequence field is not an uint32")
	// ErrBatchSequenceAndTicket is returned when a Batch specifies a non-zero Sequence and a TicketSequence.
	ErrBatchSequenceAndTicket = errors.New("batch cannot have both a non-zero sequence and a ticket sequence")
	// ErrBatchSequenceNotSet is returned when a Batch has neither a non-zero Sequence nor a TicketSequence.
	ErrBatchSequenceNotSet = errors.New("batch must have a sequence or ticket sequence")
	// ErrFlagsFieldIsNotAnUint32 is returned when the flags field is missing or is not an uint32.
	ErrFlagsFieldIsNotAnUint32 = errors.New("flags field is not an uint32")
	// ErrRawTransactionsFieldIsNotAnArray is returned when the raw transactions field is not an array.
	ErrRawTransactionsFieldIsNotAnArray = errors.New("raw transactions field is not an array")
	// ErrRawTransactionFieldIsNotAnObject is returned when the raw transaction field is not an object.
	ErrRawTransactionFieldIsNotAnObject = errors.New("raw transaction field is not an object")
)

// ErrFailedToGetTxIDFromRawTransaction is returned when getting txID from raw transaction fails.
type ErrFailedToGetTxIDFromRawTransaction struct {
	Err error
}

// Error implements the error interface for ErrFailedToGetTxIDFromRawTransaction
func (e ErrFailedToGetTxIDFromRawTransaction) Error() string {
	return fmt.Sprintf("failed to get txID from raw transaction: %v", e.Err)
}
