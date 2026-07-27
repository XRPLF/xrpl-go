package transaction

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlatTransaction_NormalizeFlagsOK(t *testing.T) {
	tests := []struct {
		name     string
		flags    any
		present  bool
		expected uint32
	}{
		{name: "missing defaults to 0", present: false, expected: 0},
		{name: "uint32 preserved", flags: uint32(131072), present: true, expected: 131072},
		{name: "uint8 coerced", flags: uint8(8), present: true, expected: 8},
		{name: "uint16 coerced", flags: uint16(8), present: true, expected: 8},
		{name: "int coerced", flags: 131072, present: true, expected: 131072},
		{name: "int8 coerced", flags: int8(8), present: true, expected: 8},
		{name: "int16 coerced", flags: int16(8), present: true, expected: 8},
		{name: "int64 coerced", flags: int64(8), present: true, expected: 8},
		{name: "uint64 coerced", flags: uint64(2147483648), present: true, expected: 2147483648},
		{name: "max uint32", flags: int64(math.MaxUint32), present: true, expected: math.MaxUint32},
		{name: "whole float64 coerced", flags: float64(131072), present: true, expected: 131072},
		{name: "whole float32 coerced", flags: float32(8), present: true, expected: 8},
		{name: "json.Number integer coerced", flags: json.Number("131072"), present: true, expected: 131072},
		{name: "json.Number whole float coerced", flags: json.Number("131072.0"), present: true, expected: 131072},
		{name: "zero int", flags: 0, present: true, expected: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := FlatTransaction{"TransactionType": string(PaymentTx)}
			if tt.present {
				tx["Flags"] = tt.flags
			}

			require.NoError(t, tx.NormalizeFlags())
			assert.Equal(t, tt.expected, tx["Flags"])
		})
	}
}

func TestFlatTransaction_NormalizeFlagsErr(t *testing.T) {
	tests := []struct {
		name  string
		flags any
	}{
		{name: "negative int", flags: -1},
		{name: "int64 above max uint32", flags: int64(math.MaxUint32) + 1},
		{name: "uint64 above max uint32", flags: uint64(math.MaxUint32) + 1},
		{name: "fractional float64", flags: float64(1.5)},
		{name: "negative float64", flags: float64(-2)},
		{name: "float64 above max uint32", flags: float64(math.MaxUint32) + 1},
		{name: "json.Number above max uint32", flags: json.Number("4294967296")},
		{name: "json.Number fractional", flags: json.Number("1.5")},
		{name: "json.Number rounds down to max uint32", flags: json.Number("4294967295.0000001")},
		{name: "json.Number rounds up to max uint32", flags: json.Number("4294967294.9999999999")},
		{name: "json.Number garbage", flags: json.Number("abc")},
		{name: "unsupported type", flags: "131072"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := FlatTransaction{"TransactionType": string(PaymentTx), "Flags": tt.flags}

			require.ErrorIs(t, tx.NormalizeFlags(), ErrInvalidFlagsValue)
		})
	}
}

func TestFlatTransaction_RequireTransactionType(t *testing.T) {
	tests := []struct {
		name    string
		tx      FlatTransaction
		wantErr error
	}{
		{
			name:    "present",
			tx:      FlatTransaction{"TransactionType": string(PaymentTx)},
			wantErr: nil,
		},
		{
			name:    "missing",
			tx:      FlatTransaction{"Flags": uint32(1)},
			wantErr: ErrTransactionTypeMissing,
		},
		{
			name:    "wrong type",
			tx:      FlatTransaction{"TransactionType": 42},
			wantErr: ErrTransactionTypeMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.RequireTransactionType()
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
