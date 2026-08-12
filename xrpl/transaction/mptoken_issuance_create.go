package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// TfMPTCanLock if set, indicates that the MPT can be locked both individually and globally.
	// If not set, the MPT cannot be locked in any way.
	TfMPTCanLock uint32 = 0x00000002
	// TfMPTRequireAuth if set, indicates that individual holders must be authorized.
	// This enables issuers to limit who can hold their assets.
	TfMPTRequireAuth uint32 = 0x00000004
	// TfMPTCanEscrow if set, indicates that individual holders can place their balances into an escrow.
	TfMPTCanEscrow uint32 = 0x00000008
	// TfMPTCanTrade if set, indicates that individual holders can trade their balances using the XRP Ledger DEX or AMM.
	TfMPTCanTrade uint32 = 0x00000010
	// TfMPTCanTransfer if set, indicates that tokens may be transferred to other accounts that are not the issuer.
	TfMPTCanTransfer uint32 = 0x00000020
	// TfMPTCanClawback if set, indicates that the issuer may use the Clawback transaction to claw back value from individual holders.
	TfMPTCanClawback uint32 = 0x00000040
	// TfMPTCanHoldConfidentialBalance if set, indicates that holders can hold confidential balances.
	TfMPTCanHoldConfidentialBalance uint32 = 0x00000080
)

// ImmutableFlags constants for MPTokenIssuanceCreate and MPTokenIssuanceSet.
// A set bit permanently prevents the related capability from being enabled or the related field from being changed.
const (
	// TifMPTCanLock makes the CanLock capability immutable.
	TifMPTCanLock uint32 = 0x00000002
	// TifMPTRequireAuth makes the RequireAuth capability immutable.
	TifMPTRequireAuth uint32 = 0x00000004
	// TifMPTCanEscrow makes the CanEscrow capability immutable.
	TifMPTCanEscrow uint32 = 0x00000008
	// TifMPTCanTrade makes the CanTrade capability immutable.
	TifMPTCanTrade uint32 = 0x00000010
	// TifMPTCanTransfer makes the CanTransfer capability immutable.
	TifMPTCanTransfer uint32 = 0x00000020
	// TifMPTCanClawback makes the CanClawback capability immutable.
	TifMPTCanClawback uint32 = 0x00000040
	// TifMPTCanHoldConfidentialBalance makes the confidential balance capability immutable.
	TifMPTCanHoldConfidentialBalance uint32 = 0x00000080
	// TifMPTMetadata makes MPTokenMetadata immutable.
	TifMPTMetadata uint32 = 0x00010000
	// TifMPTTransferFee makes TransferFee immutable.
	TifMPTTransferFee uint32 = 0x00020000

	validMPTokenIssuanceImmutableFlags = TifMPTCanLock |
		TifMPTRequireAuth |
		TifMPTCanEscrow |
		TifMPTCanTrade |
		TifMPTCanTransfer |
		TifMPTCanClawback |
		TifMPTCanHoldConfidentialBalance |
		TifMPTMetadata |
		TifMPTTransferFee
)

// MPTokenIssuanceCreateMetadata represents the resulting metadata of a succeeded MPTokenIssuanceCreate transaction.
// It extends from TxObjMeta.
type MPTokenIssuanceCreateMetadata struct {
	TxObjMeta
	MPTIssuanceID *types.MPTIssuanceID `json:"mpt_issuance_id,omitempty"`
}

