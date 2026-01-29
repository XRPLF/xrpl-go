package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestOfferCancel_TxType(t *testing.T) {
	tx := &OfferCancel{}
	assert.Equal(t, OfferCancelTx, tx.TxType())
}

func TestOfferCancel_Flatten(t *testing.T) {
	tx := &OfferCancel{
		BaseTx: BaseTx{
			Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			TransactionType:    OfferCancelTx,
			Fee:                types.XRPCurrencyAmount(10),
			Flags:              123,
			LastLedgerSequence: 7108629,
			Sequence:           7,
		},
		OfferSequence: 6,
	}

	expected := `{
		"Account":            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"TransactionType":    "OfferCancel",
		"Fee":                "10",
		"Flags":              123,
		"LastLedgerSequence": 7108629,
		"Sequence":           7,
		"OfferSequence":      6
	}`

	err := testutil.CompareFlattenAndExpected(tx.Flatten(), []byte(expected))
	if err != nil {
		t.Error(err)
	}
}
func TestOfferCancel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		tx        *OfferCancel
		wantValid bool
	}{
		{
			name: "pass - valid OfferCancel",
			tx: &OfferCancel{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCancelTx,
					Fee:                types.XRPCurrencyAmount(10),
					Flags:              123,
					LastLedgerSequence: 7108629,
					Sequence:           7,
				},
				OfferSequence: 6,
			},
			wantValid: true,
		},
		{
			name: "fail - invalid BaseTx",
			tx: &OfferCancel{
				BaseTx: BaseTx{
					Account:            "",
					TransactionType:    OfferCancelTx,
					Fee:                types.XRPCurrencyAmount(10),
					Flags:              123,
					LastLedgerSequence: 7108629,
					Sequence:           7,
				},
				OfferSequence: 6,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if valid != tt.wantValid {
				t.Errorf("expected %v, got %v, error: %v", tt.wantValid, valid, err)
			}
		})
	}
}
