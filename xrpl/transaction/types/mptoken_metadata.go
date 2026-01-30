//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"encoding/json"
	"regexp"
	"slices"
	"strings"

	"github.com/Peersyst/xrpl-go/pkg/typecheck"
)

// MaxMPTokenMetadataByteLength is the maximum byte length for MPToken metadata (1024 bytes).
const (
	MaxMPTokenMetadataByteLength = 1024
	URIRequiredFieldCount        = 3
)

const (
	// Long MPTokenMetadata JSON Keys
	tickerLongKey         = "ticker"
	nameLongKey           = "name"
	descLongKey           = "desc"
	iconLongKey           = "icon"
	assetClassLongKey     = "asset_class"
	assetSubclassLongKey  = "asset_subclass"
	issuerNameLongKey     = "issuer_name"
	urisLongKey           = "uris"
	additionalInfoLongKey = "additional_info"

	// Compact MPTokenMetadata JSON Keys
	tickerCompactKey         = "t"
	nameCompactKey           = "n"
	descCompactKey           = "d"
	iconCompactKey           = "i"
	assetClassCompactKey     = "ac"
	assetSubclassCompactKey  = "as"
	issuerNameCompactKey     = "in"
	urisCompactKey           = "us"
	additionalInfoCompactKey = "ai"

	// Long MPTokenMetadataURI JSON Keys
	uriLongKey      = "uri"
	categoryLongKey = "category"
	titleLongKey    = "title"

	// Compact MPTokenMetadataURI JSON Keys
	uriCompactKey      = "u"
	categoryCompactKey = "c"
	titleCompactKey    = "t"
)

// Uppercase letters (A-Z) and digits (0-9) only. Max 6 chars.
var tickerRegex = regexp.MustCompile(`^[A-Z0-9]{1,6}$`)

var (
	// MPTokenMetadataAssetClasses contains the allowed values for the asset class field.
	MPTokenMetadataAssetClasses = [6]string{"rwa", "memes", "wrapped", "gaming", "defi", "other"}
	// MPTokenMetadataAssetSubClasses contains the allowed values for the asset subclass field.
	MPTokenMetadataAssetSubClasses = [7]string{"stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"}
	// MPTokenMetadataURICategories contains the allowed values for the URI category field.
	MPTokenMetadataURICategories = [4]string{"website", "social", "docs", "other"}
)

// fieldDef defines the structure for metadata field definitions.
type fieldDef struct {
	long     string
	compact  string
	required bool
	validate func(meta map[string]any, assetClass string) error
}

// uriFieldDef defines the structure for URI field definitions.
type uriFieldDef struct {
	long    string
	compact string
}

// mptMetadataURIFields defines the URI fields with their long and compact forms.
var mptMetadataURIFields = []uriFieldDef{
	{long: uriLongKey, compact: uriCompactKey},
	{long: categoryLongKey, compact: categoryCompactKey},
	{long: titleLongKey, compact: titleCompactKey},
}

