package transaction

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/stretchr/testify/assert"
)

func TestOfferCreate_TxType(t *testing.T) {
	tx := &OfferCreate{}
	assert.Equal(t, OfferCreateTx, tx.TxType())
}

func TestOfferCreateFlatten(t *testing.T) {
	tests := []struct {
		name     string
		input    OfferCreate
		expected string
	}{
		{
			name: "pass - with Expiration and OfferSequence",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				Expiration:    6000000,
				OfferSequence: 10,
				TakerGets:     types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: `{
				"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "OfferCreate",
				"Fee": "12",
				"Sequence": 8,
				"LastLedgerSequence": 7108682,
				"Expiration": 6000000,
				"OfferSequence": 10,
				"TakerGets": "6000000",
				"TakerPays": {
					"issuer": "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					"currency": "GKO",
					"value": "2"
				}
			}`,
		},
		{
			name: "pass - without Expiration and OfferSequence",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: `{
				"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "OfferCreate",
				"Fee": "12",
				"Sequence": 8,
				"LastLedgerSequence": 7108682,
				"TakerGets": "6000000",
				"TakerPays": {
					"issuer": "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					"currency": "GKO",
					"value": "2"
				}
			}`,
		},
		{
			name: "pass - with DomainID",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
				DomainID: types.DomainID("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
			},
			expected: `{
				"Account": "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
				"TransactionType": "OfferCreate",
				"Fee": "12",
				"Sequence": 8,
				"LastLedgerSequence": 7108682,
				"TakerGets": "6000000",
				"TakerPays": {
					"issuer": "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					"currency": "GKO",
					"value": "2"
				},
				"DomainID": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()
			err := testutil.CompareFlattenAndExpected(result, []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOfferCreate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		input    OfferCreate
		expected bool
	}{
		{
			name: "pass - valid OfferCreate",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: true,
		},
		{
			name: "fail - invalid OfferCreate, missing Account",
			input: OfferCreate{
				BaseTx: BaseTx{
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid TakerGets",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.IssuedCurrencyAmount{
					Issuer:   "",
					Currency: "",
					Value:    "",
				},
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid TakerPays",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "",
					Currency: "",
					Value:    "",
				},
			},
			expected: false,
		},
		{
			name: "pass - valid OfferCreate with DomainID",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
				DomainID: types.DomainID("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
			},
			expected: true,
		},
		{
			name: "fail - hybrid flag without DomainID",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
					Flags:              TfHybrid,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
			},
			expected: false,
		},
		{
			name: "fail - invalid DomainID length",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
				DomainID: types.DomainID("invalid_length"),
			},
			expected: false,
		},
		{
			name: "pass - hybrid flag with valid DomainID",
			input: OfferCreate{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    OfferCreateTx,
					Fee:                types.XRPCurrencyAmount(12),
					Sequence:           8,
					LastLedgerSequence: 7108682,
					Flags:              TfHybrid,
				},
				TakerGets: types.XRPCurrencyAmount(6000000),
				TakerPays: types.IssuedCurrencyAmount{
					Issuer:   "ruazs5h1qEsqpke88pcqnaseXdm6od2xc",
					Currency: "GKO",
					Value:    "2",
				},
				DomainID: types.DomainID("1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.input.Validate()
			if valid != tt.expected {
				t.Errorf("expected %v, got %v, error: %v", tt.expected, valid, err)
			}
		})
	}
}

func TestOfferCreate_Flags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*OfferCreate)
		expected uint32
	}{
		{
			name: "pass - SetPassiveFlag",
			setter: func(a *OfferCreate) {
				a.SetPassiveFlag()
			},
			expected: TfPassive,
		},
		{
			name: "pass - SetImmediateOrCancelFlag",
			setter: func(a *OfferCreate) {
				a.SetImmediateOrCancelFlag()
			},
			expected: TfImmediateOrCancel,
		},
		{
			name: "pass - SetFillOrKillFlag",
			setter: func(a *OfferCreate) {
				a.SetFillOrKillFlag()
			},
			expected: TfFillOrKill,
		},
		{
			name: "pass - SetSellFlag",
			setter: func(a *OfferCreate) {
				a.SetSellFlag()
			},
			expected: TfSell,
		},
		{
			name: "pass - SetHybridFlag",
			setter: func(a *OfferCreate) {
				a.SetHybridFlag()
			},
			expected: TfHybrid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OfferCreate{}
			tt.setter(o)
			if o.Flags != tt.expected {
				t.Errorf("Expected OfferCreate Flags to be %d, got %d", tt.expected, o.Flags)
			}
		})
	}
}
