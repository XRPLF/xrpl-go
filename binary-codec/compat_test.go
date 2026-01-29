package binarycodec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CodecFixtures represents the structure of codec-fixtures.json
type CodecFixtures struct {
	AccountState []AccountStateTest `json:"accountState"`
}

type AccountStateTest struct {
	Binary string         `json:"binary"`
	JSON   map[string]any `json:"json"`
}

// DataDrivenTests represents the structure of data-driven-tests.json
type DataDrivenTests struct {
	Types       []TypeTest  `json:"types"`
	FieldsTests []FieldTest `json:"fields_tests"`
}

type TypeTest struct {
	Name    string `json:"name"`
	Ordinal int    `json:"ordinal"`
}

type FieldTest struct {
	TypeName    string `json:"type_name"`
	Name        string `json:"name"`
	NthOfType   int    `json:"nth_of_type"`
	Type        int    `json:"type"`
	ExpectedHex string `json:"expected_hex"`
}

func loadJSONFile(t *testing.T, filename string) map[string]any {
	data, err := os.ReadFile(filepath.Join("testdata/fixtures", filename))
	require.NoError(t, err, "Failed to read fixture file: %s", filename)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "Failed to parse fixture JSON: %s", filename)

	return result
}

func loadBinaryFile(t *testing.T, filename string) string {
	data, err := os.ReadFile(filepath.Join("testdata/fixtures", filename))
	require.NoError(t, err, "Failed to read binary fixture file: %s", filename)

	// Binary files just contain the hex string
	return strings.TrimSpace(string(data))
}

// cleanJSON removes fields that shouldn't be compared (metadata, etc.)
func cleanJSONForEncoding(json map[string]any) map[string]any {
	result := make(map[string]any)
	nonEncodingFields := map[string]bool{
		"date":         true,
		"hash":         true,
		"inLedger":     true,
		"ledger_index": true,
		"meta":         true,
		"validated":    true,
	}

	for k, v := range json {
		if !nonEncodingFields[k] {
			result[k] = v
		}
	}
	return result
}

// convertJSONTypes recursively converts JSON types to Go types expected by the binary codec
// JSON numbers (float64) are converted to uint32 for integer fields
// This is needed because Go's json.Unmarshal uses float64 for all numbers
func convertJSONTypes(data map[string]any) map[string]any {
	// Fields that should be uint32
	uint32Fields := map[string]bool{
		"Flags":              true,
		"Sequence":           true,
		"LastLedgerSequence": true,
		"DestinationTag":     true,
		"SourceTag":          true,
		"OfferSequence":      true,
		"CancelAfter":        true,
		"FinishAfter":        true,
		"SettleDelay":        true,
		"Expiration":         true,
		"TransferFee":        true,
		"QualityIn":          true,
		"QualityOut":         true,
		"SignerQuorum":       true,
		"TicketCount":        true,
		"TicketSequence":     true,
		"OwnerCount":         true,
		"PreviousTxnLgrSeq":  true,
	}

	// Fields that should be int (used by UInt16)
	intFields := map[string]bool{
		"SignerWeight": true,
		"type":         true, // Used in Paths
	}

	result := make(map[string]any)
	for k, v := range data {
		result[k] = convertValue(k, v, uint32Fields, intFields)
	}
	return result
}

func convertValue(key string, v any, uint32Fields, intFields map[string]bool) any {
	switch val := v.(type) {
	case float64:
		// Convert to uint32 if it's a uint32 field
		if uint32Fields[key] {
			return uint32(val)
		}
		// Convert to int if it's an int field
		if intFields[key] {
			return int(val)
		}
		// Check if it's a whole number that should be int
		if val == float64(int64(val)) {
			return int(val)
		}
		return val
	case map[string]any:
		// Recursively convert nested objects
		return convertJSONTypes(val)
	case []any:
		// Recursively convert arrays
		result := make([]any, len(val))
		for i, item := range val {
			if m, ok := item.(map[string]any); ok {
				result[i] = convertJSONTypes(m)
			} else {
				result[i] = convertValue("", item, uint32Fields, intFields)
			}
		}
		return result
	default:
		return v
	}
}

