package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/flag"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// VaultCreateMaxDataLength is the maximum length in characters for the Data field (256 bytes = 512 hex chars).
	VaultCreateMaxDataLength = 512
	// VaultCreateMaxMPTokenMetadataLength is the maximum length in characters for the MPTokenMetadata field (1024 bytes = 2048 hex chars).
	VaultCreateMaxMPTokenMetadataLength = 2048
	// VaultCreateMaxScale is the maximum value for Scale.
	VaultCreateMaxScale = 18

	// TfVaultPrivate indicates that the Vault is private.
	TfVaultPrivate uint32 = 0x00010000
	// TfVaultShareNonTransferable indicates that the Vault shares are non-transferable.
	TfVaultShareNonTransferable uint32 = 0x00020000
)

// VaultCreate creates a new Vault object.
//
// ```json
//
//	{
//	  "TransactionType": "VaultCreate",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "Asset": {
//	     "currency": "USD", "issuer": "rXJSJiZMxaLuH3kQBUV5DLipnYtrE6iVb"
//	  }
//	}
//
// ```
type VaultCreate struct {
	BaseTx
	// The asset (XRP, IOU or MPT) of the Vault.
	Asset ledger.Asset
	// Arbitrary metadata in hex format. The field is limited to 256 bytes (512 hex chars).
	Data *types.Data `json:",omitempty"`
	// The maximum asset amount that can be held in a vault.
	AssetsMaximum *types.XRPLNumber `json:",omitempty"`
	// Should follow the XLS-89 standard (https://github.com/XRPLF/XRPL-Standards/tree/master/XLS-0089-multi-purpose-token-metadata-schema).
	// Use EncodeMPTokenMetadata/DecodeMPTokenMetadata utility functions to convert to/from a blob.
	// While adherence to the XLS-89 format is not mandatory, non-compliant metadata
	// may not be discoverable by ecosystem tools such as explorers and indexers.
	MPTokenMetadata *string `json:",omitempty"`
	// Indicates the withdrawal strategy used by the Vault.
	WithdrawalPolicy *types.VaultWithdrawalPolicy `json:",omitempty"`
	// The PermissionedDomain object ID associated with the shares of this Vault.
	DomainID *string `json:",omitempty"`
	// The scaling factor for vault shares. Only applicable for IOU assets.
	// Valid values are between 0 and 18 inclusive. For XRP and MPT, this must not be provided.
	Scale *uint8 `json:",omitempty"`
}

// TxType returns the TxType for VaultCreate transactions.
func (tx *VaultCreate) TxType() TxType {
	return VaultCreateTx
}

// SetVaultPrivateFlag sets the TfVaultPrivate flag, indicating that the Vault is private.
func (tx *VaultCreate) SetVaultPrivateFlag() {
	tx.Flags |= TfVaultPrivate
}

// SetVaultShareNonTransferableFlag sets the TfVaultShareNonTransferable flag, indicating that the Vault shares are non-transferable.
func (tx *VaultCreate) SetVaultShareNonTransferableFlag() {
	tx.Flags |= TfVaultShareNonTransferable
}

// Flatten returns a map representation of the VaultCreate transaction for JSON-RPC submission.
func (tx *VaultCreate) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	if tx.Asset != (ledger.Asset{}) {
		flattened["Asset"] = tx.Asset.Flatten()
	}

	if tx.Data != nil && *tx.Data != "" {
		flattened["Data"] = tx.Data.Value()
	}

	if tx.AssetsMaximum != nil && *tx.AssetsMaximum != "" {
		flattened["AssetsMaximum"] = tx.AssetsMaximum.String()
	}

	if tx.MPTokenMetadata != nil && *tx.MPTokenMetadata != "" {
		flattened["MPTokenMetadata"] = *tx.MPTokenMetadata
	}

	if tx.WithdrawalPolicy != nil {
		flattened["WithdrawalPolicy"] = tx.WithdrawalPolicy.Value()
	}

	if tx.DomainID != nil && *tx.DomainID != "" {
		flattened["DomainID"] = *tx.DomainID
	}

	if tx.Scale != nil {
		flattened["Scale"] = *tx.Scale
	}

	return flattened
}

// Validate checks VaultCreate transaction fields and returns false with an error if invalid.
func (tx *VaultCreate) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.Asset == (ledger.Asset{}) {
		return false, ErrVaultCreateAssetRequired
	}

	if ok, err := IsAsset(tx.Asset); !ok {
		return false, err
	}

	if tx.Data != nil && *tx.Data != "" {
		if !ValidateHexMetadata(tx.Data.Value(), VaultCreateMaxDataLength) {
			return false, ErrVaultCreateDataInvalid
		}
	}

	if tx.AssetsMaximum != nil && *tx.AssetsMaximum != "" && !typecheck.IsXRPLNumber(tx.AssetsMaximum.String()) {
		return false, ErrVaultCreateAssetsMaximumInvalid
	}

	if tx.MPTokenMetadata != nil && *tx.MPTokenMetadata != "" {
		if !ValidateHexMetadata(*tx.MPTokenMetadata, VaultCreateMaxMPTokenMetadataLength) {
			return false, ErrVaultCreateMPTokenMetadataInvalid
		}
	}

	if tx.Scale != nil {
		if *tx.Scale > VaultCreateMaxScale {
			return false, ErrVaultCreateScaleInvalid
		}
		// Scale is only valid for IOU assets (not XRP or MPT)
		if tx.Asset.Kind() != ledger.AssetIOU {
			return false, ErrVaultCreateScaleRequiresIOU
		}
	}

	if tx.DomainID != nil && *tx.DomainID != "" {
		// DomainID requires the TfVaultPrivate flag
		if !flag.Contains(tx.Flags, TfVaultPrivate) {
			return false, ErrVaultCreateDomainIDRequiresPrivateFlag
		}
		if !IsDomainID(*tx.DomainID) {
			return false, ErrVaultCreateDomainIDInvalid
		}
	}

	return true, nil
}
