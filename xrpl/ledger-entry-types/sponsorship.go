package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// Sponsorship ledger entry flags.
const (
	// LsfSponsorshipRequireSignForFee requires a sponsor signature on every transaction where this
	// sponsorship pays the transaction fee.
	LsfSponsorshipRequireSignForFee uint32 = 0x00010000
	// LsfSponsorshipRequireSignForReserve requires a sponsor signature on every transaction where
	// this sponsorship pays an object or account reserve.
	LsfSponsorshipRequireSignForReserve uint32 = 0x00020000
)

// Sponsorship entry type represents a pre-funded sponsoring relationship between a sponsor (Owner)
// and a sponsee.
type Sponsorship struct {
	// The unique ID for this ledger entry.
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry.
	LedgerEntryType EntryType
	// A bit-map of boolean flags.
	Flags uint32
	// The sponsor associated with this relationship.
	Owner types.Address
	// The sponsee associated with this relationship.
	Sponsee types.Address
	// The remaining amount of XRP, in drops, that the sponsor has provided for the sponsee to use
	// for fees.
	FeeAmount *types.XRPCurrencyAmount `json:",omitempty"`
	// The maximum fee, in drops, that is sponsored per transaction.
	MaxFee *types.XRPCurrencyAmount `json:",omitempty"`
	// The remaining number of owner count units that the sponsor has provided for the sponsee to
	// use for reserves.
	RemainingOwnerCount *uint32 `json:",omitempty"`
	// A hint indicating which page of the sponsor's owner directory links to this entry, in case
	// the directory consists of multiple pages.
	OwnerNode string
	// A hint indicating which page of the sponsee's owner directory links to this entry, in case
	// the directory consists of multiple pages.
	SponseeNode string
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this
	// entry.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the type of the ledger entry.
func (*Sponsorship) EntryType() EntryType {
	return SponsorshipEntry
}
