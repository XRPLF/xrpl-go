package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// VaultDeposit deposits assets into an existing Vault.
//
// ```json
//
//	{
//	  "TransactionType": "VaultDeposit",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430",
//	  "Amount": "1000000"
//	}
//
// ```
type VaultDeposit struct {
	BaseTx
	// The ID of the vault to which the assets are deposited.
	VaultID string
	// Asset amount to deposit.
	Amount types.CurrencyAmount
}

// TxType returns the TxType for VaultDeposit transactions.
func (tx *VaultDeposit) TxType() TxType {
	return VaultDepositTx
}

// Flatten returns a map representation of the VaultDeposit transaction for JSON-RPC submission.
func (tx *VaultDeposit) Flatten() map[string]interface{} {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["VaultID"] = tx.VaultID

	if tx.Amount != nil {
		flattened["Amount"] = tx.Amount.Flatten()
	}

	return flattened
}

// Validate checks VaultDeposit transaction fields and returns false with an error if invalid.
func (tx *VaultDeposit) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrVaultDepositVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID) {
		return false, ErrVaultDepositVaultIDInvalid
	}

	if tx.Amount == nil {
		return false, ErrVaultDepositAmountRequired
	}

	if ok, err := IsAmount(tx.Amount, "Amount", true); !ok {
		return false, err
	}

	return true, nil
}
