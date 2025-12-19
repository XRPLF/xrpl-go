package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// TxMeta represents the metadata interface for a transaction.
type TxMeta interface {
	TxMeta()
}

// TxObjMeta holds object-level metadata for a transaction.
// TODO: Improve CurrencyAmount parsing
type TxObjMeta struct {
	AffectedNodes []AffectedNode `json:"AffectedNodes"`
	// PartialDeliveredAmount types.CurrencyAmount `json:"DeliveredAmount,omitempty"`
	PartialDeliveredAmount any    `json:"DeliveredAmount,omitempty"`
	TransactionIndex       uint64 `json:"TransactionIndex"`
	TransactionResult      string `json:"TransactionResult"`
	// DeliveredAmount        types.CurrencyAmount `json:"delivered_amount,omitempty"`
	DeliveredAmount any `json:"delivered_amount,omitempty"`

	// ParentBatchID is the hash of the parent Batch transaction when this transaction is executed as part of a batch.
	ParentBatchID *types.BatchID `json:"ParentBatchID,omitempty"`
}

// TxMeta implements the TxMeta interface for TxObjMeta.
func (TxObjMeta) TxMeta() {}

// AffectedNode represents a ledger node that was modified by a transaction.
type AffectedNode struct {
	CreatedNode  *CreatedNode  `json:"CreatedNode,omitempty"`
	ModifiedNode *ModifiedNode `json:"ModifiedNode,omitempty"`
	DeletedNode  *DeletedNode  `json:"DeletedNode,omitempty"`
}

// CreatedNode represents a ledger node that was created by a transaction.
type CreatedNode struct {
	LedgerEntryType ledger.EntryType        `json:"LedgerEntryType,omitempty"`
	LedgerIndex     string                  `json:"LedgerIndex,omitempty"`
	NewFields       ledger.FlatLedgerObject `json:"NewFields,omitempty"`
}

// ModifiedNode represents a ledger node that was modified by a transaction.
type ModifiedNode struct {
	LedgerEntryType   ledger.EntryType        `json:"LedgerEntryType,omitempty"`
	LedgerIndex       string                  `json:"LedgerIndex,omitempty"`
	FinalFields       ledger.FlatLedgerObject `json:"FinalFields,omitempty"`
	PreviousFields    ledger.FlatLedgerObject `json:"PreviousFields,omitempty"`
	PreviousTxnID     string                  `json:"PreviousTxnID,omitempty"`
	PreviousTxnLgrSeq uint64                  `json:"PreviousTxnLgrSeq,omitempty"`
}

// DeletedNode represents a ledger node that was deleted by a transaction.
type DeletedNode struct {
	LedgerEntryType ledger.EntryType        `json:"LedgerEntryType,omitempty"`
	LedgerIndex     string                  `json:"LedgerIndex,omitempty"`
	FinalFields     ledger.FlatLedgerObject `json:"FinalFields,omitempty"`
}
