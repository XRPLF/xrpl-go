package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestMPTokenIssuanceDestroy_TxType(t *testing.T) {
	tx := &MPTokenIssuanceDestroy{}
	require.Equal(t, MPTokenIssuanceDestroyTx, tx.TxType())
}

func TestMPTokenIssuanceDestroy_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *MPTokenIssuanceDestroy
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			tx: &MPTokenIssuanceDestroy{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			expected: FlatTransaction{
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":               "12",
				"TransactionType":   "MPTokenIssuanceDestroy",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattened := tt.tx.Flatten()
			require.Equal(t, tt.expected, flattened)
		})
	}
}

func TestMPTokenIssuanceDestroy_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *MPTokenIssuanceDestroy
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &MPTokenIssuanceDestroy{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceDestroyTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &MPTokenIssuanceDestroy{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: MPTokenIssuanceDestroyTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "",
			},
			wantErr: ErrInvalidMPTokenIssuanceID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.tx.Validate()
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
