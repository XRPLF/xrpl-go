package transaction

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestAMMDeposit_TxType(t *testing.T) {
	tx := &AMMDeposit{}
	assert.Equal(t, AMMDepositTx, tx.TxType())
}
func TestAMMDeposit_Flatten(t *testing.T) {
	tx := &AMMDeposit{
		BaseTx: BaseTx{
			Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			Fee:      types.XRPCurrencyAmount(10),
			Flags:    1048576,
			Sequence: 7,
		},
		Asset: ledger.Asset{
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		Asset2: ledger.Asset{
			Currency: "XRP",
		},
		Amount: types.IssuedCurrencyAmount{
			Value:    "2.5",
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		Amount2: types.XRPCurrencyAmount(30000000),
		EPrice: types.IssuedCurrencyAmount{
			Value:    "1.5",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
			Currency: "TST",
		},
		LPTokenOut: types.IssuedCurrencyAmount{
			Value:    "100",
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		TradingFee: 10,
	}

	flattened := tx.Flatten()

	expected := `{
	"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Fee":             "10",
	"Flags":           1048576,
	"Sequence":        7,
	"TransactionType": "AMMDeposit",
	"Asset": {
		"currency": "TST",
		"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	},
	"Asset2": {
		"currency": "XRP"
	},
	"Amount": {
		"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		"value":    "2.5",
		"currency": "TST"
	},
	"Amount2": "30000000",
	"EPrice": {
		"value": "1.5",
		"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		"currency": "TST"
	},
	"LPTokenOut": {
		"value": "100",
		"currency": "TST",
		"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	},
	"TradingFee": 10
}`

	err := testutil.CompareFlattenAndExpected(flattened, []byte(expected))
	if err != nil {
		t.Error(err)
	}
}

func TestAMMDeposit_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*AMMDeposit)
		expected uint32
	}{
		{
			name: "pass - SetLPTokentFlag",
			setter: func(a *AMMDeposit) {
				a.SetLPTokentFlag()
			},
			expected: TfLPToken,
		},
		{
			name: "pass - SetSingleAssetFlag",
			setter: func(a *AMMDeposit) {
				a.SetSingleAssetFlag()
			},
			expected: TfSingleAsset,
		},
		{
			name: "pass - SetTwoAssetFlag",
			setter: func(a *AMMDeposit) {
				a.SetTwoAssetFlag()
			},
			expected: TfTwoAsset,
		},
		{
			name: "pass - SetOneAssetLPTokenFlag",
			setter: func(a *AMMDeposit) {
				a.SetOneAssetLPTokenFlag()
			},
			expected: TfOneAssetLPToken,
		},
		{
			name: "pass - SetLimitLPTokenFlag",
			setter: func(a *AMMDeposit) {
				a.SetLimitLPTokenFlag()
			},
			expected: TfLimitLPToken,
		},
		{
			name: "pass - SetTwoAssetIfEmptyFlag",
			setter: func(a *AMMDeposit) {
				a.SetTwoAssetIfEmptyFlag()
			},
			expected: TfTwoAssetIfEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AMMDeposit{}
			tt.setter(a)
			if a.Flags != tt.expected {
				t.Errorf("Expected AMMDeposit Flags to be %d, got %d", tt.expected, a.Flags)
			}
		})
	}
}
func TestAMMDeposit_Validate(t *testing.T) {
	tests := []struct {
		name     string
		tx       *AMMDeposit
		expected bool
	}{
		{
			name: "pass - valid AMMDeposit with Amount and Amount2",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "2.5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.XRPCurrencyAmount(30000000),
			},
			expected: true,
		},
		{
			name: "fail - invalid AMMDeposit BaseTx without TransactionType",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:      types.XRPCurrencyAmount(10),
					Flags:    1048576,
					Sequence: 7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "2.5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.XRPCurrencyAmount(30000000),
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit with Amount2 but no Amount",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				Amount2: types.XRPCurrencyAmount(30000000),
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit with EPrice but no Amount",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value:    "1.5",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit with no LPTokenOut or Amount",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
			},
			expected: false,
		},
		{
			name: "pass - valid AMMDeposit with LPTokenOut",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: true,
		},
		{
			name: "fail - invalid AMMDeposit, invalid Asset",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit, invalid Asset2",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit, invalid Amount",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount: types.IssuedCurrencyAmount{
					Value: "1",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit, invalid Amount2",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "USD",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value: "1",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid AMMDeposit, invalid EPrice",
			tx: &AMMDeposit{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMDeposit",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           1048576,
					Sequence:        7,
				},
				Asset: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Asset2: ledger.Asset{
					Currency: "XRP",
				},
				LPTokenOut: types.IssuedCurrencyAmount{
					Value:    "100",
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "USD",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				Amount2: types.IssuedCurrencyAmount{
					Value:    "1",
					Currency: "USD",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				EPrice: types.IssuedCurrencyAmount{
					Value: "1",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if valid != tt.expected {
				t.Errorf("Expected validation result to be %v, got %v", tt.expected, valid)
			}
			if err != nil && tt.expected {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}
