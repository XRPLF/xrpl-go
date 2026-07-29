package ledger

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const largeUInt64 = uint64(9223372036854775807)

func TestMPTUInt64JSONMarshal(t *testing.T) {
	tests := []struct {
		name   string
		entry  any
		want   map[string]string // field -> expected quoted decimal value
		absent []string          // optional amount fields omitted at zero
	}{
		{
			name:   "MPToken required zero and omitted optional zero",
			entry:  MPToken{},
			want:   map[string]string{"MPTAmount": "0"},
			absent: []string{"LockedAmount"},
		},
		{
			name:  "MPToken large exact values",
			entry: MPToken{MPTAmount: largeUInt64, LockedAmount: largeUInt64},
			want: map[string]string{
				"MPTAmount":    "9223372036854775807",
				"LockedAmount": "9223372036854775807",
			},
		},
		{
			name:   "MPTokenIssuance required zero and omitted optional zeros",
			entry:  MPTokenIssuance{},
			want:   map[string]string{"OutstandingAmount": "0"},
			absent: []string{"MaximumAmount", "LockedAmount"},
		},
		{
			name: "MPTokenIssuance large exact values",
			entry: MPTokenIssuance{
				MaximumAmount:     largeUInt64,
				OutstandingAmount: largeUInt64,
				LockedAmount:      largeUInt64,
			},
			want: map[string]string{
				"MaximumAmount":     "9223372036854775807",
				"OutstandingAmount": "9223372036854775807",
				"LockedAmount":      "9223372036854775807",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := json.Marshal(test.entry)
			require.NoError(t, err)

			fields := make(map[string]json.RawMessage)
			require.NoError(t, json.Unmarshal(encoded, &fields))
			for field, want := range test.want {
				require.JSONEq(t, fmt.Sprintf("%q", want), string(fields[field]), field)
			}
			for _, field := range test.absent {
				require.NotContains(t, fields, field)
			}
		})
	}
}

func TestMPTUInt64JSONUnmarshal(t *testing.T) {
	t.Run("valid values and optional omission", func(t *testing.T) {
		var zeroMPToken MPToken
		require.NoError(t, json.Unmarshal([]byte(`{"MPTAmount":"0"}`), &zeroMPToken))
		require.Zero(t, zeroMPToken.MPTAmount)
		require.Zero(t, zeroMPToken.LockedAmount)

		var mpToken MPToken
		require.NoError(t, json.Unmarshal([]byte(`{"MPTAmount":"9223372036854775807"}`), &mpToken))
		require.Equal(t, largeUInt64, mpToken.MPTAmount)
		require.Zero(t, mpToken.LockedAmount)

		var zeroIssuance MPTokenIssuance
		require.NoError(t, json.Unmarshal([]byte(`{"MaximumAmount":"0","OutstandingAmount":"0"}`), &zeroIssuance))
		require.Zero(t, zeroIssuance.MaximumAmount)
		require.Zero(t, zeroIssuance.OutstandingAmount)

		var issuance MPTokenIssuance
		require.NoError(t, json.Unmarshal([]byte(`{
			"MaximumAmount":"9223372036854775807",
			"OutstandingAmount":"0",
			"LockedAmount":"9223372036854775807"
		}`), &issuance))
		require.Equal(t, largeUInt64, issuance.MaximumAmount)
		require.Zero(t, issuance.OutstandingAmount)
		require.Equal(t, largeUInt64, issuance.LockedAmount)
	})

	fields := []struct {
		name   string
		field  string
		target func() any
	}{
		{name: "MPToken MPTAmount", field: "MPTAmount", target: func() any { return new(MPToken) }},
		{name: "MPToken LockedAmount", field: "LockedAmount", target: func() any { return new(MPToken) }},
		{name: "MPTokenIssuance MaximumAmount", field: "MaximumAmount", target: func() any { return new(MPTokenIssuance) }},
		{name: "MPTokenIssuance OutstandingAmount", field: "OutstandingAmount", target: func() any { return new(MPTokenIssuance) }},
		{name: "MPTokenIssuance LockedAmount", field: "LockedAmount", target: func() any { return new(MPTokenIssuance) }},
	}
	invalidValues := []struct {
		name  string
		value string
	}{
		{name: "unquoted number", value: `1`},
		{name: "empty string", value: `""`},
		{name: "non-decimal string", value: `"invalid"`},
		{name: "negative string", value: `"-1"`},
		{name: "fractional string", value: `"1.5"`},
		{name: "overflow", value: `"18446744073709551616"`},
	}

	for _, field := range fields {
		for _, invalid := range invalidValues {
			t.Run(field.name+"/"+invalid.name, func(t *testing.T) {
				data := fmt.Appendf(nil, `{%q:%s}`, field.field, invalid.value)
				require.Error(t, json.Unmarshal(data, field.target()))
			})
		}
	}
}
