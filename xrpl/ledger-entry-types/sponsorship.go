package ledger

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

const (
	// LsfSponsorshipRequireSignForFee requires the sponsor's signature to use the fee budget.
	LsfSponsorshipRequireSignForFee uint32 = 0x00010000
	// LsfSponsorshipRequireSignForReserve requires the sponsor's signature to use the reserve budget.
	LsfSponsorshipRequireSignForReserve uint32 = 0x00020000
)

// Sponsorship represents a pre-funded fee and reserve relationship between an
// owner (sponsor) and a sponsee. Requires the Sponsor amendment.
type Sponsorship struct {
	// The unique ID of this ledger entry.
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry.
	LedgerEntryType EntryType
	// A set of sponsorship flags.
	Flags uint32
	// The account providing the fee and reserve budgets.
	Owner types.Address
	// The account allowed to use the budgets.
	Sponsee types.Address
	// A hexadecimal page hint for the owner's directory.
	OwnerNode string
	// A hexadecimal page hint for the sponsee's directory.
	SponseeNode string
	// The hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The ledger index of the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
	// The remaining pre-funded fee budget, in XRP drops.
	FeeAmount *types.XRPCurrencyAmount `json:",omitempty"`
	// The maximum fee allowed per transaction, in XRP drops.
	MaxFee *types.XRPCurrencyAmount `json:",omitempty"`
	// The remaining number of owner reserves available for sponsorship.
	RemainingOwnerCount *uint32 `json:",omitempty"`
}

// EntryType returns the ledger entry type for Sponsorship.
func (*Sponsorship) EntryType() EntryType {
	return SponsorshipEntry
}
