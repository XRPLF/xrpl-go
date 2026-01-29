package transaction

import (
	"testing"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestAMMDelete_TxType(t *testing.T) {
	tx := &AMMDelete{}
	assert.Equal(t, AMMDeleteTx, tx.TxType())
}

func TestAMMDelete_Flatten(t *testing.T) {
	tx := &AMMDelete{
		BaseTx: BaseTx{
			Account:  "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			Fee:      types.XRPCurrencyAmount(10),
			Sequence: 9,
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
		Asset2: ledger.Asset{
			Currency: "TST",
			Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
		},
	}

	flattened := tx.Flatten()

	expected := `{
	"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
	"Fee": "10",
	"Sequence": 9,
	"TransactionType": "AMMDelete",
	"Asset": {
		"currency": "XRP"
	},
	"Asset2": {
		"currency": "TST",
		"issuer": "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd"
	}
}`

	err := testutil.CompareFlattenAndExpected(flattened, []byte(expected))
	if err != nil {
		t.Error(err)
	}
}
func TestAMMDelete_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tx      *AMMDelete
		wantErr bool
	}{
		{
			name: "pass - valid AMMDelete",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			wantErr: false,
		},
		{
			name: "fail - invalid AMMDelete BaseTx, Account missing",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "  ",
				},
				Asset2: ledger.Asset{
					Currency: "TST",
					Issuer:   "rP9jPyP5kyvFRb6ZiRghAGw5u8SGAmU4bd",
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset2, empty currency",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "  ",
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset2, invalid xrpl address as issuer",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "USD",
					Issuer:   "invalid xrpl address",
				},
			},
			wantErr: true,
		},
		{
			name: "fail - invalid Asset2, empty issuer",
			tx: &AMMDelete{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: AMMDeleteTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        9,
				},
				Asset: ledger.Asset{
					Currency: "XRP",
				},
				Asset2: ledger.Asset{
					Currency: "USD",
					Issuer:   " ",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AMMDelete.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != !tt.wantErr {
				t.Errorf("AMMDelete.Validate() = %v, want %v", valid, !tt.wantErr)
			}
		})
	}
}
