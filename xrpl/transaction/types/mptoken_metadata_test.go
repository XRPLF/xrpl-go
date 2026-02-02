package types

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				ErrInvalidMPTokenMetadataInvalidString{Key: "icon"},
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
				ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses[:]},
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
				ErrInvalidMPTokenMetadataAssetSubClass{AssetSubclassSet: MPTokenMetadataAssetSubClasses[:]},
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
				ErrInvalidMPTokenMetadataInvalidString{Key: "ticker"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "name"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "desc"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "icon"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "issuer_name"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "asset_class"},
				ErrInvalidMPTokenMetadataInvalidString{Key: "asset_subclass"},
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
			validationMessages: []error{
				ErrInvalidMPTokenMetadataEmptyString{Key: "uri"},
			},
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
		name     string
		metadata ParsedMPTokenMetadata
		hex      string
	}{
		{
			name: "valid long MPTokenMetadata",
			metadata: ParsedMPTokenMetadata{
				Ticker:        "TBILL",
				Name:          "T-Bill Yield Token",
				Desc:          stringPtr("A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments."),
				Icon:          "https://example.org/tbill-icon.png",
				AssetClass:    "rwa",
				AssetSubclass: stringPtr("treasury"),
				IssuerName:    "Example Yield Co.",
				URIs: []ParsedMPTokenMetadataURI{
					{URI: "https://exampleyield.co/tbill", Category: "website", Title: "Product Page"},
					{URI: "https://exampleyield.co/docs", Category: "docs", Title: "Yield Token Docs"},
				},
				AdditionalInfo: map[string]any{
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
			metadata: ParsedMPTokenMetadata{
				Ticker:        "TBILL",
				Name:          "T-Bill Yield Token",
				Desc:          stringPtr("A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments."),
				Icon:          "https://example.org/tbill-icon.png",
				AssetClass:    "rwa",
				AssetSubclass: stringPtr("treasury"),
				IssuerName:    "Example Yield Co.",
				URIs: []ParsedMPTokenMetadataURI{
					{URI: "https://exampleyield.co/tbill", Category: "website", Title: "Product Page"},
					{URI: "https://exampleyield.co/docs", Category: "docs", Title: "Yield Token Docs"},
				},
				AdditionalInfo: map[string]any{
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
			metadata: ParsedMPTokenMetadata{
				Ticker:         "CRYPTO",
				Name:           "Crypto Token",
				Desc:           stringPtr("A gaming token for virtual worlds."),
				Icon:           "https://example.org/crypto-icon.png",
				AssetClass:     "gaming",
				AssetSubclass:  stringPtr("equity"),
				IssuerName:     "Gaming Studios Inc.",
				AdditionalInfo: "Gaming ecosystem token",
				URIs: []ParsedMPTokenMetadataURI{
					{URI: "https://gamingstudios.com", Category: "website", Title: "Main Website"},
					{URI: "https://gamingstudios.com", Category: "website", Title: "Main Website"},
				},
			},
			hex: "7B226163223A2267616D696E67222C226169223A2247616D696E672065636F73797374656D20746F6B656E222C226173223A22657175697479222C2264223A22412067616D696E6720746F6B656E20666F72207669727475616C20776F726C64732E222C2269223A2268747470733A2F2F6578616D706C652E6F72672F63727970746F2D69636F6E2E706E67222C22696E223A2247616D696E672053747564696F7320496E632E222C226E223A2243727970746F20546F6B656E222C2274223A2243525950544F222C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D2C7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeMPTokenMetadata(tt.metadata)
			require.NoError(t, err)
			assert.Equal(t, tt.hex, encoded)

			decoded, err := DecodeMPTokenMetadata(tt.hex)
			require.NoError(t, err)
			assert.Equal(t, tt.metadata, decoded)

			encodedAgain, err := EncodeMPTokenMetadata(decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.hex, encodedAgain)
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
			hex:  "7B226163223A2267616D696E67222C226169223A2247616D696E672065636F73797374656D20746F6B656E222C226173223A22657175697479222C2264223A22412067616D696E6720746F6B656E20666F72207669727475616C20776F726C64732E222C226578747261223A7B226578747261223A226578747261227D2C2269223A2268747470733A2F2F6578616D706C652E6F72672F63727970746F2D69636F6E2E706E67222C22696E223A2247616D696E672053747564696F7320496E632E222C226E223A2243727970746F20546F6B656E222C2274223A2243525950544F222C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
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
			expected: ParsedMPTokenMetadata{
				Ticker: "CRYPTO",
			},
		},
		{
			name: "multiple uris and us",
			hex:  "7B2274223A2243525950544F222C2275726973223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D2C227573223A5B7B2263223A2277656273697465222C2274223A224D61696E2057656273697465222C2275223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
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
			hex:  "7B227573223A5B7B2263223A224D61696E2057656273697465222C2263617465676F7279223A224D61696E2057656273697465222C2275223A2277656273697465222C22757269223A2268747470733A2F2F67616D696E6773747564696F732E636F6D227D5D7D",
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
			assert.Equal(t, tt.expected, decoded, "Decoded metadata does not match expected")
		})
	}
}

func TestDecodeMPTokenMetadata_NotCompactKeys(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected ParsedMPTokenMetadata
	}{
		{
			name: "not compact keys",
			hex:  "7B226164646974696F6E616C5F696E666F223A7B226375736970223A22393132373936525830222C22696E7465726573745F72617465223A22352E303025222C22696E7465726573745F74797065223A227661726961626C65222C226D617475726974795F64617465223A22323034352D30362D3330222C227969656C645F736F75726365223A22552E532E2054726561737572792042696C6C73227D2C2261737365745F636C617373223A22727761222C2261737365745F737562636C617373223A227472656173757279222C2264657363223A2241207969656C642D62656172696E6720737461626C65636F696E206261636B65642062792073686F72742D7465726D20552E532E205472656173757269657320616E64206D6F6E6579206D61726B657420696E737472756D656E74732E222C2269636F6E223A2268747470733A2F2F6578616D706C652E6F72672F7462696C6C2D69636F6E2E706E67222C226973737565725F6E616D65223A224578616D706C65205969656C6420436F2E222C226E616D65223A22542D42696C6C205969656C6420546F6B656E222C227469636B6572223A225442494C4C222C2275726973223A5B7B2263617465676F7279223A2277656273697465222C227469746C65223A2250726F647563742050616765222C22757269223A2268747470733A2F2F6578616D706C657969656C642E636F2F7462696C6C227D2C7B2263617465676F7279223A22646F6373222C227469746C65223A225969656C6420546F6B656E20446F6373222C22757269223A2268747470733A2F2F6578616D706C657969656C642E636F2F646F6373227D5D7D",
			expected: ParsedMPTokenMetadata{
				Ticker:        "TBILL",
				Name:          "T-Bill Yield Token",
				Desc:          stringPtr("A yield-bearing stablecoin backed by short-term U.S. Treasuries and money market instruments."),
				Icon:          "https://example.org/tbill-icon.png",
				AssetClass:    "rwa",
				AssetSubclass: stringPtr("treasury"),
				IssuerName:    "Example Yield Co.",
				URIs: []ParsedMPTokenMetadataURI{
					{
						URI:      "https://exampleyield.co/tbill",
						Category: "website",
						Title:    "Product Page",
					},
					{
						URI:      "https://exampleyield.co/docs",
						Category: "docs",
						Title:    "Yield Token Docs",
					},
				},
				AdditionalInfo: map[string]any{
					"interest_rate": "5.00%",
					"interest_type": "variable",
					"yield_source":  "U.S. Treasury Bills",
					"maturity_date": "2045-06-30",
					"cusip":         "912796RX0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeMPTokenMetadata(tt.hex)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, decoded)
		})
	}
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

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
