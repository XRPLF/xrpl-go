package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestNFTokenCancelOffer_TxType(t *testing.T) {
	tx := &NFTokenCancelOffer{}
	assert.Equal(t, NFTokenCancelOfferTx, tx.TxType())
}

func TestNFTokenCancelOffer_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *NFTokenCancelOffer
		expected string
	}{
		{
			name: "pass - Empty NFTokenOffers",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: NFTokenCancelOfferTx,
				},
				NFTokenOffers: []types.NFTokenID{},
			},
			expected: `{
				"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "NFTokenCancelOffer"
			}`,
		},
		{
			name: "pass - With NFTokenOffers",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				},
				NFTokenOffers: []types.NFTokenID{
					"9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D",
				},
			},
			expected: `{
				"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "NFTokenCancelOffer",
				"NFTokenOffers": [
					"9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D"
				]
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

func TestNFTokenCancelOffer_Validate(t *testing.T) {
	tests := []struct {
		name       string
		tx         *NFTokenCancelOffer
		wantValid  bool
		wantErr    bool
		errMessage error
	}{
		{
			name: "pass - Valid NFTokenCancelOffer",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: NFTokenCancelOfferTx,
				},
				NFTokenOffers: []types.NFTokenID{
					types.NFTokenID("9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D"),
					types.NFTokenID("0008000095CC60265CE05AB4D4281338ED985F224E7090093C1C2FC00023F125"),
				},
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - Invalid NFTokenCancelOffer - Empty NFTokenOffers",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: NFTokenCancelOfferTx,
				},
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrEmptyNFTokenOffers,
		},
		{
			name: "fail - Invalid NFTokenCancelOffer - Empty NFTokenOffers",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: NFTokenCancelOfferTx,
				},
				NFTokenOffers: []types.NFTokenID{},
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrEmptyNFTokenOffers,
		},
		{
			name: "fail - Invalid NFTokenCancelOffer BaseTx - Missing TransactionType",
			tx: &NFTokenCancelOffer{
				BaseTx: BaseTx{
					Account: "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				},
				NFTokenOffers: []types.NFTokenID{
					"9C92E061381C1EF37A8CDE0E8FC35188BFC30B1883825042A64309AC09F4C36D",
				},
			},
			wantValid:  false,
			wantErr:    true,
			errMessage: ErrInvalidTransactionType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", valid, tt.wantValid)
			}
			if (err != nil) && err != tt.errMessage {
				t.Errorf("Validate() got error message = %v, want error message %v", err, tt.errMessage)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
