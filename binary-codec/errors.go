package binarycodec

import "errors"

var (
	// ErrSigningClaimFieldNotFound is returned when the 'Channel' & 'Amount' fields are both required, but were not found.
	ErrSigningClaimFieldNotFound = errors.New("'Channel' & 'Amount' fields are both required, but were not found")

	// ErrBatchAccountFieldNotFound is returned when the outer Batch 'account' field is missing.
	ErrBatchAccountFieldNotFound = errors.New("no field `account`")
	// ErrBatchSequenceFieldNotFound is returned when the outer Batch 'sequence' value is missing.
	ErrBatchSequenceFieldNotFound = errors.New("no field `sequence`")
	// ErrBatchFlagsFieldNotFound is returned when the 'flags' field is missing.
	ErrBatchFlagsFieldNotFound = errors.New("no field `flags`")
	// ErrBatchTxIDsFieldNotFound is returned when the 'txIDs' field is missing.
	ErrBatchTxIDsFieldNotFound = errors.New("no field `txIDs`")
	// ErrBatchSignerAccountWithoutBatchAccount is returned when a nested signer account is supplied without its Batch signer account.
	ErrBatchSignerAccountWithoutBatchAccount = errors.New("field `signerAccount` requires `batchAccount`")
	// ErrBatchTxIDsNotArray is returned when the 'txIDs' field is not an array.
	ErrBatchTxIDsNotArray = errors.New("txIDs field must be an array")
	// ErrBatchFlagsNotUInt32 is returned when the 'flags' field is not a uint32.
	ErrBatchFlagsNotUInt32 = errors.New("flags field must be a uint32")
	// ErrBatchTxIDsLengthTooLong is returned when the 'txIDs' field is too long.
	ErrBatchTxIDsLengthTooLong = errors.New("txIDs length exceeds maximum uint32 value")

	// ErrInvalidUNLModifyAccount is returned when a supplied UNLModify Account is not canonical.
	ErrInvalidUNLModifyAccount = errors.New("invalid UNLModify Account: must be an empty string or the canonical XRPL zero account")

	// ErrInvalidQuality is returned when the quality is invalid.
	ErrInvalidQuality = errors.New("invalid quality")
)
