package hash

import (
	"maps"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	testPublicKey = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
	testSignature = "30440220702ABC11419AD4940969CC32EB4D1BFDBFCA651F064F30D6E1646D74FBFC493902204E5B451B447B0F69904127F04FE71634BD825A8970B9467871DA89EEC4B021F8"
	testAccount   = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	testSigner    = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
)

func TestSignTxSignedFormMatrix(t *testing.T) {
	validSigner := map[string]any{
		"Signer": map[string]any{
			"Account":       testSigner,
			"SigningPubKey": testPublicKey,
			"TxnSignature":  testSignature,
		},
	}
	base := func() map[string]any {
		return map[string]any{
			"TransactionType": "Payment",
			"Account":         testAccount,
			"Flags":           uint32(0),
		}
	}
	with := func(fields map[string]any) map[string]any {
		tx := base()
		maps.Copy(tx, fields)
		return tx
	}

	tests := []struct {
		name        string
		tx          map[string]any
		expectedErr error
	}{
		{
			name: "pass - complete single-sign",
			tx: with(map[string]any{
				"SigningPubKey": testPublicKey,
				"TxnSignature":  testSignature,
			}),
		},
		{
			name: "pass - complete multisign",
			tx: with(map[string]any{
				"SigningPubKey": "",
				"Signers":       []any{validSigner},
			}),
		},
		{
			name:        "fail - unsigned",
			tx:          base(),
			expectedErr: ErrNonSignedTransaction,
		},
		{
			name:        "fail - single-sign missing signature",
			tx:          with(map[string]any{"SigningPubKey": testPublicKey}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name:        "fail - single-sign missing public key",
			tx:          with(map[string]any{"TxnSignature": testSignature}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - single-sign empty field",
			tx: with(map[string]any{
				"SigningPubKey": "",
				"TxnSignature":  testSignature,
			}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - mixed single and multisign",
			tx: with(map[string]any{
				"SigningPubKey": testPublicKey,
				"TxnSignature":  testSignature,
				"Signers":       []any{validSigner},
			}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name:        "fail - empty signers",
			tx:          with(map[string]any{"SigningPubKey": "", "Signers": []any{}}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - signer missing account",
			tx: with(map[string]any{"Signers": []any{map[string]any{
				"Signer": map[string]any{"SigningPubKey": testPublicKey, "TxnSignature": testSignature},
			}}}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - signer missing public key",
			tx: with(map[string]any{"Signers": []any{map[string]any{
				"Signer": map[string]any{"Account": testSigner, "TxnSignature": testSignature},
			}}}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - signer missing signature",
			tx: with(map[string]any{"Signers": []any{map[string]any{
				"Signer": map[string]any{"Account": testSigner, "SigningPubKey": testPublicKey},
			}}}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name:        "fail - malformed signer wrapper",
			tx:          with(map[string]any{"Signers": []any{"not an object"}}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "pass - explicitly unsigned inner Batch",
			tx: with(map[string]any{
				"Flags":         uint32(types.TfInnerBatchTxn),
				"SigningPubKey": "",
			}),
		},
		{
			name:        "fail - inner Batch missing empty SigningPubKey",
			tx:          with(map[string]any{"Flags": uint32(types.TfInnerBatchTxn)}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - inner Batch has signature",
			tx: with(map[string]any{
				"Flags":         uint32(types.TfInnerBatchTxn),
				"SigningPubKey": "",
				"TxnSignature":  testSignature,
			}),
			expectedErr: ErrInvalidSignedTransaction,
		},
		{
			name: "fail - inner Batch has signers",
			tx: with(map[string]any{
				"Flags":         uint32(types.TfInnerBatchTxn),
				"SigningPubKey": "",
				"Signers":       []any{},
			}),
			expectedErr: ErrInvalidSignedTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := SignTx(tt.tx)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				require.Empty(t, hash)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, hash)
		})
	}
}

func TestSignTxBlob(t *testing.T) {
	t.Run("pass - complete single-sign blob", func(t *testing.T) {
		blob, err := binarycodec.Encode(map[string]any{
			"TransactionType": "Payment",
			"Account":         testAccount,
			"Flags":           uint32(0),
			"SigningPubKey":   testPublicKey,
			"TxnSignature":    testSignature,
		})
		require.NoError(t, err)

		hash, err := SignTxBlob(blob)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})

	t.Run("fail - malformed blob returns error without panic", func(t *testing.T) {
		require.NotPanics(t, func() {
			hash, err := SignTxBlob("not-a-transaction-blob")
			require.Error(t, err)
			require.Empty(t, hash)
		})
	})

	t.Run("fail - partial signed blob", func(t *testing.T) {
		blob, err := binarycodec.Encode(map[string]any{
			"TransactionType": "Payment",
			"Account":         testAccount,
			"Flags":           uint32(0),
			"SigningPubKey":   testPublicKey,
		})
		require.NoError(t, err)

		hash, err := SignTxBlob(blob)
		require.ErrorIs(t, err, ErrInvalidSignedTransaction)
		require.Empty(t, hash)
	})
}
