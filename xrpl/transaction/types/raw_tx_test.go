package types

import (
	"errors"
	"testing"
)

func TestRawTransaction_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		rawTx    *RawTransaction
		expected map[string]any
	}{
		{
			name: "pass - basic transaction data",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
					"Amount":          "1000000",
					"Flags":           TfInnerBatchTxn,
					"Fee":             "0",
					"SigningPubKey":   "",
				},
			},
			expected: map[string]any{
				"RawTransaction": map[string]any{
					"TransactionType": "Payment",
					"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
					"Amount":          "1000000",
					"Flags":           TfInnerBatchTxn,
					"Fee":             "0",
					"SigningPubKey":   "",
				},
			},
		},
		{
			name: "pass - empty transaction data",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{},
			},
			expected: map[string]any{
				"RawTransaction": map[string]any{},
			},
		},
		{
			name: "pass - transaction with additional fields",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "OfferCreate",
					"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					"TakerGets":       "1000000",
					"TakerPays": map[string]any{
						"currency": "USD",
						"issuer":   "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
						"value":    "1000",
					},
					"Flags":         TfInnerBatchTxn,
					"Fee":           "0",
					"SigningPubKey": "",
				},
			},
			expected: map[string]any{
				"RawTransaction": map[string]any{
					"TransactionType": "OfferCreate",
					"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					"TakerGets":       "1000000",
					"TakerPays": map[string]any{
						"currency": "USD",
						"issuer":   "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
						"value":    "1000",
					},
					"Flags":         TfInnerBatchTxn,
					"Fee":           "0",
					"SigningPubKey": "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rawTx.Flatten()

			// Compare the structure
			if len(result) != len(tt.expected) {
				t.Errorf("Flatten() returned map with %d keys, expected %d", len(result), len(tt.expected))
				return
			}

			// Check if RawTransaction key exists
			if _, exists := result["RawTransaction"]; !exists {
				t.Error("Flatten() result missing 'RawTransaction' key")
				return
			}

			// Verify the RawTransaction field exists and is not nil
			if result["RawTransaction"] == nil {
				t.Error("Flatten() RawTransaction field should not be nil")
			}
		})
	}
}

func TestRawTransaction_Validate(t *testing.T) {
	tests := []struct {
		name          string
		rawTx         *RawTransaction
		expectedValid bool
		expectedError error
	}{
		{
			name: "pass - valid payment transaction",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Account":         "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
					"Destination":     "rLNaPoKeeBjZe2qs6x52yVPZpZ8td4dc6w",
					"Amount":          "1000000",
					"Flags":           TfInnerBatchTxn,
					"Fee":             "0",
					"SigningPubKey":   "",
				},
			},
			expectedValid: true,
			expectedError: nil,
		},
		{
			name: "pass - valid transaction with minimal fields",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
				},
			},
			expectedValid: true,
			expectedError: nil,
		},
		{
			name: "fail - nested batch transaction",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Batch",
					"Flags":           TfInnerBatchTxn,
				},
			},
			expectedValid: false,
			expectedError: ErrBatchNestedTransaction,
		},
		{
			name: "fail - missing TfInnerBatchTxn flag",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           uint32(0),
				},
			},
			expectedValid: false,
			expectedError: ErrBatchMissingInnerFlag,
		},
		{
			name: "fail - missing flags field",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
				},
			},
			expectedValid: false,
			expectedError: ErrBatchMissingInnerFlag,
		},
		{
			name: "fail - invalid flags type",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           "invalid_flags",
				},
			},
			expectedValid: false,
			expectedError: ErrBatchMissingInnerFlag,
		},
		{
			name: "fail - non-zero fee",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"Fee":             "100",
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - invalid fee type",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"Fee":             123,
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - non-empty signing pub key",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"SigningPubKey":   "ED74D4036C6591A4BDF9C54CEFA39B996A5DCE5F86D11FDA1874481CE9D5A1CDC1",
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - invalid signing pub key type",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"SigningPubKey":   123,
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - disallowed field - LastLedgerSequence",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType":    "Payment",
					"Flags":              TfInnerBatchTxn,
					"LastLedgerSequence": 12345,
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - disallowed field - Signers",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"Signers":         []any{},
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "fail - disallowed field - TxnSignature",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn,
					"TxnSignature":    "1234567890ABCDEF",
				},
			},
			expectedValid: false,
			expectedError: ErrBatchInnerTransactionInvalid,
		},
		{
			name: "pass - valid transaction with other flags combined",
			rawTx: &RawTransaction{
				RawTransaction: map[string]any{
					"TransactionType": "Payment",
					"Flags":           TfInnerBatchTxn | 0x00020000, // TfInnerBatchTxn + some other flag
					"Fee":             "0",
					"SigningPubKey":   "",
				},
			},
			expectedValid: true,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.rawTx.Validate()

			if valid != tt.expectedValid {
				t.Errorf("Validate() returned valid = %v, expected %v", valid, tt.expectedValid)
			}

			if !errors.Is(err, tt.expectedError) {
				t.Errorf("Validate() returned error = %v, expected %v", err, tt.expectedError)
			}
		})
	}
}
