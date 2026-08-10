package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// MPTokenIssuanceSet Flags
const (
	// TfMPTLock if set, indicates that all MPT balances for this asset should be locked.
	TfMPTLock uint32 = 0x00000001
	// TfMPTUnlock if set, indicates that all MPT balances for this asset should be unlocked.
	TfMPTUnlock uint32 = 0x00000002
)

// MutableFlags constants for MPTokenIssuanceSet.
// These flags enable issuance capabilities that were declared mutable at creation.
const (
	// TmfMPTSetCanLock enables the CanLock flag.
	TmfMPTSetCanLock uint32 = 0x00000001
	// TmfMPTSetRequireAuth enables the RequireAuth flag.
	TmfMPTSetRequireAuth uint32 = 0x00000002
	// TmfMPTSetCanEscrow enables the CanEscrow flag.
	TmfMPTSetCanEscrow uint32 = 0x00000004
	// TmfMPTSetCanTrade enables the CanTrade flag.
	TmfMPTSetCanTrade uint32 = 0x00000008
	// TmfMPTSetCanTransfer enables the CanTransfer flag.
	TmfMPTSetCanTransfer uint32 = 0x00000010
	// TmfMPTSetCanClawback enables the CanClawback flag.
	TmfMPTSetCanClawback uint32 = 0x00000020

	validMPTokenIssuanceSetMutableFlags = TmfMPTSetCanLock |
		TmfMPTSetRequireAuth |
		TmfMPTSetCanEscrow |
		TmfMPTSetCanTrade |
		TmfMPTSetCanTransfer |
		TmfMPTSetCanClawback
)

// MPTokenIssuanceSet transaction is used to globally lock/unlock a MPTokenIssuance,
// lock/unlock an individual's MPToken, mutate dynamic MPT properties
// (MutableFlags, MPTokenMetadata, TransferFee), or update the DomainID.
//
// ```json
//
//	{
//	      "TransactionType": "MPTokenIssuanceSet",
//	      "Fee": "10",
//	      "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
//	      "Flags": 1
//	}
//
// ```
type MPTokenIssuanceSet struct {
	BaseTx
	// The MPTokenIssuance identifier.
	MPTokenIssuanceID string
	// (Optional) XRPL Address of an individual token holder balance to lock/unlock. If omitted, this transaction applies to all any accounts holding MPTs.
	Holder *types.Address
	// (Optional) The ledger entry ID of a permissioned domain to associate with this issuance.
	// An empty string removes the domain.
	DomainID *string `json:",omitempty"`
	// (Optional) New metadata to replace the existing value.
	MPTokenMetadata *string `json:",omitempty"`
	// (Optional) New transfer fee value between 0 and 50,000.
	TransferFee *uint16 `json:",omitempty"`
	// (Optional) Enable issuance flags that were declared mutable at creation.
	// Once enabled, these flags cannot be disabled by MPTokenIssuanceSet.
	MutableFlags *uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (MPTokenIssuanceSet).
func (*MPTokenIssuanceSet) TxType() TxType {
	return MPTokenIssuanceSetTx
}

// Flatten returns the flattened map of the MPTokenIssuanceSet transaction.
func (m *MPTokenIssuanceSet) Flatten() FlatTransaction {
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenIssuanceSet"

	flattened["MPTokenIssuanceID"] = m.MPTokenIssuanceID

	if m.Holder != nil {
		flattened["Holder"] = m.Holder.String()
	}

	if m.DomainID != nil {
		flattened["DomainID"] = *m.DomainID
	}

	if m.MPTokenMetadata != nil {
		flattened["MPTokenMetadata"] = *m.MPTokenMetadata
	}

	if m.TransferFee != nil {
		flattened["TransferFee"] = *m.TransferFee
	}

	if m.MutableFlags != nil {
		flattened["MutableFlags"] = *m.MutableFlags
	}

	return flattened
}

// SetMPTLockFlag sets the TfMPTLock flag on the transaction.
// Indicates that all MPT balances for this asset should be locked.
func (m *MPTokenIssuanceSet) SetMPTLockFlag() {
	m.Flags |= TfMPTLock
}

// SetMPTUnlockFlag sets the TfMPTUnlock flag on the transaction.
// Indicates that all MPT balances for this asset should be unlocked.
func (m *MPTokenIssuanceSet) SetMPTUnlockFlag() {
	m.Flags |= TfMPTUnlock
}

// setMutableFlag is a helper that initialises MutableFlags if nil and applies the given flag.
func (m *MPTokenIssuanceSet) setMutableFlag(f uint32) {
	if m.MutableFlags == nil {
		mf := uint32(0)
		m.MutableFlags = &mf
	}
	*m.MutableFlags |= f
}

// SetMPTSetCanLockMutableFlag enables the CanLock flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanLockMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanLock)
}

