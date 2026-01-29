package transaction

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentChannelCreate_TxType(t *testing.T) {
	tx := &PaymentChannelCreate{}
	assert.Equal(t, PaymentChannelCreateTx, tx.TxType())
}

func TestPaymentChannelCreate_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *PaymentChannelCreate
		expected string
	}{
		{
			name: "pass - All fields set",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay:    86400,
				PublicKey:      "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				CancelAfter:    533171558,
				DestinationTag: types.DestinationTag(23480),
			},
			expected: `{
				"Account":       "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelCreate",
				"Amount":        "10000",
				"Destination":   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"SettleDelay":   86400,
				"PublicKey":     "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				"CancelAfter":   533171558,
				"DestinationTag": 23480
			}`,
		},
		{
			name: "pass - All fields set with DestinationTag to 0",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay:    86400,
				PublicKey:      "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				CancelAfter:    533171558,
				DestinationTag: types.DestinationTag(0),
			},
			expected: `{
				"Account":       "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelCreate",
				"Amount":        "10000",
				"Destination":   "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"SettleDelay":   86400,
				"PublicKey":     "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				"CancelAfter":   533171558,
				"DestinationTag": 0
			}`,
		},
		{
			name: "pass - Optional fields omitted",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay: 86400,
				PublicKey:   "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
			},
			expected: `{
				"Account":     "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelCreate",
				"Amount":      "10000",
				"Destination": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"SettleDelay": 86400,
				"PublicKey":   "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.tx.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPaymentChannelCreate_Validate(t *testing.T) {
	tests := []struct {
		name        string
		tx          *PaymentChannelCreate
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - All fields valid",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay:    86400,
				PublicKey:      "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				CancelAfter:    533171558,
				DestinationTag: types.DestinationTag(23480),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - Invalid BaseTx, missing TransactionType",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Amount:         types.XRPCurrencyAmount(10000),
				Destination:    types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay:    86400,
				PublicKey:      "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
				CancelAfter:    533171558,
				DestinationTag: types.DestinationTag(23480),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name: "fail - Invalid destination address",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "invalidAddress",
				SettleDelay: 86400,
				PublicKey:   "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidDestination,
		},
		{
			name: "fail - Empty destination address",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: "",
				SettleDelay: 86400,
				PublicKey:   "32D2471DB72B27E3310F355BB33E339BF26F8392D5A93D3BC0FC3B566612DA0F0A",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidDestination,
		},
		{
			name: "fail - Invalid public key",
			tx: &PaymentChannelCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelCreateTx,
				},
				Amount:      types.XRPCurrencyAmount(10000),
				Destination: types.Address("rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"),
				SettleDelay: 86400,
				PublicKey:   "invalidPublicKey",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidHexPublicKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.expectedErr != nil {
				require.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantValid, valid)
			}
		})
	}
}

func TestPaymentChannelCreate_Unmarshal(t *testing.T) {
	tests := []struct {
		name                 string
		jsonData             string
		expectedTag          *uint32
		expectUnmarshalError bool
	}{
		{
			name: "pass - full PaymentChannelCreate with DestinationTag",
			jsonData: `{
				"TransactionType": "PaymentChannelCreate",
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
			name: "pass - partial PaymentChannelCreate with DestinationTag set to 0",
			jsonData: `{
				"TransactionType": "PaymentChannelCreate",
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
			name: "pass - partial PaymentChannelCreate with DestinationTag undefined",
			jsonData: `{
				"TransactionType": "PaymentChannelCreate",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var paymentChannelCreate PaymentChannelCreate
			err := json.Unmarshal([]byte(tt.jsonData), &paymentChannelCreate)
			fmt.Println(paymentChannelCreate.TransactionType)
			if (err != nil) != tt.expectUnmarshalError {
				t.Errorf("Unmarshal() error = %v, expectUnmarshalError %v", err, tt.expectUnmarshalError)
				return
			}
			if tt.expectedTag == nil {
				require.Nil(t, paymentChannelCreate.DestinationTag, "Expected DestinationTag to be nil")
			} else {
				require.NotNil(t, paymentChannelCreate.DestinationTag, "Expected DestinationTag not to be nil")
				require.Equal(t, *tt.expectedTag, *paymentChannelCreate.DestinationTag)
			}
		})
	}
}
