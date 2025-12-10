package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// MPTokenMetadataMaxSize defines the maximum allowed size in bytes for
	// multi-purpose token metadata when serialized to JSON.
	// https://github.com/XRPLF/XRPL-Standards/tree/master/XLS-0089d-multi-purpose-token-metadata-schema
	MPTokenMetadataMaxSize = 1024
)

// MPTokenMetadataURL represents a URL reference within multi-purpose token metadata.
// It follows the XLS-0089d standard for multi-purpose token metadata schema.
type MPTokenMetadataURL struct {
	URL   string `json:"url"`   // The URL address
	Type  string `json:"type"`  // The MIME type of the resource
	Title string `json:"title"` // Human-readable title for the URL
}

// MPTokenMetadata represents the metadata structure for multi-purpose tokens
// following the XLS-0089d standard. This structure defines the schema for
// token metadata that can be attached to XRPL tokens.
type MPTokenMetadata struct {
	Ticker         string               `json:"ticker"`                    // Short symbol for the token
	Name           string               `json:"name"`                      // Full name of the token
	Desc           string               `json:"desc,omitempty"`            // Description of the token
	Icon           string               `json:"icon,omitempty"`            // URL to token icon image
	AssetClass     string               `json:"asset_class"`               // Primary classification of the asset
	AssetSubclass  string               `json:"asset_subclass"`            // Secondary classification of the asset
	IssuerName     string               `json:"issuer_name"`               // Name of the token issuer
	URLs           []MPTokenMetadataURL `json:"urls,omitempty"`            // Additional URL references
	AdditionalInfo json.RawMessage      `json:"additional_info,omitempty"` // Custom additional metadata
}

// Validate checks if the MPTokenMetadata has all required fields populated.
// Required fields are: Ticker, AssetClass, AssetSubclass, and Name.
// Returns an error if any required field is empty or contains only whitespace.
func (m MPTokenMetadata) Validate() error {
	if strings.TrimSpace(m.Ticker) == "" {
		return ErrEmptyTicker
	}
	if strings.TrimSpace(m.AssetClass) == "" {
		return ErrEmptyAssetClass
	}
	if strings.TrimSpace(m.AssetSubclass) == "" {
		return ErrEmptyAssetSubclass
	}
	if strings.TrimSpace(m.Name) == "" {
		return ErrEmptyName
	}
	return nil
}

// MPTokenMetadataFromBlob parses a hex-encoded blob string into MPTokenMetadata.
// The blob should contain JSON data that conforms to the XLS-0089d standard
// for multi-purpose token metadata schema.
//
// Returns an error if the blob is not valid hex or if the JSON doesn't conform
// to the expected schema.
func MPTokenMetadataFromBlob(blob string) (*MPTokenMetadata, error) {
	b, err := hex.DecodeString(blob)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidHexBlob, err)
	}
	m := MPTokenMetadata{}

	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidSchema, err)
	}

	// Validate required fields
	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}

	return &m, nil
}

// Blob serializes the MPTokenMetadata to a hex-encoded string.
// The metadata is first marshaled to JSON and then encoded as a hex string.
//
// Returns an error if the JSON marshaling fails or if the resulting
// blob exceeds the maximum allowed size (MPTokenMetadataMaxSize).
func (m MPTokenMetadata) Blob() (string, error) {
	// Validate required fields before serialization
	if err := m.Validate(); err != nil {
		return "", fmt.Errorf("%w: %w", ErrValidationFailed, err)
	}

	json, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrMarshalFailed, err)
	}

	if len(json) > MPTokenMetadataMaxSize {
		return "", fmt.Errorf("%w: %d", ErrBlobTooLarge, len(json))
	}

	return hex.EncodeToString(json), nil
}