// SetMPTSetRequireAuthMutableFlag enables the RequireAuth flag.
func (m *MPTokenIssuanceSet) SetMPTSetRequireAuthMutableFlag() {
	m.setMutableFlag(TmfMPTSetRequireAuth)
}

// SetMPTSetCanEscrowMutableFlag enables the CanEscrow flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanEscrowMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanEscrow)
}

// SetMPTSetCanTradeMutableFlag enables the CanTrade flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTradeMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTrade)
}

// SetMPTSetCanTransferMutableFlag enables the CanTransfer flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTransferMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTransfer)
}

// SetMPTSetCanClawbackMutableFlag enables the CanClawback flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanClawbackMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanClawback)
}

// Validate validates the MPTokenIssuanceSet transaction ensuring all fields are correct.
func (m *MPTokenIssuanceSet) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// MPTokenIssuanceID is required and must be valid hex.
	if m.MPTokenIssuanceID == "" || !typecheck.IsHex(m.MPTokenIssuanceID) {
		return false, ErrInvalidMPTokenIssuanceIDSet
	}

	// If a Holder is specified, validate it as a proper XRPL address.
	if m.Holder != nil && !addresscodec.IsValidAddress(m.Holder.String()) {
		return false, ErrInvalidAccount
	}

	// Holder must be different from Account.
	if m.Holder != nil && m.Account.String() == m.Holder.String() {
		return false, ErrHolderAccountConflict
	}

	// Check flag conflict: TfMPTLock and TfMPTUnlock cannot both be enabled
	isLock := flag.Contains(m.Flags, TfMPTLock)
	isUnlock := flag.Contains(m.Flags, TfMPTUnlock)

	if isLock && isUnlock {
		return false, ErrMPTokenIssuanceSetFlags
	}

	hasDynamicMPTFields := m.MutableFlags != nil || m.MPTokenMetadata != nil || m.TransferFee != nil

	// At least one operation must be specified (lock/unlock, holder lock/unlock, DynamicMPT mutation, or DomainID).
	if m.Flags == 0 && m.Holder == nil && !hasDynamicMPTFields && m.DomainID == nil {
		return false, ErrMPTIssuanceSetEmpty
	}

	// Holder is mutually exclusive with DynamicMPT fields and DomainID.
	if m.Holder != nil && (hasDynamicMPTFields || m.DomainID != nil) {
		return false, ErrMPTIssuanceSetHolderMutuallyExclusive
	}

	// Non-zero Flags are mutually exclusive with DynamicMPT fields.
	if m.Flags != 0 && hasDynamicMPTFields {
		return false, ErrMPTIssuanceSetFlagsMutuallyExclusive
	}

	if m.MutableFlags != nil {
		if *m.MutableFlags == 0 {
			return false, ErrMPTIssuanceSetMutableFlagsZero
		}
		if *m.MutableFlags&^uint32(validMPTokenIssuanceSetMutableFlags) != 0 {
			return false, ErrMPTIssuanceSetInvalidMutableFlags
		}
	}

	// TransferFee must not exceed MaxTransferFee.
	if m.TransferFee != nil && *m.TransferFee > MaxTransferFee {
		return false, ErrInvalidTransferFee
	}

	// MPTokenMetadata: empty string is valid (removes the field per XLS-94),
	// otherwise must be valid hex and at most 1024 bytes (2048 hex chars).
	if m.MPTokenMetadata != nil && *m.MPTokenMetadata != "" && !ValidateHexMetadata(*m.MPTokenMetadata, 2*types.MaxMPTokenMetadataByteLength) {
		return false, ErrInvalidMPTokenMetadata
	}

	// DomainID: empty string is valid (removes domain), otherwise must be valid 64-char hex.
	if m.DomainID != nil && *m.DomainID != "" && !IsDomainID(*m.DomainID) {
		return false, ErrMPTIssuanceSetDomainIDInvalid
	}

	return true, nil
}
