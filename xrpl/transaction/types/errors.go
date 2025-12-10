package types

import "errors"

// Predefined errors for MPTokenMetadata operations
var (
	// ErrEmptyTicker indicates that the ticker field is empty or contains only whitespace
	ErrEmptyTicker = errors.New("ticker is required and cannot be empty")

	// ErrEmptyAssetClass indicates that the asset_class field is empty or contains only whitespace
	ErrEmptyAssetClass = errors.New("asset_class is required and cannot be empty")

	// ErrEmptyAssetSubclass indicates that the asset_subclass field is empty or contains only whitespace
	ErrEmptyAssetSubclass = errors.New("asset_subclass is required and cannot be empty")

	// ErrEmptyName indicates that the name field is empty or contains only whitespace
	ErrEmptyName = errors.New("name is required and cannot be empty")

	// ErrInvalidHexBlob indicates that the provided blob string is not valid hex
	ErrInvalidHexBlob = errors.New("decode from blob in hex")

	// ErrInvalidSchema indicates that the JSON data doesn't conform to the XLS-0089d schema
	ErrInvalidSchema = errors.New("metadata is not in XLS-0089d schema")

	// ErrValidationFailed indicates that metadata validation failed
	ErrValidationFailed = errors.New("metadata validation failed")

	// ErrMarshalFailed indicates that JSON marshaling failed
	ErrMarshalFailed = errors.New("marshal to json for blob")

	// ErrBlobTooLarge indicates that the blob exceeds the maximum allowed size
	ErrBlobTooLarge = errors.New("blob is too large")
)
