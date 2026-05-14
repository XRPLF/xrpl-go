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
		{
			name: "pass - nil Amount omitted",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
			},
			expected: `{
				"TransactionType": "EscrowCreate",
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
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
		name        string
		entry       *EscrowCreate
		expectedErr error
	}{
		{
			name: "fail - missing Amount",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name: "fail - malformed issued Amount",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Currency: "USD",
					Value:    "10000",
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			expectedErr: ErrInvalidTokenFields,
		},
		{
			name: "fail - zero XRP amount",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(0),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			expectedErr: ErrEscrowCreateZeroAmount,
		},
		{
			name: "fail - zero IOU amount",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Currency: "USD",
					Value:    "0",
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			expectedErr: ErrEscrowCreateZeroAmount,
		},
		{
			name: "fail - zero MPT amount",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "00002A1F8B7E0C5E0A3B5B8B5B8B5B8B5B8B5B8B5B8B5B8B",
					Value:         "0",
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				FinishAfter: 533171558,
			},
			expectedErr: ErrEscrowCreateZeroAmount,
		},
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
			expectedErr: ErrEscrowCreateNoConditionOrFinishAfterSet,
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
			expectedErr: ErrEscrowCreateNoConditionOrFinishAfterSet,
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
			expectedErr: ErrEscrowCreateNoConditionOrFinishAfterSet,
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
			expectedErr: ErrEscrowCreateInvalidDestinationAddress,
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
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name: "fail - non-hex Condition (odd length)",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
				Condition:   "not-hex",
			},
			expectedErr: ErrEscrowCreateInvalidCondition,
		},
		{
			name: "fail - non-hex Condition (even length)",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
				Condition:   "GG",
			},
			expectedErr: ErrEscrowCreateInvalidCondition,
		},
		{
			name: "fail - odd-length Condition",
			entry: &EscrowCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				CancelAfter: 533257958,
				Condition:   "F",
			},
			expectedErr: ErrEscrowCreateInvalidCondition,
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.entry.Validate()

			if tt.expectedErr != nil {
				assert.False(t, valid)
				assert.ErrorIs(t, err, tt.expectedErr)
				return
			}

			assert.True(t, valid)
			assert.NoError(t, err)
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
