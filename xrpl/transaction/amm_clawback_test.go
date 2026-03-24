package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestAMMClawback_TxType(t *testing.T) {
	tx := &AMMClawback{}
	require.Equal(t, AMMClawbackTx, tx.TxType())
}

func TestAMMClawback_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *AMMClawback
		expected FlatTransaction
	}{
		{
			name: "pass - basic",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
			},
		},
		{
			name: "pass - with holder",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
				Holder: "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
				"Holder":          "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
			},
		},
		{
			name: "pass - with asset",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
				Asset: types.IssuedCurrency{
					Currency: "USD",
					Issuer:   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
				},
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
				"Asset": map[string]any{
					"currency": "USD",
					"issuer":   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
				},
			},
		},
		{
			name: "pass - with asset2 as xrp currency amount",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
				Asset2: types.XRPCurrencyAmount(100),
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
				"Asset2":          "100",
			},
		},
		{
			name: "pass - with asset2 as issued currency amount",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
				Asset2: types.IssuedCurrencyAmount{
					Issuer:   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
					Currency: "USD",
					Value:    "100",
				},
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
				"Asset2": map[string]any{
					"currency": "USD",
					"issuer":   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
					"value":    "100",
				},
			},
		},
		{
			name: "pass - with amount",
			tx: &AMMClawback{
				BaseTx: BaseTx{
					TransactionType: AMMClawbackTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
					Currency: "USD",
					Value:    "100",
				},
			},
			expected: FlatTransaction{
				"TransactionType": AMMClawbackTx.String(),
				"Amount": map[string]any{
					"currency": "USD",
					"issuer":   "rWYkbWkCeg8dP6rXALnjgZSjjLyih5NXm",
					"value":    "100",
				},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}
