package types

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to convert JSON data to hex string
func toHexString(t *testing.T, data any) string {
	var jsonBytes []byte
	var err error

	if str, ok := data.(string); ok {
		jsonBytes = []byte(str)
	} else {
		jsonBytes, err = json.Marshal(data)
		require.NoError(t, err)
	}

	return strings.ToUpper(hex.EncodeToString(jsonBytes))
}

// Helper function to extract errors from validation errors
func extractValidationErrors(err error) []error {
	if err == nil {
		return []error{}
	}

	if validationErrs, ok := err.(MPTokenMetadataValidationErrors); ok {
		return []error(validationErrs)
	}

	return []error{err}
}

// convertToCompactKeys converts long field names to compact equivalents
func convertToCompactKeys(m map[string]any) map[string]any {
	fieldMap := map[string]string{
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

	uriFieldMap := map[string]string{
		"uri":      "u",
		"category": "c",
		"title":    "t",
	}

	result := make(map[string]any)
	for k, v := range m {
		if compact, ok := fieldMap[k]; ok {
			// Special handling for uris array
			if k == "uris" {
				if uris, ok := v.([]any); ok {
					var newURIs []any
					for _, uri := range uris {
						if uriMap, ok := uri.(map[string]any); ok {
							compactURI := make(map[string]any)
							for uk, uv := range uriMap {
								if uCompact, uOk := uriFieldMap[uk]; uOk {
									compactURI[uCompact] = uv
								} else {
									compactURI[uk] = uv
								}
							}
							newURIs = append(newURIs, compactURI)
						} else {
							newURIs = append(newURIs, uri)
						}
					}
					result[compact] = newURIs
					continue
				}
			}
			result[compact] = v
		} else {
			result[k] = v
		}
	}
	return result
}

// convertToLongKeys converts compact field names to long equivalents
func convertToLongKeys(m map[string]any) map[string]any {
	fieldMap := map[string]string{
		"t":  "ticker",
		"n":  "name",
		"d":  "desc",
		"i":  "icon",
		"ac": "asset_class",
		"as": "asset_subclass",
		"in": "issuer_name",
		"us": "uris",
		"ai": "additional_info",
	}

	uriFieldMap := map[string]string{
		"u": "uri",
		"c": "category",
		"t": "title",
	}

	result := make(map[string]any)
	for k, v := range m {
		if long, ok := fieldMap[k]; ok {
			// Special handling for uris array
			if k == "us" {
				if uris, ok := v.([]any); ok {
					var newURIs []any
					for _, uri := range uris {
						if uriMap, ok := uri.(map[string]any); ok {
							longURI := make(map[string]any)
							for uk, uv := range uriMap {
								if uLong, uOk := uriFieldMap[uk]; uOk {
									longURI[uLong] = uv
								} else {
									longURI[uk] = uv
								}
							}
							newURIs = append(newURIs, longURI)
						} else {
							newURIs = append(newURIs, uri)
						}
					}
					result[long] = newURIs
					continue
				}
			}
			result[long] = v
		} else {
			result[k] = v
		}
	}
	return result
}

func TestValidateMPTokenMetadata(t *testing.T) {
	tests := []struct {
		name               string
		mptMetadata        any
		validationMessages []error
	}{
		{
			name: "valid MPTokenMetadata",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			validationMessages: []error{},
		},
		{
			name: "valid MPTokenMetadata with all short field names",
			mptMetadata: map[string]any{
				"t":  "TBILL",
				"n":  "T-Bill Yield Token",
				"d":  "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"i":  "https://example.org/tbill-icon.png",
				"ac": "rwa",
				"as": "treasury",
				"in": "Example Yield Co.",
				"us": []any{
					map[string]any{
						"u": "https://exampleyield.co/tbill",
						"c": "website",
						"t": "Product Page",
					},
					map[string]any{
						"u": "https://exampleyield.co/docs",
						"c": "docs",
						"t": "Yield Token Docs",
					},
				},
				"ai": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			validationMessages: []error{},
		},
		{
			name: "valid MPTokenMetadata with mixed short and long field names",
			mptMetadata: map[string]any{
				"ticker":      "CRYPTO",
				"n":           "Crypto Token",
				"icon":        "https://example.org/crypto-icon.png",
				"asset_class": "gaming",
				"d":           "A gaming token for virtual worlds.",
				"issuer_name": "Gaming Studios Inc.",
				"as":          "equity",
				"uris": []any{
					map[string]any{
						"uri":   "https://gamingstudios.com",
						"c":     "website",
						"title": "Main Website",
					},
					map[string]any{
						"uri":      "https://gamingstudios.com",
						"category": "website",
						"t":        "Main Website",
					},
				},
				"ai": "Gaming ecosystem token",
			},
			validationMessages: []error{},
		},
		{
			name: "conflicting short and long fields - ticker and t",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"t":              "BILL",
				"name":           "T-Bill Token",
				"icon":           "https://example.com/icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataFieldCollision{Long: "ticker", Compact: "t"},
			},
		},
		{
			name: "missing ticker",
			mptMetadata: map[string]any{
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataMissingField{Field: "ticker"},
			},
		},
		{
			name: "ticker has lowercase letters",
			mptMetadata: map[string]any{
				"ticker":         "tbill",
				"name":           "T-Bill Token",
				"icon":           "https://example.com/icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataTicker,
			},
		},
		{
			name: "icon not present",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Token",
				"icon":           nil,
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataMissingField{Field: "icon"},
			},
		},
		{
			name: "invalid asset_class",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Token",
				"icon":           "https://example.com/icon.png",
				"asset_class":    "invalid",
				"asset_subclass": "treasury",
				"issuer_name":    "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses},
			},
		},
		{
			name: "invalid asset_subclass not in set",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Token",
				"icon":           "https://example.com/icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "junk",
				"issuer_name":    "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses},
			},
		},
		{
			name: "missing asset_subclass for rwa",
			mptMetadata: map[string]any{
				"ticker":      "TBILL",
				"name":        "T-Bill Token",
				"icon":        "https://example.com/icon.png",
				"asset_class": "rwa",
				"issuer_name": "Issuer",
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataRWASubClassRequired,
			},
		},
		{
			name: "uris empty",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Token",
				"icon":           "https://example.com/icon.png",
				"asset_class":    "defi",
				"issuer_name":    "Issuer",
				"asset_subclass": "stablecoin",
				"uris":           []any{},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name: "additional_info is invalid type - array",
			mptMetadata: map[string]any{
				"ticker":          "TBILL",
				"name":            "T-Bill Token",
				"icon":            "https://example.com/icon.png",
				"asset_class":     "defi",
				"issuer_name":     "Issuer",
				"additional_info": []any{"not", "valid"},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataAdditionalInfo,
			},
		},
		{
			name: "additional_info is invalid type - number",
			mptMetadata: map[string]any{
				"ticker":          "TBILL",
				"name":            "T-Bill Token",
				"icon":            "https://example.com/icon.png",
				"asset_class":     "defi",
				"issuer_name":     "Issuer",
				"additional_info": 123,
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataAdditionalInfo,
			},
		},
		{
			name: "multiple warnings",
			mptMetadata: map[string]any{
				"ticker":         "TBILLLLLLL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https/example.org/tbill-icon.png",
				"asset_class":    "rwamemes",
				"asset_subclass": "treasurymemes",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":   "http://notsecure.com",
						"type":  "website",
						"title": "Homepage",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataTicker,
				ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses},
				ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses},
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name:        "null mptMetadata",
			mptMetadata: nil,
			validationMessages: []error{
				ErrInvalidMPTokenMetadataMissingField{Field: "ticker"},
				ErrInvalidMPTokenMetadataMissingField{Field: "name"},
				ErrInvalidMPTokenMetadataMissingField{Field: "icon"},
				ErrInvalidMPTokenMetadataMissingField{Field: "issuer_name"},
				ErrInvalidMPTokenMetadataMissingField{Field: "asset_class"},
			},
		},
		{
			name:        "empty mptMetadata",
			mptMetadata: map[string]any{},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataMissingField{Field: "ticker"},
				ErrInvalidMPTokenMetadataMissingField{Field: "name"},
				ErrInvalidMPTokenMetadataMissingField{Field: "icon"},
				ErrInvalidMPTokenMetadataMissingField{Field: "issuer_name"},
				ErrInvalidMPTokenMetadataMissingField{Field: "asset_class"},
			},
		},
		{
			name:        "incorrect JSON",
			mptMetadata: "not a json",
			validationMessages: []error{
				ErrInvalidMPTokenMetadataJSON,
			},
		},
		{
			name: "more than 9 fields",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"issuer_address": "123 Example Yield Co.",
				"issuer_account": "321 Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "http://notsecure.com",
						"category": "website",
						"title":    "Homepage",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataFieldCount{Count: 9},
				ErrInvalidMPTokenMetadataUnknownField{Field: "issuer_account"},
				ErrInvalidMPTokenMetadataUnknownField{Field: "issuer_address"},
			},
		},
		{
			name: "more than 3 uri fields",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://notsecure.com",
						"category": "website",
						"title":    "Homepage",
						"footer":   "footer",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name: "invalid uris structure",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris":           "uris",
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name: "invalid uri inner structure",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris":           []any{1, 2},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataURIs,
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name: "conflicting uri long and compact forms",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":   "https://exampleyield.co/tbill",
						"u":     "website",
						"title": "Product Page",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataFieldCollision{Long: "uri", Compact: "u"},
				ErrInvalidMPTokenMetadataURIs,
			},
		},
		{
			name: "exceeds 1024 bytes",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataSize,
			},
		},
		{
			name: "null values",
			mptMetadata: map[string]any{
				"ticker":          nil,
				"name":            nil,
				"desc":            nil,
				"icon":            nil,
				"asset_class":     nil,
				"asset_subclass":  nil,
				"issuer_name":     nil,
				"uris":            nil,
				"additional_info": nil,
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataTicker,
				ErrInvalidMPTokenMetadataMissingField{Field: "name"},
				ErrInvalidMPTokenMetadataMissingField{Field: "icon"},
				ErrInvalidMPTokenMetadataMissingField{Field: "issuer_name"},
				ErrInvalidMPTokenMetadataAssetClass{AssetClassSet: MPTokenMetadataAssetClasses},
				ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses},
				ErrInvalidMPTokenMetadataURIs,
				ErrInvalidMPTokenMetadataAdditionalInfo,
			},
		},
		{
			name: "empty string in URI fields",
			mptMetadata: map[string]any{
				"ticker":      "TEST",
				"name":        "Test Token",
				"icon":        "icon.png",
				"asset_class": "other",
				"issuer_name": "Issuer",
				"uris": []any{
					map[string]any{
						"uri":      "",
						"category": "website",
						"title":    "Title",
					},
				},
			},
			validationMessages: []error{}, // Empty strings are valid (validation only checks type, not content)
		},
		{
			name: "unknown field in URI object",
			mptMetadata: map[string]any{
				"ticker":      "TEST",
				"name":        "Test Token",
				"icon":        "icon.png",
				"asset_class": "other",
				"issuer_name": "Issuer",
				"uris": []any{
					map[string]any{
						"uri":      "https://example.com",
						"category": "website",
						"title":    "Title",
						"extra":    "unknown",
					},
				},
			},
			validationMessages: []error{
				ErrInvalidMPTokenMetadataURIs,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hexStr := toHexString(t, tt.mptMetadata)
			err := ValidateMPTokenMetadata(hexStr)
			actualErrors := extractValidationErrors(err)

			// Compare error messages for easier debugging
			expectedMessages := make([]string, len(tt.validationMessages))
			for i, e := range tt.validationMessages {
				expectedMessages[i] = e.Error()
			}
			actualMessages := make([]string, len(actualErrors))
			for i, e := range actualErrors {
				actualMessages[i] = e.Error()
			}

			assert.ElementsMatch(t, expectedMessages, actualMessages,
				"Validation errors do not match for test: %s", tt.name)
		})
	}
}

