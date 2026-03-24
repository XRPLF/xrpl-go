// Package types contains data structures for v1 subscription stream types.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// TransactionStream represents a transaction notification from a v1 subscription stream.
type TransactionStream struct {
	// `transaction` indicates this is the notification of a transaction, which could
	// come from several possible streams.
	Type Type `json:"type"`
	// String Transaction result code
	EngineResult string `json:"engine_result"`
	// Numeric transaction response code, if applicable.
	EngineResultCode int `json:"engine_result_code"`
	// Human-readable explanation for the transaction response.
	EngineResultMessage string `json:"engine_result_message"`
	// (Unvalidated transactions only) The ledger index of the current in-progress ledger
	// version for which this transaction is currently proposed.
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`
	// (Validated transactions only) The identifying hash of the ledger version that includes
	// this transaction
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`
	// (Validated transactions only) The ledger index of the ledger version that includes
	// this transaction.
	LedgerIndex common.LedgerIndex `json:"ledger_index,omitempty"`
	// (Validated transactions only) The transaction metadata, which shows the exact outcome
	// of the transaction in detail.
	Meta transactions.TxObjMeta `json:"meta,omitzero"`
	// The definition of the transaction in JSON format.
	Transaction transactions.FlatTransaction `json:"transaction"`
	// If true, this transaction is included in a validated ledger and its outcome is final.
	// Responses from the transaction stream should always be validated.
	Validated bool `json:"validated"`
}
