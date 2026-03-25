package hash

import (
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/require"
)

func TestSignTxBlob(t *testing.T) {
	tests := []struct {
		name        string
		tx          map[string]any
		expected    bool
		expectedErr error
	}{
		{
			name: "pass - has TxnSignature",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"TxnSignature":    "30440220702ABC11419AD4940969CC32EB4D1BFDBFCA651F064F30D6E1646D74FBFC493902204E5B451B447B0F69904127F04FE71634BD825A8970B9467871DA89EEC4B021F8",
				"Flags":           uint32(0),
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "pass - has SigningPubKey",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"SigningPubKey":   "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
				"Flags":           uint32(0),
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "pass - has Signers",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Signers": []any{
					map[string]any{
						"Signer": map[string]any{
							"Account":       "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
							"SigningPubKey": "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
							"TxnSignature":  "30440220702ABC11419AD4940969CC32EB4D1BFDBFCA651F064F30D6E1646D74FBFC493902204E5B451B447B0F69904127F04FE71634BD825A8970B9467871DA89EEC4B021F8",
						},
					},
				},
				"Flags": uint32(0),
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "fail - no signature fields",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Destination":     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"Amount":          "1000000",
				"Flags":           uint32(0),
			},
			expected:    false,
			expectedErr: ErrMissingSignature,
		},
		{
			name: "pass - inner-batch skips signature",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Destination":     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"Amount":          "1000000",
				"Flags":           uint32(types.TfInnerBatchTxn),
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "pass - inner-batch + other flags skips signature",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Destination":     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
				"Amount":          "1000000",
				"Flags":           uint32(types.TfInnerBatchTxn | 0x00010000),
			},
			expected:    true,
			expectedErr: nil,
		},
		{
			name: "fail - other flags without inner-batch",
			tx: map[string]any{
				"TransactionType": "Payment",
				"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
				"Flags":           uint32(0x00010000),
			},
			expected:    false,
			expectedErr: ErrMissingSignature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob, err := binarycodec.Encode(tt.tx)
			require.NoError(t, err, "Encode failed")

			hash, err := SignTxBlob(blob)

			if tt.expected {
				require.NoError(t, err)
				require.NotEmpty(t, hash)
			} else {
				require.Error(t, err)
				require.Empty(t, hash)
				if tt.expectedErr != nil {
					require.Equal(t, tt.expectedErr, err)
				}
			}
		})
	}
}
