package transaction

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// VaultDelete deletes an existing Vault object.
//
// ```json
//
//	{
//	  "TransactionType": "VaultDelete",
//	  "Account": "rNGHoQwNG753zyfDrib4qDvvswbrtmV8Es",
//	  "VaultID": "B91CD2033E73E0DD17AF043FBD458CE7D996850A83DCED23FB122A3BFAA7F430"
//	}
//
// ```
type VaultDelete struct {
	BaseTx
	// The ID of the Vault to be deleted.
	VaultID types.Hash256
}

// TxType returns the TxType for VaultDelete transactions.
func (tx *VaultDelete) TxType() TxType {
	return VaultDeleteTx
}

// Flatten returns a map representation of the VaultDelete transaction for JSON-RPC submission.
func (tx *VaultDelete) Flatten() FlatTransaction {
	flattened := tx.BaseTx.Flatten()

	flattened["TransactionType"] = tx.TxType().String()

	flattened["VaultID"] = tx.VaultID.String()

	return flattened
}

// Validate checks VaultDelete transaction fields and returns false with an error if invalid.
func (tx *VaultDelete) Validate() (bool, error) {
	if ok, err := tx.BaseTx.Validate(); !ok {
		return false, err
	}

	if tx.VaultID == "" {
		return false, ErrVaultDeleteVaultIDRequired
	}

	if !IsLedgerEntryID(tx.VaultID.String()) {
		return false, ErrVaultDeleteVaultIDInvalid
	}

	return true, nil
}
