package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// VaultWithdraw withdraws assets from an existing Vault.
//
// ```json
//
//	{
//	  "TransactionType": "VaultWithdraw",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": "1000000"
//	}
//
// ```
type VaultWithdraw struct {
	BaseTx
	// The ID of the vault from which assets are withdrawn.
	VaultID string
	// The exact amount of Vault asset to withdraw.
	Amount types.CurrencyAmount
	// An account to receive the assets. It must be able to receive the asset.
	Destination *types.Address `json:",omitempty"`
	// Arbitrary tag identifying the reason for the withdrawal to the destination.
	DestinationTag *uint32 `json:",omitempty"`
}

// TxType returns the TxType for VaultWithdraw transactions.
func (tx *VaultWithdraw) TxType() TxType {
	return VaultWithdrawTx
}

// Flatten returns a map representation of the VaultWithdraw transaction for JSON-RPC submission.
func (tx *VaultWithdraw) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["VaultID"] = tx.VaultID

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	if tx.Destination != nil {
		flattened["Destination"] = tx.Destination.String()
	}

	if tx.DestinationTag != nil {
		flattened["DestinationTag"] = *tx.DestinationTag
	}

	return flattened
}

// Validate checks VaultWithdraw transaction fields and returns false with an error if invalid.
func (tx *VaultWithdraw) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrVaultWithdrawVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID) {
		return false, ErrVaultWithdrawVaultIDInvalid
	}

	if tx.Amount == nil {
		return false, ErrVaultWithdrawAmountRequired
	}

	if ok, err := IsAmount(tx.Amount, "Amount", true); !ok {
		return false, err
	}

	if tx.Destination != nil {
		if !addresscodec.IsValidAddress(tx.Destination.String()) {
			return false, ErrInvalidDestination
		}
	}

	return true, nil
}
