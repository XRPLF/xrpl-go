// Package types contains data structures for wallet operations and batch signing.
//
//revive:disable:var-naming
package types

import (
	"slices"

	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// BatchSignable contains the fields needed to perform a Batch transactions signature.
// It contains the Flags of all transactions in the batch and the IDs of the transactions.
type BatchSignable struct {
	Flags uint32
	TxIDs []string
}

// FromFlatBatchTransaction creates a BatchSignable from a Batch transaction.
// It returns an error if the transaction is invalid.
func FromFlatBatchTransaction(transaction *transaction.FlatTransaction) (*BatchSignable, error) {
	flags, ok := (*transaction)["Flags"].(uint32)
	if !ok {
		return nil, ErrFlagsFieldIsNotAnUint32
	}

	rawTxs, ok := (*transaction)["RawTransactions"].([]map[string]any)
	if !ok {
		return nil, ErrRawTransactionsFieldIsNotAnArray
	}

	batchSignable := &BatchSignable{
		Flags: flags,
		TxIDs: make([]string, len(rawTxs)),
	}

	for i, rawTx := range rawTxs {
		innerRawTx, ok := rawTx["RawTransaction"].(map[string]any)
		if !ok {
			return nil, ErrRawTransactionFieldIsNotAnObject
		}
		txID, err := hash.SignTx(innerRawTx)
		if err != nil {
			return nil, ErrFailedToGetTxIDFromRawTransaction{
				Err: err,
			}
		}
		batchSignable.TxIDs[i] = txID
	}

	return batchSignable, nil
}

// FromBatchTransaction creates a BatchSignable from a Batch transaction.
// It returns an error if the transaction is invalid.
func FromBatchTransaction(transaction *transaction.Batch) (*BatchSignable, error) {
	rawTxs := transaction.RawTransactions

	batchSignable := &BatchSignable{
		Flags: transaction.Flags,
		TxIDs: make([]string, len(rawTxs)),
	}

	for i, rawTx := range rawTxs {
		txID, err := hash.SignTx(rawTx.RawTransaction)
		if err != nil {
			return nil, ErrFailedToGetTxIDFromRawTransaction{
				Err: err,
			}
		}
		batchSignable.TxIDs[i] = txID
	}

	return batchSignable, nil
}

// Equals checks if the BatchSignable is equal to another BatchSignable.
// It returns true if the flags and txIDs are equal, false otherwise.
func (b *BatchSignable) Equals(other *BatchSignable) bool {
	return b.Flags == other.Flags && slices.Equal(b.TxIDs, other.TxIDs)
}

// Flatten returns the BatchSignable as a map[string]interface{} for encoding.
func (b *BatchSignable) Flatten() map[string]any {
	flattened := make(map[string]any)

	flattened["flags"] = b.Flags

	if len(b.TxIDs) > 0 {
		flattened["txIDs"] = b.TxIDs
	}

	return flattened
}
