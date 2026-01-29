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
