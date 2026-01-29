package transaction

import (
	"reflect"
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestClawback_Flatten(t *testing.T) {
	s := Clawback{
		BaseTx: BaseTx{
			Account:         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			TransactionType: ClawbackTx,
			Fee:             types.XRPCurrencyAmount(1),
			Sequence:        1234,
		},
		Amount: types.IssuedCurrencyAmount{
			Issuer:   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			Currency: "USD",
			Value:    "1",
		},
	}

	flattened := s.Flatten()

	expected := FlatTransaction{
		"Account":         "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"TransactionType": "Clawback",
		"Fee":             "1",
		"Sequence":        uint32(1234),
		"Amount": map[string]interface{}{
			"issuer":   "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			"currency": "USD",
			"value":    "1",
		},
	}

	if !reflect.DeepEqual(flattened, expected) {
		t.Errorf("Flatten result differs from expected: %v, %v", flattened, expected)
	}
}
func TestClawback_Validate(t *testing.T) {
	tests := []struct {
		name       string
		clawback   Clawback
		shouldPass bool
	}{
		{
			name: "pass - valid Clawback transaction",
			clawback: Clawback{
				BaseTx: BaseTx{
					Account:         "rnLYcEcYw2r3w6BDsFDSScoFmvZXbwa6EQ",
					TransactionType: ClawbackTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rhsTg7mm7v3oEGrF85n1KdB3JjCk5KPT4M",
					Currency: "USD",
					Value:    "1",
				},
			},
			shouldPass: true,
		},
		{
			name: "fail - clawback transaction with missing Amount field",
			clawback: Clawback{
				BaseTx: BaseTx{
					Account:         "rnLYcEcYw2r3w6BDsFDSScoFmvZXbwa6EQ",
					TransactionType: ClawbackTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
			},
			shouldPass: false,
		},
		{
			name: "fail - clawback transaction with invalid Amount",
			clawback: Clawback{
				BaseTx: BaseTx{
					Account:         "rL2ek7KyeTk6NiyJcxYFrfiPcv8PFVoqgR",
					TransactionType: ClawbackTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rnLYcEcYw2r3w6BDsFDSScoFmvZXbwa6EQ",
					Currency: "USD",
					Value:    "invalid",
				},
			},
			shouldPass: false,
		},
		{
			name: "fail - clawback transaction with Account same as the issuer",
			clawback: Clawback{
				BaseTx: BaseTx{
					Account:         "rL2ek7KyeTk6NiyJcxYFrfiPcv8PFVoqgR",
					TransactionType: ClawbackTx,
					Fee:             types.XRPCurrencyAmount(1),
					Sequence:        1234,
					SigningPubKey:   "ghijk",
					TxnSignature:    "A1B2C3D4E5F6",
				},
				Amount: types.IssuedCurrencyAmount{
					Issuer:   "rL2ek7KyeTk6NiyJcxYFrfiPcv8PFVoqgR",
					Currency: "USD",
					Value:    "1",
				},
			},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.clawback.Validate()
			if tt.shouldPass {
				if err != nil {
					t.Errorf("Validation failed for valid Clawback transaction with error: %s", err.Error())
				}
				if !valid {
					t.Error("Validation should pass for valid Clawback transaction")
				}
			} else {
				if err == nil || valid {
					t.Error("Validation should fail for invalid Clawback transaction")
				}
			}
		})
	}
}
