package transaction

import (
	"reflect"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestTicketCreate_Flatten(t *testing.T) {
	s := TicketCreate{
		BaseTx: BaseTx{
			Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			TransactionType: TicketCreateTx,
			Fee:             types.XRPCurrencyAmount(10),
			Sequence:        50,
		},
		TicketCount: 5,
	}

	flattened := s.Flatten()

	expected := FlatTransaction{
		"Account":         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"TransactionType": "TicketCreate",
		"Fee":             "10",
		"Sequence":        uint32(50),
		"TicketCount":     uint32(5),
	}

	if !reflect.DeepEqual(flattened, expected) {
		t.Errorf("Flatten result differs from expected: %v, %v", flattened, expected)
	}
}

func TestTicketCreate_TxType(t *testing.T) {
	tx := &TicketCreate{}
	assert.Equal(t, TicketCreateTx, tx.TxType())
}

func TestTicketCreate_Validate(t *testing.T) {
	tests := []struct {
		name      string
		ticket    TicketCreate
		wantValid bool
	}{
		{
			name: "pass - valid ticket count",
			ticket: TicketCreate{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: TicketCreateTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        50,
				},
				TicketCount: 5,
			},
			wantValid: true,
		},
		{
			name: "fail - invalid BaseTx",
			ticket: TicketCreate{
				BaseTx: BaseTx{
					Account:         "",
					TransactionType: TicketCreateTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        50,
				},
				TicketCount: 5,
			},
			wantValid: false,
		},
		{
			name: "fail - ticket count zero",
			ticket: TicketCreate{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: TicketCreateTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        50,
				},
				TicketCount: 0,
			},
			wantValid: false,
		},
		{
			name: "fail - ticket count exceeds limit",
			ticket: TicketCreate{
				BaseTx: BaseTx{
					Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType: TicketCreateTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        50,
				},
				TicketCount: 251,
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.ticket.Validate()
			if valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v, err: %v", valid, tt.wantValid, err)
			}
		})
	}
}
