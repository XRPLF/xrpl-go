package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// VaultClawback performs a Clawback from the Vault, exchanging the shares of an account.
//
// Conceptually, the transaction performs VaultWithdraw on behalf of the Holder, sending the funds to the
// Issuer account of the asset. In case there are insufficient funds for the entire Amount the transaction
// will perform a partial Clawback, up to the Vault.AssetsAvailable. The Clawback transaction must respect
// any future fees or penalties.
//
// ```json
//
//	{
//	  "TransactionType": "VaultClawback",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Holder": "rwCNM7SeUHTajEBQDiNqxDG8p1Mreizw85"
//	}
//
// ```
type VaultClawback struct {
	BaseTx
	// The ID of the vault from which assets are withdrawn.
	VaultID types.Hash256
	// The account ID from which to clawback the assets.
	Holder types.Address
	// The asset amount to clawback. When Amount is 0 clawback all funds, up to the total shares the Holder owns.
	Amount types.CurrencyAmount `json:",omitempty"`
}

// TxType returns the TxType for VaultClawback transactions.
func (tx *VaultClawback) TxType() TxType {
	return VaultClawbackTx
}

// Flatten returns a map representation of the VaultClawback transaction for JSON-RPC submission.
func (tx *VaultClawback) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["VaultID"] = tx.VaultID.String()
	flattened["Holder"] = tx.Holder.String()

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	return flattened
}

// Validate checks VaultClawback transaction fields and returns false with an error if invalid.
func (tx *VaultClawback) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrVaultClawbackVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID.String()) {
		return false, ErrVaultClawbackVaultIDInvalid
	}

	if tx.Holder == "" {
		return false, ErrVaultClawbackHolderRequired
	}

	if !addresscodec.IsValidAddress(tx.Holder.String()) {
		return false, ErrVaultClawbackHolderInvalid
	}

	if tx.Amount != nil {
		switch tx.Amount.Kind() { //nolint:exhaustive // ISSUED is handled by default case
		case types.XRP:
			return false, ErrVaultClawbackAmountInvalidType
		case types.MPT:
			if ok, err := IsMPTCurrency(tx.Amount); !ok {
				return false, err
			}
		default:
			if ok, err := IsIssuedCurrency(tx.Amount); !ok {
				return false, err
			}
		}
	}

	return true, nil
}