func TestEncodeDecodeMPTokenMetadata(t *testing.T) {
	tests := []struct {
		name             string
		mptMetadata      map[string]any
		expectedLongForm map[string]any
		hex              string
	}{
		{
			name: "valid long MPTokenMetadata",
			mptMetadata: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			expectedLongForm: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			hex: "7B226163223A22727761222C226169223A7B226375736970223A22393132373936525830222C22696E7465726573745F72617465223A22352E303025222C22696E7465726573745F74797065223A227661726961626C65222C226D617475726974795F64617465223A22323034352D30362D3330222C227969656C645F736F75726365223A22552E532E2054726561737572792042696C6C73227D2C226173223A227472656173757279222C2264223A2241207969656C642D62656172696E6720737461626C65636F696E206261636B65642062792073686F72742D7465726D20552E532E205472656173757269657320616E64206D6F6E6579206D61726B657420696E737472756D656E74732E222C2269223A2268747470733A2F2F6578616D706C652E6F72672F7462696C6C2D69636F6E2E706E67222C22696E223A224578616D706C65205969656C6420436F2E222C226E223A22542D42696C6C205969656C6420546F6B656E222C2274223A225442494C4C222C227573223A5B7B2263223A2277656273697465222C2274223A2250726F647563742050616765222C2275223A2268747470733A2F2F6578616D706C657969656C642E636F2F7462696C6C227D2C7B2263223A22646F6373222C2274223A225969656C6420546F6B656E20446F6373222C2275223A2268747470733A2F2F6578616D706C657969656C642E636F2F646F6373227D5D7D",
		},
		{
			name: "valid MPTokenMetadata with all short field names",
			mptMetadata: map[string]any{
				"t":  "TBILL",
				"n":  "T-Bill Yield Token",
				"d":  "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"i":  "https://example.org/tbill-icon.png",
				"ac": "rwa",
				"as": "treasury",
				"in": "Example Yield Co.",
				"us": []any{
					map[string]any{
						"u": "https://exampleyield.co/tbill",
						"c": "website",
						"t": "Product Page",
					},
					map[string]any{
						"u": "https://exampleyield.co/docs",
						"c": "docs",
						"t": "Yield Token Docs",
					},
				},
				"ai": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			expectedLongForm: map[string]any{
				"ticker":         "TBILL",
				"name":           "T-Bill Yield Token",
				"desc":           "A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments.",
				"icon":           "https://example.org/tbill-icon.png",
				"asset_class":    "rwa",
				"asset_subclass": "treasury",
				"issuer_name":    "Example Yield Co.",
				"uris": []any{
					map[string]any{
						"uri":      "https://exampleyield.co/tbill",
						"category": "website",
						"title":    "Product Page",
					},
					map[string]any{
						"uri":      "https://exampleyield.co/docs",
						"category": "docs",
						"title":    "Yield Token Docs",
					},
				},
				"additional_info": map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
			hex: "7B226163223A22727761222C226169223A7B226375736970223A22393132373936525830222C22696E7465726573745F72617465223A22352E303025222C22696E7465726573745F74797065223A227661726961626C65222C226D617475726974795F64617465223A22323034352D30362D3330222C227969656C645F736F75726365223A22552E532E2054726561737572792042696C6C73227D2C226173223A227472656173757279222C2264223A2241207969656C642D62656172696E6720737461626C65636F696E206261636B65642062792073686F72742D7465726D20552E532E205472656173757269657320616E64206D6F6E6579206D61726B657420696E737472756D656E74732E222C2269223A2268747470733A2F2F6578616D706C652E6F72672F7462696C6C2D69636F6E2E706E67222C22696E223A224578616D706C65205969656C6420436F2E222C226E223A22542D42696C6C205969656C6420546F6B656E222C2274223A225442494C4C222C227573223A5B7B2263223A2277656273697465222C2274223A2250726F647563742050616765222C2275223A2268747470733A2F2F6578616D706C657969656C642E636F2F7462696C6C227D2C7B2263223A22646F6373222C2274223A225969656C6420546F6B656E20446F6373222C2275223A2268747470733A2F2F6578616D706C657969656C642E636F2F646F6373227D5D7D",
		},
		{
			name: "valid MPTokenMetadata with mixed short and long field names",
			mptMetadata: map[string]any{
				"ticker":      "CRYPTO",
				"n":           "Crypto Token",
				"icon":        "https://example.org/crypto-icon.png",
				"asset_class": "gaming",
				"d":           "A gaming token for virtual worlds.",
				"issuer_name": "Gaming Studios Inc.",
				"as":          "equity",
				"uris": []any{
					map[string]any{
						"uri":   "https://gamingstudios.com",
						"c":     "website",
						"title": "Main Website",
					},
					map[string]any{
						"uri":      "https://gamingstudios.com",
						"category": "website",
						"t":        "Main Website",
					},
				},
				"ai": "Gaming ecosystem token",
			},
			expectedLongForm: map[string]any{
				"ticker":         "CRYPTO",
				"name":           "Crypto Token",
				"icon":           "https://example.org/crypto-icon.png",
				"asset_class":    "gaming",
				"desc":           "A gaming token for virtual worlds.",
				"issuer_name":    "Gaming Studios Inc.",
				"asset_subclass": "equity",
				"uris": []any{
					map[string]any{
						"uri":      "https://gamingstudios.com",
						"category": "website",
						"title":    "Main Website",
					},
					map[string]any{
						"uri":      "https://gamingstudios.com",
						"category": "website",
						"title":    "Main Website",
					},
				},
				"additional_info": "Gaming ecosystem token",
			},
			hex: "7B226163223A2267616D696E67222C226169223A2247616D696E672065636F73797374656D20746F6B656E222C226173223A22657175697479222C2264223A22412067616D696E6720746F6B656E20666F72207669727475616C20776F726C64732E222C2269223A2268747470733A2F2F6578616D706C652E6F72672F63727970746F2D69636F6E2E706E67222C22696E223A2247616D696E672053747564696F7320496E632E222C226E223A2243727970746F20546F6B656E222C2274223A2243525950544F222C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D2C7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert map with long field names to compact field names for struct unmarshaling
			compactMap := convertToCompactKeys(tt.mptMetadata)

			// Convert compact map to ParsedMPTokenMetadata struct
			jsonBytes, err := json.Marshal(compactMap)
			require.NoError(t, err)

			var meta ParsedMPTokenMetadata
			err = json.Unmarshal(jsonBytes, &meta)
			require.NoError(t, err)

			// Test encoding
			encodedHex, err := EncodeMPTokenMetadata(meta)
			require.NoError(t, err)
			assert.Equal(t, tt.hex, encodedHex, "Encoded hex does not match")

			// Test decoding
			decoded, err := DecodeMPTokenMetadata(encodedHex)
			require.NoError(t, err)

			// Convert decoded struct back to map for comparison
			decodedBytes, err := json.Marshal(decoded)
			require.NoError(t, err)

			var decodedMap map[string]any
			err = json.Unmarshal(decodedBytes, &decodedMap)
			require.NoError(t, err)

			// Convert decoded map (which has compact keys) back to long form for comparison
			decodedLongForm := convertToLongKeys(decodedMap)

			// Compare
			assert.Equal(t, tt.expectedLongForm, decodedLongForm, "Decoded metadata does not match expected long form")
		})
	}
}

func TestDecodeMPTokenMetadata_EdgeCases(t *testing.T) {
	// These tests verify that decoding handles edge cases (collisions, extra fields)
	// that cannot round-trip through struct encoding because:
	// - Extra fields are lost when unmarshaling into a struct
	// - Field collisions cannot exist in a struct (only one form can be present)
	tests := []struct {
		name     string
		hex      string
		expected ParsedMPTokenMetadata
	}{
		{
			name: "with extra fields",
			// JSON contains: {"extra":{"extra":"extra"}, ...valid fields...} - extra field is ignored
			hex: "7B226163223A2267616D696E67222C226169223A2247616D696E672065636F73797374656D20746F6B656E222C226173223A22657175697479222C2264223A22412067616D696E6720746F6B656E20666F72207669727475616C20776F726C64732E222C226578747261223A7B226578747261223A226578747261227D2C2269223A2268747470733A2F2F6578616D706C652E6F72672F63727970746F2D69636F6E2E706E67222C22696E223A2247616D696E672053747564696F7320496E632E222C226E223A2243727970746F20546F6B656E222C2274223A2243525950544F222C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
			expected: ParsedMPTokenMetadata{
				Ticker:         "CRYPTO",
				Name:           "Crypto Token",
				Icon:           "https://example.org/crypto-icon.png",
				AssetClass:     "gaming",
				AssetSubclass:  stringPtr("equity"),
				IssuerName:     "Gaming Studios Inc.",
				Desc:           stringPtr("A gaming token for virtual worlds."),
				AdditionalInfo: "Gaming ecosystem token",
				URIs: []ParsedMPTokenMetadataURI{
					{
						URI:      "https://gamingstudios.com",
						Category: "website",
						Title:    "Main Website",
					},
				},
			},
		},
		{
			name: "with unknown null fields",
			hex:  "7B226578747261223A6E756C6C2C2274223A2243525950544F227D",
			// JSON contains: {"extra":null,"t":"CRYPTO"} - null extra field is ignored
			expected: ParsedMPTokenMetadata{
				Ticker: "CRYPTO",
			},
		},
		{
			name: "multiple uris and us",
			// JSON contains: {"t":"CRYPTO","uris":[...],"us":[...]} - both forms present, compactKeys resolves collision
			hex: "7B2274223A2243525950544F222C2275726973223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D2C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
			expected: ParsedMPTokenMetadata{
				Ticker: "CRYPTO",
				URIs: []ParsedMPTokenMetadataURI{
					{
						URI:      "https://gamingstudios.com",
						Category: "website",
						Title:    "Main Website",
					},
				},
			},
		},
		{
			name: "multiple keys in uri",
			// JSON contains: {"us":[{"uri":"https://...","u":"website","category":"Main","c":"Main"}]} - nested collisions
			hex: "7B227573223A5B7B2263223A224D61696E2057656273697465222C2263617465676F7279223A224D61696E2057656273697465222C2275223A2277656273697465222C22757269223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
			expected: ParsedMPTokenMetadata{
				URIs: []ParsedMPTokenMetadataURI{
					{
						URI:      "website",
						Category: "Main Website",
						Title:    "",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test decoding from hex (these cases cannot be encoded through struct)
			decoded, err := DecodeMPTokenMetadata(tt.hex)
			require.NoError(t, err)

			// Compare struct fields directly
			assert.Equal(t, tt.expected.Ticker, decoded.Ticker)
			assert.Equal(t, tt.expected.Name, decoded.Name)
			assert.Equal(t, tt.expected.Icon, decoded.Icon)
			assert.Equal(t, tt.expected.AssetClass, decoded.AssetClass)
			if tt.expected.AssetSubclass != nil {
				require.NotNil(t, decoded.AssetSubclass)
				assert.Equal(t, *tt.expected.AssetSubclass, *decoded.AssetSubclass)
			} else {
				assert.Nil(t, decoded.AssetSubclass)
			}
			assert.Equal(t, tt.expected.IssuerName, decoded.IssuerName)
			if tt.expected.Desc != nil {
				require.NotNil(t, decoded.Desc)
				assert.Equal(t, *tt.expected.Desc, *decoded.Desc)
			} else {
				assert.Nil(t, decoded.Desc)
			}
			assert.Equal(t, tt.expected.AdditionalInfo, decoded.AdditionalInfo)
			assert.Equal(t, len(tt.expected.URIs), len(decoded.URIs))
			for i, expectedURI := range tt.expected.URIs {
				if i < len(decoded.URIs) {
					assert.Equal(t, expectedURI.URI, decoded.URIs[i].URI)
					assert.Equal(t, expectedURI.Category, decoded.URIs[i].Category)
					assert.Equal(t, expectedURI.Title, decoded.URIs[i].Title)
				}
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

func TestDecodeMPTokenMetadata_Errors(t *testing.T) {
	t.Run("invalid hex", func(t *testing.T) {
		_, err := DecodeMPTokenMetadata("invalid")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidMPTokenMetadataHex, err)
	})

	t.Run("invalid JSON underneath hex", func(t *testing.T) {
		_, err := DecodeMPTokenMetadata("464F4F")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidMPTokenMetadataJSON, err)
	})
}
