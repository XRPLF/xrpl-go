package transaction

import (
	"fmt"
	"strconv"
	"strings"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	bctypes "github.com/Peersyst/xrpl-go/binary-codec/types"
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
	// Hex256Length is the number of characters in a 256-bit hexadecimal value.
	Hex256Length = 64
	// MPTIssuanceIDLength is the hex-encoded length of a 24-byte MPT issuance ID (48 hex chars).
	MPTIssuanceIDLength = 48
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
// It is a string for an XRP amount, a map for an IssuedCurrency amount, or an MPT amount.
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

	if field.Kind() == types.MPT {
		if ok, err := IsMPTCurrency(field); !ok {
			return false, err
		}
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

	numOfKeys := len(maputils.GetKeys(issuedAmount.Flatten().(map[string]any)))
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

	// Check that the value is an XRPL String Number (same gate the binary codec
	// applies at encode time), then reject negative amounts.
	// Zero is a valid token amount; "-0" parses as zero and is not treated as negative.
	isZero, err := bctypes.VerifyIOUValue(issuedAmount.Value)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInvalidTokenValue, err)
	}
	if strings.HasPrefix(issuedAmount.Value, "-") && !isZero {
		return false, ErrInvalidTokenValue
	}

	return true, nil
}

// IsMPTCurrency checks if the given object is a valid MPTCurrencyAmount object.
func IsMPTCurrency(input types.CurrencyAmount) (bool, error) {
	if input.Kind() != types.MPT {
		return false, ErrInvalidMPTType
	}

	mptAmount, _ := input.(types.MPTCurrencyAmount)

	if strings.TrimSpace(mptAmount.MPTIssuanceID) == "" {
		return false, ErrMissingMPTIssuanceID
	}

	if !typecheck.IsHex(mptAmount.MPTIssuanceID) {
		return false, ErrInvalidMPTIssuanceID
	}

	// Check if the value is a valid positive integer in the range 0 to 0x7FFFFFFFFFFFFFFF
	value, err := strconv.ParseInt(mptAmount.Value, 10, 64)
	if err != nil || value < 0 {
		return false, ErrInvalidMPTValue
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
	// MPT asset: only MPTIssuanceID should be set
	if asset.MPTIssuanceID != "" {
		if asset.Currency != "" || asset.Issuer != "" {
			return false, ErrInvalidMPTIssuanceIDAsset
		}
		if !typecheck.IsHex(asset.MPTIssuanceID) {
			return false, ErrInvalidMPTIssuanceIDAsset
		}
		return true, nil
	}

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
	return IsHex256(id)
}

// IsHex256 checks if the input is a 256-bit value encoded as hexadecimal.
func IsHex256(input string) bool {
	return len(input) == Hex256Length && typecheck.IsHex(input)
}

// IsLedgerEntryID checks if the input is a valid ledger entry id.
// A valid ledger entry id is a 256-bit value encoded as hexadecimal.
func IsLedgerEntryID(input string) bool {
	return IsHex256(input)
}

// IsMPTIssuanceID checks if the given hex string is a valid 24-byte MPT issuance ID (48 hex chars).
func IsMPTIssuanceID(id string) bool {
	return len(id) == MPTIssuanceIDLength && typecheck.IsHex(id)
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
