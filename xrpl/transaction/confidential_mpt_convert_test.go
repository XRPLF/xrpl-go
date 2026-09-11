package transaction

import (
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestConfidentialMPTConvert_TxType(t *testing.T) {
	tx := &ConfidentialMPTConvert{}
	require.Equal(t, ConfidentialMPTConvertTx, tx.TxType())
}

func TestConfidentialMPTConvert_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *ConfidentialMPTConvert
		expected FlatTransaction
	}{
		{
			name: "pass - without optional fields",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			expected: FlatTransaction{
				"Account":               "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                   "12",
				"TransactionType":       "ConfidentialMPTConvert",
				"MPTokenIssuanceID":     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"MPTAmount":             "1000",
				"HolderEncryptedAmount": testCiphertext,
				"IssuerEncryptedAmount": testCiphertext2,
				"BlindingFactor":        testBlindingFactor,
			},
		},
		{
			name: "pass - with key and proof",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:     types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(500),
				HolderEncryptionKey:    types.EncryptionKey(testCompressedPoint1),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				AuditorEncryptedAmount: types.HexBlob(testCiphertext3),
				BlindingFactor:         testBlindingFactor,
				ZKProof:                types.HexBlob(testSchnorrProof),
			},
			expected: FlatTransaction{
				"Account":                "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                    "12",
				"TransactionType":        "ConfidentialMPTConvert",
				"MPTokenIssuanceID":      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				"MPTAmount":              "500",
				"HolderEncryptionKey":    testCompressedPoint1,
				"HolderEncryptedAmount":  testCiphertext,
				"IssuerEncryptedAmount":  testCiphertext2,
				"AuditorEncryptedAmount": testCiphertext3,
				"BlindingFactor":         testBlindingFactor,
				"ZKProof":                testSchnorrProof,
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

func TestConfidentialMPTConvert_BinaryRoundTrip(t *testing.T) {
	const (
		prefix = "1200552400000001301A00000000000000645028ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB70242102ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB702540"
		suffix = "70264202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD70274202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD702B4202ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB03CDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCDCD81145B812C9D57731E27A2DA8B1830195F88EF32A3B60115000004C463C52827307480341125DA0577DEFC38405B0E3E"
	)
	point := "02" + strings.Repeat("AB", 32)
	ciphertext := point + "03" + strings.Repeat("CD", 32)
	tx := &ConfidentialMPTConvert{
		BaseTx: BaseTx{
			Account:         "r9LqNeG6qHxjeUocjvVki2XR35weJ9mZgQ",
			TransactionType: ConfidentialMPTConvertTx,
			Sequence:        1,
		},
		MPTokenIssuanceID:      "000004C463C52827307480341125DA0577DEFC38405B0E3E",
		MPTAmount:              100,
		HolderEncryptionKey:    types.EncryptionKey(point),
		HolderEncryptedAmount:  ciphertext,
		IssuerEncryptedAmount:  ciphertext,
		AuditorEncryptedAmount: types.HexBlob(ciphertext),
		BlindingFactor:         strings.Repeat("AB", 32),
		ZKProof:                types.HexBlob(strings.Repeat("AB", SchnorrProofLen/2)),
	}
	expected := prefix + strings.Repeat("AB", SchnorrProofLen/2) + suffix

	encoded, err := binarycodec.Encode(tx.Flatten())
	require.NoError(t, err)
	require.Equal(t, expected, encoded)
	decoded, err := binarycodec.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, map[string]any(tx.Flatten()), decoded)
}

func TestConfidentialMPTConvert_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *ConfidentialMPTConvert
		wantErr error
	}{
		{
			name: "pass - without key registration",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: nil,
		},
		{
			name: "pass - with key registration",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedPoint1),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(testSchnorrProof),
			},
			wantErr: nil,
		},
		{
			name: "pass - with valid AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				AuditorEncryptedAmount: types.HexBlob(testCiphertext3),
				BlindingFactor:         testBlindingFactor,
			},
			wantErr: nil,
		},
		{
			name: "fail - empty MPTokenIssuanceID",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialMPTInvalidIssuanceID,
		},
		{
			name: "fail - key without proof",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedPoint1),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialConvertKeyProofMismatch,
		},
		{
			name: "fail - proof without key",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(testSchnorrProof),
			},
			wantErr: ErrConfidentialConvertKeyProofMismatch,
		},
		{
			name: "fail - invalid key length",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.HexBlob("AABB"),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(testSchnorrProof),
			},
			wantErr: ErrConfidentialConvertInvalidEncryptionKey,
		},
		{
			name: "fail - encryption key is not on curve",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey("02" + strings.Repeat("00", 32)),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(testSchnorrProof),
			},
			wantErr: ErrConfidentialConvertInvalidEncryptionKey,
		},
		{
			name: "fail - invalid proof length",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedPoint1),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob("AABB"),
			},
			wantErr: ErrConfidentialConvertInvalidProofLength,
		},
		{
			name: "fail - invalid proof hex",
			tx: &ConfidentialMPTConvert{
				BaseTx:                BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTConvertTx, Fee: types.XRPCurrencyAmount(12)},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedPoint1),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(strings.Repeat("AA", 63) + "ZZ"),
			},
			wantErr: ErrConfidentialConvertInvalidProofLength,
		},
		{
			name: "fail - invalid blinding factor",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        "tooshort",
			},
			wantErr: ErrConfidentialConvertInvalidBlindingFactor,
		},
		{
			name: "fail - invalid blinding factor hex",
			tx: &ConfidentialMPTConvert{
				BaseTx:                BaseTx{Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD", TransactionType: ConfidentialMPTConvertTx, Fee: types.XRPCurrencyAmount(12)},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        strings.Repeat("AA", 31) + "ZZ",
			},
			wantErr: ErrConfidentialConvertInvalidBlindingFactor,
		},
		{
			name: "fail - empty HolderEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "",
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
		},
		{
			name: "fail - invalid AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				BlindingFactor:         testBlindingFactor,
				AuditorEncryptedAmount: types.HexBlob("not-hex!"),
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
		},
		{
			name: "fail - empty IssuerEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: "",
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
		},
		{
			name: "fail - wrong length HolderEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: "AABB",
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
		},
		{
			name: "fail - wrong length IssuerEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: "AABB",
				BlindingFactor:        testBlindingFactor,
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
		},
		{
			name: "fail - wrong length AuditorEncryptedAmount",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:      "00000001B5F762798A53D543A014CAF8B297CFF8F2F937E8",
				MPTAmount:              types.MPTPlainAmount(1000),
				HolderEncryptedAmount:  testCiphertext,
				IssuerEncryptedAmount:  testCiphertext2,
				BlindingFactor:         testBlindingFactor,
				AuditorEncryptedAmount: types.HexBlob("AABB"),
			},
			wantErr: ErrConfidentialConvertInvalidCiphertext,
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
