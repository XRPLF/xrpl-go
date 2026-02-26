package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestPaymentChannelClaim_TxType(t *testing.T) {
	tx := &PaymentChannelClaim{}
	assert.Equal(t, PaymentChannelClaimTx, tx.TxType())
}

func TestPaymentChannelClaimFlags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*PaymentChannelClaim)
		expected uint32
	}{
		{
			name: "pass - SetRenewFlag",
			setter: func(p *PaymentChannelClaim) {
				p.SetRenewFlag()
			},
			expected: TfRenew,
		},
		{
			name: "pass - SetCloseFlag",
			setter: func(p *PaymentChannelClaim) {
				p.SetCloseFlag()
			},
			expected: TfClose,
		},
		{
			name: "pass - SetRenewFlag and SetCloseFlag",
			setter: func(p *PaymentChannelClaim) {
				p.SetRenewFlag()
				p.SetCloseFlag()
			},
			expected: TfRenew | TfClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PaymentChannelClaim{}
			tt.setter(p)
			assert.Equal(t, tt.expected, p.Flags)
		})
	}
}

func TestPaymentChannelClaim_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		claim    PaymentChannelClaim
		expected string
	}{
		{
			name: "pass - PaymentChannelClaim with Channel",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel: types.Hash256("ABC123"),
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelClaim",
				"Channel": "ABC123"
			}`,
		},
		{
			name: "pass - PaymentChannelClaim with Balance and Amount",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Balance: types.XRPCurrencyAmount(1000),
				Amount:  types.XRPCurrencyAmount(2000),
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelClaim",
				"Balance": "1000",
				"Amount": "2000"
			}`,
		},
		{
			name: "pass - PaymentChannelClaim with Signature and PublicKey",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Signature: "ABCDEF",
				PublicKey: "123456",
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelClaim",
				"Signature": "ABCDEF",
				"PublicKey": "123456"
			}`,
		},
		{
			name: "pass - PaymentChannelClaim with all fields",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel:       types.Hash256("ABC123"),
				Balance:       types.XRPCurrencyAmount(1000),
				Amount:        types.XRPCurrencyAmount(2000),
				Signature:     "ABCDEF",
				PublicKey:     "123456",
				CredentialIDs: types.CredentialIDs{"1234567890abcdef"},
			},
			expected: `{
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "PaymentChannelClaim",
				"Channel": "ABC123",
				"Balance": "1000",
				"Amount": "2000",
				"Signature": "ABCDEF",
				"PublicKey": "123456",
				"CredentialIDs": ["1234567890abcdef"]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.claim.Flatten(), []byte(tt.expected))
			assert.NoError(t, err)
		})
	}
}

func TestPaymentChannelClaim_Validate(t *testing.T) {
	tests := []struct {
		name        string
		claim       PaymentChannelClaim
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - all fields valid",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Balance:       types.XRPCurrencyAmount(1000),
				Channel:       "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				Signature:     "ABCDEF",
				PublicKey:     "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				CredentialIDs: types.CredentialIDs{"1234567890abcdef"},
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - missing Account in BaseTx",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					TransactionType: PaymentChannelClaimTx,
				},
				Balance:   types.XRPCurrencyAmount(1000),
				Channel:   "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				Signature: "ABCDEF",
				PublicKey: "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - empty Channel",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidChannel,
		},
		{
			name: "fail - invalid Signature",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel:   "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				Signature: "INVALID_SIGNATURE",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidSignature,
		},
		{
			name: "pass - no Signature",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel: "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid PublicKey",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel:   "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				PublicKey: "INVALID",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidHexPublicKey,
		},
		{
			name: "fail - invalid CredentialIDs",
			claim: PaymentChannelClaim{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: PaymentChannelClaimTx,
				},
				Channel:       "ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC123ABC1",
				CredentialIDs: types.CredentialIDs{"invalid"},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCredentialIDs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.claim.Validate()
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil && err != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}
