// Package types contains data structures for wallet operations and batch signing.
//
//revive:disable:var-naming
package types

import (
	"slices"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/hash"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// BatchSignable contains the fields used by a BatchV1_1 signature.
type BatchSignable struct {
	Account       string
	Sequence      uint32
	Flags         uint32
	TxIDs         []string
	BatchAccount  string
	SignerAccount string
}

// FromFlatBatchTransaction creates a BatchSignable from a Batch transaction.
// It returns an error if the transaction is invalid.
func FromFlatBatchTransaction(transaction *transaction.FlatTransaction) (*BatchSignable, error) {
	account, ok := typecheck.ToString((*transaction)["Account"])
	if !ok || account == "" {
		return nil, ErrAccountFieldIsNotAString
	}
	sequence, err := flatBatchSequenceValue(*transaction)
	if err != nil {
		return nil, err
	}
	flags, ok := typecheck.ToUint32((*transaction)["Flags"])
	if !ok {
		return nil, ErrFlagsFieldIsNotAnUint32
	}

	rawTxs, ok := (*transaction)["RawTransactions"].([]map[string]any)
	if !ok {
		return nil, ErrRawTransactionsFieldIsNotAnArray
	}

	batchSignable := &BatchSignable{
		Account:  account,
		Sequence: sequence,
		Flags:    flags,
		TxIDs:    make([]string, len(rawTxs)),
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
	if transaction.Account == "" {
		return nil, ErrAccountFieldIsNotAString
	}
	sequence, err := batchSequenceValue(transaction.Sequence, transaction.TicketSequence)
	if err != nil {
		return nil, err
	}
	rawTxs := transaction.RawTransactions

	batchSignable := &BatchSignable{
		Account:  transaction.Account.String(),
		Sequence: sequence,
		Flags:    transaction.Flags,
		TxIDs:    make([]string, len(rawTxs)),
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

func flatBatchSequenceValue(tx transaction.FlatTransaction) (uint32, error) {
	// An absent field is equivalent to its zero value, as in the struct-based
	// FromBatchTransaction path.
	var sequence uint32
	if value, exists := tx["Sequence"]; exists {
		var ok bool
		if sequence, ok = typecheck.ToUint32(value); !ok {
			return 0, ErrSequenceFieldIsNotAnUint32
		}
	}

	var ticketSequence uint32
	if value, exists := tx["TicketSequence"]; exists {
		var ok bool
		if ticketSequence, ok = typecheck.ToUint32(value); !ok {
			return 0, ErrTicketSequenceFieldIsNotAnUint32
		}
	}

	return batchSequenceValue(sequence, ticketSequence)
}

func batchSequenceValue(sequence, ticketSequence uint32) (uint32, error) {
	if sequence != 0 {
		if ticketSequence != 0 {
			return 0, ErrBatchSequenceAndTicket
		}
		return sequence, nil
	}
	if ticketSequence == 0 {
		return 0, ErrBatchSequenceNotSet
	}
	return ticketSequence, nil
}

// Equals reports whether two fragments bind the same common BatchV1_1 data.
func (b *BatchSignable) Equals(other *BatchSignable) bool {
	return b.Account == other.Account &&
		b.Sequence == other.Sequence &&
		b.Flags == other.Flags &&
		slices.Equal(b.TxIDs, other.TxIDs)
}

// Flatten returns the BatchSignable as a map[string]any for encoding.
func (b *BatchSignable) Flatten() map[string]any {
	flattened := map[string]any{
		"account":  b.Account,
		"sequence": b.Sequence,
		"flags":    b.Flags,
		"txIDs":    b.TxIDs,
	}
	if b.BatchAccount != "" {
		flattened["batchAccount"] = b.BatchAccount
	}
	if b.SignerAccount != "" {
		flattened["signerAccount"] = b.SignerAccount
	}
	return flattened
}
