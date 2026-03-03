package transaction

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestIsSigner(t *testing.T) {
	tests := []struct {
		name     string
		input    types.SignerData
		expected bool
	}{
		{
			name: "pass - valid Signer object",
			input: types.SignerData{
				Account:       "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				TxnSignature:  "0123456789abcdef",
				SigningPubKey: "abcdef0123456789",
			},
			expected: true,
		},
		{
			name: "fail - Signer object with missing fields",
			input: types.SignerData{
				Account:       "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				SigningPubKey: "abcdef0123456789",
			},
			expected: false,
		},
		{
			name: "fail - invalid Signer object with empty XRPL account",
			input: types.SignerData{
				Account:       "  ",
				SigningPubKey: "abcdef0123456789",
				TxnSignature:  "0123456789abcdef",
			},
			expected: false,
		},
		{
			name: "fail - invalid Signer object with invalid XRPL account",
			input: types.SignerData{
				Account:       "invalid",
				SigningPubKey: "abcdef0123456789",
				TxnSignature:  "0123456789abcdef",
			},
			expected: false,
		},
		{
			name: "fail - invalid Signer object with empty TxnSignature",
			input: types.SignerData{
				Account:       "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				TxnSignature:  "  ",
				SigningPubKey: "abcdef0123456789",
			},
			expected: false,
		},
		{
			name: "fail - invalid Signer object with empty SigningPubKey",
			input: types.SignerData{
				Account:       "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				TxnSignature:  "0123456789abcdef",
				SigningPubKey: "  ",
			},
			expected: false,
		},
		{
			name:     "fail - nil object",
			input:    types.SignerData{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsSigner(tt.input); ok != tt.expected {
				t.Errorf("Expected IsSigner to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}
func TestIsAmount(t *testing.T) {
	tests := []struct {
		name            string
		input           types.CurrencyAmount
		fieldName       string
		isFieldRequired bool
		expected        bool
	}{
		{
			name:            "pass - valid XRP amount",
			input:           types.XRPCurrencyAmount(1000000),
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        true,
		},
		{
			name: "pass - valid IssuedCurrency amount",
			input: types.IssuedCurrencyAmount{
				Value:    "100",
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "USD",
			},
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        true,
		},
		{
			name: "pass - valid MPT amount",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
				Value:         "100",
			},
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        true,
		},
		{
			name:            "fail - required field is nil",
			input:           nil,
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        false,
		},
		{
			name:            "pass - optional field is nil",
			input:           nil,
			fieldName:       "Amount",
			isFieldRequired: false,
			expected:        true,
		},
		{
			name: "fail - invalid MPT amount with non-hex issuance ID",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "not-hex",
				Value:         "100",
			},
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        false,
		},
		{
			name: "fail - invalid MPT amount with missing value",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
			},
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        false,
		},
		{
			name: "fail - invalid IssuedCurrency with XRP currency",
			input: types.IssuedCurrencyAmount{
				Value:    "100",
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "XRP",
			},
			fieldName:       "Amount",
			isFieldRequired: true,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsAmount(tt.input, tt.fieldName, tt.isFieldRequired); ok != tt.expected {
				t.Errorf("Expected IsAmount to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}

func TestIsIssuedCurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    types.CurrencyAmount
		expected bool
	}{
		{
			name: "pass - valid IssuedCurrency object",
			input: types.IssuedCurrencyAmount{
				Value:    "100",
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "USD",
			},
			expected: true,
		},
		{
			name:     "fail - invalid IssuedCurrency object",
			input:    types.XRPCurrencyAmount(100), // should be non XRP
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with missing currency and issuer fields",
			input: types.IssuedCurrencyAmount{
				Value: "100",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with missing issuer and value fields",
			input: types.IssuedCurrencyAmount{
				Currency: "USD",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with missing currency and value fields",
			input: types.IssuedCurrencyAmount{
				Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with empty currency",
			input: types.IssuedCurrencyAmount{
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "   ",
				Value:    "100",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with XRP currency",
			input: types.IssuedCurrencyAmount{
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "XRp", // will be uppercased during validation
				Value:    "100",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with empty value",
			input: types.IssuedCurrencyAmount{
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "USD",
				Value:    "  ",
			},
			expected: false,
		},
		{
			name: "fail - issuedCurrency object with invalid issuer",
			input: types.IssuedCurrencyAmount{
				Issuer:   "invalid",
				Currency: "USD",
				Value:    "100",
			},
			expected: false,
		},
		{
			name:     "fail - empty object",
			input:    types.IssuedCurrencyAmount{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsIssuedCurrency(tt.input); ok != tt.expected {
				t.Errorf("Expected IsIssuedCurrency to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}

func TestIsMPTCurrency(t *testing.T) {
	tests := []struct {
		name     string
		input    types.CurrencyAmount
		expected bool
	}{
		{
			name: "pass - valid MPTCurrencyAmount",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
				Value:         "100",
			},
			expected: true,
		},
		{
			name:     "fail - non-MPT type (XRP)",
			input:    types.XRPCurrencyAmount(100),
			expected: false,
		},
		{
			name: "fail - non-MPT type (IssuedCurrency)",
			input: types.IssuedCurrencyAmount{
				Value:    "100",
				Issuer:   "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW",
				Currency: "USD",
			},
			expected: false,
		},
		{
			name: "fail - missing MPTIssuanceID",
			input: types.MPTCurrencyAmount{
				Value: "100",
			},
			expected: false,
		},
		{
			name: "fail - non-hex MPTIssuanceID",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "not-a-hex-string",
				Value:         "100",
			},
			expected: false,
		},
		{
			name: "fail - empty value",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
			},
			expected: false,
		},
		{
			name: "fail - negative value",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
				Value:         "-5",
			},
			expected: false,
		},
		{
			name: "fail - fractional value",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
				Value:         "10.5",
			},
			expected: false,
		},
		{
			name: "fail - value exceeds max int64",
			input: types.MPTCurrencyAmount{
				MPTIssuanceID: "00000001A407AF5856CEF3379FAB85D584F3AA7C0E8B8C4A",
				Value:         "9223372036854775808",
			},
			expected: false,
		},
		{
			name:     "fail - empty object",
			input:    types.MPTCurrencyAmount{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsMPTCurrency(tt.input); ok != tt.expected {
				t.Errorf("Expected IsMPTCurrency to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}

func TestIsMemo(t *testing.T) {
	t.Run("pass - valid Memo object with all fields", func(t *testing.T) {
		obj := types.Memo{
			MemoData:   "0123456789abcdef",
			MemoFormat: "abcdef0123456789",
			MemoType:   "abcdef0123456789",
		}

		ok, _ := IsMemo(obj)

		if !(ok) {
			t.Errorf("Expected IsMemo to return true, but got false")
		}
	})

	t.Run("pass - valid memo object with missing fields", func(t *testing.T) {
		obj := types.Memo{
			MemoData: "0123456789abcdef",
		}

		ok, err := IsMemo(obj)

		if !ok {
			t.Errorf("Expected IsMemo to return true, but got false with error: %v", err)
		}
	})

	t.Run("fail - memo object with MemoData non hex value", func(t *testing.T) {
		obj := types.Memo{
			MemoData: "bob",
		}

		if ok, _ := IsMemo(obj); ok {
			t.Errorf("Expected IsMemo to return false, but got true")
		}
	})

	t.Run("fail - memo object with MemoFormat non hex value", func(t *testing.T) {
		obj := types.Memo{
			MemoData:   "0123456789abcdef",
			MemoFormat: "non-hex",
		}

		if ok, _ := IsMemo(obj); ok {
			t.Errorf("Expected IsMemo to return false, but got true")
		}
	})

	t.Run("fail - memo object with MemoType non hex value", func(t *testing.T) {
		obj := types.Memo{
			MemoData:   "0123456789abcdef",
			MemoFormat: "0123456789abcdef",
			MemoType:   "non-hex",
		}

		if ok, _ := IsMemo(obj); ok {
			t.Errorf("Expected IsMemo to return false, but got true")
		}
	})

	t.Run("fail - empty object", func(t *testing.T) {
		obj := types.Memo{}
		if ok, _ := IsMemo(obj); ok {
			t.Errorf("Expected IsMemo to return false, but got true")
		}
	})
}
func TestIsAsset(t *testing.T) {
	t.Run("pass - valid Asset object with currency XRP only", func(t *testing.T) {
		obj := ledger.Asset{
			Currency: "xrP", // will be converted to XRP in the Validate function
		}

		ok, err := IsAsset(obj)

		if !ok {
			t.Errorf("Expected IsAsset to return true, but got false with error: %v", err)
		}
	})

	t.Run("fail - invalid Asset object with currency XRP and an issuer defined", func(t *testing.T) {
		obj := ledger.Asset{
			Currency: "xrP", // will be converted to XRP in the Validate function
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}

		ok, err := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return true, but got false with error: %v", err)
		}
	})

	t.Run("fail - invalid Asset object with currency only and different than XRP", func(t *testing.T) {
		obj := ledger.Asset{
			Currency: "USD", // missing issuer
		}

		ok, err := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return true, but got false with error: %v", err)
		}
	})

	t.Run("pass - valid Asset object with currency and issuer", func(t *testing.T) {
		obj := ledger.Asset{
			Currency: "USD",
			Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}

		ok, err := IsAsset(obj)

		if !ok {
			t.Errorf("Expected IsAsset to return true, but got false with error: %v", err)
		}
	})

	t.Run("fail - Asset object with missing currency", func(t *testing.T) {
		obj := ledger.Asset{
			Issuer: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}

		ok, err := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return false, but got true")
		} else if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	})

	t.Run("fail - empty Asset object", func(t *testing.T) {
		obj := ledger.Asset{}

		ok, err := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return false, but got true")
		} else if err == nil {
			t.Errorf("Expected an error, but got nil")
		}
	})

	t.Run("pass - valid MPT asset", func(t *testing.T) {
		obj := ledger.Asset{
			MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
		}

		ok, err := IsAsset(obj)

		if !ok {
			t.Errorf("Expected IsAsset to return true, but got false with error: %v", err)
		}
	})

	t.Run("fail - MPT asset with currency set", func(t *testing.T) {
		obj := ledger.Asset{
			MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
			Currency:      "USD",
		}

		ok, _ := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return false, but got true")
		}
	})

	t.Run("fail - MPT asset with issuer set", func(t *testing.T) {
		obj := ledger.Asset{
			MPTIssuanceID: "983F536DBB46D5BBF43A0B5890576874EE1CF48CE31CA508A529EC17CD1A90EF",
			Issuer:        "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		}

		ok, _ := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return false, but got true")
		}
	})

	t.Run("fail - MPT asset with non-hex ID", func(t *testing.T) {
		obj := ledger.Asset{
			MPTIssuanceID: "not-a-hex-string",
		}

		ok, _ := IsAsset(obj)

		if ok {
			t.Errorf("Expected IsAsset to return false, but got true")
		}
	})
}
func TestIsPath(t *testing.T) {
	tests := []struct {
		name     string
		input    []PathStep
		expected bool
	}{
		{
			name: "pass - valid path with account only",
			input: []PathStep{
				{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
			},
			expected: true,
		},
		{
			name: "pass - valid path with currency only",
			input: []PathStep{
				{Currency: "USD"},
			},
			expected: true,
		},
		{
			name: "pass - valid path with issuer only",
			input: []PathStep{
				{Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
			},
			expected: true,
		},
		{
			name: "pass - valid path with currency and issuer",
			input: []PathStep{
				{Currency: "USD", Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
			},
			expected: true,
		},
		{
			name: "fail - invalid path with account and currency",
			input: []PathStep{
				{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW", Currency: "USD"},
			},
			expected: false,
		},
		{
			name: "fail - invalid path with account and issuer",
			input: []PathStep{
				{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW", Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
			},
			expected: false,
		},
		{
			name: "fail - invalid path with currency XRP and issuer",
			input: []PathStep{
				{Currency: "XRP", Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
			},
			expected: false,
		},
		{
			name:     "fail - empty path",
			input:    []PathStep{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsPath(tt.input); ok != tt.expected {
				t.Errorf("Expected IsPath to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}
func TestIsPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]PathStep
		expected bool
	}{
		{
			name: "pass - valid paths with single path and single step",
			input: [][]PathStep{
				{
					{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
				},
			},
			expected: true,
		},
		{
			name: "pass - valid paths with multiple paths and steps",
			input: [][]PathStep{
				{
					{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
					{Currency: "USD"},
				},
				{
					{Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
					{Currency: "EUR", Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
				},
			},
			expected: true,
		},
		{
			name: "fail - invalid paths with empty path",
			input: [][]PathStep{
				{},
			},
			expected: false,
		},
		{
			name: "fail - invalid paths with empty path step",
			input: [][]PathStep{
				{
					{},
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid paths with invalid path step, account and currency cannot be together",
			input: [][]PathStep{
				{
					{Account: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW", Currency: "USD"},
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid paths with invalid path step having currency XRP and issuer",
			input: [][]PathStep{
				{
					{Currency: "XRP", Issuer: "r4ES5Mmnz4HGbu2asdicuECBaBWo4knhXW"},
				},
			},
			expected: false,
		},
		{
			name:     "fail - empty paths",
			input:    [][]PathStep{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok, err := IsPaths(tt.input); ok != tt.expected {
				t.Errorf("Expected IsPaths to return %v, but got %v with error: %v", tt.expected, ok, err)
			}
		})
	}
}

func TestIsDomainID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "pass - valid 64 character DomainID",
			input:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expected: true,
		},
		{
			name:     "fail - too short DomainID",
			input:    "1234567890abcdef",
			expected: false,
		},
		{
			name:     "fail - too long DomainID",
			input:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef12",
			expected: false,
		},
		{
			name:     "fail - empty DomainID",
			input:    "",
			expected: false,
		},
		{
			name:     "pass - valid DomainID with all uppercase hex",
			input:    "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
			expected: true,
		},
		{
			name:     "pass - valid DomainID with mixed case hex",
			input:    "1234567890abcDEF1234567890ABcdef1234567890ABcdef1234567890ABcdef",
			expected: true,
		},
		{
			name:     "pass - valid DomainID with all numbers",
			input:    "1234567890123456789012345678901234567890123456789012345678901234",
			expected: true,
		},
		{
			name:     "pass - valid DomainID with all letters",
			input:    "abcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := IsDomainID(tt.input); result != tt.expected {
				t.Errorf("Expected IsDomainID to return %v, but got %v", tt.expected, result)
			}
		})
	}
}
