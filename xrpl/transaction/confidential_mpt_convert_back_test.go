package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// Test helpers for ConfidentialMPTConvertBack.
var (
	testConvertBackCiphertext1 = strings.Repeat("A1", 66)
	testConvertBackCiphertext2 = strings.Repeat("B2", 66)
	testConvertBackCiphertext3 = strings.Repeat("C3", 66)
	testConvertBackCommitment  = strings.Repeat("D4", 33)
	testConvertBackBF          = strings.Repeat("EF", 32)
)

func TestConfidentialMPTConvertBack_TxType(t *testing.T) {
	tx := &ConfidentialMPTConvertBack{}
	require.Equal(t, ConfidentialMPTConvertBackTx, tx.TxType())
}

func TestConfidentialMPTConvertBack_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTConvertBack
		expected FlatTransaction
	}{
		{
			name: "pass - without optional fields",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			expected: FlatTransaction{
				"Account":               "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                   "12",
				"TransactionType":       "ConfidentialMPTConvertBack",
				"MPTokenIssuanceID":     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTAmount":             "1000",
				"HolderEncryptedAmount": testConvertBackCiphertext1,
				"IssuerEncryptedAmount": testConvertBackCiphertext2,
				"BlindingFactor":        testConvertBackBF,
				"BalanceCommitment":     testConvertBackCommitment,
				"ZKProof":               "1122",
			},
		},
		{
			name: "pass - with auditor encrypted amount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:              types.MPTPlainAmount(500),
				HolderEncryptedAmount:  testConvertBackCiphertext1,
				IssuerEncryptedAmount:  testConvertBackCiphertext2,
				BlindingFactor:         testConvertBackBF,
				AuditorEncryptedAmount: types.HexBlob(testConvertBackCiphertext3),
				BalanceCommitment:      testConvertBackCommitment,
				ZKProof:                "1122",
			},
			expected: FlatTransaction{
				"Account":                "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                    "12",
				"TransactionType":        "ConfidentialMPTConvertBack",
				"MPTokenIssuanceID":      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTAmount":              "500",
				"HolderEncryptedAmount":  testConvertBackCiphertext1,
				"IssuerEncryptedAmount":  testConvertBackCiphertext2,
				"BlindingFactor":         testConvertBackBF,
				"AuditorEncryptedAmount": testConvertBackCiphertext3,
				"BalanceCommitment":      testConvertBackCommitment,
				"ZKProof":                "1122",
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

func TestConfidentialMPTConvertBack_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTConvertBack
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
		{
			name: "fail - zero MPTAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(0),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidAmount,
		},
		{
			name: "fail - invalid blinding factor",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        "short",
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidBlindingFactor,
		},
		{
			name: "fail - empty HolderEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "",
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - empty IssuerEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: "",
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - empty ZKProof",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "",
			},
			wantErr: ErrConfidentialConvertBackInvalidProof,
		},
		{
			name: "fail - invalid AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testConvertBackCiphertext1,
				IssuerEncryptedAmount:  testConvertBackCiphertext2,
				BlindingFactor:         testConvertBackBF,
				AuditorEncryptedAmount: types.HexBlob("not-hex!"),
				BalanceCommitment:      testConvertBackCommitment,
				ZKProof:                "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - empty BalanceCommitment",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     "",
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCommitment,
		},
		{
			name: "fail - wrong length HolderEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "AABB",
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - wrong length IssuerEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: "CCDD",
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     testConvertBackCommitment,
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - wrong length AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testConvertBackCiphertext1,
				IssuerEncryptedAmount:  testConvertBackCiphertext2,
				BlindingFactor:         testConvertBackBF,
				AuditorEncryptedAmount: types.HexBlob("AABB"),
				BalanceCommitment:      testConvertBackCommitment,
				ZKProof:                "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - wrong length BalanceCommitment",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testConvertBackCiphertext1,
				IssuerEncryptedAmount: testConvertBackCiphertext2,
				BlindingFactor:        testConvertBackBF,
				BalanceCommitment:     "EEFF",
				ZKProof:               "1122",
			},
			wantErr: ErrConfidentialConvertBackInvalidCommitment,
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
