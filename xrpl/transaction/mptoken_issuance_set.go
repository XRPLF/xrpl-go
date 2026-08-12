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
	// TfMPTSetCanLock enables CanLock.
	TfMPTSetCanLock uint32 = 0x00000004
	// TfMPTSetRequireAuth enables RequireAuth.
	TfMPTSetRequireAuth uint32 = 0x00000008
	// TfMPTSetCanEscrow enables CanEscrow.
	TfMPTSetCanEscrow uint32 = 0x00000010
	// TfMPTSetCanTrade enables CanTrade.
	TfMPTSetCanTrade uint32 = 0x00000020
	// TfMPTSetCanTransfer enables CanTransfer.
	TfMPTSetCanTransfer uint32 = 0x00000040
	// TfMPTSetCanClawback enables CanClawback.
	TfMPTSetCanClawback uint32 = 0x00000080
	// TfMPTSetCanHoldConfidentialBalance enables confidential balances.
	TfMPTSetCanHoldConfidentialBalance uint32 = 0x00000100

	mpTokenIssuanceSetEnableFlagMask = TfMPTSetCanLock |
		TfMPTSetRequireAuth |
		TfMPTSetCanEscrow |
		TfMPTSetCanTrade |
		TfMPTSetCanTransfer |
		TfMPTSetCanClawback |
		TfMPTSetCanHoldConfidentialBalance
)

// MPTokenIssuanceSet transaction is used to lock or unlock an MPT and to update Dynamic MPT properties.
//
// ```json
//
//	{
//	      "TransactionType": "MPTokenIssuanceSet",
//	      "Fee": "10",
//	      "MPTokenIssuanceID": "000004C463C52827307480341125DA0577DEFC38405B0E3E",
//	      "Flags": 1
//	}
//
// ```
type MPTokenIssuanceSet struct {
	BaseTx
	// The MPTokenIssuance identifier.
	MPTokenIssuanceID string
	// Holder is the optional XRPL address of a token holder balance to lock or unlock.
	Holder *types.Address
	// DomainID is the optional permissioned domain to associate with this issuance.
	// An empty string removes the domain.
	DomainID *string `json:",omitempty"`
	// MPTokenMetadata is the optional new metadata. An empty string removes the metadata.
	MPTokenMetadata *string `json:",omitempty"`
	// TransferFee is the optional new transfer fee value between 0 and 50,000.
	TransferFee *uint16 `json:",omitempty"`
	// ImmutableFlags adds permanent restrictions to issuance capabilities and fields.
	ImmutableFlags *uint32 `json:",omitempty"`
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
	if m.ImmutableFlags != nil {
		flattened["ImmutableFlags"] = *m.ImmutableFlags
	}

	return flattened
}

// SetMPTLockFlag sets the TfMPTLock flag on the transaction.
func (m *MPTokenIssuanceSet) SetMPTLockFlag() {
	m.Flags |= TfMPTLock
}

// SetMPTUnlockFlag sets the TfMPTUnlock flag on the transaction.
func (m *MPTokenIssuanceSet) SetMPTUnlockFlag() {
	m.Flags |= TfMPTUnlock
}

// SetMPTCanLockFlag enables CanLock.
func (m *MPTokenIssuanceSet) SetMPTCanLockFlag() {
	m.Flags |= TfMPTSetCanLock
}

// SetMPTRequireAuthFlag enables RequireAuth.
func (m *MPTokenIssuanceSet) SetMPTRequireAuthFlag() {
	m.Flags |= TfMPTSetRequireAuth
}

// SetMPTCanEscrowFlag enables CanEscrow.
func (m *MPTokenIssuanceSet) SetMPTCanEscrowFlag() {
	m.Flags |= TfMPTSetCanEscrow
}

// SetMPTCanTradeFlag enables CanTrade.
func (m *MPTokenIssuanceSet) SetMPTCanTradeFlag() {
	m.Flags |= TfMPTSetCanTrade
}

// SetMPTCanTransferFlag enables CanTransfer.
func (m *MPTokenIssuanceSet) SetMPTCanTransferFlag() {
	m.Flags |= TfMPTSetCanTransfer
}

// SetMPTCanClawbackFlag enables CanClawback.
func (m *MPTokenIssuanceSet) SetMPTCanClawbackFlag() {
	m.Flags |= TfMPTSetCanClawback
}

// SetMPTCanHoldConfidentialBalanceFlag enables confidential balances.
func (m *MPTokenIssuanceSet) SetMPTCanHoldConfidentialBalanceFlag() {
	m.Flags |= TfMPTSetCanHoldConfidentialBalance
}

func (m *MPTokenIssuanceSet) setImmutableFlag(f uint32) {
	if m.ImmutableFlags == nil {
		m.ImmutableFlags = new(uint32)
	}
	*m.ImmutableFlags |= f
}