// mptMetadataFields defines all MPToken metadata fields in a table format.
// Each field has a long form, compact form, required flag, and validation function.
var mptMetadataFields = []fieldDef{
	{
		long:     tickerLongKey,
		compact:  tickerCompactKey,
		required: true,
		validate: func(meta map[string]any, _ string) error {
			v, exists, err := getStringField(meta, tickerLongKey, tickerCompactKey)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataMissingField{Field: tickerLongKey}
			}
			if !tickerRegex.MatchString(v) {
				return ErrInvalidMPTokenMetadataTicker
			}
			return nil
		},
	},
	{
		long:     nameLongKey,
		compact:  nameCompactKey,
		required: true,
		validate: func(meta map[string]any, _ string) error {
			_, exists, err := getStringField(meta, nameLongKey, nameCompactKey)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataMissingField{Field: nameLongKey}
			}
			return nil
		},
	},
	{
		long:     descLongKey,
		compact:  descCompactKey,
		required: false,
		validate: func(meta map[string]any, _ string) error {
			_, _, err := getStringField(meta, descLongKey, descCompactKey)
			return err
		},
	},
	{
		long:     iconLongKey,
		compact:  iconCompactKey,
		required: true,
		validate: func(meta map[string]any, _ string) error {
			_, exists, err := getStringField(meta, iconLongKey, iconCompactKey)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataMissingField{Field: iconLongKey}
			}
			return nil
		},
	},
	{
		long:     assetClassLongKey,
		compact:  assetClassCompactKey,
		required: true,
		validate: func(meta map[string]any, _ string) error {
			v, exists, err := getStringField(meta, assetClassLongKey, assetClassCompactKey)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataMissingField{Field: assetClassLongKey}
			}
			if !slices.Contains(MPTokenMetadataAssetClasses[:], v) {
				return ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses}
			}
			return nil
		},
	},
	{
		long:     assetSubclassLongKey,
		compact:  assetSubclassCompactKey,
		required: false,
		validate: func(meta map[string]any, assetClass string) error {
			v, exists, err := getStringField(meta, assetSubclassLongKey, assetSubclassCompactKey)
			if err != nil {
				return err
			}
			if assetClass == "rwa" && !exists {
				return ErrInvalidMPTokenMetadataRWASubClassRequired
			}
			if exists && !slices.Contains(MPTokenMetadataAssetSubClasses[:], v) {
				return ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses[:]}
			}
			return nil
		},
	},
	{
		long:     issuerNameLongKey,
		compact:  issuerNameCompactKey,
		required: true,
		validate: func(meta map[string]any, _ string) error {
			_, exists, err := getStringField(meta, issuerNameLongKey, issuerNameCompactKey)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataMissingField{Field: issuerNameLongKey}
			}
			return nil
		},
	},
	{
		long:     urisLongKey,
		compact:  urisCompactKey,
		required: false,
		validate: func(meta map[string]any, _ string) error {
			return validateURIs(meta)
		},
	},
	{
		long:     additionalInfoLongKey,
		compact:  additionalInfoCompactKey,
		required: false,
		validate: func(meta map[string]any, _ string) error {
			val, exists, err := getField(meta, additionalInfoLongKey, additionalInfoCompactKey)
			if err != nil {
				return err
			}
			if exists && !typecheck.IsString(val) && !typecheck.IsMap(val) {
				return ErrInvalidMPTokenMetadataAdditionalInfo
			}
			return nil
		},
	},
}

// buildValidKeySet creates a set of all valid keys (both long and compact).
func buildValidKeySet() map[string]bool {
	validKeys := make(map[string]bool, len(mptMetadataFields)*2)
	for _, f := range mptMetadataFields {
		validKeys[f.long] = true
		validKeys[f.compact] = true
	}
	return validKeys
}

// ParsedMPTokenMetadata represents the MPToken metadata defined as per the XLS-89 standard.
// Fields are ordered alphabetically by JSON key for consistent encoding.
type ParsedMPTokenMetadata struct {
	// Top-level classification of token purpose.
	// Allowed values: "rwa", "memes", "wrapped", "gaming", "defi", "other"
	// Example: "rwa"
	AssetClass string `json:"ac"`
	// Freeform field for key token details like interest rate, maturity date, term, or other relevant info.
	// Can be any valid JSON object or UTF-8 string.
	// Example: { "interest_rate": "5.00%", "maturity_date": "2045-06-30" }
	AdditionalInfo any `json:"ai,omitempty"`
	// Optional subcategory of the asset class.
	// Required if AssetClass is "rwa".
	// Allowed values: "stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"
	// Example: "treasury"
	AssetSubclass *string `json:"as,omitempty"`
	// Short description of the token.
	// Any UTF-8 string.
	// Example: "A sample token used for demonstration"
	Desc *string `json:"d,omitempty"`
	// URI to the token icon.
	// Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://).
	// Example: example.org/token-icon, ipfs://token-icon.png
	Icon string `json:"i"`
	// The name of the issuer account.
	// Any UTF-8 string.
	// Example: "Example Issuer"
	IssuerName string `json:"in"`
	// Display name of the token.
	// Any UTF-8 string.
	// Example: "Example Token"
	Name string `json:"n"`
	// Ticker symbol used to represent the token.
	// Uppercase letters (A-Z) and digits (0-9) only. Max 6 chars.
	// Example: "EXMPL"
	Ticker string `json:"t"`
	// List of related URIs (site, dashboard, social media, documentation, etc.).
	// Each URI object contains the link, its category, and a human-readable title.
	URIs []ParsedMPTokenMetadataURI `json:"us,omitempty"`
}