// MPTokenIssuanceCreate represents a transaction to create a new MPTokenIssuance object.
// This is the only opportunity an issuer has to specify immutable token fields.
//
// Example:
//
// ```json
//
//	{
//	   "TransactionType": "MPTokenIssuanceCreate",
//	   "Account": "rajgkBmMxmz161r8bWYH7CQAFZP5bA9oSG",
//	   "AssetScale": 2,
//	   "TransferFee": 314,
//	   "MaximumAmount": "50000000",
//	   "Flags": 83659,
//	   "MPTokenMetadata": "FOO",
//	   "Fee": "10"
//	}
//
// ```
type MPTokenIssuanceCreate struct {
	BaseTx
	// An asset scale is the difference, in orders of magnitude, between a standard unit and
	// a corresponding fractional unit. More formally, the asset scale is a non-negative integer
	// (0, 1, 2, …) such that one standard unit equals 10^(-scale) of a corresponding
	// fractional unit. If the fractional unit equals the standard unit, then the asset scale is 0.
	// Note that this value is optional, and will default to 0 if not supplied.
	AssetScale *uint8 `json:",omitempty"`
	// Specifies the fee to charged by the issuer for secondary sales of the Token,
	// if such sales are allowed. Valid values for this field are between 0 and 50,000 inclusive,
	// allowing transfer rates of between 0.000% and 50.000% in increments of 0.001.
	// The field must NOT be present if the `TfMPTCanTransfer` flag is not set.
	TransferFee *uint16 `json:",omitempty"`
	// Specifies the maximum number of MPT units that may be issued.
	// When present, the value must be between 1 and 2^63-1. If omitted, the
	// protocol currently defaults to 2^63-1.
	MaximumAmount *types.MPTAmount `json:",omitempty"`
	// MPTokenMetadata is arbitrary metadata about this issuance in hex format.
	// The limit for this field is 1024 bytes.
	MPTokenMetadata *string `json:",omitempty"`
	// DomainID is the ledger entry ID of a permissioned domain that grants access to the MPT.
	// Requires the TfMPTRequireAuth flag to be set.
	DomainID *string `json:",omitempty"`
	// ImmutableFlags identifies capabilities and fields that can no longer change after creation.
	ImmutableFlags *uint32 `json:",omitempty"`
}

// TxType returns the type of the transaction (MPTokenIssuanceCreate).
func (*MPTokenIssuanceCreate) TxType() TxType {
	return MPTokenIssuanceCreateTx
}

// Flatten returns the flattened map of the MPTokenIssuanceCreate transaction.
func (m *MPTokenIssuanceCreate) Flatten() FlatTransaction {
	flattened := m.BaseTx.Flatten()

	flattened["TransactionType"] = "MPTokenIssuanceCreate"

	if m.AssetScale != nil {
		flattened["AssetScale"] = *m.AssetScale
	}

	if m.TransferFee != nil {
		flattened["TransferFee"] = *m.TransferFee
	}

	if m.MaximumAmount != nil {
		flattened["MaximumAmount"] = m.MaximumAmount.Flatten()
	}

	if m.MPTokenMetadata != nil {
		flattened["MPTokenMetadata"] = *m.MPTokenMetadata
	}

	if m.DomainID != nil {
		flattened["DomainID"] = *m.DomainID
	}

	if m.ImmutableFlags != nil {
		flattened["ImmutableFlags"] = *m.ImmutableFlags
	}

	return flattened
}

// SetMPTCanLockFlag sets the TfMPTCanLock flag to allow the MPT to be locked both individually and globally.
func (m *MPTokenIssuanceCreate) SetMPTCanLockFlag() {
	m.Flags |= TfMPTCanLock
}

// SetMPTRequireAuthFlag sets the TfMPTRequireAuth flag to require individual holders to be authorized.
func (m *MPTokenIssuanceCreate) SetMPTRequireAuthFlag() {
	m.Flags |= TfMPTRequireAuth
}

// SetMPTCanEscrowFlag sets the TfMPTCanEscrow flag to allow individual holders to place their balances into an escrow.
func (m *MPTokenIssuanceCreate) SetMPTCanEscrowFlag() {
	m.Flags |= TfMPTCanEscrow
}

// SetMPTCanTradeFlag sets the TfMPTCanTrade flag to allow individual holders to trade their balances via DEX or AMM.
func (m *MPTokenIssuanceCreate) SetMPTCanTradeFlag() {
	m.Flags |= TfMPTCanTrade
}

// SetMPTCanTransferFlag sets the TfMPTCanTransfer flag to allow tokens to be transferred to non-issuer accounts.
func (m *MPTokenIssuanceCreate) SetMPTCanTransferFlag() {
	m.Flags |= TfMPTCanTransfer
}

// SetMPTCanClawbackFlag sets the TfMPTCanClawback flag to allow the issuer to claw back tokens from individual holders.
func (m *MPTokenIssuanceCreate) SetMPTCanClawbackFlag() {
	m.Flags |= TfMPTCanClawback
}

// SetMPTCanHoldConfidentialBalanceFlag sets the confidential balance capability.
func (m *MPTokenIssuanceCreate) SetMPTCanHoldConfidentialBalanceFlag() {
	m.Flags |= TfMPTCanHoldConfidentialBalance
}

