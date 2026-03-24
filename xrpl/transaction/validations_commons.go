package transaction

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ValidateOptionalField validates an optional field in the transaction map.
func ValidateOptionalField(tx FlatTransaction, paramName string, checkValidity func(any) bool) error {
	// Check if the field is present in the transaction map.
	if value, ok := tx[paramName]; ok {
		// Check if the field is valid.
		if !checkValidity(value) {
			transactionType, _ := tx["TransactionType"].(string)
			return ErrTransactionInvalidField{
				Type:  transactionType,
				Field: paramName,
			}
		}
	}

	return nil
}

// validateMemos validates the Memos field in the transaction map.
func validateMemos(memoWrapper []types.MemoWrapper) error {
	// loop through each memo and validate it
	for _, memo := range memoWrapper {
		isMemo, err := IsMemo(memo.Memo)
		if !isMemo {
			return err
		}
	}

	return nil
}

// validateSigners validates the Signers field in the transaction map.
func validateSigners(signers []types.Signer) error {
	// loop through each signer and validate it
	for _, signer := range signers {
		isSigner, err := IsSigner(signer.SignerData)
		if !isSigner {
			return err
		}
	}

	return nil
}
