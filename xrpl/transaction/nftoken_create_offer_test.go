package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestNFTokenCreateOffer_TxType(t *testing.T) {
	tx := &NFTokenCreateOffer{}
	assert.Equal(t, NFTokenCreateOfferTx, tx.TxType())
}

func TestNFTokenCreateOffer_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*NFTokenCreateOffer)
		expected uint32
	}{
		{
			name: "pass - SetSellNFTokenFlag",
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expected: TfSellNFToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NFTokenCreateOffer{}
			tt.setter(n)
			if n.Flags != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, n.Flags)
			}
		})
	}
}

func TestNFTokenCreateOffer_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    *NFTokenCreateOffer
		expected string
	}{
		{
			name: "pass - all fields set",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:       "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID:   "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:      types.XRPCurrencyAmount(1000000),
				Expiration:  600000000,
				Destination: "r3G8r9hV1J8r9hV1J8r9hV1J8r9hV1J8r9",
			},
			expected: `{
				"TransactionType": "NFTokenCreateOffer",
				"Account": "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
				"Owner": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"NFTokenID": "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				"Amount": "1000000",
				"Expiration": 600000000,
				"Destination": "r3G8r9hV1J8r9hV1J8r9hV1J8r9hV1J8r9"
			}`,
		},
		{
			name: "pass - optional fields omitted",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			expected: `{
				"TransactionType": "NFTokenCreateOffer",
				"Account": "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
				"NFTokenID": "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				"Amount": "1000000"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.input.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}
func TestNFTokenCreateOffer_Validate(t *testing.T) {
	tests := []struct {
		name       string
		input      *NFTokenCreateOffer
		setter     func(*NFTokenCreateOffer)
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "pass - valid sell offer",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid BaseTx, missing account",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidAccount,
		},
		{
			name: "pass - valid buy offer",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - owner and account are equal",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrOwnerAccountConflict,
		},
		{
			name: "fail - destination and account are equal",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID:   "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:      types.XRPCurrencyAmount(1000000),
				Destination: "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrDestinationAccountConflict,
		},
		{
			name: "fail - invalid owner address",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "invalidAddress",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidOwner,
		},
		{
			name: "fail - invalid destination address",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID:   "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:      types.XRPCurrencyAmount(1000000),
				Destination: "invalidAddress",
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidDestination,
		},
		{
			name: "fail - owner present for sell offer",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrOwnerPresentForSellOffer,
		},
		{
			name: "invalid - owner not present for buy offer",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrOwnerNotPresentForBuyOffer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setter != nil {
				tt.setter(tt.input)
			}
			valid, err := tt.input.Validate()
			if valid != tt.wantValid {
				t.Errorf("expected valid to be %v, got %v", tt.wantValid, valid)
			}
			if (err != nil) && err != tt.errMessage {
				t.Errorf("Validate() got error message = %v, want error message %v", err, tt.errMessage)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error presence to be %v, got %v, err: %s", tt.wantErr, err != nil, err)
			}
		})
	}
}