// TestCompat_EncodeTransaction tests encoding transactions
func TestCompat_EncodeTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		txFile     string
		binaryFile string
	}{
		{"DeliverMin", "delivermin-tx.json", "delivermin-tx-binary.json"},
		{"EscrowCreate", "escrow-create-tx.json", "escrow-create-binary.json"},
		{"EscrowCancel", "escrow-cancel-tx.json", "escrow-cancel-binary.json"},
		{"EscrowFinish", "escrow-finish-tx.json", "escrow-finish-binary.json"},
		{"PaymentChannelCreate", "payment-channel-create-tx.json", "payment-channel-create-binary.json"},
		{"PaymentChannelClaim", "payment-channel-claim-tx.json", "payment-channel-claim-binary.json"},
		{"PaymentChannelFund", "payment-channel-fund-tx.json", "payment-channel-fund-binary.json"},
		{"SignerListSet", "signerlistset-tx.json", "signerlistset-tx-binary.json"},
		{"DepositPreauth", "deposit-preauth-tx.json", "deposit-preauth-tx-binary.json"},
		{"TicketCreate", "ticket-create-tx.json", "ticket-create-binary.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_Encode", func(t *testing.T) {
			txJSON := loadJSONFile(t, tc.txFile)
			expectedBinary := loadBinaryFile(t, tc.binaryFile)

			// Clean the JSON before encoding (remove non-encoding fields)
			cleanedJSON := cleanJSONForEncoding(txJSON)
			// Convert JSON types (float64 -> uint32 for integer fields)
			convertedJSON := convertJSONTypes(cleanedJSON)

			encoded, err := Encode(convertedJSON)
			require.NoError(t, err, "Failed to encode transaction")

			// Strip quotes from expected binary if present
			expectedBinary = strings.Trim(expectedBinary, "\"")

			require.Equal(t, strings.ToUpper(expectedBinary), strings.ToUpper(encoded),
				"Encoding mismatch for %s", tc.name)
		})
	}
}

// TestCompat_DecodeTransaction tests decoding transactions
func TestCompat_DecodeTransaction(t *testing.T) {
	testCases := []struct {
		name       string
		txFile     string
		binaryFile string
	}{
		{"DeliverMin", "delivermin-tx.json", "delivermin-tx-binary.json"},
		{"EscrowCreate", "escrow-create-tx.json", "escrow-create-binary.json"},
		{"EscrowCancel", "escrow-cancel-tx.json", "escrow-cancel-binary.json"},
		{"EscrowFinish", "escrow-finish-tx.json", "escrow-finish-binary.json"},
		{"PaymentChannelCreate", "payment-channel-create-tx.json", "payment-channel-create-binary.json"},
		{"PaymentChannelClaim", "payment-channel-claim-tx.json", "payment-channel-claim-binary.json"},
		{"PaymentChannelFund", "payment-channel-fund-tx.json", "payment-channel-fund-binary.json"},
		{"SignerListSet", "signerlistset-tx.json", "signerlistset-tx-binary.json"},
		{"DepositPreauth", "deposit-preauth-tx.json", "deposit-preauth-tx-binary.json"},
		{"TicketCreate", "ticket-create-tx.json", "ticket-create-binary.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name+"_Decode", func(t *testing.T) {
			binary := loadBinaryFile(t, tc.binaryFile)

			// Strip quotes from binary if present
			binary = strings.Trim(binary, "\"")

			decoded, err := Decode(binary)
			require.NoError(t, err, "Failed to decode transaction")

			// Load expected JSON for comparison
			expectedJSON := loadJSONFile(t, tc.txFile)
			cleanedExpected := cleanJSONForEncoding(expectedJSON)

			// Compare key fields
			for key, expectedVal := range cleanedExpected {
				actualVal, exists := decoded[key]
				if !exists {
					t.Errorf("Missing field %s in decoded result", key)
					continue
				}

				// Deep compare values
				if !deepEqual(expectedVal, actualVal) {
					t.Errorf("Field %s mismatch:\n  expected: %v (%T)\n  actual:   %v (%T)",
						key, expectedVal, expectedVal, actualVal, actualVal)
				}
			}
		})
	}
}

// TestCompat_AccountState tests encoding/decoding ledger entries from codec-fixtures.json
func TestCompat_AccountState(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/codec-fixtures.json")
	require.NoError(t, err, "Failed to read codec-fixtures.json")

	var fixtures CodecFixtures
	err = json.Unmarshal(data, &fixtures)
	require.NoError(t, err, "Failed to parse codec-fixtures.json")

	for i, tc := range fixtures.AccountState {
		t.Run("AccountState_"+string(rune(i)), func(t *testing.T) {
			// Test decoding
			t.Run("Decode", func(t *testing.T) {
				decoded, err := Decode(tc.Binary)
				require.NoError(t, err, "Failed to decode account state")

				// Compare key fields
				for key, expectedVal := range tc.JSON {
					actualVal, exists := decoded[key]
					if !exists {
						t.Errorf("Missing field %s in decoded result", key)
						continue
					}

					if !deepEqual(expectedVal, actualVal) {
						t.Errorf("Field %s mismatch:\n  expected: %v\n  actual:   %v", key, expectedVal, actualVal)
					}
				}
			})

			// Test encoding
			t.Run("Encode", func(t *testing.T) {
				// Convert JSON types for encoding
				convertedJSON := convertJSONTypes(tc.JSON)
				encoded, err := Encode(convertedJSON)
				require.NoError(t, err, "Failed to encode account state")

				require.Equal(t, strings.ToUpper(tc.Binary), strings.ToUpper(encoded),
					"Encoding mismatch for account state %d", i)
			})
		})
	}
}

