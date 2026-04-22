package transaction

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

// Test helper: 66-char hex string (33-byte compressed key).
var testCompressedKey = strings.Repeat("AB", 33)

// Test helper: 128-char hex string (64-byte Schnorr proof).
var testSchnorrProof = strings.Repeat("CD", 64)

// Test helper: 64-char hex string (32-byte blinding factor).
var testBlindingFactor = strings.Repeat("EF", 32)

// Test helper: 132-char hex string (66-byte ElGamal ciphertext).
var testCiphertext = strings.Repeat("A1", 66)

// Test helper: alternate 132-char hex string (66-byte ElGamal ciphertext).
var testCiphertext2 = strings.Repeat("B2", 66)

// Test helper: third 132-char hex string (66-byte ElGamal ciphertext).
var testCiphertext3 = strings.Repeat("C3", 66)

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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
			},
			expected: FlatTransaction{
				"Account":               "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Fee":                   "12",
				"TransactionType":       "ConfidentialMPTConvert",
				"MPTokenIssuanceID":     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:              types.MPTPlainAmount(500),
				HolderEncryptionKey:    types.EncryptionKey(testCompressedKey),
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
				"MPTokenIssuanceID":      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				"MPTAmount":              "500",
				"HolderEncryptionKey":    testCompressedKey,
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedKey),
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
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedKey),
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.HexBlob("AABB"),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob(testSchnorrProof),
			},
			wantErr: ErrConfidentialConvertInvalidKeyLength,
		},
		{
			name: "fail - invalid proof length",
			tx: &ConfidentialMPTConvert{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: ConfidentialMPTConvertTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptionKey:   types.EncryptionKey(testCompressedKey),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        testBlindingFactor,
				ZKProof:               types.HexBlob("AABB"),
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
				MPTAmount:             types.MPTPlainAmount(1000),
				HolderEncryptedAmount: testCiphertext,
				IssuerEncryptedAmount: testCiphertext2,
				BlindingFactor:        "tooshort",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:     "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
				MPTokenIssuanceID:      "00070C4495F14B0E44F78A264E41713C64B5F89242540EE255534400000000000000",
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
			_, err := tt.tx.Validate()
			if tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
