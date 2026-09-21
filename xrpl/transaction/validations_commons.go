package transaction

import (
	"bytes"
	"encoding/hex"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
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

// assetIssuerAccountID returns the AccountID that issues asset. It reports false for
// XRP, which has no issuer, and for an asset whose issuer cannot be decoded.
func assetIssuerAccountID(asset ledger.Asset) ([]byte, bool) {
	switch asset.Kind() {
	case ledger.AssetIOU:
		issuerID, _, err := decodeAddressAccountID(asset.Issuer)
		return issuerID, err == nil
	case ledger.AssetMPT:
		return mptIssuerAccountID(asset.MPTIssuanceID)
	case ledger.AssetXRP:
	}
	return nil, false
}

// Issue keys exist because one token has several spellings. "USD" and its hex form
// are the same currency, and a classic address and an X-address are the same issuer,
// so tokens are compared as canonical bytes and never as strings. The parts are joined
// into one key so a single comparison also tells token kinds apart, because an issued
// key is 40 bytes and an MPT key is 24. Failing to build a key means the token is
// malformed, which callers report separately from a mismatch.

// assetIssueKey returns the issue key of asset, which must already satisfy IsAsset.
func assetIssueKey(asset ledger.Asset) ([]byte, bool) {
	switch asset.Kind() {
	case ledger.AssetIOU:
		currencyBytes, err := (&bctypes.Currency{}).FromJSON(asset.Currency)
		issuerID, ok := assetIssuerAccountID(asset)
		if err != nil || !ok {
			return nil, false
		}
		return append(currencyBytes, issuerID...), true
	case ledger.AssetMPT:
		return decodeMPTIssuanceID(asset.MPTIssuanceID)
	case ledger.AssetXRP:
	}
	return bctypes.XRPBytes, true
}

// amountIssueKey returns the issue key of the token that amount is denominated in.
// XRP amounts have no key here because no caller matches them against an asset.
func amountIssueKey(amount types.CurrencyAmount) ([]byte, bool) {
	switch amount := amount.(type) {
	case types.IssuedCurrencyAmount:
		currencyBytes, err := bctypes.SerializeIssuedCurrencyCode(amount.Currency)
		issuerID, hasTag, issuerErr := decodeAddressAccountID(amount.Issuer)
		if err != nil || issuerErr != nil || hasTag {
			return nil, false
		}
		return append(currencyBytes, issuerID...), true
	case types.MPTCurrencyAmount:
		return decodeMPTIssuanceID(amount.MPTIssuanceID)
	}
	return nil, false
}

// isPositiveTokenAmount reports whether amount is a well-formed issued or MPT amount
// greater than zero. XRP is not a token amount.
func isPositiveTokenAmount(amount types.CurrencyAmount) bool {
	if !IsTokenAmount(amount) {
		return false
	}
	ok, _ := IsAmount(amount, "Amount", true)
	return ok && !amount.IsZero()
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

// isPublicKey reports whether signingPubKey is a well-formed public key. An empty value fails.
func isPublicKey(signingPubKey string) bool {
	_, err := keypairs.DeriveClassicAddress(signingPubKey)
	return err == nil
}

// isSignaturePair reports whether a public key and signature are well formed.
// It never verifies the signature, because the signed payload is not available
// during field validation. An empty value fails either check.
func isSignaturePair(signingPubKey, txnSignature string) bool {
	return isPublicKey(signingPubKey) && typecheck.IsHexBlob(txnSignature)
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
