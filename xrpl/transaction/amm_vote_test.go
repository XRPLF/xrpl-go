package transaction

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestAMMVote_TxType(t *testing.T) {
	tx := &AMMVote{}
	assert.Equal(t, AMMVoteTx, tx.TxType())
}

func TestAMMVote_Flatten(t *testing.T) {
	tx := &AMMVote{
		BaseTx: BaseTx{
			Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			TransactionType: "AMMVote",
			Fee:             types.XRPCurrencyAmount(10),
			Flags:           2147483648,
			Sequence:        8,
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
		Asset2: ledger.Asset{
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
		TradingFee: 600,
	}

	flattened := tx.Flatten()

	expected := `{
		"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
		"Fee":             "10",
		"Flags":           2147483648,
		"Sequence":        8,
		"TransactionType": "AMMVote",
		"Asset": {
			"currency": "XRP"
		},
		"Asset2": {
			"currency": "TST",
			"issuer":   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
		},
		"TradingFee": 600
	}`

	err := testutil.CompareFlattenAndExpected(flattened, []byte(expected))
	if err != nil {
		t.Error(err)
	}
}

func TestAMMVote_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *AMMVote
		wantErr bool
	}{
		{
			name: "pass - valid AMMVote",
			tx: &AMMVote{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMVote",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           2147483648,
					Sequence:        8,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				TradingFee: 600,
			},
			wantErr: false,
		},
		{
			name: "fail - invalid AMMVote BaseTx, TransactionType missing",
			tx: &AMMVote{
				BaseTx: BaseTx{
					Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					Fee:      types.XRPCurrencyAmount(10),
					Flags:    2147483648,
					Sequence: 8,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				TradingFee: 600,
			},
			wantErr: true,
		},
		{
			name: "fail - invalid TradingFee",
			tx: &AMMVote{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMVote",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           2147483648,
					Sequence:        8,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				TradingFee: 1200,
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset",
			tx: &AMMVote{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMVote",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           2147483648,
					Sequence:        8,
				},
				Asset: ledger.Asset{
					Currency: " ",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
				TradingFee: 600,
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset2",
			tx: &AMMVote{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: "AMMVote",
					Fee:             types.XRPCurrencyAmount(10),
					Flags:           2147483648,
					Sequence:        8,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   " ",
				},
				TradingFee: 600,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AMMVote.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != !tt.wantErr {
				t.Errorf("AMMVote.Validate() = %v, want %v", valid, !tt.wantErr)
			}
		})
	}
}
