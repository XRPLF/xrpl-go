//revive:disable:var-naming
package types

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// xchain bridge

	// ErrInvalidIssuingChainDoorAddress is returned when the issuing chain door address is invalid.
	ErrInvalidIssuingChainDoorAddress = errors.New("xchain bridge: invalid issuing chain door address")
	// ErrInvalidIssuingChainIssueAddress is returned when the issuing chain issue address is invalid.
	ErrInvalidIssuingChainIssueAddress = errors.New("xchain bridge: invalid issuing chain issue address")
	// ErrInvalidLockingChainDoorAddress is returned when the locking chain door address is invalid.
	ErrInvalidLockingChainDoorAddress = errors.New("xchain bridge: invalid locking chain door address")
	// ErrInvalidLockingChainIssueAddress is returned when the locking chain issue address is invalid.
	ErrInvalidLockingChainIssueAddress = errors.New("xchain bridge: invalid locking chain issue address")

	// raw tx

	// ErrBatchRawTransactionMissing is returned when the RawTransaction field is missing from an array element.
	ErrBatchRawTransactionMissing = errors.New("batch RawTransaction field is missing")
	// ErrBatchRawTransactionFieldNotObject is returned when the RawTransaction field is not an object.
	ErrBatchRawTransactionFieldNotObject = errors.New("batch RawTransaction field is not an object")
	// ErrBatchNestedTransaction is returned when trying to include a Batch transaction within another Batch.
	ErrBatchNestedTransaction = errors.New("batch cannot contain nested Batch transactions")
	// ErrBatchMissingInnerFlag is returned when an inner transaction lacks the TfInnerBatchTxn flag.
	ErrBatchMissingInnerFlag = errors.New("batch RawTransaction must contain the TfInnerBatchTxn flag")
	// ErrBatchInnerTransactionInvalid is returned when an inner transaction fails its own validation.
	ErrBatchInnerTransactionInvalid = errors.New("batch inner transaction validation failed")

	// permission

	// ErrInvalidPermissionValue is returned when PermissionValue is empty or undefined.
	ErrInvalidPermissionValue = errors.New("permission value cannot be empty or undefined")

	// batch signer

	// ErrBatchSignerAccountMissing is returned when a BatchSigner lacks the required Account field.
	ErrBatchSignerAccountMissing = errors.New("batch BatchSigner Account is missing")
	// ErrBatchSignerSigningPubKeyMissing is returned when a BatchSigner lacks the required SigningPubKey field.
	ErrBatchSignerSigningPubKeyMissing = errors.New("batch BatchSigner SigningPubKey is missing")
	// ErrBatchSignerInvalidTxnSignature is returned when a BatchSigner has an invalid TxnSignature field.
	ErrBatchSignerInvalidTxnSignature = errors.New("batch BatchSigner TxnSignature is invalid")

	// credential

	// ErrInvalidCredentialType is returned when the credential type is invalid; it must be a hexadecimal string between 1 and 64 bytes.
	ErrInvalidCredentialType = errors.New("invalid credential type, must be a hexadecimal string between 1 and 64 bytes")
	// ErrInvalidCredentialIssuer is returned when the credential Issuer field is missing.
	ErrInvalidCredentialIssuer = errors.New("credential type: missing field Issuer")

	// ErrEmptyCredentials is returned when the credential list is empty.
	ErrEmptyCredentials = errors.New("credentials list cannot be empty")
	// ErrInvalidCredentialCount is returned when the credential list size is out of allowed range.
	ErrInvalidCredentialCount = errors.New("accepted credentials list must contain at least one and no more than the maximum allowed number of items")
	// ErrDuplicateCredentials is returned when duplicate credentials are present in the list.
	ErrDuplicateCredentials = errors.New("credentials list cannot contain duplicate elements")

	// mptoken metadata

	// ErrInvalidMPTokenMetadataHex is returned when the MPTokenMetadata field is not a hex string.
	ErrInvalidMPTokenMetadataHex = errors.New("mptoken metadata should be a hex string")

	// ErrInvalidMPTokenMetadataJSON is returned when the MPTokenMetadata field is not a valid JSON object.
	ErrInvalidMPTokenMetadataJSON = errors.New("mptoken metadata should be a valid JSON object")

	// ErrInvalidMPTokenMetadataSize is returned when the MPTokenMetadata field is longer than 1024 bytes.
	ErrInvalidMPTokenMetadataSize = errors.New("mptoken metadata byte length should be at most 1024 bytes")

	// ErrInvalidMPTokenMetadataTicker is returned when the ticker does not match the required format (uppercase letters A-Z and digits 0-9, max 6 characters).
	ErrInvalidMPTokenMetadataTicker = errors.New("mptoken metadata ticker should contain only uppercase letters (A-Z) and digits (0-9), max 6 characters")

	// ErrInvalidMPTokenMetadataRWASubClassRequired is returned when the asset subclass is required when the asset class is rwa.
	ErrInvalidMPTokenMetadataRWASubClassRequired = errors.New("mptoken metadata asset subclass is required when asset class is rwa")

	// ErrInvalidMPTokenMetadataAdditionalInfo is returned when the additional info is not a string or a map.
	ErrInvalidMPTokenMetadataAdditionalInfo = errors.New("mptoken metadata additional info must be a string or a map")

	// ErrInvalidMPTokenMetadataURIs is returned when the URIs is not an array of objects each with uri/u, category/c, and title/t properties.
	ErrInvalidMPTokenMetadataURIs = errors.New("mptoken metadata URIs should be an array of objects each with uri/u, category/c, and title/t properties")
)

