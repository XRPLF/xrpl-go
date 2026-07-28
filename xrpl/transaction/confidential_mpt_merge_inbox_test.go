package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTMergeInbox_TxType(t *testing.T) {
	tx := &ConfidentialMPTMergeInbox{}
	require.Equal(t, ConfidentialMPTMergeInboxTx, tx.TxType())
}

func TestConfidentialMPTMergeInbox_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTMergeInbox
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			expected: FlatTransaction{
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":               "12",
				"TransactionType":   "ConfidentialMPTMergeInbox",
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

func TestConfidentialMPTMergeInbox_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTMergeInbox
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTMergeInboxTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTMergeInbox{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTMergeInboxTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID: "",
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
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
