package wallet

import "errors"

var (
	// address

	// ErrAddressTagNotZero is returned when the address tag is not zero.
	ErrAddressTagNotZero = errors.New("address tag is not zero")

	// batch

	// ErrBatchAccountNotFound is returned when the batch account is not found in the transaction.
	ErrBatchAccountNotFound = errors.New("batch account not found in transaction")
	// ErrTransactionMustBeBatch is returned when the transaction is not a batch transaction.
	ErrTransactionMustBeBatch = errors.New("transaction must be a batch transaction")
	// ErrNoTransactionsProvided is returned when no transactions are provided.
	ErrNoTransactionsProvided = errors.New("no transactions provided")
	// ErrTxMustIncludeBatchSigner is returned when the transaction does not include a batch signer.
	ErrTxMustIncludeBatchSigner = errors.New("transaction must include a batch signer")
	// ErrTransactionAlreadySigned is returned when the transaction has already been signed.
	ErrTransactionAlreadySigned = errors.New("transaction has already been signed")
	// ErrBatchSignableNotEqual is returned when the batch signable is not equal.
	ErrBatchSignableNotEqual = errors.New("batch signable is not equal")

	// counterparty

	// ErrTxMustBeLoanSet is returned when the transaction is not a LoanSet.
	ErrTxMustBeLoanSet = errors.New("transaction must be a LoanSet")
	// ErrCounterpartyAlreadySigned is returned when CounterpartySignature is already set.
	ErrCounterpartyAlreadySigned = errors.New("counterparty has already signed this transaction")
	// ErrBrokerMustSignFirst is returned when the first party has not yet signed.
	ErrBrokerMustSignFirst = errors.New("transaction must be signed by the first party before the counterparty can sign")
	// ErrNoTransactionsToSign is returned when CombineLoanSetCounterpartySigners receives an empty slice.
	ErrNoTransactionsToSign = errors.New("no transactions provided to combine")
	// ErrTxMustIncludeCounterpartySigners is returned when CounterpartySignature.Signers is missing or empty.
	ErrTxMustIncludeCounterpartySigners = errors.New("transaction must include counterparty signers in CounterpartySignature.Signers")
	// ErrLoanSetTxNotEqual is returned when blobs do not represent equivalent transactions.
	ErrLoanSetTxNotEqual = errors.New("transactions are not equivalent (excluding CounterpartySignature.Signers)")
)
