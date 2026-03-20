// Package types provides types for the subscription streams, including book changes events.
//
//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/queries/common"

// A BookUpdate represents the changes to a single order book in a single ledger version.
type BookUpdate struct {
	// An identifier for the first of the two currencies in the order book. For XRP, this is the
	// string XRP_drops. For tokens, this is formatted as the address of the issuer in base58,
	// followed by a forward-slash (/), followed by the Currency Code for the token, which can
	// be a 3-character standard code or a 20-character hexadecimal code.
	CurrencyA string `json:"currency_a"`
	// An identifier for the second of two currencies in the order book. This is in the same
	// format as currency_a, except currency_b can never be XRP.
	CurrencyB string `json:"currency_b"`
	// The total amount, or volume, of the first currency (that is, currency_a) that moved as
	// a result of trades through this order book in this ledger.
	VolumeA any `json:"volume_a"`
	// The volume of the second currency (that is, currency_b) that moved as a result of trades
	// through this order book in this ledger.
	VolumeB any `json:"volume_b"`
	// The highest exchange rate among all offers matched in this ledger, as a ratio of the first
	// currency to the second currency. (In other words, currency_a : currency_b.)
	High any `json:"high"`
	// The lowest exchange rate among all offers matched in this ledger, as a ratio of the first
	// currency to the second currency.
	Low any `json:"low"`
	// The exchange rate at the top of this order book before processing the transactions in this
	// ledger, as a ratio of the first currency to the second currency.
	Open any `json:"open"`
	// The exchange rate at the top of this order book after processing the transactions in this
	// ledger, as a ratio of the first currency to the second currency.
	Close any `json:"close"`
}

// BookChangesStream sends bookChanges messages whenever a new ledger is validated.
// It contains a summary of all changes to order books in the decentralized exchange for that ledger.
type BookChangesStream struct {
	// The value bookChanges indicates this is from the Book Changes stream.
	Type Type `json:"type"`
	// The ledger index of the ledger with these changes.
	LedgerIndex common.LedgerIndex `json:"ledger_index"`
	// The identifying hash of the ledger with these changes.
	LedgerHash common.LedgerHash `json:"ledger_hash"`
	// The official close time of the ledger with these changes, in seconds since the Ripple Epoch.
	LedgerTime uint64 `json:"ledger_time"`
	// List of BookUpdateObject, containing one entry for each order book that was updated in this
	// ledger version. The array is empty if no order books were updated.
	Changes []BookUpdate `json:"changes"`
}