// ParsedMPTokenMetadataURI represents a URI entry within MPTokenMetadata as per XLS-89 standard.
// Fields are ordered alphabetically by JSON key for consistent encoding.
type ParsedMPTokenMetadataURI struct {
	// The category of the link.
	// Allowed values: "website", "social", "docs", "other"
	// Example: "website"
	Category string `json:"c"`
	// A human-readable label for the link.
	// Any UTF-8 string.
	// Example: "Product Page"
	Title string `json:"t"`
	// URI to the related resource.
	// Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://).
	// Example: "exampleyield.com/tbill" or "ipfs://QmXxxx"
	URI string `json:"u"`
}

// UnmarshalJSON implements custom JSON unmarshaling for ParsedMPTokenMetadataURI.
// It handles both long-form and compact-form keys.
func (u *ParsedMPTokenMetadataURI) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, f := range mptMetadataURIFields {
		if v, ok := raw[f.compact]; ok {
			switch f.long {
			case "uri":
				u.URI = v
			case "category":
				u.Category = v
			case "title":
				u.Title = v
			}
		} else if v, ok := raw[f.long]; ok {
			switch f.long {
			case "uri":
				u.URI = v
			case "category":
				u.Category = v
			case "title":
				u.Title = v
			}
		}
	}

	return nil
}

// MPTokenMetadata returns a pointer to a string containing metadata for an MPToken.
func MPTokenMetadata(value string) *string {
	return &value
}

func getValue(raw map[string]json.RawMessage, compact, long string) json.RawMessage {
	if v, ok := raw[compact]; ok {
		return v
	}
	return raw[long]
}

// UnmarshalJSON implements custom JSON unmarshaling for ParsedMPTokenMetadata.
// It handles both long-form and compact-form keys.
func (m *ParsedMPTokenMetadata) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, f := range mptMetadataFields {
		v := getValue(raw, f.compact, f.long)
		if v == nil {
			continue
		}

		var err error
		switch f.long {
		case "ticker":
			err = json.Unmarshal(v, &m.Ticker)
		case "name":
			err = json.Unmarshal(v, &m.Name)
		case "desc":
			err = json.Unmarshal(v, &m.Desc)
		case "icon":
			err = json.Unmarshal(v, &m.Icon)
		case "asset_class":
			err = json.Unmarshal(v, &m.AssetClass)
		case "asset_subclass":
			err = json.Unmarshal(v, &m.AssetSubclass)
		case "issuer_name":
			err = json.Unmarshal(v, &m.IssuerName)
		case "uris":
			err = json.Unmarshal(v, &m.URIs)
		case "additional_info":
			err = json.Unmarshal(v, &m.AdditionalInfo)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// EncodeMPTokenMetadata encodes the ParsedMPTokenMetadata struct into a hex string compliant with XLS-89.
// Returns the encoded hex string and an error if encoding fails.
func EncodeMPTokenMetadata(meta ParsedMPTokenMetadata) (string, error) {
	// When Marshaling all the keys are sorted alphabetically by default
	bytes, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(bytes)), nil
}

// DecodeMPTokenMetadata decodes a hex string into a ParsedMPTokenMetadata struct.
// It handles input with either long or compact keys via custom UnmarshalJSON methods.
// Returns a ParsedMPTokenMetadata and an error if decoding fails.
func DecodeMPTokenMetadata(hexInput string) (ParsedMPTokenMetadata, error) {
	bytes, err := hex.DecodeString(hexInput)
	if err != nil {
		return ParsedMPTokenMetadata{}, ErrInvalidMPTokenMetadataHex
	}

	var result ParsedMPTokenMetadata
	if err := json.Unmarshal(bytes, &result); err != nil {
		return ParsedMPTokenMetadata{}, ErrInvalidMPTokenMetadataJSON
	}

	return result, nil
}

// ValidateMPTokenMetadata validates MPToken metadata according to XLS-89 standard.
func ValidateMPTokenMetadata(input string) error {
	// Should be a valid hex string
	bytes, err := hex.DecodeString(input)
	if err != nil {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataHex})
	}

	// By the XLS-89 standard, the metadata should have a max length
	if len(bytes) > MaxMPTokenMetadataByteLength {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataSize})
	}

	var rawData map[string]any
	if err := json.Unmarshal(bytes, &rawData); err != nil {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataJSON})
	}

	var errs []error

	// Check field count
	if len(rawData) > len(mptMetadataFields) {
		errs = append(errs, ErrInvalidMPTokenMetadataFieldCount{Count: len(mptMetadataFields)})
	}

	// Validate all keys are known
	validKeys := buildValidKeySet()
	for key := range rawData {
		if !validKeys[key] {
			errs = append(errs, ErrInvalidMPTokenMetadataUnknownField{Field: key})
		}
	}

	// Get asset class for conditional validation
	assetClass, _, _ := getStringField(rawData, "asset_class", "ac")

	// Run all field validations from the table
	for _, f := range mptMetadataFields {
		if err := f.validate(rawData, assetClass); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return MPTokenMetadataValidationErrors(errs)
	}

	return nil
}

