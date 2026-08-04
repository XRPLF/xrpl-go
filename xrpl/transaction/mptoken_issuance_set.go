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

// MutableFlags constants for MPTokenIssuanceSet. Each operation enables a capability and is one-way.
const (
	// TmfMPTSetCanLock enables CanLock.
	TmfMPTSetCanLock uint32 = 0x00000001
	// TmfMPTSetRequireAuth enables RequireAuth.
	TmfMPTSetRequireAuth uint32 = 0x00000002
	// TmfMPTSetCanEscrow enables CanEscrow.
	TmfMPTSetCanEscrow uint32 = 0x00000004
	// TmfMPTSetCanTrade enables CanTrade.
	TmfMPTSetCanTrade uint32 = 0x00000008
	// TmfMPTSetCanTransfer enables CanTransfer.
	TmfMPTSetCanTransfer uint32 = 0x00000010
	// TmfMPTSetCanClawback enables CanClawback.
	TmfMPTSetCanClawback uint32 = 0x00000020
	// TmfMPTSetCanHoldConfidentialBalance enables confidential balances.
	TmfMPTSetCanHoldConfidentialBalance uint32 = 0x00000040
	// MPTokenIssuanceSetMutableFlagsMask contains every supported MutableFlags bit.
	MPTokenIssuanceSetMutableFlagsMask uint32 = 0x0000007F
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
//	      "MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE2",
//	      "Flags": 1
//	}
//
// ```
type MPTokenIssuanceSet struct {
	BaseTx
	// The MPTokenIssuance identifier as exactly 48 hexadecimal characters.
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
	// (Optional) One-way enable operations for capabilities permitted by the issuance's MutableFlags.
	// Capabilities enabled here cannot be disabled.
	MutableFlags *uint32 `json:",omitempty"`
	// (Optional) A 33-byte compressed ElGamal public key for the issuer.
	// Required to use the confidential transfer feature. Must be 66 hex characters.
	IssuerEncryptionKey *string `json:",omitempty"`
	// (Optional) A 33-byte compressed ElGamal public key for an on-chain auditor.
	// Must be 66 hex characters. Requires IssuerEncryptionKey to also be set.
	AuditorEncryptionKey *string `json:",omitempty"`
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

	if m.IssuerEncryptionKey != nil {
		flattened["IssuerEncryptionKey"] = *m.IssuerEncryptionKey
	}

	if m.AuditorEncryptionKey != nil {
		flattened["AuditorEncryptionKey"] = *m.AuditorEncryptionKey
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

// SetMPTSetRequireAuthMutableFlag sets the RequireAuth mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetRequireAuthMutableFlag() {
	m.setMutableFlag(TmfMPTSetRequireAuth)
}

// SetMPTSetCanEscrowMutableFlag sets the CanEscrow mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanEscrowMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanEscrow)
}

// SetMPTSetCanTradeMutableFlag sets the CanTrade mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTradeMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTrade)
}

// SetMPTSetCanTransferMutableFlag sets the CanTransfer mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanTransferMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanTransfer)
}

// SetMPTSetCanClawbackMutableFlag sets the CanClawback mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanClawbackMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanClawback)
}

// SetMPTSetCanHoldConfidentialBalanceMutableFlag sets the CanHoldConfidentialBalance mutable flag.
func (m *MPTokenIssuanceSet) SetMPTSetCanHoldConfidentialBalanceMutableFlag() {
	m.setMutableFlag(TmfMPTSetCanHoldConfidentialBalance)
}

// Validate validates the MPTokenIssuanceSet transaction ensuring all fields are correct.
func (m *MPTokenIssuanceSet) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// MPTokenIssuanceID is required and must be exactly 24 bytes of hexadecimal.
	if !IsMPTIssuanceID(m.MPTokenIssuanceID) {
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
	hasEncryptionKeys := m.IssuerEncryptionKey != nil || m.AuditorEncryptionKey != nil

	// At least one operation must be specified (lock/unlock, holder lock/unlock, DynamicMPT mutation, DomainID, or encryption keys).
	if m.Flags == 0 && m.Holder == nil && !hasDynamicMPTFields && m.DomainID == nil && !hasEncryptionKeys {
		return false, ErrMPTIssuanceSetEmpty
	}

	// Holder is mutually exclusive with DynamicMPT fields and DomainID.
	if m.Holder != nil && (hasDynamicMPTFields || m.DomainID != nil) {
		return false, ErrMPTIssuanceSetHolderMutuallyExclusive
	}

	// Encryption keys are mutually exclusive with Holder.
	if m.Holder != nil && hasEncryptionKeys {
		return false, ErrMPTIssuanceSetKeyConflict
	}

	// Non-zero Flags are mutually exclusive with DynamicMPT fields.
	if m.Flags != 0 && hasDynamicMPTFields {
		return false, ErrMPTIssuanceSetFlagsMutuallyExclusive
	}

	// MutableFlags must contain at least one supported enable operation.
	if m.MutableFlags != nil && (*m.MutableFlags == 0 || *m.MutableFlags&^MPTokenIssuanceSetMutableFlagsMask != 0) {
		return false, ErrMPTIssuanceSetInvalidMutableFlags
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

	// Confidential balances are encrypted, so a transfer fee cannot be enabled with them.
	if m.TransferFee != nil && *m.TransferFee != 0 && m.MutableFlags != nil && flag.Contains(*m.MutableFlags, TmfMPTSetCanHoldConfidentialBalance) {
		return false, ErrMPTIssuanceSetTransferFeeWithConfidentialBalance
	}

	// AuditorEncryptionKey requires IssuerEncryptionKey.
	if m.AuditorEncryptionKey != nil && m.IssuerEncryptionKey == nil {
		return false, ErrMPTIssuanceSetAuditorRequiresIssuerKey
	}

	// Validate encryption key lengths (issuer and auditor keys must be 33-byte compressed).
	if m.IssuerEncryptionKey != nil && !IsValidCompressedEncryptionKey(*m.IssuerEncryptionKey) {
		return false, ErrMPTIssuanceSetInvalidKeyLength
	}
	if m.AuditorEncryptionKey != nil && !IsValidCompressedEncryptionKey(*m.AuditorEncryptionKey) {
		return false, ErrMPTIssuanceSetInvalidKeyLength
	}

	return true, nil
}
