package transaction

import (
	"encoding/hex"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// decodeAddressAccountID returns the AccountID represented by a classic or
// X-address and reports whether the X-address carries a tag.
func decodeAddressAccountID(address types.Address) (accountID []byte, hasTag bool, err error) {
	_, accountID, err = addresscodec.DecodeClassicAddressToAccountID(address.String())
	if err == nil {
		return accountID, false, nil
	}

	accountID, _, hasTag, _, err = addresscodec.DecodeXAddress(address.String())
	return accountID, hasTag, err
}

func decodeMPTIssuanceID(issuanceID string) ([]byte, bool) {
	idBytes, err := hex.DecodeString(issuanceID)
	if err != nil || len(idBytes) != bctypes.MPTIssuanceIDByteLength {
		return nil, false
	}
	return idBytes, true
}

// mptIssuerAccountID returns the issuer AccountID encoded in the trailing bytes of
// issuanceID. It reports false when issuanceID is not a well-formed MPT issuance ID.
func mptIssuerAccountID(issuanceID string) ([]byte, bool) {
	issuanceIDBytes, ok := decodeMPTIssuanceID(issuanceID)
	if !ok {
		return nil, false
	}

	return issuanceIDBytes[len(issuanceIDBytes)-addresscodec.AccountAddressLength:], true
}

// ValidateOptionalField validates an optional field in the transaction map.
func ValidateOptionalField(tx FlatTransaction, paramName string, checkValidity func(any) bool) error {
	// Check if the field is present in the transaction map.
	if value, ok := tx[paramName]; ok {
		// Check if the field is valid.
		if !checkValidity(value) {
			return ErrTransactionInvalidField{
				Type:  tx.TxType().String(),
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
