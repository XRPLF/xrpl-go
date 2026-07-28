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
	// TfMPTCanHoldConfidentialBalance if set, indicates that confidential transfers are enabled for this token issuance.
	TfMPTCanHoldConfidentialBalance uint32 = 0x00000080
)

// MutableFlags constants for MPTokenIssuanceCreate declare future permissions and an opt-out.
// CanEnable permissions allow only a later transition from disabled to enabled; they never
// permit disabling. CanMutate permissions allow repeated metadata or transfer-fee updates.
// CannotEnableCanHoldConfidentialBalance permanently declines the confidential capability.
const (
	// TmfMPTCanEnableCanLock grants one-way permission to enable CanLock after creation.
	TmfMPTCanEnableCanLock uint32 = 0x00000002
	// TmfMPTCanEnableRequireAuth grants one-way permission to enable RequireAuth after creation.
	TmfMPTCanEnableRequireAuth uint32 = 0x00000004
	// TmfMPTCanEnableCanEscrow grants one-way permission to enable CanEscrow after creation.
	TmfMPTCanEnableCanEscrow uint32 = 0x00000008
	// TmfMPTCanEnableCanTrade grants one-way permission to enable CanTrade after creation.
	TmfMPTCanEnableCanTrade uint32 = 0x00000010
	// TmfMPTCanEnableCanTransfer grants one-way permission to enable CanTransfer after creation.
	TmfMPTCanEnableCanTransfer uint32 = 0x00000020
	// TmfMPTCanEnableCanClawback grants one-way permission to enable CanClawback after creation.
	TmfMPTCanEnableCanClawback uint32 = 0x00000040
	// TmfMPTCanMutateMetadata grants repeatable permission to replace MPTokenMetadata after creation.
	TmfMPTCanMutateMetadata uint32 = 0x00010000
	// TmfMPTCanMutateTransferFee grants repeatable permission to replace TransferFee after creation.
	TmfMPTCanMutateTransferFee uint32 = 0x00020000
	// TmfMPTCannotEnableCanHoldConfidentialBalance permanently opts out of enabling confidential balances after creation.
	TmfMPTCannotEnableCanHoldConfidentialBalance uint32 = 0x00000080
	// MPTokenIssuanceCreateMutableFlagsMask contains every supported MutableFlags bit. Bit 0x01 is reserved.
	MPTokenIssuanceCreateMutableFlagsMask uint32 = 0x000300FE
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
	// Specifies the maximum asset amount of this token that should ever be issued.
	// It is a non-negative integer string that can store a range of up to 63 bits. If not set, the max
	// amount will default to the largest unsigned 63-bit integer (0x7FFFFFFFFFFFFFFF or 9223372036854775807)
	//
	// Example:
	// ```
	// MaximumAmount: '9223372036854775807'
	// ```
	MaximumAmount *types.XRPCurrencyAmount `json:",omitempty"`
	// MPTokenMetadata is arbitrary metadata about this issuance in hex format.
	// The limit for this field is 1024 bytes.
	MPTokenMetadata *string `json:",omitempty"`
	// DomainID is the ledger entry ID of a permissioned domain that grants access to the MPT.
	// Requires the TfMPTRequireAuth flag to be set.
	DomainID *string `json:",omitempty"`
	// MutableFlags grants one-way future-enable permissions for ordinary capabilities,
	// repeatable mutation permission for metadata and transfer fees, and a permanent
	// opt-out from enabling confidential balances.
	MutableFlags *uint32 `json:",omitempty"`
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

	if m.MutableFlags != nil {
		flattened["MutableFlags"] = *m.MutableFlags
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

// SetMPTCanHoldConfidentialBalanceFlag sets the TfMPTCanHoldConfidentialBalance flag to enable confidential transfers for this token issuance.
func (m *MPTokenIssuanceCreate) SetMPTCanHoldConfidentialBalanceFlag() {
	m.Flags |= TfMPTCanHoldConfidentialBalance
}

// setMutableFlag is a helper that initialises MutableFlags if nil and applies the given flag.
func (m *MPTokenIssuanceCreate) setMutableFlag(f uint32) {
	if m.MutableFlags == nil {
		mf := uint32(0)
		m.MutableFlags = &mf
	}
	*m.MutableFlags |= f
}

// SetMPTCanEnableCanLockFlag grants one-way permission to enable CanLock after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableCanLockFlag() {
	m.setMutableFlag(TmfMPTCanEnableCanLock)
}

// SetMPTCanEnableRequireAuthFlag grants one-way permission to enable RequireAuth after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableRequireAuthFlag() {
	m.setMutableFlag(TmfMPTCanEnableRequireAuth)
}

// SetMPTCanEnableCanEscrowFlag grants one-way permission to enable CanEscrow after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableCanEscrowFlag() {
	m.setMutableFlag(TmfMPTCanEnableCanEscrow)
}

// SetMPTCanEnableCanTradeFlag grants one-way permission to enable CanTrade after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableCanTradeFlag() {
	m.setMutableFlag(TmfMPTCanEnableCanTrade)
}

// SetMPTCanEnableCanTransferFlag grants one-way permission to enable CanTransfer after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableCanTransferFlag() {
	m.setMutableFlag(TmfMPTCanEnableCanTransfer)
}

// SetMPTCanEnableCanClawbackFlag grants one-way permission to enable CanClawback after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanEnableCanClawbackFlag() {
	m.setMutableFlag(TmfMPTCanEnableCanClawback)
}

// SetMPTCanMutateMetadataFlag grants repeatable permission to replace MPTokenMetadata after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanMutateMetadataFlag() {
	m.setMutableFlag(TmfMPTCanMutateMetadata)
}

// SetMPTCanMutateTransferFeeFlag grants repeatable permission to replace TransferFee after creation.
func (m *MPTokenIssuanceCreate) SetMPTCanMutateTransferFeeFlag() {
	m.setMutableFlag(TmfMPTCanMutateTransferFee)
}

// SetMPTCannotEnableCanHoldConfidentialBalanceFlag permanently opts out of enabling confidential balances after creation.
func (m *MPTokenIssuanceCreate) SetMPTCannotEnableCanHoldConfidentialBalanceFlag() {
	m.setMutableFlag(TmfMPTCannotEnableCanHoldConfidentialBalance)
}

// Validate validates the MPTokenIssuanceCreate transaction ensuring all fields are correct.
func (m *MPTokenIssuanceCreate) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	if m.MutableFlags != nil && (*m.MutableFlags == 0 || *m.MutableFlags&^MPTokenIssuanceCreateMutableFlagsMask != 0) {
		return false, ErrMPTIssuanceCreateInvalidMutableFlags
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

	if m.MaximumAmount != nil {
		if ok, err := IsAmount(*m.MaximumAmount, "MaximumAmount", true); !ok {
			return false, err
		}
	}

	// Validate MPTokenMetadata: ensure it's in hex format and at most 1024 bytes (2048 chars).
	// This assumes m.MPTokenMetadata.String() returns its hex representation.
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

	return true, nil
}
