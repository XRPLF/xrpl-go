//revive:disable:var-naming
package types

import (
	"encoding/hex"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
)

// MaxMPTokenMetadataByteLength is the maximum byte length for MPToken metadata (1024 bytes).
const (
	MaxMPTokenMetadataByteLength = 1024
	URIRequiredFieldCount        = 3
)

// Uppercase letters (A-Z) and digits (0-9) only. Max 6 chars.
var tickerRegex = regexp.MustCompile(`^[A-Z0-9]{1,6}$`)

var (
	// MPTokenMetadataAssetClasses contains the allowed values for the asset class field.
	MPTokenMetadataAssetClasses = []string{"rwa", "memes", "wrapped", "gaming", "defi", "other"}
	// MPTokenMetadataAssetSubClasses contains the allowed values for the asset subclass field.
	MPTokenMetadataAssetSubClasses = []string{"stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"}
	// MPTokenMetadataURICategories contains the allowed values for the URI category field.
	MPTokenMetadataURICategories = []string{"website", "social", "docs", "other"}
)

// Field Mappings (Long <-> Compact)
var mptTokenMetadataFieldMap = map[string]string{
	"ticker":          "t",
	"name":            "n",
	"desc":            "d",
	"icon":            "i",
	"asset_class":     "ac",
	"asset_subclass":  "as",
	"issuer_name":     "in",
	"uris":            "us",
	"additional_info": "ai",
}

// Field Mappings (Long <-> Compact)
var mptTokenMetadataURIFieldMap = map[string]string{
	"uri":      "u",
	"category": "c",
	"title":    "t",
}

// ParsedMPTokenMetadataURI represents a URI entry within MPTokenMetadata as per XLS-89 standard.
type ParsedMPTokenMetadataURI struct {
	// URI to the related resource.
	// Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://).
	// Example: "exampleyield.com/tbill" or "ipfs://QmXxxx"
	URI string `json:"u"`
	// The category of the link.
	// Allowed values: "website", "social", "docs", "other"
	// Example: "website"
	Category string `json:"c"`
	// A human-readable label for the link.
	// Any UTF-8 string.
	// Example: "Product Page"
	Title string `json:"t"`
}

// ParsedMPTokenMetadata represents the MPToken metadata defined as per the XLS-89 standard.
type ParsedMPTokenMetadata struct {
	// Ticker symbol used to represent the token.
	// Uppercase letters (A-Z) and digits (0-9) only. Max 6 chars.
	// Example: "EXMPL"
	Ticker string `json:"t"`
	// Display name of the token.
	// Any UTF-8 string.
	// Example: "Example Token"
	Name string `json:"n"`
	// Short description of the token.
	// Any UTF-8 string.
	// Example: "A sample token used for demonstration"
	Desc *string `json:"d,omitempty"`
	// URI to the token icon.
	// Can be a hostname/path (HTTPS assumed) or full URI for other protocols (e.g., ipfs://).
	// Example: example.org/token-icon, ipfs://token-icon.png
	Icon string `json:"i"`
	// Top-level classification of token purpose.
	// Allowed values: "rwa", "memes", "wrapped", "gaming", "defi", "other"
	// Example: "rwa"
	AssetClass string `json:"ac"`
	// Optional subcategory of the asset class.
	// Required if AssetClass is "rwa".
	// Allowed values: "stablecoin", "commodity", "real_estate", "private_credit", "equity", "treasury", "other"
	// Example: "treasury"
	AssetSubclass *string `json:"as,omitempty"`
	// The name of the issuer account.
	// Any UTF-8 string.
	// Example: "Example Issuer"
	IssuerName string `json:"in"`
	// List of related URIs (site, dashboard, social media, documentation, etc.).
	// Each URI object contains the link, its category, and a human-readable title.
	URIs []ParsedMPTokenMetadataURI `json:"us,omitempty"`
	// Freeform field for key token details like interest rate, maturity date, term, or other relevant info.
	// Can be any valid JSON object or UTF-8 string.
	// Example: { "interest_rate": "5.00%", "maturity_date": "2045-06-30" }
	AdditionalInfo any `json:"ai,omitempty"`
}

// MPTokenMetadata returns a pointer to a string containing metadata for an MPToken.
func MPTokenMetadata(value string) *string {
	return &value
}

