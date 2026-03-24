package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// Test helpers for ConfidentialMPTSend.
var (
	testSendCiphertext1 = strings.Repeat("A1", 66)
	testSendCiphertext2 = strings.Repeat("B2", 66)
	testSendCiphertext3 = strings.Repeat("C3", 66)
	testSendCiphertext4 = strings.Repeat("D4", 66)
	testSendCommitment1 = strings.Repeat("E5", 33)
	testSendCommitment2 = strings.Repeat("F6", 33)
)

func TestConfidentialMPTSend_TxType(t *testing.T) {
	tx := &ConfidentialMPTSend{}
	require.Equal(t, ConfidentialMPTSendTx, tx.TxType())
}

func TestConfidentialMPTSend_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTSend
		expected FlatTransaction
	}{
		{
			name: "pass - without optional fields",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID":          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"SenderEncryptedAmount":      testSendCiphertext1,
				"DestinationEncryptedAmount": testSendCiphertext2,
				"IssuerEncryptedAmount":      testSendCiphertext3,
				"ZKProof":                    "DD44",
				"BalanceCommitment":          testSendCommitment1,
				"AmountCommitment":           testSendCommitment2,
			},
		},
		{
			name: "pass - with optional fields",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
				AuditorEncryptedAmount:     types.HexBlob(testSendCiphertext4),
				CredentialIDs:              types.CredentialIDs{"AABBCCDD"},
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID":          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"SenderEncryptedAmount":      testSendCiphertext1,
				"DestinationEncryptedAmount": testSendCiphertext2,
				"IssuerEncryptedAmount":      testSendCiphertext3,
				"ZKProof":                    "DD44",
				"BalanceCommitment":          testSendCommitment1,
				"AmountCommitment":           testSendCommitment2,
				"AuditorEncryptedAmount":     testSendCiphertext4,
				"CredentialIDs":              []string{"AABBCCDD"},
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

func TestConfidentialMPTSend_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTSend
		wantErr error
	}{
		{
			name: "pass - valid transaction",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: nil,
		},
		{
			name: "pass - with credential IDs",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
				CredentialIDs:              types.CredentialIDs{"AABBCCDD"},
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
		{
			name: "fail - invalid Destination address",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "invalidAddress",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidDestination,
		},
		{
			name: "fail - Destination same as Account",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendSelfSend,
		},
		{
			name: "fail - empty SenderEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      "",
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - empty DestinationEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: "",
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - empty IssuerEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      "",
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - empty ZKProof",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidProof,
		},
		{
			name: "fail - empty BalanceCommitment",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          "",
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCommitment,
		},
		{
			name: "fail - empty AmountCommitment",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           "",
			},
			wantErr: ErrConfidentialSendInvalidCommitment,
		},
		{
			name: "fail - invalid AuditorEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
				AuditorEncryptedAmount:     types.HexBlob("not-hex!"),
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - wrong length SenderEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      "AABB",
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - wrong length DestinationEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: "BBCC",
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - wrong length AuditorEncryptedAmount",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
				AuditorEncryptedAmount:     types.HexBlob("AABB"),
			},
			wantErr: ErrConfidentialSendInvalidCiphertext,
		},
		{
			name: "fail - wrong length BalanceCommitment",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          "EEFF",
				AmountCommitment:           testSendCommitment2,
			},
			wantErr: ErrConfidentialSendInvalidCommitment,
		},
		{
			name: "fail - wrong length AmountCommitment",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           "FF11",
			},
			wantErr: ErrConfidentialSendInvalidCommitment,
		},
		{
			name: "fail - invalid CredentialIDs",
			tx: &ConfidentialMPTSend{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTSendTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				Destination:                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				MPTokenIssuanceID:          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				SenderEncryptedAmount:      testSendCiphertext1,
				DestinationEncryptedAmount: testSendCiphertext2,
				IssuerEncryptedAmount:      testSendCiphertext3,
				ZKProof:                    "DD44",
				BalanceCommitment:          testSendCommitment1,
				AmountCommitment:           testSendCommitment2,
				CredentialIDs:              types.CredentialIDs{"not-hex!"},
			},
			wantErr: ErrInvalidCredentialIDs,
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
