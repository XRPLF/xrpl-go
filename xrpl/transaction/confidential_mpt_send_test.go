package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID":          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"SenderEncryptedAmount":      "AA11",
				"DestinationEncryptedAmount": "BB22",
				"IssuerEncryptedAmount":      "CC33",
				"ZKProof":                    "DD44",
				"BalanceCommitment":          "EE55",
				"AmountCommitment":           "FF66",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
				AuditorEncryptedAmount:     types.HexBlob("7788"),
				CredentialIDs:              types.CredentialIDs{"AABBCCDD"},
			},
			expected: FlatTransaction{
				"Account":                    "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                        "12",
				"TransactionType":            "ConfidentialMPTSend",
				"Destination":                "rDgHn3T2P7eNAaoHh43iRudhAUjAHmDgEP",
				"MPTokenIssuanceID":          "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"SenderEncryptedAmount":      "AA11",
				"DestinationEncryptedAmount": "BB22",
				"IssuerEncryptedAmount":      "CC33",
				"ZKProof":                    "DD44",
				"BalanceCommitment":          "EE55",
				"AmountCommitment":           "FF66",
				"AuditorEncryptedAmount":     "7788",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "",
				AmountCommitment:           "FF66",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "",
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
				AuditorEncryptedAmount:     types.HexBlob("not-hex!"),
			},
			wantErr: ErrConfidentialSendMissingFields,
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
				SenderEncryptedAmount:      "AA11",
				DestinationEncryptedAmount: "BB22",
				IssuerEncryptedAmount:      "CC33",
				ZKProof:                    "DD44",
				BalanceCommitment:          "EE55",
				AmountCommitment:           "FF66",
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