// EncodeMPTokenMetadata encodes the ParsedMPTokenMetadata struct into a hex string compliant with XLS-89.
// It ensures keys are compact and fields are sorted alphabetically.
// Returns the encoded hex string and an error if encoding fails.
func EncodeMPTokenMetadata(meta ParsedMPTokenMetadata) (string, error) {
	// 1. Marshal struct to JSON (this applies the Compact tags `json:"t"` etc)
	bytes, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}

	// 2. Unmarshal into map[string]any to ensure we can sort keys and handle nested cleaning if necessary
	var asMap map[string]any
	if err := json.Unmarshal(bytes, &asMap); err != nil {
		return "", err
	}

	// 3. Ensure keys are strictly compact (though struct tags should have handled this, this is a safety net)
	compactMap := compactKeys(asMap, mptTokenMetadataFieldMap)

	// 4. Handle nested URIs shortening
	if uris, ok := compactMap["us"].([]any); ok {
		var newURIs []any
		for _, uri := range uris {
			// If the URI is a map, we need to compact the keys
			if uriMap, ok := uri.(map[string]any); ok {
				newURIs = append(newURIs, compactKeys(uriMap, mptTokenMetadataURIFieldMap))
			} else {
				// If the URI is not a map, we keep it as is
				newURIs = append(newURIs, uri)
			}
		}
		compactMap["us"] = newURIs
	}

	// 5. Marshal map back to JSON. json.Marshal sorts map keys lexicographically (as an implementation detail),
	// producing deterministic output, which is required by the XLS-89 standard.
	finalJSON, err := json.Marshal(compactMap)
	if err != nil {
		return "", err
	}

	return strings.ToUpper(hex.EncodeToString(finalJSON)), nil
}

// DecodeMPTokenMetadata decodes a hex string into a ParsedMPTokenMetadata struct.
// It handles input with either long or compact keys by normalizing them to compact form before mapping.
// Returns a pointer to ParsedMPTokenMetadata and an error if decoding fails.
func DecodeMPTokenMetadata(hexInput string) (*ParsedMPTokenMetadata, error) {
	bytes, err := hex.DecodeString(hexInput)
	if err != nil {
		return nil, ErrInvalidMPTokenMetadataHex
	}

	var rawMap map[string]any
	if err := json.Unmarshal(bytes, &rawMap); err != nil {
		return nil, ErrInvalidMPTokenMetadataJSON
	}

	// Normalize keys to Compact form so they match the Struct tags
	compactMap := compactKeys(rawMap, mptTokenMetadataFieldMap)

	// Handle nested URIs
	if uris, ok := compactMap["us"].([]any); ok {
		var newURIs []any
		for _, u := range uris {
			if uMap, ok := u.(map[string]any); ok {
				newURIs = append(newURIs, compactKeys(uMap, mptTokenMetadataURIFieldMap))
			} else {
				newURIs = append(newURIs, u) // Keep as is if not a map (will fail validation later)
			}
		}
		compactMap["us"] = newURIs
	}

	// Marshal cleaned map back to bytes, then unmarshal into Struct
	cleanBytes, err := json.Marshal(compactMap)
	if err != nil {
		return nil, err
	}

	var result ParsedMPTokenMetadata
	if err := json.Unmarshal(cleanBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ValidateMPTokenMetadata validates if the hex string adheres to all XLS-89 rules.
// Returns MPTokenMetadataValidationErrors if the hex string is invalid, or nil if valid.
func ValidateMPTokenMetadata(hexInput string) error {
	// 1. Validate Hex
	bytes, err := hex.DecodeString(hexInput)
	if err != nil {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataHex})
	}

	// 2. Validate Byte Length
	if len(bytes) > MaxMPTokenMetadataByteLength {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataSize})
	}

	// 3. Parse JSON
	var rawData map[string]any
	if err := json.Unmarshal(bytes, &rawData); err != nil {
		return MPTokenMetadataValidationErrors([]error{ErrInvalidMPTokenMetadataJSON})
	}

	// This var will be used to concatenate all errors
	var errs []error

	// 4. Validate Top Level Fields count
	if len(rawData) > len(mptTokenMetadataFieldMap) {
		errs = append(errs, ErrInvalidMPTokenMetadataFieldCount{Count: len(mptTokenMetadataFieldMap)})
	}

	// + validate all keys are known (either long or compact form)
	validKeys := make(map[string]bool)
	for long, compact := range mptTokenMetadataFieldMap {
		validKeys[long] = true
		validKeys[compact] = true
	}
	for key := range rawData {
		if !validKeys[key] {
			errs = append(errs, ErrInvalidMPTokenMetadataUnknownField{Field: key})
		}
	}

	// 5. Field Validations
	errs = append(errs, validateField(rawData, "ticker", "t", validateTicker)...)
	errs = append(errs, validateField(rawData, "name", "n", validateNonEmptyString("name"))...)
	errs = append(errs, validateField(rawData, "icon", "i", validateNonEmptyString("icon"))...)
	errs = append(errs, validateField(rawData, "issuer_name", "in", validateNonEmptyString("issuer_name"))...)
	errs = append(errs, validateField(rawData, "asset_class", "ac", validateAssetClass)...)
	errs = append(errs, validateField(rawData, "desc", "d", validateOptionalString("desc"))...)
	errs = append(errs, validateSubClass(rawData)...)
	errs = append(errs, validateURIs(rawData)...)
	errs = append(errs, validateAdditionalInfo(rawData)...)

	finalErrs := filterErrors(errs)

	if len(finalErrs) > 0 {
		return MPTokenMetadataValidationErrors(finalErrs)
	}

	return nil
}

