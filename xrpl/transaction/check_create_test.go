package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckCreate_TxType(t *testing.T) {
	tx := &CheckCreate{}
	assert.Equal(t, CheckCreateTx, tx.TxType())
}

func TestCheckCreate_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		tx       *CheckCreate
		expected FlatTransaction
	}{
		{
			name: "pass - All fields",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: CheckCreateTx,
				},
				Destination:    "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				SendMax:        types.XRPCurrencyAmount(10000),
				DestinationTag: types.DestinationTag(23480),
				Expiration:     533257958,
				InvoiceID:      "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			expected: FlatTransaction{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "CheckCreate",
				"Destination":     "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"SendMax":         "10000",
				"DestinationTag":  uint32(23480),
				"Expiration":      uint32(533257958),
				"InvoiceID":       "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
		},
		{
			name: "pass - Optional fields omitted",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: CheckCreateTx,
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				SendMax:     types.XRPCurrencyAmount(10000),
			},
			expected: FlatTransaction{
				"Account":         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				"TransactionType": "CheckCreate",
				"Destination":     "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				"SendMax":         "10000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tx.Flatten())
		})
	}
}

func TestCheckCreate_Validate(t *testing.T) {
	tests := []struct {
		name        string
		tx          *CheckCreate
		wantValid   bool
		wantErr     bool
		expectedErr error
	}{
		{
			name: "pass - All fields valid",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: CheckCreateTx,
				},
				Destination:    "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				SendMax:        types.XRPCurrencyAmount(10000),
				DestinationTag: types.DestinationTag(23480),
				Expiration:     533257958,
				InvoiceID:      "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			wantValid:   true,
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name: "fail - BaseTx missing TransactionType",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account: "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
				},
				Destination:    "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				SendMax:        types.XRPCurrencyAmount(10000),
				DestinationTag: types.DestinationTag(23480),
				Expiration:     533257958,
				InvoiceID:      "A0258020E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855810100",
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidTransactionType,
		},
		{
			name: "fail - Invalid destination address",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: CheckCreateTx,
				},
				Destination: "invalidAddress",
				SendMax:     types.XRPCurrencyAmount(10000),
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidDestination,
		},
		{
			name: "fail - Invalid SendMax amount, missing Issuer",
			tx: &CheckCreate{
				BaseTx: BaseTx{
					Account:         "rLUEXYuLiQptky37CqLcm9USQpPiz5rkpD",
					TransactionType: CheckCreateTx,
				},
				Destination: "rsA2LpzuawewSBQXkiju3YQTMzW13pAAdW",
				SendMax: types.IssuedCurrencyAmount{
					Currency: "USD",
					Value:    "10000",
				},
			},
			wantValid:   false,
			wantErr:     true,
			expectedErr: ErrInvalidTokenFields,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.tx.Validate()
			assert.Equal(t, tt.wantValid, valid)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil && err != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}
