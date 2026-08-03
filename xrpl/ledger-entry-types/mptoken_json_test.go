package ledger

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const largeUInt64 = uint64(9223372036854775807)

func TestMPTJSONMarshal(t *testing.T) {
	tests := []struct {
		name   string
		entry  any
		want   map[string]string // field -> expected quoted value
		absent []string          // default or optional amount fields omitted at zero
	}{
		{
			name:  "pass - MPToken omitted default amounts and hexadecimal OwnerNode",
			entry: MPToken{},
			want: map[string]string{
				"OwnerNode": "0000000000000000",
			},
			absent: []string{"MPTAmount", "LockedAmount"},
		},
		{
			name: "pass - MPToken large exact values",
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
			name:  "pass - MPTokenIssuance required zero and omitted optional zeros",
			entry: MPTokenIssuance{},
			want: map[string]string{
				"OutstandingAmount": "0",
				"OwnerNode":         "0000000000000000",
			},
			absent: []string{"MaximumAmount", "LockedAmount"},
		},
		{
			name: "pass - MPTokenIssuance large exact values",
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

func TestMPTJSONUnmarshal(t *testing.T) {
	t.Run("pass - canonical strings and optional omission", func(t *testing.T) {
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

	t.Run("pass - absent fields preserve existing values", func(t *testing.T) {
		mpToken := MPToken{
			Flags:        1,
			MPTAmount:    1,
			LockedAmount: 2,
			OwnerNode:    3,
		}
		require.NoError(t, json.Unmarshal([]byte(`{"MPTAmount":"4"}`), &mpToken))
		require.Equal(t, uint32(1), mpToken.Flags)
		require.Equal(t, uint64(4), mpToken.MPTAmount)
		require.Equal(t, uint64(2), mpToken.LockedAmount)
		require.Equal(t, uint64(3), mpToken.OwnerNode)

		issuance := MPTokenIssuance{
			Flags:             1,
			MaximumAmount:     2,
			OutstandingAmount: 3,
			LockedAmount:      4,
			OwnerNode:         5,
		}
		require.NoError(t, json.Unmarshal([]byte(`{"OutstandingAmount":"6"}`), &issuance))
		require.Equal(t, uint32(1), issuance.Flags)
		require.Equal(t, uint64(2), issuance.MaximumAmount)
		require.Equal(t, uint64(6), issuance.OutstandingAmount)
		require.Equal(t, uint64(4), issuance.LockedAmount)
		require.Equal(t, uint64(5), issuance.OwnerNode)
	})
}
