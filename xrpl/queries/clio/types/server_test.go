package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLedgerInfoFeeFieldPresence(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		wantBaseFee       *float64
		wantReserveIncXRP *float64
	}{
		{name: "missing", input: `{}`},
		{name: "null", input: `{"base_fee_xrp":null,"reserve_inc_xrp":null}`},
		{name: "explicit zero", input: `{"base_fee_xrp":0,"reserve_inc_xrp":0}`, wantBaseFee: float64Pointer(0), wantReserveIncXRP: float64Pointer(0)},
		{name: "non-zero", input: `{"base_fee_xrp":0.00001,"reserve_inc_xrp":2.5}`, wantBaseFee: float64Pointer(0.00001), wantReserveIncXRP: float64Pointer(2.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ledger LedgerInfo
			require.NoError(t, json.Unmarshal([]byte(tt.input), &ledger))
			require.Equal(t, tt.wantBaseFee, ledger.BaseFeeXRP)
			require.Equal(t, tt.wantReserveIncXRP, ledger.ReserveIncXRP)
		})
	}
}

func float64Pointer(value float64) *float64 {
	return &value
}
