// Package types defines types for subscription streams in XRPL.
// revive:disable:var-naming
package types

// Type represents a subscription stream type for server events.
type Type string

// Stream message types returned by subscription notifications.
const (
	LedgerStreamType      Type = "ledgerClosed"
	ValidationStreamType  Type = "validationReceived"
	TransactionStreamType Type = "transaction"
	PeerStatusStreamType  Type = "peerStatusChange"
	OrderBookStreamType   Type = TransactionStreamType
	BookChangesStreamType Type = "bookChanges"
	ConsensusStreamType   Type = "consensusPhase"
)
