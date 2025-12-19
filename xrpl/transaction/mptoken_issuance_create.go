package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// If set, indicates that the MPT can be locked both individually and globally.
	// If not set, the MPT cannot be locked in any way.
	tfMPTCanLock uint32 = 0x00000002
	// If set, indicates that individual holders must be authorized.
	// This enables issuers to limit who can hold their assets.
	tfMPTRequireAuth uint32 = 0x00000004
	// If set, indicates that individual holders can place their balances into an escrow.
	tfMPTCanEscrow uint32 = 0x00000008
	// If set, indicates that individual holders can trade their balances
	//  using the XRP Ledger DEX or AMM.
	tfMPTCanTrade uint32 = 0x00000010
	// If set, indicates that tokens may be transferred to other accounts
	//  that are not the issuer.
	tfMPTCanTransfer uint32 = 0x00000020
	// If set, indicates that the issuer may use the Clawback transaction
	// to clawback value from individual holders.
	tfMPTCanClawback uint32 = 0x00000040
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
	// The field must NOT be present if the `tfMPTCanTransfer` flag is not set.
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
	MPTokenMetadata *string
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
		flattened["AssetScale"] = int(*m.AssetScale)
	}

	if m.TransferFee != nil {
		flattened["TransferFee"] = int(*m.TransferFee)
	}

	if m.MaximumAmount != nil {
		flattened["MaximumAmount"] = m.MaximumAmount.Flatten()
	}

	if m.MPTokenMetadata != nil {
		flattened["MPTokenMetadata"] = *m.MPTokenMetadata
	}

	return flattened
}

// SetMPTCanLockFlag sets the tfMPTCanLock flag to allow the MPT to be locked both individually and globally.
func (m *MPTokenIssuanceCreate) SetMPTCanLockFlag() {
	m.Flags |= tfMPTCanLock
}

// SetMPTRequireAuthFlag sets the tfMPTRequireAuth flag to require individual holders to be authorized.
func (m *MPTokenIssuanceCreate) SetMPTRequireAuthFlag() {
	m.Flags |= tfMPTRequireAuth
}

// SetMPTCanEscrowFlag sets the tfMPTCanEscrow flag to allow individual holders to place their balances into an escrow.
func (m *MPTokenIssuanceCreate) SetMPTCanEscrowFlag() {
	m.Flags |= tfMPTCanEscrow
}

// SetMPTCanTradeFlag sets the tfMPTCanTrade flag to allow individual holders to trade their balances via DEX or AMM.
func (m *MPTokenIssuanceCreate) SetMPTCanTradeFlag() {
	m.Flags |= tfMPTCanTrade
}

// SetMPTCanTransferFlag sets the tfMPTCanTransfer flag to allow tokens to be transferred to non-issuer accounts.
func (m *MPTokenIssuanceCreate) SetMPTCanTransferFlag() {
	m.Flags |= tfMPTCanTransfer
}

// SetMPTCanClawbackFlag sets the tfMPTCanClawback flag to allow the issuer to claw back tokens from individual holders.
func (m *MPTokenIssuanceCreate) SetMPTCanClawbackFlag() {
	m.Flags |= tfMPTCanClawback
}

// Validate validates the MPTokenIssuanceCreate transaction ensuring all fields are correct.
func (m *MPTokenIssuanceCreate) Validate() (bool, error) {
	ok, err := m.BaseTx.Validate()
	if err != nil || !ok {
		return false, err
	}

	// Validate TransferFee: must not exceed MAX_TRANSFER_FEE and requires tfMPTCanTransfer flag.
	if m.TransferFee != nil && *m.TransferFee > 0 {
		if *m.TransferFee > MaxTransferFee {
			return false, ErrInvalidTransferFee
		}
		if !types.IsFlagEnabled(m.Flags, tfMPTCanTransfer) {
			return false, ErrTransferFeeRequiresCanTransfer
		}
	}

	if m.MaximumAmount != nil {
		if ok, err := IsAmount(*m.MaximumAmount, "MaximumAmount", true); !ok {
			return false, err
		}
	}

	// Validate MPTokenMetadata: ensure it's in hex format.
	// This assumes m.MPTokenMetadata.String() returns its hex representation.
	if m.MPTokenMetadata != nil && !typecheck.IsHex(*m.MPTokenMetadata) {
		return false, ErrInvalidMPTokenMetadata
	}

	return true, nil
}
