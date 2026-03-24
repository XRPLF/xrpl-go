package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTClawback_TxType(t *testing.T) {
	tx := &ConfidentialMPTClawback{}
	require.Equal(t, ConfidentialMPTClawbackTx, tx.TxType())
}

func TestConfidentialMPTClawback_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTClawback
		expected FlatTransaction
	}{
		{
			name: "pass - all fields",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "AABBCCDD",
			},
			expected: FlatTransaction{
				"Account":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":               "12",
				"TransactionType":   "ConfidentialMPTClawback",
				"Holder":            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID": "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTAmount":         "1000",
				"ZKProof":           "AABBCCDD",
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

func TestConfidentialMPTClawback_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTClawback
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "AABBCCDD",
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "AABBCCDD",
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
		{
			name: "fail - invalid Holder address",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "invalidAddress",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "AABBCCDD",
			},
			wantErr: ErrConfidentialClawbackInvalidHolder,
		},
		{
			name: "fail - Holder same as Account",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "AABBCCDD",
			},
			wantErr: ErrConfidentialClawbackSelfClawback,
		},
		{
			name: "fail - zero MPTAmount",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(0),
				ZKProof:           "AABBCCDD",
			},
			wantErr: ErrConfidentialClawbackInvalidAmount,
		},
		{
			name: "fail - empty ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "",
			},
			wantErr: ErrConfidentialClawbackBadProof,
		},
		{
			name: "fail - invalid hex ZKProof",
			tx: &ConfidentialMPTClawback{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTClawbackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Holder:            "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID: "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:         types.MPTPlainAmount(1000),
				ZKProof:           "ZZZZ",
			},
			wantErr: ErrConfidentialClawbackBadProof,
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
