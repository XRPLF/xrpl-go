package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// Check represents a check ledger entry, similar to a paper personal check, which can be cashed by its destination to debit the sender's balance. (Added by the Checks amendment.)
type Check struct {
	// The unique ID for this ledger entry.
	// In JSON, this field is represented with different names depending on the context and API method.
	// (Note, even though this is specified as "optional" in the code, every ledger entry should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value 0x0043, mapped to the string Check, indicates that this object is a Check object.
	LedgerEntryType EntryType
	Flags           uint32
	// The sender of the Check. Cashing the Check debits this address's balance.
	Account types.Address
	// The intended recipient of the Check. Only this address can cash the Check,
	// using a CheckCash transaction.
	Destination types.Address
	// A hint indicating which page of the destination's owner directory links to this object,
	// in case the directory consists of multiple pages.
	DestinationNode string `json:",omitempty"`
	// An arbitrary tag to further specify the destination for this Check,
	// such as a hosted recipient at the destination address.
	DestinationTag uint32 `json:",omitempty"`
	// Indicates the time after which this Check is considered expired. See Specifying Time for details.
	Expiration uint32 `json:",omitempty"`
	// Arbitrary 256-bit hash provided by the sender as a specific reason or identifier for this Check.
	InvoiceID types.Hash256 `json:",omitempty"`

	// A hint indicating which page of the sender's owner directory links to this object, in
	// case the directory consists of multiple pages.
	OwnerNode string
	// The identifying hash of the transaction that most recently modified this object.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// The maximum amount of currency this Check can debit the sender. If the Check is successfully cashed,
	// the destination is credited in the same currency for up to this amount.
	SendMax types.CurrencyAmount
	// The sequence number of the CheckCreate transaction that created this check.
	Sequence uint32
	// An arbitrary tag to further specify the source for this Check, such as a hosted recipient at the sender's address.
	SourceTag uint32 `json:",omitempty"`
	// The account paying this object's owner reserve. Requires the Sponsor amendment.
	Sponsor types.Address `json:",omitempty"`
}

// EntryType returns the ledger entry type for Check.
func (*Check) EntryType() EntryType {
	return CheckEntry
}
