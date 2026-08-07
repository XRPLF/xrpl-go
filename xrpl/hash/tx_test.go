package hash

import (
	"maps"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

const (
	testPublicKey = "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A"
	testSignature = "30440220702ABC11419AD4940969CC32EB4D1BFDBFCA651F064F30D6E1646D74FBFC493902204E5B451B447B0F69904127F04FE71634BD825A8970B9467871DA89EEC4B021F8"
	testAccount   = "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH"
	testSigner    = "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
)

func TestErrMissingSignatureAliasesErrNonSignedTransaction(t *testing.T) {
	require.ErrorIs(t, ErrNonSignedTransaction, ErrMissingSignature)
}

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
			name: "pass - named transaction type",
			tx: with(map[string]any{
				"TransactionType": transaction.PaymentTx,
				"SigningPubKey":   testPublicKey,
				"TxnSignature":    testSignature,
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

func TestSignTxPseudoTransactions(t *testing.T) {
	tests := []struct {
		name     string
		txType   transaction.TxType
		tx       map[string]any
		expected string
	}{
		{
			name:   "EnableAmendment",
			txType: transaction.EnableAmendmentTx,
			tx: map[string]any{
				"Account":         "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"Amendment":       "AE35ABDEFBDE520372B31C957020B34A7A4A9DC3115A69803A44016477C84D6E",
				"Fee":             "0",
				"LedgerSequence":  uint32(84206081),
				"Sequence":        uint32(0),
				"SigningPubKey":   "",
				"TransactionType": transaction.EnableAmendmentTx.String(),
			},
			expected: "CA4562711E4679FE9317DD767871E90A404C7A8B84FAFD35EC2CF0231F1F6DAF",
		},
		{
			name:   "SetFee",
			txType: transaction.SetFeeTx,
			tx: map[string]any{
				"Account":           "rrrrrrrrrrrrrrrrrrrrrhoLvTp",
				"BaseFee":           "000000000000000A",
				"Fee":               "0",
				"ReferenceFeeUnits": uint32(10),
				"ReserveBase":       uint32(20000000),
				"ReserveIncrement":  uint32(5000000),
				"Sequence":          uint32(0),
				"SigningPubKey":     "",
				"TransactionType":   transaction.SetFeeTx.String(),
			},
			expected: "1C15FEA3E1D50F96B6598607FC773FF1F6E0125F30160144BE0C5CBC52F5151B",
		},
		{
			name:   "UNLModify",
			txType: transaction.UNLModifyTx,
			tx: map[string]any{
				"Account":            "",
				"Fee":                "0",
				"LedgerSequence":     uint32(67850752),
				"Sequence":           uint32(0),
				"SigningPubKey":      "",
				"TransactionType":    transaction.UNLModifyTx.String(),
				"UNLModifyDisabling": uint8(1),
				"UNLModifyValidator": "EDB6FC8E803EE8EDC2793F1EC917B2EE41D35255618DEB91D3F9B1FC89B75D4539",
			},
			expected: "80CDD04AC3C26F02C678881C546280C116648C9B116F87320B1CE68490F13907",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := SignTx(tt.tx)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)

			namedTx := maps.Clone(tt.tx)
			namedTx["TransactionType"] = tt.txType
			actual, err = SignTx(namedTx)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
			require.IsType(t, transaction.TxType(""), namedTx["TransactionType"])

			blob, err := binarycodec.Encode(tt.tx)
			require.NoError(t, err)
			actual, err = SignTxBlob(blob)
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestSignTxPseudoTransactionSigningFields(t *testing.T) {
	validSigner := map[string]any{
		"Signer": map[string]any{
			"Account":       testSigner,
			"SigningPubKey": testPublicKey,
			"TxnSignature":  testSignature,
		},
	}
	tests := []struct {
		name     string
		fields   map[string]any
		wantErr  bool
		blobPath bool
	}{
		{name: "pass - absent SigningPubKey", blobPath: true},
		{name: "pass - empty SigningPubKey", fields: map[string]any{"SigningPubKey": ""}, blobPath: true},
		{name: "fail - nonempty SigningPubKey", fields: map[string]any{"SigningPubKey": testPublicKey}, wantErr: true, blobPath: true},
		{name: "fail - false SigningPubKey", fields: map[string]any{"SigningPubKey": false}, wantErr: true},
		{name: "fail - empty TxnSignature", fields: map[string]any{"TxnSignature": ""}, wantErr: true, blobPath: true},
		{name: "fail - nonempty TxnSignature", fields: map[string]any{"TxnSignature": testSignature}, wantErr: true, blobPath: true},
		{name: "fail - false TxnSignature", fields: map[string]any{"TxnSignature": false}, wantErr: true},
		{name: "fail - empty Signers", fields: map[string]any{"Signers": []any{}}, wantErr: true, blobPath: true},
		{name: "fail - nonempty Signers", fields: map[string]any{"Signers": []any{validSigner}}, wantErr: true, blobPath: true},
		{name: "fail - false Signers", fields: map[string]any{"Signers": false}, wantErr: true},
	}

	for _, txType := range []transaction.TxType{
		transaction.EnableAmendmentTx,
		transaction.SetFeeTx,
		transaction.UNLModifyTx,
	} {
		t.Run(txType.String(), func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					tx := map[string]any{"TransactionType": txType}
					maps.Copy(tx, tt.fields)

					hash, err := SignTx(tx)
					if tt.wantErr {
						require.ErrorIs(t, err, ErrInvalidSignedTransaction)
						require.Empty(t, hash)
					} else {
						require.NoError(t, err)
						require.NotEmpty(t, hash)
					}

					if !tt.blobPath {
						return
					}
					blobTx := maps.Clone(tx)
					blobTx["TransactionType"] = txType.String()
					blob, err := binarycodec.Encode(blobTx)
					require.NoError(t, err)

					hash, err = SignTxBlob(blob)
					if tt.wantErr {
						require.ErrorIs(t, err, ErrInvalidSignedTransaction)
						require.Empty(t, hash)
					} else {
						require.NoError(t, err)
						require.NotEmpty(t, hash)
					}
				})
			}
		})
	}
}

func TestSignTxMalformedTransactionType(t *testing.T) {
	tests := []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "boolean", value: false},
		{name: "unknown string", value: "Unknown"},
		{name: "unknown named string", value: transaction.TxType("Unknown")},
		{name: "slice", value: []string{transaction.EnableAmendmentTx.String()}},
		{name: "map", value: map[string]string{"type": transaction.EnableAmendmentTx.String()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				hash, err := SignTx(map[string]any{
					"TransactionType": tt.value,
					"Account":         testAccount,
					"SigningPubKey":   testPublicKey,
					"TxnSignature":    testSignature,
				})
				require.Error(t, err)
				require.Empty(t, hash)
			})
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