// ErrInvalidMPTokenMetadataUnknownField is returned when a field is unknown in MPToken metadata.
type ErrInvalidMPTokenMetadataUnknownField struct {
	Field string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataUnknownField
func (e ErrInvalidMPTokenMetadataUnknownField) Error() string {
	return fmt.Sprintf("mptoken metadata unknown field: %s", e.Field)
}

// ErrInvalidMPTokenMetadataFieldCount is returned when the MPToken metadata has an invalid field count.
type ErrInvalidMPTokenMetadataFieldCount struct {
	Count int
}

// Error implements the error interface for ErrMarshalPayload
func (e ErrInvalidMPTokenMetadataFieldCount) Error() string {
	return fmt.Sprintf("mptoken metadata field count should be at most %d", e.Count)
}

// ErrInvalidMPTokenMetadataFieldCollision is returned when both long and compact forms of a field are present in MPToken metadata.
type ErrInvalidMPTokenMetadataFieldCollision struct {
	Long    string
	Compact string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataFieldCollision
func (e ErrInvalidMPTokenMetadataFieldCollision) Error() string {
	return fmt.Sprintf("mptoken metadata field collision: %s and %s both present", e.Long, e.Compact)
}

// ErrInvalidMPTokenMetadataMissingField is returned when a required field is missing from MPToken metadata.
type ErrInvalidMPTokenMetadataMissingField struct {
	Field string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataMissingField
func (e ErrInvalidMPTokenMetadataMissingField) Error() string {
	return fmt.Sprintf("mptoken metadata field missing: %s", e.Field)
}

// ErrInvalidMPTokenMetadataInvalidString is returned when a string value is invalid.
type ErrInvalidMPTokenMetadataInvalidString struct {
	Key string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataInvalidString
func (e ErrInvalidMPTokenMetadataInvalidString) Error() string {
	return fmt.Sprintf("mptoken metadata field %s must be a valid string", e.Key)
}

// ErrInvalidMPTokenMetadataEmptyString is returned when a string value is empty.
type ErrInvalidMPTokenMetadataEmptyString struct {
	Key string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataEmptyString
func (e ErrInvalidMPTokenMetadataEmptyString) Error() string {
	return fmt.Sprintf("mptoken metadata field %s cannot be empty", e.Key)
}

// ErrInvalidMPTokenMetadataAssetClass is returned when the asset class is invalid.
type ErrInvalidMPTokenMetadataAssetClass struct {
	AssetClassSet [6]string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataAssetClass
func (e ErrInvalidMPTokenMetadataAssetClass) Error() string {
	return fmt.Sprintf("mptoken metadata asset class should be one of: %s", strings.Join(e.AssetClassSet[:], ", "))
}

// ErrInvalidMPTokenMetadataAssetSubClass is returned when the asset subclass is invalid.
type ErrInvalidMPTokenMetadataAssetSubClass struct {
	AssetSubclassSet []string
}

// Error implements the error interface for ErrInvalidMPTokenMetadataAssetSubClass
func (e ErrInvalidMPTokenMetadataAssetSubClass) Error() string {
	return fmt.Sprintf("mptoken metadata asset subclass should be one of: %s", strings.Join(e.AssetSubclassSet, ", "))
}

// MPTokenMetadataValidationErrors is a custom error type that holds a list of validation failures.
// It stores actual error objects to support wrapping/unwrapping.
type MPTokenMetadataValidationErrors []error

// Error implements the error interface.
func (v MPTokenMetadataValidationErrors) Error() string {
	var msgs []string
	for _, err := range v {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("mptoken metadata validation failed with %d errors:\n- %s", len(v), strings.Join(msgs, "\n- "))
}

// Unwrap returns the list of errors, allowing "errors.Is" support for lists.
// Example: errors.Is(err, types.ErrInvalidMPTokenMetadataTicker)
func (v MPTokenMetadataValidationErrors) Unwrap() []error {
	return v
}