// validateURIs validates the uris field.
func validateURIs(meta map[string]any) error {
	val, exists, err := getField(meta, "uris", "us")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	urisList, ok := val.([]any)
	if !ok || len(urisList) == 0 {
		return ErrInvalidMPTokenMetadataURIs
	}

	for _, item := range urisList {
		uriObject, ok := item.(map[string]any)
		if !ok || len(uriObject) != URIRequiredFieldCount {
			return ErrInvalidMPTokenMetadataURIs
		}

		// Validate each URI field from the table
		for _, f := range mptMetadataURIFields {
			val, exists, err := getStringField(uriObject, f.long, f.compact)
			if err != nil {
				return err
			}
			if !exists {
				return ErrInvalidMPTokenMetadataURIs
			}
			// Validate category against allowed values
			if f.long == "category" && !slices.Contains(MPTokenMetadataURICategories[:], val) {
				return ErrInvalidMPTokenMetadataURIs
			}
		}
	}
	return nil
}

// getField retrieves a value from a map using either the long-form or compact-form key.
// If both keys are present, it returns an error.
// If only one key is present, it returns the value.
// If no key is present, it returns an empty string and false.
func getField(meta map[string]any, longKey, compactKey string) (any, bool, error) {
	compactValue, compactExists := meta[compactKey]
	longValue, longExists := meta[longKey]

	if longExists && compactExists {
		return "", false, ErrInvalidMPTokenMetadataFieldCollision{Long: longKey, Compact: compactKey}
	}

	if compactExists {
		return compactValue, true, nil
	}

	if longExists {
		return longValue, true, nil
	}

	return nil, false, nil
}

// getStringField retrieves a string value from a map using either the long-form or compact-form key.
// If both keys are present, it returns an error.
// If only one key is present, it returns the value.
// If no key is present, it returns an empty string and false.
func getStringField(meta map[string]any, longKey, compactKey string) (string, bool, error) {
	value, exists, err := getField(meta, longKey, compactKey)
	if err != nil {
		return "", false, err
	}
	if !exists {
		return "", false, nil
	}

	val, ok := value.(string)

	if !ok {
		return "", true, ErrInvalidMPTokenMetadataInvalidString{Key: longKey}
	}

	if val == "" {
		return "", true, ErrInvalidMPTokenMetadataEmptyString{Key: longKey}
	}

	return val, true, nil
}
