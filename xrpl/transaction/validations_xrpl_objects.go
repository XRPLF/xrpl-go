package transaction

import (
	"strconv"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	maputils "github.com/Peersyst/xrpl-go/pkg/map_utils"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

const (
	// MemoSize is the expected number of fields in a Memo object (MemoData, MemoFormat, MemoType).
	MemoSize = 3
	// SignerSize is the expected number of fields in a Signer object (Account, TxnSignature, SigningPubKey).
	SignerSize = 3
	// IssuedCurrencySize is the expected number of fields in an IssuedCurrency object (currency, issuer, value).
	IssuedCurrencySize = 3
	// StandardCurrencyCodeLen is the required length of a standard three-character currency code.
	StandardCurrencyCodeLen = 3
	// DomainIDLength is the required length of a domain id
	DomainIDLength = 64
	// SHA512HalfLength is the length of a SHA-512 half hash (64 hex characters).
	SHA512HalfLength = 64
)

// *************************
// Validations
// *************************

// IsMemo checks if the given object is a valid Memo object.
func IsMemo(memo types.Memo) (bool, error) {
	// Get the size of the Memo object.
	size := len(maputils.GetKeys(memo.Flatten()))

	if size == 0 {
		return false, ErrMemoShouldHaveAtLeastOneField
	}

	validData := memo.MemoData == "" || typecheck.IsHex(memo.MemoData)
	if !validData {
		return false, ErrMemoDataShouldBeHex
	}

	validFormat := memo.MemoFormat == "" || typecheck.IsHex(memo.MemoFormat)
	if !validFormat {
		return false, ErrMemoFormatShouldBeHex
	}

	validType := memo.MemoType == "" || typecheck.IsHex(memo.MemoType)
	if !validType {
		return false, ErrMemoTypeShouldBeHex
	}

	return true, nil
}

// IsSigner checks if the given object is a valid Signer object.
func IsSigner(signerData types.SignerData) (bool, error) {
	size := len(maputils.GetKeys(signerData.Flatten()))
	if size != SignerSize {
		return false, ErrSignerShouldHaveThreeFields
	}

	validAccount := strings.TrimSpace(signerData.Account.String()) != "" && addresscodec.IsValidAddress(signerData.Account.String())
	if !validAccount {
		return false, ErrSignerAccountShouldBeString
	}

	if strings.TrimSpace(signerData.TxnSignature) == "" {
		return false, ErrSignerTxnSignatureShouldBeNonEmpty
	}

	if strings.TrimSpace(signerData.SigningPubKey) == "" {
		return false, ErrSignerSigningPubKeyShouldBeNonEmpty
	}

	return true, nil

}

// IsAmount checks if the given object is a valid Amount object.
// It is a string for an XRP amount or a map for an IssuedCurrency amount.
func IsAmount(field types.CurrencyAmount, fieldName string, isFieldRequired bool) (bool, error) {
	if isFieldRequired && field == nil {
		return false, ErrMissingField{
			Field: fieldName,
		}
	}

	if !isFieldRequired && field == nil {
		// no need to check further properties on a nil field, will create a panic with tests otherwise
		return true, nil
	}

	if field.Kind() == types.XRP {
		return true, nil
	}

	if ok, err := IsIssuedCurrency(field); !ok {
		return false, err
	}

	return true, nil
}

// IsIssuedCurrency checks if the given object is a valid IssuedCurrency object.
func IsIssuedCurrency(input types.CurrencyAmount) (bool, error) {
	if input.Kind() == types.XRP {
		return false, ErrInvalidTokenType
	}

	// Get the size of the IssuedCurrency object.
	issuedAmount, _ := input.(types.IssuedCurrencyAmount)

	numOfKeys := len(maputils.GetKeys(issuedAmount.Flatten().(map[string]interface{})))
	if numOfKeys != IssuedCurrencySize {
		return false, ErrInvalidTokenFields
	}

	if strings.TrimSpace(issuedAmount.Currency) == "" {
		return false, ErrMissingTokenCurrency
	}
	if strings.ToUpper(issuedAmount.Currency) == currency.NativeCurrencySymbol {
		return false, ErrInvalidTokenCurrency
	}

	if !addresscodec.IsValidAddress(issuedAmount.Issuer.String()) {
		return false, ErrInvalidIssuer
	}

	// Check if the value is a valid positive number
	value, err := strconv.ParseFloat(issuedAmount.Value, 64)
	if err != nil || value < 0 {
		return false, ErrInvalidTokenValue
	}

	return true, nil
}

// IsPath checks if the given pathstep is valid.
func IsPath(path []PathStep) (bool, error) {
	for _, pathStep := range path {

		hasAccount := pathStep.Account != ""
		hasCurrency := pathStep.Currency != ""
		hasIssuer := pathStep.Issuer != ""

		/**
		In summary, the following combination of fields are valid, optionally with type, type_hex, or both (but these two are deprecated):

		- account by itself
		- currency by itself
		- currency and issuer as long as the currency is not XRP
		- issuer by itself

		Any other use of account, currency, and issuer fields in a path step is invalid.

		https://xrpl.org/docs/concepts/tokens/fungible-tokens/paths#path-specifications
		*/
		switch {
		case hasAccount && !hasCurrency && !hasIssuer:
			return true, nil
		case hasCurrency && !hasAccount && !hasIssuer:
			return true, nil
		case hasIssuer && !hasAccount && !hasCurrency:
			return true, nil
		case hasIssuer && hasCurrency && pathStep.Currency != currency.NativeCurrencySymbol:
			return true, nil
		default:
			return false, ErrInvalidPathStepCombination
		}

	}
	return true, nil
}

// IsPaths checks if the given slice of slices of maps is a valid Paths.
func IsPaths(pathsteps [][]PathStep) (bool, error) {
	if len(pathsteps) == 0 {
		return false, ErrEmptyPath
	}

	for _, path := range pathsteps {
		if len(path) == 0 {
			return false, ErrEmptyPath
		}

		if ok, err := IsPath(path); !ok {
			return false, err
		}
	}

	return true, nil
}

// IsAsset checks if the given object is a valid Asset object.
func IsAsset(asset ledger.Asset) (bool, error) {
	// Get the size of the Asset object.
	lenKeys := len(maputils.GetKeys(asset.Flatten()))

	if lenKeys == 0 {
		return false, ErrInvalidAssetFields
	}

	if strings.TrimSpace(asset.Currency) == "" {
		return false, ErrMissingAssetCurrency
	}

	if strings.ToUpper(asset.Currency) == currency.NativeCurrencySymbol && strings.TrimSpace(asset.Issuer.String()) == "" {
		return true, nil
	}

	if strings.ToUpper(asset.Currency) == currency.NativeCurrencySymbol && asset.Issuer != "" {
		return false, ErrInvalidAssetIssuer
	}

	if asset.Currency != "" && !addresscodec.IsValidAddress(asset.Issuer.String()) {
		return false, ErrInvalidAssetIssuer
	}

	return true, nil
}

// IsDomainID checks if the given domain ID is valid.
func IsDomainID(id string) bool {
	return len(id) == DomainIDLength
}

// IsLedgerEntryID checks if the input is a valid ledger entry id.
// A valid ledger entry id is a 64-character hexadecimal string (SHA-512 half length).
func IsLedgerEntryID(input string) bool {
	return len(input) == SHA512HalfLength && typecheck.IsHex(input)
}

// ValidateHexMetadata validates input is non-empty hex string of up to a certain length.
// Returns true if the input is a valid non-empty hex string up to the specified length.
func ValidateHexMetadata(input string, maxLength int) bool {
	return len(input) > 0 && len(input) <= maxLength && typecheck.IsHex(input)
}

// IsTokenAmount checks if the given amount is a token amount (IssuedCurrencyAmount or MPTCurrencyAmount).
// Returns true if the amount is either an IssuedCurrencyAmount or MPTCurrencyAmount.
func IsTokenAmount(amount types.CurrencyAmount) bool {
	if amount == nil {
		return false
	}
	kind := amount.Kind()
	return kind == types.ISSUED || kind == types.MPT
}
