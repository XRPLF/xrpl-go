package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckCash_TxType(t *testing.T) {
	tx := &CheckCash{}
	assert.Equal(t, CheckCashTx, tx.TxType())
}

func TestCheckCash_Flatten(t *testing.T) {
	tests := []struct {
		name      string
		checkCash CheckCash
		expected  FlatTransaction
	}{
		{
			name: "pass - CheckCash with Amount",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: types.Hash256("838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334"),
				Amount:  types.XRPCurrencyAmount(100000000),
			},
			expected: FlatTransaction{
				"Account":         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
				"TransactionType": "CheckCash",
				"CheckID":         "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				"Amount":          "100000000",
				"Fee":             "12",
			},
		},
		{
			name: "pass - CheckCash with DeliverMin",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: types.Hash256("838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334"),
				DeliverMin: types.IssuedCurrencyAmount{
					Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Value:    "50000000",
					Currency: "USD",
				},
			},
			expected: FlatTransaction{
				"Account":         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
				"TransactionType": "CheckCash",
				"CheckID":         "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				"DeliverMin": map[string]interface{}{
					"issuer":   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					"value":    "50000000",
					"currency": "USD",
				},
				"Fee": "12",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.checkCash.Flatten())
		})
	}
}

func TestCheckCash_Validate(t *testing.T) {
	tests := []struct {
		name        string
		checkCash   CheckCash
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - CheckCash with Amount",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				Amount:  types.XRPCurrencyAmount(100000000),
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - CheckCash BaseTx without Account",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				Amount:  types.XRPCurrencyAmount(100000000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "pass - CheckCash with DeliverMin",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				DeliverMin: types.IssuedCurrencyAmount{
					Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Value:    "50000000",
					Currency: "USD",
				},
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "fail - both Amount and DeliverMin provided",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				Amount:  types.XRPCurrencyAmount(100000000),
				DeliverMin: types.IssuedCurrencyAmount{
					Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Value:    "50000000",
					Currency: "USD",
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrAmountOrDeliverMinNotProvided,
		},
		{
			name: "invalid - neither Amount nor DeliverMin provided",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrMutuallyExclusiveAmountDeliverMin,
		},
		{
			name: "invalid - invalid CheckID",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "invalidCheckID",
				Amount:  types.XRPCurrencyAmount(100000000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCheckID,
		},
		{
			name: "invalid - invalid CheckID, length is not 64 characters",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B",
				Amount:  types.XRPCurrencyAmount(100000000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidCheckID,
		},
		{
			name: "fail - Invalid Amount",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Value:    "invalid",
					Currency: "USD",
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidTokenValue,
		},
		{
			name: "fail - Invalid DeliverMin",
			checkCash: CheckCash{
				BaseTx: BaseTx{
					Account:         "rfkE1aSy9G8Upk4JssnwBxhEv5p4mn2KTy",
					TransactionType: "CheckCash",
					Fee:             types.XRPCurrencyAmount(12),
				},
				CheckID: "838766BA2B995C00744175F69A1B11E32C3DBC40E64801A4056FCBD657F57334",
				DeliverMin: types.IssuedCurrencyAmount{
					Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					Value:    "invalid",
					Currency: "USD",
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidTokenValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.checkCash.Validate()
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil && err != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}
