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
		want   map[string]string // field -> expected quoted value
		absent []string          // default or optional amount fields omitted at zero
	}{
		{
			name:  "MPToken omitted default amounts and hexadecimal OwnerNode",
			entry: MPToken{},
			want: map[string]string{
				"OwnerNode": "0000000000000000",
			},
			absent: []string{"MPTAmount", "LockedAmount"},
		},
		{
			name: "MPToken large exact values",
			entry: MPToken{
				MPTAmount:    largeUInt64,
				LockedAmount: largeUInt64,
				OwnerNode:    largeUInt64,
			},
			want: map[string]string{
				"MPTAmount":    "9223372036854775807",
				"LockedAmount": "9223372036854775807",
				"OwnerNode":    "7FFFFFFFFFFFFFFF",
			},
		},
		{
			name:  "MPTokenIssuance required zero and omitted optional zeros",
			entry: MPTokenIssuance{},
			want: map[string]string{
				"OutstandingAmount": "0",
				"OwnerNode":         "0000000000000000",
			},
			absent: []string{"MaximumAmount", "LockedAmount"},
		},
		{
			name: "MPTokenIssuance large exact values",
			entry: MPTokenIssuance{
				MaximumAmount:     largeUInt64,
				OutstandingAmount: largeUInt64,
				LockedAmount:      largeUInt64,
				OwnerNode:         largeUInt64,
			},
			want: map[string]string{
				"MaximumAmount":     "9223372036854775807",
				"OutstandingAmount": "9223372036854775807",
				"LockedAmount":      "9223372036854775807",
				"OwnerNode":         "7FFFFFFFFFFFFFFF",
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
	t.Run("canonical strings and optional omission", func(t *testing.T) {
		var zeroMPToken MPToken
		require.NoError(t, json.Unmarshal([]byte(`{"OwnerNode":"0000000000000000"}`), &zeroMPToken))
		require.Zero(t, zeroMPToken.MPTAmount)
		require.Zero(t, zeroMPToken.LockedAmount)
		require.Zero(t, zeroMPToken.OwnerNode)

		var mpToken MPToken
		require.NoError(t, json.Unmarshal([]byte(`{
			"MPTAmount":"9223372036854775807",
			"OwnerNode":"7FFFFFFFFFFFFFFF"
		}`), &mpToken))
		require.Equal(t, largeUInt64, mpToken.MPTAmount)
		require.Zero(t, mpToken.LockedAmount)
		require.Equal(t, largeUInt64, mpToken.OwnerNode)

		var rippledOwnerNode MPToken
		require.NoError(t, json.Unmarshal([]byte(`{"OwnerNode":"a"}`), &rippledOwnerNode))
		require.Equal(t, uint64(10), rippledOwnerNode.OwnerNode)

		var zeroIssuance MPTokenIssuance
		require.NoError(t, json.Unmarshal([]byte(`{
			"MaximumAmount":"0",
			"OutstandingAmount":"0",
			"OwnerNode":"0000000000000000"
		}`), &zeroIssuance))
		require.Zero(t, zeroIssuance.MaximumAmount)
		require.Zero(t, zeroIssuance.OutstandingAmount)
		require.Zero(t, zeroIssuance.OwnerNode)

		var issuance MPTokenIssuance
		require.NoError(t, json.Unmarshal([]byte(`{
			"MaximumAmount":"9223372036854775807",
			"OutstandingAmount":"0",
			"LockedAmount":"9223372036854775807",
			"OwnerNode":"7FFFFFFFFFFFFFFF"
		}`), &issuance))
		require.Equal(t, largeUInt64, issuance.MaximumAmount)
		require.Zero(t, issuance.OutstandingAmount)
		require.Equal(t, largeUInt64, issuance.LockedAmount)
		require.Equal(t, largeUInt64, issuance.OwnerNode)
	})

	t.Run("xrpl.js-compatible unsigned integer inputs", func(t *testing.T) {
		var mpToken MPToken
		require.NoError(t, json.Unmarshal([]byte(`{"MPTAmount":1,"LockedAmount":2,"OwnerNode":3}`), &mpToken))
		require.Equal(t, uint64(1), mpToken.MPTAmount)
		require.Equal(t, uint64(2), mpToken.LockedAmount)
		require.Equal(t, uint64(3), mpToken.OwnerNode)

		var issuance MPTokenIssuance
		require.NoError(t, json.Unmarshal([]byte(`{
			"MaximumAmount":4,
			"OutstandingAmount":5,
			"LockedAmount":6,
			"OwnerNode":7
		}`), &issuance))
		require.Equal(t, uint64(4), issuance.MaximumAmount)
		require.Equal(t, uint64(5), issuance.OutstandingAmount)
		require.Equal(t, uint64(6), issuance.LockedAmount)
		require.Equal(t, uint64(7), issuance.OwnerNode)
	})

	type uint64Field struct {
		name   string
		field  string
		target func() any
	}
	type invalidValue struct {
		name  string
		value string
	}

	// A 16-digit hex string tops out at exactly MaxUint64, so hex has no
	// reachable overflow case beyond the digit cap; decimal does.
	invalidGroups := []struct {
		fields  []uint64Field
		invalid []invalidValue
	}{
		{
			fields: []uint64Field{
				{name: "MPToken MPTAmount", field: "MPTAmount", target: func() any { return new(MPToken) }},
				{name: "MPToken LockedAmount", field: "LockedAmount", target: func() any { return new(MPToken) }},
				{name: "MPTokenIssuance MaximumAmount", field: "MaximumAmount", target: func() any { return new(MPTokenIssuance) }},
				{name: "MPTokenIssuance OutstandingAmount", field: "OutstandingAmount", target: func() any { return new(MPTokenIssuance) }},
				{name: "MPTokenIssuance LockedAmount", field: "LockedAmount", target: func() any { return new(MPTokenIssuance) }},
			},
			invalid: []invalidValue{
				{name: "empty string", value: `""`},
				{name: "non-decimal string", value: `"invalid"`},
				{name: "negative string", value: `"-1"`},
				{name: "fractional string", value: `"1.5"`},
				{name: "too many decimal digits", value: `"000000000000000000001"`},
				{name: "quoted overflow", value: `"18446744073709551616"`},
				{name: "negative number", value: `-1`},
				{name: "fractional number", value: `1.5`},
				{name: "numeric overflow", value: `18446744073709551616`},
			},
		},
		{
			fields: []uint64Field{
				{name: "MPToken OwnerNode", field: "OwnerNode", target: func() any { return new(MPToken) }},
				{name: "MPTokenIssuance OwnerNode", field: "OwnerNode", target: func() any { return new(MPTokenIssuance) }},
			},
			invalid: []invalidValue{
				{name: "empty string", value: `""`},
				{name: "non-hex string", value: `"invalid"`},
				{name: "too many hexadecimal digits", value: `"00000000000000001"`},
				{name: "negative number", value: `-1`},
				{name: "fractional number", value: `1.5`},
				{name: "numeric overflow", value: `18446744073709551616`},
			},
		},
	}

	for _, group := range invalidGroups {
		for _, field := range group.fields {
			for _, invalid := range group.invalid {
				t.Run(field.name+"/"+invalid.name, func(t *testing.T) {
					data := fmt.Appendf(nil, `{%q:%s}`, field.field, invalid.value)
					require.Error(t, json.Unmarshal(data, field.target()))
				})
			}
		}
	}
}
