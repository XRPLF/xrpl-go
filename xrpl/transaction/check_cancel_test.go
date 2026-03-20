package transaction

import (
	"errors"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCancel_TxType(t *testing.T) {
	tx := &CheckCancel{}
	assert.Equal(t, CheckCancelTx, tx.TxType())
}

func TestCheckCancel_Flatten(t *testing.T) {
	tx := &CheckCancel{
		BaseTx: BaseTx{
			Account:  "rMbnCbYfoFo47di25iD2tsJXW83Wpq85Xe",
			Fee:      types.XRPCurrencyAmount(10),
			Sequence: 1,
		},
		CheckID: "ABC123DEF456",
	}

	expected := FlatTransaction{
		"TransactionType": "CheckCancel",
		"Account":         "rMbnCbYfoFo47di25iD2tsJXW83Wpq85Xe",
		"Fee":             "10",
		"Sequence":        uint32(1),
		"CheckID":         "ABC123DEF456",
	}

	require.Equal(t, expected, tx.Flatten())
}

func TestCheckCancel_Validate(t *testing.T) {
	tests := []struct {
		name        string
		tx          *CheckCancel
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - valid CheckCancel",
			tx: &CheckCancel{
				BaseTx: BaseTx{
					Account:         "rMbnCbYfoFo47di25iD2tsJXW83Wpq85Xe",
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        1,
					TransactionType: CheckCancelTx,
				},
				CheckID: "ABC123DEF4567890ABC123DEF4567890ABC123DEF4567890ABC123DEF4567890",
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - invalid CheckID",
			tx: &CheckCancel{
				BaseTx: BaseTx{
					Account:         "rMbnCbYfoFo47di25iD2tsJXW83Wpq85Xe",
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        1,
					TransactionType: CheckCancelTx,
				},
				CheckID: "INVALID_CHECK_ID",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCheckID,
		},
		{
			name: "fail - missing Account in BaseTx",
			tx: &CheckCancel{
				BaseTx: BaseTx{
					TransactionType: CheckCancelTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        1,
				},
				CheckID: "ABC123DEF4567890ABC123DEF4567890ABC123DEF4567890ABC123DEF4567890",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - CheckID length is not 64",
			tx: &CheckCancel{
				BaseTx: BaseTx{
					Account:         "rMbnCbYfoFo47di25iD2tsJXW83Wpq85Xe",
					TransactionType: CheckCancelTx,
					Fee:             types.XRPCurrencyAmount(10),
					Sequence:        1,
				},
				CheckID: "ABC123DEF4567890ABC123D",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCheckID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}
