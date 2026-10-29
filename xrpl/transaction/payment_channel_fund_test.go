package transaction

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	rippletime "github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestPaymentChannelFund_TxType(t *testing.T) {
	tx := &PaymentChannelFund{}
	assert.Equal(t, PaymentChannelFundTx, tx.TxType())
}

func TestPaymentChannelFund_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *PaymentChannelFund
		expected string
	}{
		{
			name: "pass - without Expiration",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: "ABC123",
				Amount:  types.XRPCurrencyAmount(200000),
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelFund",
				"Channel": "ABC123",
				"Amount":  "200000"
			}`,
		},
		{
			name: "pass - with Expiration",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel:    "DEF456",
				Amount:     types.XRPCurrencyAmount(300000),
				Expiration: 543171558,
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelFund",
				"Channel": "DEF456",
				"Amount": "300000",
				"Expiration": 543171558
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

func TestPaymentChannelFund_Validate(t *testing.T) {
	tests := []struct {
		name             string
		tx               *PaymentChannelFund
		expirationSetter func(tx *PaymentChannelFund)
		wantValid        bool
		wantErr          bool
		expectedErr      error
	}{
		{
			name: "pass - valid Transaction",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: testChannelID,
				Amount:  types.XRPCurrencyAmount(200000),
			},
			expirationSetter: func(tx *PaymentChannelFund) {
				tx.Expiration = uint32(time.Now().Unix()) + 5000
			},
			wantValid:   true,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "fail - zero Amount",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: testChannelID,
				Amount:  0,
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrPaymentChannelFundAmountInvalid,
		},
		{
			name: "fail - Amount exceeds native maximum",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: testChannelID,
				Amount:  types.XRPCurrencyAmount(currency.MaxNativeDrops + 1),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrPaymentChannelFundAmountInvalid,
		},
		{
			name: "fail - invalid BaseTx, missing Account",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					TransactionType: PaymentChannelFundTx,
				},
				Channel: testChannelID,
				Amount:  types.XRPCurrencyAmount(200000),
			},
			expirationSetter: func(tx *PaymentChannelFund) {
				tx.Expiration = uint32(rippletime.UnixTimeToRippleTime(time.Now().Unix()) + 5000)
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - invalid Expiration",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel:    testChannelID,
				Amount:     types.XRPCurrencyAmount(300000),
				Expiration: 1, // Invalid expiration time
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidExpiration,
		},
		{
			name: "fail - empty Channel",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Amount: types.XRPCurrencyAmount(200000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidChannel,
		},
		{
			name: "fail - short Channel",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: "ABC123",
				Amount:  types.XRPCurrencyAmount(200000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidChannel,
		},
		{
			name: "fail - non-hex Channel",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: types.Hash256(strings.Repeat("G", 64)),
				Amount:  types.XRPCurrencyAmount(200000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidChannel,
		},
		{
			name: "fail - zero Channel",
			tx: &PaymentChannelFund{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelFundTx,
				},
				Channel: types.Hash256(strings.Repeat("0", 64)),
				Amount:  types.XRPCurrencyAmount(200000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidChannel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if tt.expirationSetter != nil {
				tt.expirationSetter(tt.tx)
			}

			assert.Equal(t, tt.wantValid, valid)
			if (err != nil) && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Validate() got error message = %v, want error message %v", err, tt.expectedErr)
				return
			}
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
