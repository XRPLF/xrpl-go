package transaction

import (
	"bytes"
	"encoding/hex"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// decodeAddressAccountID returns the AccountID represented by a classic or
// X-address and reports whether the X-address carries a tag.
func decodeAddressAccountID(address types.Address) (accountID []byte, hasTag bool, err error) {
	decoded, err := addresscodec.DecodeAddress(address.String())
	if err != nil {
		return nil, false, err
	}
	return decoded.AccountID[:], decoded.HasTag, nil
}

// sameAccountAddress compares account identity, ignoring X-address tags and network
// prefixes. Invalid addresses are not equal. Callers retain field-specific address
// validation and tag policies.
func sameAccountAddress(a, b types.Address) bool {
	aID, _, err := decodeAddressAccountID(a)
	if err != nil {
		return false
	}
	bID, _, err := decodeAddressAccountID(b)
	return err == nil && bytes.Equal(aID, bID)
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

// decodeCounterparty decodes a nonzero counterparty AccountID and compares it
// with accountID. Callers own tag restrictions, self-reference rules, and
// field-specific error wrapping. Identity ignores X-address tags and networks.
func decodeCounterparty(accountID []byte, counterparty types.Address) (counterpartyID []byte, same, hasTag bool, err error) {
	counterpartyID, hasTag, err = decodeAddressAccountID(counterparty)
	if err != nil {
		return nil, false, false, err
	}
	if addresscodec.IsZeroAccountID(counterpartyID) {
		return nil, false, false, ErrZeroAccountID
	}

	return counterpartyID, bytes.Equal(accountID, counterpartyID), hasTag, nil
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

// maxTransactionSigners is STTx::kMaxMultiSigners in rippled 21890d9d.
// This bounds transaction signatures, independently of SignerListSet entries.
const maxTransactionSigners = 32

// validateSigners checks entries and strict ascending decoded AccountID order.
// Nil and empty lists are allowed here. Callers enforce required list presence.
// Validation never sorts or otherwise changes the supplied signers.
func validateSigners(signers []types.Signer) error {
	if len(signers) > maxTransactionSigners {
		return errTooManyTransactionSigners
	}
	var previous []byte
	for _, signer := range signers {
		accountID, err := validateSignerData(signer.SignerData)
		if err != nil {
			return err
		}
		if previous != nil {
			switch bytes.Compare(previous, accountID) {
			case 0:
				return errDuplicateTransactionSigner
			case 1:
				return errUnsortedTransactionSigners
			}
		}
		previous = accountID
	}
	return nil
}
