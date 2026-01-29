package transaction

import (
	"encoding/json"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscrowCreate_TxType(t *testing.T) {
	entry := &EscrowCreate{}
	assert.Equal(t, EscrowCreateTx, entry.TxType())
}

func TestEscrowCreate_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		entry    *EscrowCreate
		expected string
	}{
		{
			name: "pass - all fields set",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter:    533257958,
				FinishAfter:    533171558,
				Condition:      "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
				DestinationTag: types.DestinationTag(23480),
			},
			expected: `{
				"TransactionType": "EscrowCreate",
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Amount":          "10000",
				"Destination":     "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"CancelAfter":     533257958,
				"FinishAfter":     533171558,
				"Condition":       "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
				"DestinationTag":  23480
			}`,
		},
		{
			name: "pass - all fields set with DestinationTag to 0",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter:    533257958,
				FinishAfter:    533171558,
				Condition:      "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
				DestinationTag: types.DestinationTag(0),
			},
			expected: `{
				"TransactionType": "EscrowCreate",
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Amount":          "10000",
				"Destination":     "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"CancelAfter":     533257958,
				"FinishAfter":     533171558,
				"Condition":       "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
				"DestinationTag":  0
			}`,
		},
		{
			name: "pass - optional fields omitted",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
			},
			expected: `{
				"TransactionType": "EscrowCreate",
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Amount":          "10000",
				"Destination":     "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.entry.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEscrowCreate_Validate(t *testing.T) {
	tests := []struct {
		name      string
		entry     *EscrowCreate
		wantValid bool
		wantErr   bool
	}{

		{
			name: "fail - invalid transaction with only CancelAfter",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid transaction with only Condition",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				Condition:   "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid transaction with no Condition and FinishAfter",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid transaction with invalid destination address",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "invalidAddress",
				CancelAfter: 533257958,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid BaseTx, missing TransactionType",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "pass - valid transaction - Conditional with expiration",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
				Condition:   "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid transaction - Time based",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid transaction - Time based with expiration",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
				CancelAfter: 533257958,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid transaction - Timed conditional",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
				Condition:   "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - valid transaction - Timed conditional with Expiration",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
				Condition:   "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
				CancelAfter: 533257958,
			},
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.entry.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("escrowCreate.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("escrowCreate.Validate() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestEscrowCreate_Unmarshal(t *testing.T) {
	tests := []struct {
		name                 string
		jsonData             string
		expectedTag          *uint32
		expectUnmarshalError bool
	}{
		{
			name: "pass - full EscrowCreate with DestinationTag",
			jsonData: `{
				"TransactionType": "EscrowCreate",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Destination": "rDEST123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Amount": "1000000",
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"CancelAfter": 695123456,
				"FinishAfter": 695000000,
				"Condition": "A0258020C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B474680103080000000000000000000000000000000000000000000000000000000000000000",
				"DestinationTag": 12345,
				"SourceTag": 54321,
				"OwnerNode": "0000000000000000",
				"PreviousTxnID": "C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B47468",
				"LastLedgerSequence": 12345678,
				"NetworkID": 1024,
				"Memos": [
					{
					"Memo": {
						"MemoType": "657363726F77",
						"MemoData": "457363726F77206372656174656420666F72207061796D656E74"
					}
					}
				],
				"Signers": [
					{
					"Signer": {
						"Account": "rSIGNER123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
						"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
						"TxnSignature": "3045022100D7F67A81F343...B87D"
					}
					}
				],
				"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
				"TxnSignature": "3045022100D7F67A81F343...B87D"
			}`,
			expectedTag:          func() *uint32 { v := uint32(12345); return &v }(),
			expectUnmarshalError: false,
		},
		{
			name: "pass - partial EscrowCreate with DestinationTag set to 0",
			jsonData: `{
				"TransactionType": "EscrowCreate",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Destination": "rDEST123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Amount": "1000000",
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"CancelAfter": 695123456,
				"FinishAfter": 695000000,
				"Condition": "A0258020C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B474680103080000000000000000000000000000000000000000000000000000000000000000",
				"DestinationTag": 0
			}`,
			expectedTag:          func() *uint32 { v := uint32(0); return &v }(),
			expectUnmarshalError: false,
		},
		{
			name: "pass - partial EscrowCreate with DestinationTag undefined",
			jsonData: `{
				"TransactionType": "EscrowCreate",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Destination": "rDEST123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Amount": "1000000",
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"CancelAfter": 695123456,
				"FinishAfter": 695000000,
				"Condition": "A0258020C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B474680103080000000000000000000000000000000000000000000000000000000000000000"			}`,
			expectedTag:          nil,
			expectUnmarshalError: false,
		},
		{
			name: "pass - full EscrowCreate with MPTAmount",
			jsonData: `{
				"TransactionType": "EscrowCreate",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Destination": "rDEST123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Amount": {
					"mpt_issuance_id": "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
					"value": "1000000"
				},
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"CancelAfter": 695123456,
				"FinishAfter": 695000000,
				"Condition": "A0258020C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B474680103080000000000000000000000000000000000000000000000000000000000000000",
				"DestinationTag": 12345,
				"SourceTag": 54321,
				"OwnerNode": "0000000000000000",
				"PreviousTxnID": "C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B47468",
				"LastLedgerSequence": 12345678,
				"NetworkID": 1024,
				"Memos": [
					{
					"Memo": {
						"MemoType": "657363726F77",
						"MemoData": "457363726F77206372656174656420666F72207061796D656E74"
					}
					}
				],
				"Signers": [
					{
					"Signer": {
						"Account": "rSIGNER123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
						"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
						"TxnSignature": "3045022100D7F67A81F343...B87D"
					}
					}
				],
				"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
				"TxnSignature": "3045022100D7F67A81F343...B87D"
			}`,
			expectedTag:          func() *uint32 { v := uint32(12345); return &v }(),
			expectUnmarshalError: false,
		},
		{
			name: "pass - full EscrowCreate with IssuedAmount",
			jsonData: `{
				"TransactionType": "EscrowCreate",
				"Account": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Destination": "rDEST123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
				"Amount": {
					"issuer": "rEXAMPLE123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
					"currency": "USD",
					"value": "1000000"
				},
				"Fee": "10",
				"Sequence": 1,
				"Flags": 2147483648,
				"CancelAfter": 695123456,
				"FinishAfter": 695000000,
				"Condition": "A0258020C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B474680103080000000000000000000000000000000000000000000000000000000000000000",
				"DestinationTag": 12345,
				"SourceTag": 54321,
				"OwnerNode": "0000000000000000",
				"PreviousTxnID": "C4F71E9B01F5A78023E932ABF6B2C1F020986E6C9E55678FFBAE67A2F5B47468",
				"LastLedgerSequence": 12345678,
				"NetworkID": 1024,
				"Memos": [
					{
					"Memo": {
						"MemoType": "657363726F77",
						"MemoData": "457363726F77206372656174656420666F72207061796D656E74"
					}
					}
				],
				"Signers": [
					{
					"Signer": {
						"Account": "rSIGNER123456789ABCDEFGHJKLMNPQRSTUVWXYZ",
						"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
						"TxnSignature": "3045022100D7F67A81F343...B87D"
					}
					}
				],
				"SigningPubKey": "ED5F93AB1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF12345678",
				"TxnSignature": "3045022100D7F67A81F343...B87D"
			}`,
			expectedTag:          func() *uint32 { v := uint32(12345); return &v }(),
			expectUnmarshalError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var escrowCreate EscrowCreate
			err := json.Unmarshal([]byte(tt.jsonData), &escrowCreate)
			if (err != nil) != tt.expectUnmarshalError {
				t.Errorf("Unmarshal() error = %v, expectUnmarshalError %v", err, tt.expectUnmarshalError)
				return
			}
			if tt.expectedTag == nil {
				require.Nil(t, escrowCreate.DestinationTag, "Expected DestinationTag to be nil")
			} else {
				require.NotNil(t, escrowCreate.DestinationTag, "Expected DestinationTag not to be nil")
				require.Equal(t, *tt.expectedTag, *escrowCreate.DestinationTag)
			}
		})
	}
}
