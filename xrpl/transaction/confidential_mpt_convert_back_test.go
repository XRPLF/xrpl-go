package transaction

import (
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// testConvertBackProof is a well-formed ConvertBack proof bundle of the required length.
var testConvertBackProof = strings.Repeat("12", ConvertBackProofLen/2)

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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
			},
			expected: FlatTransaction{
				"Account":               "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                   "12",
				"TransactionType":       "ConfidentialMPTConvertBack",
				"MPTokenIssuanceID":     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"MPTAmount":             "1000",
				"HolderEncryptedAmount": testCiphertext,
				"IssuerEncryptedAmount": testCiphertext2,
				"BlindingFactor":        testBlindingFactor,
				"BalanceCommitment":     testCompressedPoint2,
				"ZKProof":               testConvertBackProof,
			},
		},
		{
			name: "pass - with auditor encrypted amount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(500),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				BlindingFactor:         testBlindingFactor,
				AuditorEncryptedAmount: types.HexBlob(testCiphertext3),
				BalanceCommitment:      testCompressedPoint2,
				ZKProof:                testConvertBackProof,
			},
			expected: FlatTransaction{
				"Account":                "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                    "12",
				"TransactionType":        "ConfidentialMPTConvertBack",
				"MPTokenIssuanceID":      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"MPTAmount":              "500",
				"HolderEncryptedAmount":  testCiphertext,
				"IssuerEncryptedAmount":  testCiphertext2,
				"BlindingFactor":         testBlindingFactor,
				"AuditorEncryptedAmount": testCiphertext3,
				"BalanceCommitment":      testCompressedPoint2,
				"ZKProof":                testConvertBackProof,
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

func TestConfidentialMPTConvertBack_BinaryRoundTrip(t *testing.T) {
	const (
		prefix = "1200572400000002301A00000000000000FA5028ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB7025C36F"
		suffix = "70264202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD70274202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD702B4202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD702E2102ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB81145B812C9D57731E27A2DA8B1830195F88EF32A3B60115000004C463C52827307480341125DA0577DEFC38405B0E3E"
	)
	point := "02" + strings.Repeat("AB", 32)
	ciphertext := point + "03" + strings.Repeat("CD", 32)
	tx := &ConfidentialMPTConvertBack{
		BaseTx: BaseTx{
			Account:         "r9LqNeG6qHxjeUocjvVki2XR35weJ9mZgQ",
			TransactionType: ConfidentialMPTConvertBackTx,
			Sequence:        2,
		},
		MPTokenIssuanceID:      "000004C463C52827307480341125DA0577DEFC38405B0E3E",
		MPTAmount:              250,
		HolderEncryptedAmount:  ciphertext,
		IssuerEncryptedAmount:  ciphertext,
		AuditorEncryptedAmount: types.HexBlob(ciphertext),
		BlindingFactor:         strings.Repeat("AB", 32),
		ZKProof:                strings.Repeat("AB", ConvertBackProofLen/2),
		BalanceCommitment:      point,
	}
	expected := prefix + strings.Repeat("AB", ConvertBackProofLen/2) + suffix

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	require.Equal(t, expected, encoded)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, map[string]any(tx.Flatten()), decoded)
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
			},
			wantErr: nil,
		},
		{
			name: "pass - with valid AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				AuditorEncryptedAmount: types.HexBlob(testCiphertext3),
				BlindingFactor:         testBlindingFactor,
				BalanceCommitment:      testCompressedPoint2,
				ZKProof:                testConvertBackProof,
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
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(0),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        "short",
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
			},
			wantErr: ErrConfidentialConvertBackInvalidBlindingFactor,
		},
		{
			name: "fail - invalid blinding factor hex",
			tx: &ConfidentialMPTConvertBack{
				BaseTx:                BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTConvertBackTx, Fee: types.XRPCurrencyAmount(12)},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        strings.Repeat("AA", 31) + "ZZ",
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "",
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: "",
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
			},
			wantErr: ErrConfidentialConvertBackInvalidCiphertext,
		},
		{
			name: "fail - short ZKProof",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               strings.Repeat("AA", 10),
			},
			wantErr: ErrConfidentialConvertBackInvalidProof,
		},
		{
			name: "fail - invalid ZKProof hex",
			tx: &ConfidentialMPTConvertBack{
				BaseTx:                BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTConvertBackTx, Fee: types.XRPCurrencyAmount(12)},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               strings.Repeat("AA", ConvertBackProofLen/2-1) + "ZZ",
			},
			wantErr: ErrConfidentialConvertBackInvalidProof,
		},
		{
			name: "fail - empty ZKProof",
			tx: &ConfidentialMPTConvertBack{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertBackTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
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
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				BlindingFactor:         testBlindingFactor,
				AuditorEncryptedAmount: types.HexBlob("not-hex!"),
				BalanceCommitment:      testCompressedPoint2,
				ZKProof:                testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     "",
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "AABB",
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: "CCDD",
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     testCompressedPoint2,
				ZKProof:               testConvertBackProof,
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
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				BlindingFactor:         testBlindingFactor,
				AuditorEncryptedAmount: types.HexBlob("AABB"),
				BalanceCommitment:      testCompressedPoint2,
				ZKProof:                testConvertBackProof,
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
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				BalanceCommitment:     "EEFF",
				ZKProof:               testConvertBackProof,
			},
			wantErr: ErrConfidentialConvertBackInvalidCommitment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.wantErr != nil {
				require.False(t, valid)
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.True(t, valid)
				require.NoError(t, err)
			}
		})
	}
}
