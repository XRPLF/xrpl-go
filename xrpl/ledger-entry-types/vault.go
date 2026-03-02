package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// LsfVaultPrivate if set, indicates that the vault is private.
	LsfVaultPrivate uint32 = 0x00010000

	// VaultStrategyFirstComeFirstServe is the default withdrawal policy.
	// Withdrawals are processed on a first-come, first-served basis.
	VaultStrategyFirstComeFirstServe uint8 = 0x01
)

// Vault represents a Single Asset Vault instance.
//
// ```json
//
//	{
//	  "LedgerEntryType": "Vault",
//	  "Flags": 0,
//	  "Sequence": 200370,
//	  "OwnerNode": "0",
//	  "Owner": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "Account": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85",
//	  "Asset": {
//	      "currency": "USD",
//	      "issuer": "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"
//	   },
//	  "ShareMPTID": "0000000169F415C9F1AB6796AB9224CE635818AFD74F8175",
//	  "WithdrawalPolicy": 1,
//	  "PreviousTxnID": "C44F2EB84196B9AD820313DBEBA6316A15C9A2D35787579ED172B87A30131DA7",
//	  "PreviousTxnLgrSeq": 28991004
//	}
//
// ```
type Vault struct {
	// The unique ID for this ledger entry. In JSON, this field is represented with different names depending on the
	// context and API method. (Note, even though this is specified as "optional" in the code, every ledger entry
	// should have one unless it's legacy data from very early in the XRP Ledger's history.)
	Index types.Hash256 `json:"index,omitempty"`
	// The value "Vault", mapped to the string Vault, indicates that this object is a Vault object.
	LedgerEntryType EntryType
	// A bit-map of boolean flags.
	Flags uint32
	// The transaction sequence number that created the vault.
	Sequence uint32
	// Identifies the page where this item is referenced in the owner's directory.
	OwnerNode string
	// The account address of the Vault Owner.
	Owner types.Address
	// The address of the Vault's pseudo-account.
	Account types.Address
	// The asset of the vault. The vault supports XRP, IOU and MPT.
	Asset Asset
	// The total value of the vault.
	AssetsTotal *types.XRPLNumber `json:",omitempty"`
	// The asset amount that is available in the vault.
	AssetsAvailable *types.XRPLNumber `json:",omitempty"`
	// The potential loss amount that is not yet realized expressed as the vault's asset.
	LossUnrealized *types.XRPLNumber `json:",omitempty"`
	// The identifier of the share MPTokenIssuance object.
	ShareMPTID string
	// Indicates the withdrawal strategy used by the Vault.
	WithdrawalPolicy uint8
	// The maximum asset amount that can be held in the vault. Zero value 0 indicates there is no cap.
	AssetsMaximum *types.XRPLNumber `json:",omitempty"`
	// Arbitrary metadata about the Vault. Limited to 256 bytes.
	Data string `json:",omitempty"`
	// The scaling factor for vault shares. Only applicable for IOU assets.
	// Valid values are between 0 and 18 inclusive. For XRP and MPT, this is always 0.
	Scale *uint8 `json:",omitempty"`
	// The identifying hash of the transaction that most recently modified this entry.
	PreviousTxnID types.Hash256
	// The index of the ledger that contains the transaction that most recently modified this entry.
	PreviousTxnLgrSeq uint32
}

// EntryType returns the ledger entry type for Vault.
func (*Vault) EntryType() EntryType {
	return VaultEntry
}

// SetLsfVaultPrivate sets the LsfVaultPrivate flag, indicating that the Vault is private.
func (v *Vault) SetLsfVaultPrivate() {
	v.Flags |= LsfVaultPrivate
}
