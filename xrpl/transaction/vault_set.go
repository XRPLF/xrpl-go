package transaction

import (
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// VaultSetMaxDataLength is the maximum length in characters for the Data field (256 bytes = 512 hex chars).
	VaultSetMaxDataLength = 512
)

// VaultSet updates an existing Vault object.
//
// ```json
//
//	{
//	  "TransactionType": "VaultSet",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
//	}
//
// ```
type VaultSet struct {
	BaseTx
	// The ID of the Vault to be modified. Must be included when updating the Vault.
	VaultID string
	// Arbitrary metadata in hex format. The field is limited to 256 bytes (512 hex chars).
	Data *types.Data `json:",omitempty"`
	// The maximum asset amount that can be held in a vault.
	AssetsMaximum *types.XRPLNumber `json:",omitempty"`
	// The PermissionedDomain object ID associated with the shares of this Vault.
	DomainID *string `json:",omitempty"`
}

// TxType returns the TxType for VaultSet transactions.
func (tx *VaultSet) TxType() TxType {
	return VaultSetTx
}

// Flatten returns a map representation of the VaultSet transaction for JSON-RPC submission.
func (tx *VaultSet) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["VaultID"] = tx.VaultID

	if tx.Data != nil && *tx.Data != "" {
		flattened["Data"] = string(*tx.Data)
	}

	if tx.AssetsMaximum != nil && *tx.AssetsMaximum != "" {
		flattened["AssetsMaximum"] = tx.AssetsMaximum.String()
	}

	if tx.DomainID != nil && *tx.DomainID != "" {
		flattened["DomainID"] = *tx.DomainID
	}

	return flattened
}

// Validate checks VaultSet transaction fields and returns false with an error if invalid.
func (tx *VaultSet) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrVaultSetVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID) {
		return false, ErrVaultSetVaultIDInvalid
	}

	if tx.Data != nil && *tx.Data != "" {
		if !ValidateHexMetadata(tx.Data.Value(), VaultSetMaxDataLength) {
			return false, ErrVaultSetDataInvalid
		}
	}

	if tx.AssetsMaximum != nil && *tx.AssetsMaximum != "" && !typecheck.IsXRPLNumber(tx.AssetsMaximum.String()) {
		return false, ErrVaultSetAssetsMaximumInvalid
	}

	if tx.DomainID != nil && *tx.DomainID != "" {
		if !IsDomainID(*tx.DomainID) {
			return false, ErrVaultSetDomainIDInvalid
		}
	}

	return true, nil
}
