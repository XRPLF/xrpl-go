package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

const (
	// LsfMPTLocked if enabled, indicates that the MPT owned by this account is currently locked and cannot be used in any XRP transactions other than sending value back to the issuer.
	LsfMPTLocked uint32 = 0x00000001

	// LsfMPTAuthorized if set, indicates that the issuer has authorized the holder for the MPT. (Only applicable for allow-listing).
	// This flag can be set using a MPTokenAuthorize transaction; it can also be "un-set" using a MPTokenAuthorize transaction specifying the TfMPTUnauthorize flag.
	LsfMPTAuthorized uint32 = 0x00000002
)

// An MPToken entry tracks MPTs held by an account that is not the token issuer. You can create or delete an empty MPToken entry by sending an MPTokenAuthorize transaction.
// You can send and receive MPTs using several other transaction types including Payment and OfferCreate transactions.
type MPToken struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry.
	LedgerEntryType EntryType
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The owner (holder) of these MPTs.
	Account types.Address
	// 	The MPTokenIssuance identifier.
	MPTokenIssuanceID types.Hash192
	// The amount of tokens currently held by the owner. The minimum is 0 and the maximum is 263-1.
	MPTAmount uint64
	// The amount of tokens currently locked up (for example, in escrow or payment channels). (Requires the TokenEscrow amendment .)
	LockedAmount uint64 `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The sequence of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// A hint indicating which page of the owner directory links to this entry, in case the directory consists of multiple pages.
	OwnerNode uint64
	// The holder's encryption key for confidential transfers.
	// Required for participating in confidential transfers.
	HolderEncryptionKey string `json:",omitempty"`
	// Encrypted balance value for issuer tracking purposes.
	// This allows the issuer to track confidential balances.
	IssuerEncryptedBalance string `json:",omitempty"`
	// Encrypted balance value for auditor tracking purposes (if configured).
	// This allows an auditor to track confidential balances.
	AuditorEncryptedBalance string `json:",omitempty"`
	// The confidential balance inbox for pending incoming transfers.
	// Contains encrypted amounts that have not yet been merged.
	ConfidentialBalanceInbox string `json:",omitempty"`
	// The confidential balance available for spending.
	// Contains the merged encrypted balance.
	ConfidentialBalanceSpending string `json:",omitempty"`
	// Version counter for the confidential balance.
	// Incremented with each update to prevent replay attacks.
	ConfidentialBalanceVersion uint32 `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*MPToken) EntryType() EntryType {
	return MPTokenEntry
}

// SetLsfMPTLocked sets the LsfMPTLocked flag.
func (c *MPToken) SetLsfMPTLocked() {
	c.Flags |= LsfMPTLocked
}

// SetLsfMPTAuthorized sets the LsfMPTAuthorized flag.
func (c *MPToken) SetLsfMPTAuthorized() {
	c.Flags |= LsfMPTAuthorized
}
