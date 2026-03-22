package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LsfMPTCanLock if set, indicates that the issuer can lock an individual balance or all balances of this MPT.
	// If not set, the MPT cannot be locked in any way.
	LsfMPTCanLock uint32 = 0x00000002
	// LsfMPTRequireAuth if set, indicates that individual holders must be authorized. This enables issuers to limit who can hold their assets.
	LsfMPTRequireAuth uint32 = 0x00000004
	// LsfMPTCanEscrow if set, indicates that individual holders can place their balances into an escrow.
	LsfMPTCanEscrow uint32 = 0x00000008
	// LsfMPTCanTrade if set, indicates that individual holders can trade their balances using the XRP Ledger DEX or AMM.
	LsfMPTCanTrade uint32 = 0x00000010
	// LsfMPTCanTransfer if set, indicates that tokens held by non-issuers can be transferred to other accounts.
	// If not set, indicates that tokens held by non-issuers cannot be transferred except back to the issuer; this enables use cases such as store credit.
	LsfMPTCanTransfer uint32 = 0x00000020
	// LsfMPTCanClawback if set, indicates that the issuer may use the Clawback transaction to claw back value from individual holders.
	LsfMPTCanClawback uint32 = 0x00000040
)

// Ledger-state mutable flags for MPTokenIssuance (Lsmf prefix).
const (
	// LsmfMPTCanMutateCanLock indicates the CanLock property can be mutated.
	LsmfMPTCanMutateCanLock uint32 = 0x00000002
	// LsmfMPTCanMutateRequireAuth indicates the RequireAuth property can be mutated.
	LsmfMPTCanMutateRequireAuth uint32 = 0x00000004
	// LsmfMPTCanMutateCanEscrow indicates the CanEscrow property can be mutated.
	LsmfMPTCanMutateCanEscrow uint32 = 0x00000008
	// LsmfMPTCanMutateCanTrade indicates the CanTrade property can be mutated.
	LsmfMPTCanMutateCanTrade uint32 = 0x00000010
	// LsmfMPTCanMutateCanTransfer indicates the CanTransfer property can be mutated.
	LsmfMPTCanMutateCanTransfer uint32 = 0x00000020
	// LsmfMPTCanMutateCanClawback indicates the CanClawback property can be mutated.
	LsmfMPTCanMutateCanClawback uint32 = 0x00000040
	// LsmfMPTCanMutateMetadata indicates the MPTokenMetadata can be mutated.
	LsmfMPTCanMutateMetadata uint32 = 0x00010000
	// LsmfMPTCanMutateTransferFee indicates the TransferFee can be mutated.
	LsmfMPTCanMutateTransferFee uint32 = 0x00020000
)

// An MPTokenIssuance entry represents a single MPT issuance and holds data associated with the issuance itself.
// You can create an MPTokenIssuance using an MPTokenIssuanceCreate transaction, and can delete it with an MPTokenIssuanceDestroy transaction.
type MPTokenIssuance struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The type of ledger entry.
	LedgerEntryType EntryType
	// Set of bit-flags for this ledger entry.
	Flags uint32
	// The address of the account that controls both the issuance amounts and characteristics of a particular fungible token.
	Issuer types.Address
	// Where to put the decimal place when displaying amounts of this MPT.
	// More formally, the asset scale is a non-negative integer (0, 1, 2, …) such that one standard unit equals 10^(-scale) of a corresponding fractional unit.
	// For example, if a US Dollar Stablecoin has an asset scale of 2, then 1 unit of that MPT would equal 0.01 US Dollars.
	// This indicates to how many decimal places the MPT can be subdivided. The default is 0, meaning that the MPT cannot be divided into smaller than 1 unit.
	AssetScale uint8
	// The maximum number of MPTs that can exist at one time. If omitted, the maximum is currently limited to 263-1.
	MaximumAmount uint64
	// The total amount of MPTs of this issuance currently in circulation. This value increases when the issuer sends MPTs to a non-issuer, and decreases whenever the issuer receives MPTs.
	OutstandingAmount uint64
	// This value specifies the fee, in tenths of a basis point, charged by the issuer for secondary sales of the token, if such sales are allowed at all.
	// Valid values for this field are between 0 and 50,000 inclusive. A value of 1 is equivalent to 1/10 of a basis point or 0.001%, allowing transfer rates between 0% and 50%.
	// A TransferFee of 50,000 corresponds to 50%. The default value for this field is 0. Any decimals in the transfer fee are rounded down.
	// The fee can be rounded down to zero if the payment is small. Issuers should make sure that their MPT's AssetScale is large enough.
	TransferFee uint16
	// Arbitrary metadata about this issuance, in hex format. The limit for this field is 1024 bytes.
	MPTokenMetadata string
	// A hint indicating which page of the owner directory links to this entry, in case the directory consists of multiple pages.
	OwnerNode uint64
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this object.
	PreviousTxnLgrSeq uint32
	// The Sequence (or Ticket) number of the transaction that created this issuance.
	// This helps to uniquely identify the issuance and distinguish it from any other later MPT issuances created by this account.
	Sequence uint32
	// The amount of tokens currently locked up (for example, in escrow or payment channels). (Requires the TokenEscrow amendment .)
	LockedAmount uint64 `json:",omitempty"`
	// DomainID is the ledger entry ID of a permissioned domain that grants access to the MPT.
	DomainID string `json:",omitempty"`
	// MutableFlags indicates which properties of this MPT can be mutated after creation.
	MutableFlags uint32 `json:",omitempty"`
}

