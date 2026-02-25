package transaction

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestAMMWithdraw_TxType(t *testing.T) {
	tx := &AMMWithdraw{}
	assert.Equal(t, AMMWithdrawTx, tx.TxType())
}

func TestAMMWithdraw_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AMMWithdraw)
		expected uint32
	}{
		{
			name: "pass - SetLPTokentFlag",
			setter: func(a *AMMWithdraw) {
				a.SetLPTokentFlag()
			},
			expected: TfLPToken,
		},
		{
			name: "pass - SetWithdrawAllFlag",
			setter: func(a *AMMWithdraw) {
				a.SetWithdrawAllFlag()
			},
			expected: TfWithdrawAll,
		},
		{
			name: "pass - SetOneAssetWithdrawAllFlag",
			setter: func(a *AMMWithdraw) {
				a.SetOneAssetWithdrawAllFlag()
			},
			expected: TfOneAssetWithdrawAll,
		},
		{
			name: "pass - SetSingleAssetFlag",
			setter: func(a *AMMWithdraw) {
				a.SetSingleAssetFlag()
			},
			expected: TfSingleAsset,
		},
		{
			name: "pass - SetTwoAssetFlag",
			setter: func(a *AMMWithdraw) {
				a.SetTwoAssetFlag()
			},
			expected: TfTwoAsset,
		},
		{
			name: "pass - SetOneAssetLPTokenFlag",
			setter: func(a *AMMWithdraw) {
				a.SetOneAssetLPTokenFlag()
			},
			expected: TfOneAssetLPToken,
		},
		{
			name: "pass - SetLimitLPTokenFlag",
			setter: func(a *AMMWithdraw) {
				a.SetLimitLPTokenFlag()
			},
			expected: TfLimitLPToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AMMWithdraw{}
			tt.setter(a)
			if a.Flags != tt.expected {
				t.Errorf("Expected AMMWithdraw Flags to be %d, got %d", tt.expected, a.Flags)
			}
		})
	}
}

func TestAMMWithdraw_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *AMMWithdraw
		expected string
	}{
		{
			name: "pass - Full AMMWithdraw",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "TST",
					Issuer:   "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv",
				},
				LPTokenIn: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr",
				},
			},
			expected: `{
				"TransactionType": "AMMWithdraw",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "10",
				"Flags": 1048576,
				"Sequence": 10,
				"Asset": {
					"currency": "TST",
					"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Asset2": {
					"currency": "XRP"
				},
				"Amount": {
					"value":    "5",
					"currency": "TST",
					"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Amount2": {
					"value": "50000000",
					"currency": "ABC",
					"issuer": "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur"
				},
				"EPrice": {
					"value": "1",
					"currency": "TST",
					"issuer": "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv"
				},
				"LPTokenIn": {
					"value": "100",
					"currency": "TST",
					"issuer": "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr"
				}
			}`,
		},
		{
			name: "pass - Minimal AMMWithdraw",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
			},
			expected: `{
				"TransactionType": "AMMWithdraw",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee": "10",
				"Flags": 1048576,
				"Sequence": 10,
				"Asset": {
					"currency": "TST",
					"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
				},
				"Asset2": {
					"currency": "XRP"
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()

			err := testutil.CompareFlattenAndExpected(result, []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}
func TestAMMWithdraw_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    *AMMWithdraw
		expected bool
	}{
		{
			name: "pass - valid AMMWithdraw",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "TST",
					Issuer:   "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv",
				},
				LPTokenIn: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr",
				},
			},
			expected: true,
		},
		{
			name: "fail - invalid AMMWithdraw BaseTx, TransactionType missing",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:      types.XRPCurrencyAmount(10),
					Flags:    1048576,
					Sequence: 10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "TST",
					Issuer:   "rJhPKEN1m6FDGy9FZ85Ek2n3tAyUBR4KBv",
				},
				LPTokenIn: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rQH2Rhja1YRC3spuVukZBu9WzRiD1R9Dcr",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid Asset",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
			},
			expected: false,
		},
		{
			name: "fail - Amount2 without Amount",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "UST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
			},
			expected: false,
		},
		{
			name: "fail - EPrice set without Amount",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "UST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "50000000",
					Currency: "ABC",
					Issuer:   "rKswUGcm3wXPSaMfEHdUAQvE8otkZCd1ur",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid Asset2",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "USDC",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid Amount",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid Amount2",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "10",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:  "",
					Issuer: "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid EPrice",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "10",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "12",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:  "12",
					Issuer: "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid LPTokenIn",
			input: &AMMWithdraw{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMWithdraw",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        10,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "10",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "12",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "12",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				LPTokenIn: types.IssuedCurrencyAmount{
					Value:  "12",
					Issuer: "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			if valid != tt.expected {
				t.Errorf("Expected validation result to be %v, got %v", tt.expected, valid)
			}
			if (err != nil) != !tt.expected {
				t.Errorf("Expected error presence to be %v, got %v", !tt.expected, err != nil)
			}
		})
	}
}
