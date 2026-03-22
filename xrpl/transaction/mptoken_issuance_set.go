package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
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

// MutableFlags constants for MPTokenIssuanceSet (Set/Clear pairs).
const (
	// TmfMPTSetCanLock sets the CanLock flag.
	TmfMPTSetCanLock uint32 = 0x00000001
	// TmfMPTClearCanLock clears the CanLock flag.
	TmfMPTClearCanLock uint32 = 0x00000002
	// TmfMPTSetRequireAuth sets the RequireAuth flag.
	TmfMPTSetRequireAuth uint32 = 0x00000004
	// TmfMPTClearRequireAuth clears the RequireAuth flag.
	TmfMPTClearRequireAuth uint32 = 0x00000008
	// TmfMPTSetCanEscrow sets the CanEscrow flag.
	TmfMPTSetCanEscrow uint32 = 0x00000010
	// TmfMPTClearCanEscrow clears the CanEscrow flag.
	TmfMPTClearCanEscrow uint32 = 0x00000020
	// TmfMPTSetCanTrade sets the CanTrade flag.
	TmfMPTSetCanTrade uint32 = 0x00000040
	// TmfMPTClearCanTrade clears the CanTrade flag.
	TmfMPTClearCanTrade uint32 = 0x00000080
	// TmfMPTSetCanTransfer sets the CanTransfer flag.
	TmfMPTSetCanTransfer uint32 = 0x00000100
	// TmfMPTClearCanTransfer clears the CanTransfer flag.
	TmfMPTClearCanTransfer uint32 = 0x00000200
	// TmfMPTSetCanClawback sets the CanClawback flag.
	TmfMPTSetCanClawback uint32 = 0x00000400
	// TmfMPTClearCanClawback clears the CanClawback flag.
	TmfMPTClearCanClawback uint32 = 0x00000800
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
	// (Optional) Set or clear the flags which were marked as mutable.
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
		flattened["TransferFee"] = int(*m.TransferFee)
	}

	if m.MutableFlags != nil {
		flattened["MutableFlags"] = int(*m.MutableFlags)
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

// SetMPTSetCanLockMutableFlag sets the CanLock mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanLockMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanLock)
}

// SetMPTClearCanLockMutableFlag clears the CanLock mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearCanLockMutableFlag() {
	m.setMutableFlag(TmfMPTClearCanLock)
}

// SetMPTSetRequireAuthMutableFlag sets the RequireAuth mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetRequireAuthMutableFlag() {
	m.setMutableFlag(TmfMPTSetRequireAuth)
}

// SetMPTClearRequireAuthMutableFlag clears the RequireAuth mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearRequireAuthMutableFlag() {
	m.setMutableFlag(TmfMPTClearRequireAuth)
}

// SetMPTSetCanEscrowMutableFlag sets the CanEscrow mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanEscrowMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanEscrow)
}

// SetMPTClearCanEscrowMutableFlag clears the CanEscrow mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearCanEscrowMutableFlag() {
	m.setMutableFlag(TmfMPTClearCanEscrow)
}

// SetMPTSetCanTradeMutableFlag sets the CanTrade mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTradeMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTrade)
}

// SetMPTClearCanTradeMutableFlag clears the CanTrade mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearCanTradeMutableFlag() {
	m.setMutableFlag(TmfMPTClearCanTrade)
}

// SetMPTSetCanTransferMutableFlag sets the CanTransfer mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTransferMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTransfer)
}

// SetMPTClearCanTransferMutableFlag clears the CanTransfer mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearCanTransferMutableFlag() {
	m.setMutableFlag(TmfMPTClearCanTransfer)
}

// SetMPTSetCanClawbackMutableFlag sets the CanClawback mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanClawbackMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanClawback)
}

// SetMPTClearCanClawbackMutableFlag clears the CanClawback mutable flag.
func (m *MPTokenIssuanceSet) SetMPTClearCanClawbackMutableFlag() {
	m.setMutableFlag(TmfMPTClearCanClawback)
}

// Validate validates the MPTokenIssuanceSet transaction ensuring all fields are correct.
func (m *MPTokenIssuanceSet) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// MPTokenIssuanceID is required and must not be empty.
	if m.MPTokenIssuanceID == "" {
		return false, ErrInvalidMPTokenIssuanceID
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

	// MutableFlags cannot be zero when set.
	if m.MutableFlags != nil && *m.MutableFlags == 0 {
		return false, ErrMPTIssuanceSetMutableFlagsZero
	}

	// Validate MutableFlags: cannot set and clear the same flag simultaneously.
	if m.MutableFlags != nil {
		if ok, err := validateMutableFlagsNoConflict(*m.MutableFlags); !ok {
			return false, err
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

	// Non-zero TransferFee cannot be set together with tmfMPTClearCanTransfer (XLS-94).
	if m.TransferFee != nil && *m.TransferFee != 0 && m.MutableFlags != nil && flag.Contains(*m.MutableFlags, TmfMPTClearCanTransfer) {
		return false, ErrMPTIssuanceSetTransferFeeWithClearCanTransfer
	}

	return true, nil
}

// validateMutableFlagsNoConflict checks that no set/clear pair is active simultaneously.
func validateMutableFlagsNoConflict(mf uint32) (bool, error) {
	pairs := [][2]uint32{
		{TmfMPTSetCanLock, TmfMPTClearCanLock},
		{TmfMPTSetRequireAuth, TmfMPTClearRequireAuth},
		{TmfMPTSetCanEscrow, TmfMPTClearCanEscrow},
		{TmfMPTSetCanTrade, TmfMPTClearCanTrade},
		{TmfMPTSetCanTransfer, TmfMPTClearCanTransfer},
		{TmfMPTSetCanClawback, TmfMPTClearCanClawback},
	}
	for _, p := range pairs {
		if flag.Contains(mf, p[0]) && flag.Contains(mf, p[1]) {
			return false, ErrMPTIssuanceSetMutableFlagsConflict
		}
	}
	return true, nil
}