// compactKeys replaces long-form map keys with their compact aliases.
// If both a long key and its compact form are present in the input,
// the long key is preserved to avoid overwriting and to surface collisions.
// Returns a new map with compact keys.
func compactKeys(input map[string]any, mapping map[string]string) map[string]any {
	result := make(map[string]any)

	for inputKey, value := range input {
		// Check whether the key has a compact alias
		if compactAlias, hasAlias := mapping[inputKey]; hasAlias {
			// If the compact alias already exists in the input, keep the long key
			// to avoid overwriting and allow collision detection.
			if _, aliasPresent := input[compactAlias]; aliasPresent {
				result[inputKey] = value
			} else {
				// Otherwise, replace the long key with its compact alias.
				result[compactAlias] = value
			}
		} else {
			// Keys without a compact alias are copied as-is.
			result[inputKey] = value
		}
	}

	return result
}

// fieldValidator is a function type that validates a field value.
// Returns an error if the value is invalid, or nil if valid.
type fieldValidator func(val any) error

// validateField validates a field in the metadata object, checking for collisions between long and compact keys.
// Parameters:
//   - obj: The metadata object to validate
//   - longKey: The long-form key name (e.g., "ticker")
//   - compactKey: The compact-form key name (e.g., "t")
//   - validator: The validation function to apply to the field value
//
// Returns a slice of errors if validation fails, or nil if the field is valid or optional and missing.
func validateField(obj map[string]any, longKey, compactKey string, validator fieldValidator) []error {
	longValue, hasLong := obj[longKey]
	compactValue, hasCompact := obj[compactKey]

	if hasLong && hasCompact {
		return []error{ErrInvalidMPTokenMetadataFieldCollision{Long: longKey, Compact: compactKey}}
	}

	var val any
	switch {
	case hasLong:
		val = longValue
	case hasCompact:
		val = compactValue
	default:
		switch longKey {
		// Optional fields are fine if they are missing
		case "desc", "asset_subclass", "uris", "additional_info":
			return nil
		}
		return []error{ErrInvalidMPTokenMetadataMissingField{Field: longKey}}
	}

	err := validator(val)
	if err != nil {
		return []error{err}
	}
	return nil
}

// validateTicker validates the ticker string.
// The ticker must be a non-empty string containing only uppercase letters (A-Z) and digits (0-9), with a maximum length of 6 characters.
// Returns an error if validation fails, or nil if valid.
func validateTicker(v any) error {
	s, ok := v.(string)
	if !ok || s == "" {
		return ErrInvalidMPTokenMetadataTicker
	}
	if !tickerRegex.MatchString(s) {
		return ErrInvalidMPTokenMetadataTicker
	}
	return nil
}

// validateNonEmptyString returns a validator function that checks if a field is a non-empty string.
func validateNonEmptyString(fieldName string) fieldValidator {
	return func(v any) error {
		s, ok := v.(string)
		if !ok || len(s) == 0 {
			return ErrInvalidMPTokenMetadataMissingField{Field: fieldName}
		}
		return nil
	}
}

// validateOptionalString returns a validator function for an optional string field.
// The field is valid if it is nil (missing) or if it is a valid non-empty string.
func validateOptionalString(fieldName string) fieldValidator {
	return func(v any) error {
		// If not present, it's optional and valid
		if v == nil {
			return nil
		}
		// If present, it must be a valid non-empty string
		return validateNonEmptyString(fieldName)(v)
	}
}

// validateAssetClass validates the asset class field value.
// The value must be a string from the allowed asset class set.
// Returns an error if validation fails, or nil if valid.
func validateAssetClass(v any) error {
	s, ok := v.(string)
	if !ok {
		return ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses}
	}
	// If the asset class is a known value, return nil
	isKnownValue := slices.Contains(MPTokenMetadataAssetClasses, s)
	if !isKnownValue {
		return ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses}
	}
	return nil
}