// SetMPTCanLockImmutableFlag makes the CanLock capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanLockImmutableFlag() {
	m.setImmutableFlag(TifMPTCanLock)
}

// SetMPTRequireAuthImmutableFlag makes the RequireAuth capability immutable.
func (m *MPTokenIssuanceSet) SetMPTRequireAuthImmutableFlag() {
	m.setImmutableFlag(TifMPTRequireAuth)
}

// SetMPTCanEscrowImmutableFlag makes the CanEscrow capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanEscrowImmutableFlag() {
	m.setImmutableFlag(TifMPTCanEscrow)
}

// SetMPTCanTradeImmutableFlag makes the CanTrade capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanTradeImmutableFlag() {
	m.setImmutableFlag(TifMPTCanTrade)
}

// SetMPTCanTransferImmutableFlag makes the CanTransfer capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanTransferImmutableFlag() {
	m.setImmutableFlag(TifMPTCanTransfer)
}

// SetMPTCanClawbackImmutableFlag makes the CanClawback capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanClawbackImmutableFlag() {
	m.setImmutableFlag(TifMPTCanClawback)
}

// SetMPTCanHoldConfidentialBalanceImmutableFlag makes the confidential balance capability immutable.
func (m *MPTokenIssuanceSet) SetMPTCanHoldConfidentialBalanceImmutableFlag() {
	m.setImmutableFlag(TifMPTCanHoldConfidentialBalance)
}

// SetMPTMetadataImmutableFlag makes MPTokenMetadata immutable.
func (m *MPTokenIssuanceSet) SetMPTMetadataImmutableFlag() {
	m.setImmutableFlag(TifMPTMetadata)
}

// SetMPTTransferFeeImmutableFlag makes TransferFee immutable.
func (m *MPTokenIssuanceSet) SetMPTTransferFeeImmutableFlag() {
	m.setImmutableFlag(TifMPTTransferFee)
}

// Validate validates the MPTokenIssuanceSet transaction ensuring all fields are correct.
func (m *MPTokenIssuanceSet) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if _, ok := decodeMPTIssuanceID(m.MPTokenIssuanceID); !ok {
		return false, ErrInvalidMPTokenIssuanceIDSet
	}
	if m.Holder != nil && !addresscodec.IsValidAddress(m.Holder.String()) {
		return false, ErrInvalidAccount
	}
	if m.Holder != nil && m.Account.String() == m.Holder.String() {
		return false, ErrHolderAccountConflict
	}

	allowedFlags := types.TfUniversal | TfMPTLock | TfMPTUnlock | mpTokenIssuanceSetEnableFlagMask
	if !flag.ContainsOnly(m.Flags, allowedFlags) {
		return false, ErrMPTIssuanceSetInvalidFlags
	}

	isLock := flag.Contains(m.Flags, TfMPTLock)
	isUnlock := flag.Contains(m.Flags, TfMPTUnlock)
	if isLock && isUnlock {
		return false, ErrMPTokenIssuanceSetFlags
	}

	hasEnableFlag := m.Flags&mpTokenIssuanceSetEnableFlagMask != 0
	isMutate := hasEnableFlag || m.ImmutableFlags != nil || m.MPTokenMetadata != nil || m.TransferFee != nil

	if m.Flags == 0 && !isMutate && m.DomainID == nil {
		return false, ErrMPTIssuanceSetEmpty
	}
	if m.Holder != nil && (isMutate || m.DomainID != nil) {
		return false, ErrMPTIssuanceSetHolderMutuallyExclusive
	}
	if isMutate && (isLock || isUnlock) {
		return false, ErrMPTIssuanceSetFlagsMutuallyExclusive
	}

	if m.ImmutableFlags != nil {
		if *m.ImmutableFlags == 0 {
			return false, ErrMPTIssuanceSetImmutableFlagsZero
		}
		if *m.ImmutableFlags&^uint32(validMPTokenIssuanceImmutableFlags) != 0 {
			return false, ErrMPTIssuanceSetInvalidImmutableFlags
		}
	}

	if m.TransferFee != nil {
		if *m.TransferFee > MaxTransferFee {
			return false, ErrInvalidTransferFee
		}
		if *m.TransferFee > 0 && flag.Contains(m.Flags, TfMPTSetCanHoldConfidentialBalance) {
			return false, ErrMPTIssuanceSetTransferFeeWithConfidentialBalance
		}
	}
	if m.MPTokenMetadata != nil && *m.MPTokenMetadata != "" && !ValidateHexMetadata(*m.MPTokenMetadata, 2*types.MaxMPTokenMetadataByteLength) {
		return false, ErrInvalidMPTokenMetadata
	}
	if m.DomainID != nil && *m.DomainID != "" && !IsDomainID(*m.DomainID) {
		return false, ErrMPTIssuanceSetDomainIDInvalid
	}

	return true, nil
}
