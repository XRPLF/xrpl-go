package transaction

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestNFTokenAcceptOffer_TxType(t *testing.T) {
	tx := &NFTokenAcceptOffer{}
	assert.Equal(t, NFTokenAcceptOfferTx, tx.TxType())
}

func TestNFTokenAcceptOffer_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *NFTokenAcceptOffer
		expected string
	}{
		{
			name: "pass - BaseTx only NFTokenAcceptOffer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "NFTokenAcceptOffer"
			}`,
		},
		{
			name: "pass - NFTokenAcceptOffer with Sell Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "NFTokenAcceptOffer",
				"NFTokenSellOffer": "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77"
			}`,
		},
		{
			name: "pass - NFTokenAcceptOffer with Buy Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBuyOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "NFTokenAcceptOffer",
				"NFTokenBuyOffer": "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77"
			}`,
		},
		{
			name: "pass - NFTokenAcceptOffer with Broker Fee",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "NFTokenAcceptOffer",
				"NFTokenBrokerFee": "1000"
			}`,
		},
		{
			name: "pass - NFTokenAcceptOffer with all fields",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBuyOffer:  "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			expected: `{
				"Account": "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
				"TransactionType": "NFTokenAcceptOffer",
				"NFTokenSellOffer": "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				"NFTokenBuyOffer": "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				"NFTokenBrokerFee": "1000"
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

func TestNFTokenAcceptOffer_Validate(t *testing.T) {
	tests := []struct {
		name       string
		tx         *NFTokenAcceptOffer
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "fail - invalid NFTokenBrokerFee",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBrokerFee: types.XRPCurrencyAmount(0),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrNFTokenBrokerFeeZero,
		},
		{
			name: "fail - invalid NFTokenBrokerFee as issued amount",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBrokerFee: types.IssuedCurrencyAmount{
					Issuer:   "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					Currency: "USD",
					Value:    "0",
				},
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrNFTokenBrokerFeeZero,
		},
		{
			name: "pass - Valid with Sell Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - Valid with Buy Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBuyOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "pass - Valid with Broker Fee",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBuyOffer:  "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - Invalid without Offers",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMissingOffer,
		},
		{
			name: "fail - Invalid with Broker Fee but missing Offers",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMissingBothOffers,
		},
		{
			name: "fail - Invalid with Broker Fee but missing Buy Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMissingBothOffers,
		},
		{
			name: "fail - Invalid with Broker Fee but missing Sell Offer",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					Account:         "rNCFjv8Ek5oDrNiMJ3pw6eLLFtMjZLJnf2",
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenBuyOffer:  "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
				NFTokenBrokerFee: types.XRPCurrencyAmount(1000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrMissingBothOffers,
		},
		{
			name: "fail - Invalid BaseTx, missing Account",
			tx: &NFTokenAcceptOffer{
				BaseTx: BaseTx{
					TransactionType: NFTokenAcceptOfferTx,
				},
				NFTokenSellOffer: "68CD1F6F906494EA08C9CB5CAFA64DFA90D4E834B7151899B73231DE5A0C3B77",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && !errors.Is(err, tt.errMessage) {
				t.Errorf("Validate() got error message = %v, want error message %v", err, tt.errMessage)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, wantValid %v", valid, tt.wantValid)
			}
		})
	}
}