// validateSubClass validates the asset subclass field according to XLS-89 rules.
// It checks for field collisions between long and compact forms, ensures asset_subclass is present
// when asset_class is "rwa", and validates that the subclass value is from the allowed list.
// Returns a slice of errors if validation fails, or nil if valid.
func validateSubClass(obj map[string]any) []error {
	assetClassVal, assetClassExists := lookupField(obj, "asset_class", "ac")
	assetSubclassVal, assetSubclassExists := lookupField(obj, "asset_subclass", "as")

	// Check duplicates
	if areBothFieldsPresent(obj, "asset_subclass", "as") {
		return []error{ErrInvalidMPTokenMetadataFieldCollision{Long: "asset_subclass", Compact: "as"}}
	}

	// If asset_class == rwa, asset_subclass is required
	if assetClassExists {
		if assetClassStr, ok := assetClassVal.(string); ok && assetClassStr == "rwa" {
			if !assetSubclassExists {
				return []error{ErrInvalidMPTokenMetadataRWASubClassRequired}
			}
		}
	}

	// If asset_subclass is present, it must be a valid value from the allowed list
	if assetSubclassExists {
		s, ok := assetSubclassVal.(string)
		if !ok || len(s) == 0 || !slices.Contains(MPTokenMetadataAssetSubClasses, s) {
			return []error{ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses}}
		}
	}
	return nil
}

// validateURIs validates the URIs array field in the metadata object.
// It checks for field collisions, ensures each URI object has exactly 3 fields (uri, category, title),
// validates that all fields are strings, and verifies that category values are from the allowed list.
// Returns a slice of errors if validation fails, or nil if the field is missing (optional) or valid.
func validateURIs(obj map[string]any) []error {
	val, exists := lookupField(obj, "uris", "us")
	if !exists {
		return nil
	}

	// Check duplicates
	if areBothFieldsPresent(obj, "uris", "us") {
		return []error{ErrInvalidMPTokenMetadataFieldCollision{Long: "uris", Compact: "us"}}
	}

	list, ok := val.([]any)
	if !ok || len(list) == 0 {
		return []error{ErrInvalidMPTokenMetadataURIs}
	}

	var errs []error

	for _, item := range list {
		uriObject, ok := item.(map[string]any)
		if !ok || len(uriObject) != URIRequiredFieldCount {
			errs = append(errs, ErrInvalidMPTokenMetadataURIs)
			continue
		}

		for long, compact := range mptTokenMetadataURIFieldMap {
			if areBothFieldsPresent(uriObject, long, compact) {
				errs = append(errs, ErrInvalidMPTokenMetadataFieldCollision{Long: long, Compact: compact})
			}
		}

		// uri is only validated if it is a string not a hostname/path validation
		uri, uriExists := lookupField(uriObject, "uri", "u")
		category, categoryExists := lookupField(uriObject, "category", "c")
		title, titleExists := lookupField(uriObject, "title", "t")

		if !uriExists || !isString(uri) || !categoryExists || !isString(category) || !titleExists || !isString(title) {
			errs = append(errs, ErrInvalidMPTokenMetadataURIs)
			continue
		}

		if categoryStr, ok := category.(string); ok {
			validCat := slices.Contains(MPTokenMetadataURICategories, categoryStr)
			if !validCat {
				errs = append(errs, ErrInvalidMPTokenMetadataURIs)
				continue
			}
		}
	}
	return errs
}

// validateAdditionalInfo validates the additional_info field in the metadata object.
// It checks for field collisions between long and compact forms, and ensures the value is either a string or a map.
// Returns a slice of errors if validation fails, or nil if the field is missing (optional) or valid.
func validateAdditionalInfo(obj map[string]any) []error {
	val, exists := lookupField(obj, "additional_info", "ai")
	if !exists {
		return nil
	}

	// Check duplicates
	if areBothFieldsPresent(obj, "additional_info", "ai") {
		return []error{ErrInvalidMPTokenMetadataFieldCollision{Long: "additional_info", Compact: "ai"}}
	}

	// additional_info must be a string or a map
	if !isString(val) && !isMap(val) {
		return []error{ErrInvalidMPTokenMetadataAdditionalInfo}
	}
	return nil
}

// areBothFieldsPresent checks if both the long and compact forms of a field are present in the map.
// Returns true if both fields are present, false otherwise.
func areBothFieldsPresent(obj map[string]any, long, compact string) bool {
	_, hasLong := obj[long]
	_, hasCompact := obj[compact]
	return hasLong && hasCompact
}

// lookupField retrieves a value from a map using either the long-form or compact-form key.
// It first checks for the long-form key, then falls back to the compact-form key.
func lookupField(m map[string]any, long, compact string) (any, bool) {
	if v, ok := m[long]; ok {
		return v, true
	}
	if v, ok := m[compact]; ok {
		return v, true
	}
	return nil, false
}

// isString checks if a value is a string type.
func isString(v any) bool {
	_, ok := v.(string)
	return ok
}

// isMap checks if a value is a map type.
func isMap(v any) bool {
	_, ok := v.(map[string]any)
	return ok
}

// filterErrors removes nil errors from a slice of errors.
func filterErrors(errs []error) []error {
	var filtered []error
	for _, e := range errs {
		if e != nil {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
