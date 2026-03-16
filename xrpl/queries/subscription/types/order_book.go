// Package types contains data structures for subscription stream types.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// OrderBook specifies the currencies and optional filters for subscribing to an order book stream.
type OrderBook struct {
	TakerGets types.IssuedCurrencyAmount `json:"taker_gets"`
	TakerPays types.IssuedCurrencyAmount `json:"taker_pays"`
	Taker     types.Address              `json:"taker"`
	Snapshot  bool                       `json:"snapshot,omitempty"`
	Both      bool                       `json:"both,omitempty"`
	Domain    *string                    `json:"domain,omitempty"`
}

// OrderBookStream documented as identical to TransactionStream
type OrderBookStream struct {
	// The ledger close time represented in ISO 8601 time format.
	CloseTimeISO string `json:"close_time_iso"`
	// `transaction` indicates this is the notification of a transaction, which could
	// come from several possible streams.
	Type Type `json:"type"`
	// String Transaction result code
	EngineResult string `json:"engine_result"`
	// Numeric transaction response code, if applicable.
	EngineResultCode int `json:"engine_result_code"`
	// Human-readable explanation for the transaction response.
	EngineResultMessage string `json:"engine_result_message"`
	// The unique has identifier of the transaction.
	Hash common.LedgerHash `json:"hash"`
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
	Transaction transactions.FlatTransaction `json:"tx_json"`
	// If true, this transaction is included in a validated ledger and its outcome is final.
	// Responses from the transaction stream should always be validated.
	Validated bool `json:"validated"`
}
