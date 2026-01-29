package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEscrowCancel_TxType(t *testing.T) {
	entry := &EscrowCancel{}
	assert.Equal(t, EscrowCancelTx, entry.TxType())
}

func TestEscrowCancel_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		escrow   *EscrowCancel
		expected string
	}{
		{
			name: "pass - complete EscrowCancel",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Owner:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				OfferSequence: 7,
			},
			expected: `{
				"TransactionType": "EscrowCancel",
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Owner":           "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"OfferSequence":   7
			}`,
		},
		{
			name: "pass - EscrowCancel without Owner",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				OfferSequence: 7,
			},
			expected: `{
				"TransactionType": "EscrowCancel",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"OfferSequence": 7
			}`,
		},
		{
			name: "pass - EscrowCancel without OfferSequence",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Owner: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
			expected: `{
				"TransactionType": "EscrowCancel",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"Owner": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
			}`,
		},
		{
			name: "pass - EscrowCancel with only BaseTx",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
			},
			expected: `{
				"TransactionType": "EscrowCancel",
				"Account": "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.CompareFlattenAndExpected(tt.escrow.Flatten(), []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestEscrowCancel_Validate(t *testing.T) {
	tests := []struct {
		name      string
		escrow    *EscrowCancel
		wantValid bool
		wantErr   bool
	}{
		{
			name: "pass - valid EscrowCancel",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCancelTx,
				},
				Owner:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				OfferSequence: 7,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid EscrowCancel BaseTx",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					TransactionType: EscrowCancelTx,
				},
				Owner:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				OfferSequence: 7,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - invalid Owner address",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCancelTx,
				},
				Owner:         "invalidAddress",
				OfferSequence: 7,
			},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "fail - missing OfferSequence",
			escrow: &EscrowCancel{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: EscrowCancelTx,
				},
				Owner: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.escrow.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("escrowCancel.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("escrowCancel.Validate() = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}