// EntryType returns the type of the ledger entry.
func (*MPTokenIssuance) EntryType() EntryType {
	return MPTokenIssuanceEntry
}

// SetLsfMPTLocked sets the LsfMPTLocked flag.
func (c *MPTokenIssuance) SetLsfMPTLocked() {
	c.Flags |= LsfMPTLocked
}

// SetLsfMPTCanLock sets the LsfMPTCanLock flag.
func (c *MPTokenIssuance) SetLsfMPTCanLock() {
	c.Flags |= LsfMPTCanLock
}

// SetLsfMPTRequireAuth sets the LsfMPTRequireAuth flag.
func (c *MPTokenIssuance) SetLsfMPTRequireAuth() {
	c.Flags |= LsfMPTRequireAuth
}

// SetLsfMPTCanEscrow sets the LsfMPTCanEscrow flag.
func (c *MPTokenIssuance) SetLsfMPTCanEscrow() {
	c.Flags |= LsfMPTCanEscrow
}

// SetLsfMPTCanTrade sets the LsfMPTCanTrade flag.
func (c *MPTokenIssuance) SetLsfMPTCanTrade() {
	c.Flags |= LsfMPTCanTrade
}

// SetLsfMPTCanTransfer sets the LsfMPTCanTransfer flag.
func (c *MPTokenIssuance) SetLsfMPTCanTransfer() {
	c.Flags |= LsfMPTCanTransfer
}

// SetLsfMPTCanClawback sets the LsfMPTCanClawback flag.
func (c *MPTokenIssuance) SetLsfMPTCanClawback() {
	c.Flags |= LsfMPTCanClawback
}

// SetLsmfMPTCanMutateCanLock sets the LsmfMPTCanMutateCanLock flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateCanLock() {
	c.MutableFlags |= LsmfMPTCanMutateCanLock
}

// SetLsmfMPTCanMutateRequireAuth sets the LsmfMPTCanMutateRequireAuth flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateRequireAuth() {
	c.MutableFlags |= LsmfMPTCanMutateRequireAuth
}

// SetLsmfMPTCanMutateCanEscrow sets the LsmfMPTCanMutateCanEscrow flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateCanEscrow() {
	c.MutableFlags |= LsmfMPTCanMutateCanEscrow
}

// SetLsmfMPTCanMutateCanTrade sets the LsmfMPTCanMutateCanTrade flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateCanTrade() {
	c.MutableFlags |= LsmfMPTCanMutateCanTrade
}

// SetLsmfMPTCanMutateCanTransfer sets the LsmfMPTCanMutateCanTransfer flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateCanTransfer() {
	c.MutableFlags |= LsmfMPTCanMutateCanTransfer
}

// SetLsmfMPTCanMutateCanClawback sets the LsmfMPTCanMutateCanClawback flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateCanClawback() {
	c.MutableFlags |= LsmfMPTCanMutateCanClawback
}

// SetLsmfMPTCanMutateMetadata sets the LsmfMPTCanMutateMetadata flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateMetadata() {
	c.MutableFlags |= LsmfMPTCanMutateMetadata
}

// SetLsmfMPTCanMutateTransferFee sets the LsmfMPTCanMutateTransferFee flag on MutableFlags.
func (c *MPTokenIssuance) SetLsmfMPTCanMutateTransferFee() {
	c.MutableFlags |= LsmfMPTCanMutateTransferFee
}