func (m *MPTokenIssuanceCreate) setImmutableFlag(f uint32) {
	if m.ImmutableFlags == nil {
		m.ImmutableFlags = new(uint32)
	}
	*m.ImmutableFlags |= f
}

// SetMPTCanLockImmutableFlag makes the CanLock capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanLockImmutableFlag() {
	m.setImmutableFlag(TifMPTCanLock)
}

// SetMPTRequireAuthImmutableFlag makes the RequireAuth capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTRequireAuthImmutableFlag() {
	m.setImmutableFlag(TifMPTRequireAuth)
}

// SetMPTCanEscrowImmutableFlag makes the CanEscrow capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanEscrowImmutableFlag() {
	m.setImmutableFlag(TifMPTCanEscrow)
}

// SetMPTCanTradeImmutableFlag makes the CanTrade capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanTradeImmutableFlag() {
	m.setImmutableFlag(TifMPTCanTrade)
}

// SetMPTCanTransferImmutableFlag makes the CanTransfer capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanTransferImmutableFlag() {
	m.setImmutableFlag(TifMPTCanTransfer)
}

// SetMPTCanClawbackImmutableFlag makes the CanClawback capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanClawbackImmutableFlag() {
	m.setImmutableFlag(TifMPTCanClawback)
}

// SetMPTCanHoldConfidentialBalanceImmutableFlag makes the confidential balance capability immutable.
func (m *MPTokenIssuanceCreate) SetMPTCanHoldConfidentialBalanceImmutableFlag() {
	m.setImmutableFlag(TifMPTCanHoldConfidentialBalance)
}

// SetMPTMetadataImmutableFlag makes MPTokenMetadata immutable.
func (m *MPTokenIssuanceCreate) SetMPTMetadataImmutableFlag() {
	m.setImmutableFlag(TifMPTMetadata)
}

// SetMPTTransferFeeImmutableFlag makes TransferFee immutable.
func (m *MPTokenIssuanceCreate) SetMPTTransferFeeImmutableFlag() {
	m.setImmutableFlag(TifMPTTransferFee)
}

// Validate validates the MPTokenIssuanceCreate transaction ensuring all fields are correct.
func (m *MPTokenIssuanceCreate) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// Validate TransferFee: must not exceed MAX_TRANSFER_FEE and requires TfMPTCanTransfer flag.
	if m.TransferFee != nil && *m.TransferFee > 0 {
		if *m.TransferFee > MaxTransferFee {
			return false, ErrInvalidTransferFee
		}
		if !flag.Contains(m.Flags, TfMPTCanTransfer) {
			return false, ErrTransferFeeRequiresCanTransfer
		}
		if flag.Contains(m.Flags, TfMPTCanHoldConfidentialBalance) {
			return false, ErrMPTIssuanceCreateTransferFeeWithConfidentialBalance
		}
	}

	if m.MaximumAmount != nil && (m.MaximumAmount.IsZero() || !m.MaximumAmount.IsValid()) {
		return false, ErrMPTIssuanceCreateMaximumAmountInvalid
	}

	// Validate MPTokenMetadata: ensure it's in hex format and at most 1024 bytes (2048 chars).
	if m.MPTokenMetadata != nil && !ValidateHexMetadata(*m.MPTokenMetadata, 2*types.MaxMPTokenMetadataByteLength) {
		return false, ErrInvalidMPTokenMetadata
	}

	// DomainID must be a valid 64-char hex string and requires TfMPTRequireAuth flag.
	if m.DomainID != nil {
		if !IsDomainID(*m.DomainID) {
			return false, ErrMPTIssuanceCreateDomainIDInvalid
		}
		if !flag.Contains(m.Flags, TfMPTRequireAuth) {
			return false, ErrMPTIssuanceCreateDomainIDRequiresRequireAuth
		}
	}

	if m.ImmutableFlags != nil {
		if *m.ImmutableFlags == 0 {
			return false, ErrMPTIssuanceCreateImmutableFlagsZero
		}
		if *m.ImmutableFlags&^uint32(validMPTokenIssuanceImmutableFlags) != 0 {
			return false, ErrMPTIssuanceCreateInvalidImmutableFlags
		}
	}

	return true, nil
}
