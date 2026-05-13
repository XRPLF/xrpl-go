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
		{
			name: "pass - nil Amount omitted",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
			},
			expected: `{
				"TransactionType": "NFTokenCreateOffer",
				"Account": "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
				"NFTokenID": "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007"
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
		name        string
		input       *NFTokenCreateOffer
		setter      func(*NFTokenCreateOffer)
		expectedErr error
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
		},
		{
			name: "pass - valid sell offer with zero XRP amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(0),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
		},
		{
			name: "pass - valid sell offer with non-zero IOU amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					Currency: "USD",
					Value:    "100",
				},
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
		},
		{
			name: "pass - valid sell offer with non-zero MPT amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "00002A1F8B7E0C5E0A3B5B8B5B8B5B8B5B8B5B8B5B8B5B8B",
					Value:         "42",
				},
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
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
			expectedErr: ErrInvalidAccount,
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
		},
		{
			name: "pass - valid buy offer with issued amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					Currency: "USD",
					Value:    "1",
				},
			},
		},
		{
			name: "pass - valid buy offer with MPT amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
					Value:         "1",
				},
			},
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
			expectedErr: ErrOwnerAccountConflict,
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
			expectedErr: ErrDestinationAccountConflict,
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
			expectedErr: ErrInvalidOwner,
		},
		{
			name: "fail - missing Amount",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrMissingField{Field: "Amount"},
		},
		{
			name: "fail - missing NFTokenID",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Amount: types.XRPCurrencyAmount(1000000),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrInvalidNFTokenID,
		},
		{
			name: "fail - invalid NFTokenID",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "invalidNFTokenID",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrInvalidNFTokenID,
		},
		{
			name: "fail - short hex NFTokenID",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "ABC123",
				Amount:    types.XRPCurrencyAmount(1000000),
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrInvalidNFTokenID,
		},
		{
			name: "fail - buy offer XRP amount cannot be zero",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount:    types.XRPCurrencyAmount(0),
			},
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "fail - buy offer issued amount cannot be zero",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					Currency: "USD",
					Value:    "0",
				},
			},
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "fail - buy offer MPT amount cannot be zero",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				Owner:     "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
					Value:         "0",
				},
			},
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "fail - sell offer issued amount cannot be zero",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					Currency: "USD",
					Value:    "0",
				},
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "fail - sell offer MPT amount cannot be zero",
			input: &NFTokenCreateOffer{
				BaseTx: BaseTx{
					Account:         "rs8jBmmfpwgmrSPgwMsh7CvKRmRt1JTVSX",
					TransactionType: NFTokenCreateOfferTx,
				},
				NFTokenID: "000100001E962F495F07A990F4ED55ACCFEEF365DBAA76B6A048C0A200000007",
				Amount: types.MPTCurrencyAmount{
					MPTIssuanceID: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
					Value:         "0",
				},
			},
			setter: func(n *NFTokenCreateOffer) {
				n.SetSellNFTokenFlag()
			},
			expectedErr: ErrInvalidTokenValue,
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
			expectedErr: ErrInvalidDestination,
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
			expectedErr: ErrOwnerPresentForSellOffer,
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
			expectedErr: ErrOwnerNotPresentForBuyOffer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setter != nil {
				tt.setter(tt.input)
			}
			valid, err := tt.input.Validate()

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