// TestCompat_RoundTrip tests that encode(decode(binary)) == binary
func TestCompat_RoundTrip(t *testing.T) {
	testCases := []struct {
		name       string
		binaryFile string
	}{
		{"DeliverMin", "delivermin-tx-binary.json"},
		{"EscrowCreate", "escrow-create-binary.json"},
		{"EscrowCancel", "escrow-cancel-binary.json"},
		{"PaymentChannelCreate", "payment-channel-create-binary.json"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binary := loadBinaryFile(t, tc.binaryFile)
			binary = strings.Trim(binary, "\"")

			// Decode
			decoded, err := Decode(binary)
			require.NoError(t, err, "Failed to decode")

			// Encode back
			reEncoded, err := Encode(decoded)
			require.NoError(t, err, "Failed to re-encode")

			require.Equal(t, strings.ToUpper(binary), strings.ToUpper(reEncoded),
				"Round-trip mismatch for %s", tc.name)
		})
	}
}

// TestCompat_XCodecFixtures tests X-address handling in binary codec
func TestCompat_XCodecFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/fixtures/x-codec-fixtures.json")
	if err != nil {
		t.Skip("x-codec-fixtures.json not found")
		return
	}

	var fixtures struct {
		Transactions []struct {
			RJSON map[string]any `json:"rjson"`
			XJSON map[string]any `json:"xjson"`
		} `json:"transactions"`
	}

	err = json.Unmarshal(data, &fixtures)
	require.NoError(t, err, "Failed to parse x-codec-fixtures.json")

	for i, tx := range fixtures.Transactions {
		t.Run("XAddress_"+string(rune(i)), func(t *testing.T) {
			// Convert JSON types for encoding
			rConverted := convertJSONTypes(tx.RJSON)
			xConverted := convertJSONTypes(tx.XJSON)

			// Encode rjson and xjson - they should produce the same binary
			rEncoded, err := Encode(rConverted)
			require.NoError(t, err, "Failed to encode rjson")

			xEncoded, err := Encode(xConverted)
			require.NoError(t, err, "Failed to encode xjson")

			require.Equal(t, rEncoded, xEncoded,
				"rjson and xjson should encode to same binary")
		})
	}
}

// deepEqual compares two values with special handling for numbers
func deepEqual(expected, actual any) bool {
	// Handle nil cases
	if expected == nil && actual == nil {
		return true
	}
	if expected == nil || actual == nil {
		return false
	}

	// Handle numeric comparisons (JSON numbers can be float64 or int)
	switch e := expected.(type) {
	case float64:
		switch a := actual.(type) {
		case float64:
			return e == a
		case int:
			return e == float64(a)
		case int64:
			return e == float64(a)
		case uint32:
			return e == float64(a)
		case uint64:
			return e == float64(a)
		}
	case int:
		switch a := actual.(type) {
		case float64:
			return float64(e) == a
		case int:
			return e == a
		case int64:
			return int64(e) == a
		case uint32:
			return uint32(e) == a
		}
	case string:
		if a, ok := actual.(string); ok {
			return strings.EqualFold(e, a)
		}
	case map[string]any:
		if a, ok := actual.(map[string]any); ok {
			if len(e) != len(a) {
				return false
			}
			for k, v := range e {
				av, exists := a[k]
				if !exists || !deepEqual(v, av) {
					return false
				}
			}
			return true
		}
	case []any:
		switch a := actual.(type) {
		case []any:
			if len(e) != len(a) {
				return false
			}
			for i := range e {
				if !deepEqual(e[i], a[i]) {
					return false
				}
			}
			return true
		case []string:
			// Handle []any (from JSON) vs []string (from Go)
			if len(e) != len(a) {
				return false
			}
			for i := range e {
				if !deepEqual(e[i], a[i]) {
					return false
				}
			}
			return true
		}
	case []string:
		switch a := actual.(type) {
		case []string:
			if len(e) != len(a) {
				return false
			}
			for i := range e {
				if !strings.EqualFold(e[i], a[i]) {
					return false
				}
			}
			return true
		case []any:
			// Handle []string vs []any
			if len(e) != len(a) {
				return false
			}
			for i := range e {
				if !deepEqual(e[i], a[i]) {
					return false
				}
			}
			return true
		}
	}

	return reflect.DeepEqual(expected, actual)
}
