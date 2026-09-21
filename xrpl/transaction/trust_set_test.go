package transaction

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func trustSetQuality(value uint32) *uint32 {
	return &value
}

func TestTrustSetFlatten(t *testing.T) {
	s := TrustSet{
		BaseTx: BaseTx{
			Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
			TransactionType:    TrustSetTx,
			Fee:                types.XRPCurrencyAmount(12),
			Flags:              262144,
			Sequence:           12,
			LastLedgerSequence: 8007750,
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Issuer:   "rsP3mgGb2tcYUrxiLFiHJiQXhsziegtwBc",
			Currency: "USD",
			Value:    "100",
		},
	}

	flattened := s.Flatten()

	expected := FlatTransaction{
		"Account":            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
		"TransactionType":    "TrustSet",
		"Fee":                "12",
		"Flags":              uint32(262144),
		"Sequence":           uint32(12),
		"LastLedgerSequence": uint32(8007750),
		"LimitAmount": map[string]any{
			"issuer":   "rsP3mgGb2tcYUrxiLFiHJiQXhsziegtwBc",
			"currency": "USD",
			"value":    "100",
		},
	}

	// Existing DeepEqual check
	if !reflect.DeepEqual(flattened, expected) {
		t.Errorf("Flatten result differs from expected: %v, %v", flattened, expected)
	}
}

func TestTrustSetFlattenQualityPresence(t *testing.T) {
	tx := TrustSet{
		QualityIn:  trustSetQuality(0),
		QualityOut: trustSetQuality(math.MaxUint32),
	}
	flattened := tx.Flatten()

	if value, ok := flattened["QualityIn"]; !ok || value != uint32(0) {
		t.Fatalf("expected explicit QualityIn zero, got %#v", value)
	}
	if value, ok := flattened["QualityOut"]; !ok || value != uint32(math.MaxUint32) {
		t.Fatalf("expected explicit maximum QualityOut, got %#v", value)
	}
	if _, err := binarycodec.Encode(flattened); err != nil {
		t.Fatalf("encode quality fields: %v", err)
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("marshal quality fields: %v", err)
	}
	var decoded TrustSet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal quality fields: %v", err)
	}
	if decoded.QualityIn == nil || *decoded.QualityIn != 0 {
		t.Fatalf("explicit QualityIn zero was not preserved: %#v", decoded.QualityIn)
	}
	if decoded.QualityOut == nil || *decoded.QualityOut != math.MaxUint32 {
		t.Fatalf("maximum QualityOut was not preserved: %#v", decoded.QualityOut)
	}
}

func TestTrustSetFlags(t *testing.T) {
	tests := []struct {
		name     string
		setter   func(*TrustSet)
		expected uint32
	}{
		{
			name: "pass - SetSetAuthFlag",
			setter: func(ts *TrustSet) {
				ts.SetSetAuthFlag()
			},
			expected: TfSetAuth,
		},
		{
			name: "pass - SetSetNoRippleFlag",
			setter: func(ts *TrustSet) {
				ts.SetSetNoRippleFlag()
			},
			expected: TfSetNoRipple,
		},
		{
			name: "pass - SetClearNoRippleFlag",
			setter: func(ts *TrustSet) {
				ts.SetClearNoRippleFlag()
			},
			expected: TfClearNoRipple,
		},
		{
			name: "pass - SetSetfAuthFlag and SetSetNoRippleFlag",
			setter: func(ts *TrustSet) {
				ts.SetSetAuthFlag()
				ts.SetSetNoRippleFlag()
			},
			expected: TfSetAuth | TfSetNoRipple,
		},
		{
			name: "pass - SetSetfAuthFlag and SetClearNoRippleFlag",
			setter: func(ts *TrustSet) {
				ts.SetSetAuthFlag()
				ts.SetClearNoRippleFlag()
			},
			expected: TfSetAuth | TfClearNoRipple,
		},
		{
			name: "pass - SetSetFreezeFlag",
			setter: func(ts *TrustSet) {
				ts.SetSetFreezeFlag()
			},
			expected: TfSetFreeze,
		},
		{
			name: "pass - SetClearFreezeFlag",
			setter: func(ts *TrustSet) {
				ts.SetClearFreezeFlag()
			},
			expected: TfClearFreeze,
		},
		{
			name: "pass - All flags",
			setter: func(ts *TrustSet) {
				ts.SetSetAuthFlag()
				ts.SetSetNoRippleFlag()
				ts.SetClearNoRippleFlag()
				ts.SetSetFreezeFlag()
				ts.SetClearFreezeFlag()
				ts.SetSetDeepFreezeFlag()
				ts.SetClearDeepFreezeFlag()
			},
			expected: TfSetAuth | TfSetNoRipple | TfClearNoRipple | TfSetFreeze | TfClearFreeze | TfSetDeepFreeze | TfClearDeepFreeze,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TrustSet{}
			tt.setter(ts)
			if ts.Flags != tt.expected {
				t.Errorf("Expected Flags to be %d, got %d", tt.expected, ts.Flags)
			}
		})
	}
}

func TestTrustSetValidate(t *testing.T) {
	tests := []struct {
		name     string
		trustSet *TrustSet
		valid    bool
		err      error
	}{
		{
			name: "pass - valid TrustSet",
			trustSet: &TrustSet{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    TrustSetTx,
					Fee:                types.XRPCurrencyAmount(12),
					Flags:              262144,
					Sequence:           12,
					LastLedgerSequence: 8007750,
				},
				LimitAmount: types.IssuedCurrencyAmount{
					Issuer:   "rsP3mgGb2tcYUrxiLFiHJiQXhsziegtwBc",
					Currency: "USD",
					Value:    "100",
				},
				QualityIn:  trustSetQuality(100),
				QualityOut: trustSetQuality(200),
			},
			valid: true,
		},
		{
			name: "fail - missing LimitAmount",
			trustSet: &TrustSet{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    TrustSetTx,
					Fee:                types.XRPCurrencyAmount(12),
					Flags:              262144,
					Sequence:           12,
					LastLedgerSequence: 8007750,
				},
				QualityIn:  trustSetQuality(100),
				QualityOut: trustSetQuality(200),
			},
			valid: false,
		},
		{
			name: "fail - invalid LimitAmount",
			trustSet: &TrustSet{
				BaseTx: BaseTx{
					Account:            "rU6K7V3Po4snVhBBaU29sesqs2qTQJWDw1",
					TransactionType:    TrustSetTx,
					Fee:                types.XRPCurrencyAmount(12),
					Flags:              262144,
					Sequence:           12,
					LastLedgerSequence: 8007750,
				},
				LimitAmount: types.IssuedCurrencyAmount{
					Issuer:   "r123",
					Currency: "USD",
				},
				QualityIn:  trustSetQuality(100),
				QualityOut: trustSetQuality(200),
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.trustSet.Validate()
			if valid != tt.valid {
				t.Errorf("Expected valid to be %v, got %v", tt.valid, valid)
			}
			if (err != nil && tt.valid) || (err == nil && !tt.valid) {
				t.Errorf("Got error: %v", err)
			}
		})
	}
}
