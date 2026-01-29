package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

// Test transaction constants
var (
	// Standard valid payment transaction for batch inner transactions
	paymentTx = Payment{
		BaseTx: BaseTx{
			Account:         "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
			TransactionType: PaymentTx,
			Fee:             types.XRPCurrencyAmount(0),
			Flags:           types.TfInnerBatchTxn,
			SigningPubKey:   "",
			Sequence:        5,
		},
		Amount:      types.XRPCurrencyAmount(6000000),
		Destination: "rUser2fDds782Bd6eK15RDnGMtxf7m",
	}

	// Standard valid offer create transaction for batch inner transactions
	offerCreateTx = OfferCreate{
		BaseTx: BaseTx{
			Account:         "rUser3ABC123456789DEF456GHI789JKL",
			TransactionType: OfferCreateTx,
			Fee:             types.XRPCurrencyAmount(0),
			Flags:           types.TfInnerBatchTxn,
			SigningPubKey:   "",
			Sequence:        10,
		},
		TakerGets: types.XRPCurrencyAmount(1000000),
		TakerPays: types.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   "rIssuer123456789ABC456DEF789GHI012",
			Value:    "100",
		},
	}

	// Edge case transactions for negative tests
	paymentTxNoFlag = Payment{
		BaseTx: BaseTx{
			Account:         "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
			TransactionType: PaymentTx,
			Fee:             types.XRPCurrencyAmount(0),
			Flags:           0, // Missing types.TfInnerBatchTxn flag
			SigningPubKey:   "",
			Sequence:        5,
		},
		Amount:      types.XRPCurrencyAmount(6000000),
		Destination: "rUser2fDds782Bd6eK15RDnGMtxf7m",
	}

	paymentTxWithFee = Payment{
		BaseTx: BaseTx{
			Account:         "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
			TransactionType: PaymentTx,
			Fee:             types.XRPCurrencyAmount(12), // Non-zero fee
			Flags:           types.TfInnerBatchTxn,
			SigningPubKey:   "",
			Sequence:        5,
		},
		Amount:      types.XRPCurrencyAmount(6000000),
		Destination: "rUser2fDds782Bd6eK15RDnGMtxf7m",
	}

	paymentTxWithSigning = Payment{
		BaseTx: BaseTx{
			Account:         "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
			TransactionType: PaymentTx,
			Fee:             types.XRPCurrencyAmount(0),
			Flags:           types.TfInnerBatchTxn,
			SigningPubKey:   "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A", // Non-empty signing pub key
			Sequence:        5,
		},
		Amount:      types.XRPCurrencyAmount(6000000),
		Destination: "rUser2fDds782Bd6eK15RDnGMtxf7m",
	}
)

func TestBatch_TxType(t *testing.T) {
	tx := &Batch{}
	assert.Equal(t, BatchTx, tx.TxType())
}

func TestBatchFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    Batch
		expected string
	}{
		{
			name: "pass - batch transaction with payment",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
					Flags:           tfAllOrNothing,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "Batch",
				"Fee": "12",
				"Flags": 65536,
				"RawTransactions": [
					{
						"RawTransaction": {
							"Account": "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
							"TransactionType": "Payment",
							"Flags": 1073741824,
							"Sequence": 5,
							"Amount": "6000000",
							"Destination": "rUser2fDds782Bd6eK15RDnGMtxf7m"
						}
					}
				]
			}`,
		},
		{
			name: "pass - batch with offer create and payment transactions",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rUserBSM7T3b6nHX3Jjua62wgX9unH8s9b",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(40),
					Flags:           tfAllOrNothing,
					Sequence:        3,
					SigningPubKey:   "022D40673B44C82DEE1DDB8B9BB53DCCE4F97B27404DB850F068DD91D685E337EA",
					TxnSignature:    "3045022100EC5D367FAE2B461679AD446FBBE7BA260506579AF4ED5EFC3EC25F4DD1885B38022018C2327DB281743B12553C7A6DC0E45B07D3FC6983F261D7BCB474D89A0EC5B8",
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: offerCreateTx.Flatten(),
					},
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
			},
			expected: `{
				"Account": "rUserBSM7T3b6nHX3Jjua62wgX9unH8s9b",
				"TransactionType": "Batch",
				"Fee": "40",
				"Flags": 65536,
				"Sequence": 3,
				"SigningPubKey": "022D40673B44C82DEE1DDB8B9BB53DCCE4F97B27404DB850F068DD91D685E337EA",
				"TxnSignature": "3045022100EC5D367FAE2B461679AD446FBBE7BA260506579AF4ED5EFC3EC25F4DD1885B38022018C2327DB281743B12553C7A6DC0E45B07D3FC6983F261D7BCB474D89A0EC5B8",
				"RawTransactions": [
					{
						"RawTransaction": {
							"Account": "rUser3ABC123456789DEF456GHI789JKL",
							"TransactionType": "OfferCreate",
							"Flags": 1073741824,
							"Sequence": 10,
							"TakerGets": "1000000",
							"TakerPays": {
								"currency": "USD",
								"issuer": "rIssuer123456789ABC456DEF789GHI012",
								"value": "100"
							}
						}
					},
					{
						"RawTransaction": {
							"Account": "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
							"TransactionType": "Payment",
							"Flags": 1073741824,
							"Sequence": 5,
							"Amount": "6000000",
							"Destination": "rUser2fDds782Bd6eK15RDnGMtxf7m"
						}
					}
				]
			}`,
		},
		{
			name: "pass - batch with batch signers",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
					Flags:           tfAllOrNothing,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
				BatchSigners: []types.BatchSigner{
					{
						BatchSigner: types.BatchSignerData{
							Account:       "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
							SigningPubKey: "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
							TxnSignature:  "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8",
						},
					},
				},
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "Batch",
				"Fee": "12",
				"Flags": 65536,
				"RawTransactions": [
					{
						"RawTransaction": {
							"Account": "rUser1fcu9RJa5W1ncAuEgLJF2oJC6",
							"TransactionType": "Payment",
							"Flags": 1073741824,
							"Sequence": 5,
							"Amount": "6000000",
							"Destination": "rUser2fDds782Bd6eK15RDnGMtxf7m"
						}
					}
				],
				"BatchSigners": [
					{
						"BatchSigner": {
							"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
							"SigningPubKey": "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
							"TxnSignature": "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8"
						}
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()
			err := testutil.CompareFlattenAndExpected(result, []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestBatch_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    Batch
		expected bool
	}{
		{
			name: "pass - valid batch transaction with payments",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
					Flags:           tfAllOrNothing,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
			},
			expected: true,
		},
		{
			name: "pass - valid batch with multiple payment transactions",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
					Flags:           tfIndependent,
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTx.Flatten(),
					},
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
			},
			expected: true,
		},
		{
			name: "fail - empty raw transactions",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{},
			},
			expected: false,
		},
		{
			name: "fail - inner transaction missing types.TfInnerBatchTxn flag",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTxNoFlag.Flatten(),
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - inner transaction with nested batch",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: FlatTransaction{
							"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
							"TransactionType": "Batch", // Nested batch not allowed
							"Fee":             "0",
							"Flags":           uint32(types.TfInnerBatchTxn),
							"SigningPubKey":   "",
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - inner transaction with non-zero fee",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTxWithFee.Flatten(),
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - inner transaction with non-empty SigningPubKey",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTxWithSigning.Flatten(),
					},
				},
			},
			expected: false,
		},
		{
			name: "fail - batch signer with empty account",
			input: Batch{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: BatchTx,
					Fee:             types.XRPCurrencyAmount(12),
				},
				RawTransactions: []types.RawTransaction{
					{
						RawTransaction: paymentTx.Flatten(),
					},
				},
				BatchSigners: []types.BatchSigner{
					{
						BatchSigner: types.BatchSignerData{
							Account: "", // Empty account not allowed
						},
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			if valid != tt.expected {
				t.Errorf("expected %v, got %v, error: %v", tt.expected, valid, err)
			}
		})
	}
}

func TestBatch_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*Batch)
		expected uint32
	}{
		{
			name: "pass - SetAllOrNothingFlag",
			setter: func(b *Batch) {
				b.SetAllOrNothingFlag()
			},
			expected: tfAllOrNothing,
		},
		{
			name: "pass - SetOnlyOneFlag",
			setter: func(b *Batch) {
				b.SetOnlyOneFlag()
			},
			expected: tfOnlyOne,
		},
		{
			name: "pass - SetUntilFailureFlag",
			setter: func(b *Batch) {
				b.SetUntilFailureFlag()
			},
			expected: tfUntilFailure,
		},
		{
			name: "pass - SetIndependentFlag",
			setter: func(b *Batch) {
				b.SetIndependentFlag()
			},
			expected: tfIndependent,
		},
		{
			name: "pass - SetAllOrNothingFlag and SetOnlyOneFlag",
			setter: func(b *Batch) {
				b.SetAllOrNothingFlag()
				b.SetOnlyOneFlag()
			},
			expected: tfAllOrNothing | tfOnlyOne,
		},
		{
			name: "pass - all flags",
			setter: func(b *Batch) {
				b.SetAllOrNothingFlag()
				b.SetOnlyOneFlag()
				b.SetUntilFailureFlag()
				b.SetIndependentFlag()
			},
			expected: tfAllOrNothing | tfOnlyOne | tfUntilFailure | tfIndependent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Batch{}
			tt.setter(b)
			if b.Flags != tt.expected {
				t.Errorf("Expected Batch Flags to be %d, got %d", tt.expected, b.Flags)
			}
		})
	}
}
