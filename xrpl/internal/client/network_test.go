package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func TestResolveNetworkIdentity(t *testing.T) {
	tests := []struct {
		name             string
		override         *uint32
		discovered       NetworkIdentity
		expectedID       *uint32
		expectedErr      error
		preserveOverride bool
	}{
		{
			name:        "missing network ID",
			discovered:  NetworkIdentity{BuildVersion: "1.12.0"},
			expectedErr: ErrNetworkIDUnavailable,
		},
		{
			name: "valid zero",
			discovered: NetworkIdentity{
				NetworkID:    uint32Pointer(0),
				BuildVersion: "1.12.0",
			},
			expectedID: uint32Pointer(0),
		},
		{
			name:     "matching override is preserved",
			override: uint32Pointer(21337),
			discovered: NetworkIdentity{
				NetworkID:    uint32Pointer(21337),
				BuildVersion: "1.12.0",
			},
			expectedID:       uint32Pointer(21337),
			preserveOverride: true,
		},
		{
			name:     "mismatching override",
			override: uint32Pointer(1),
			discovered: NetworkIdentity{
				NetworkID:    uint32Pointer(2),
				BuildVersion: "1.12.0",
			},
			expectedErr: ErrNetworkIDOverrideMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := ResolveNetworkIdentity(tt.override, tt.discovered)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resolved.NetworkID)
			require.Equal(t, *tt.expectedID, *resolved.NetworkID)
			if tt.preserveOverride {
				require.Same(t, tt.override, resolved.NetworkID)
			}
		})
	}
}

func TestNetworkIDRequired(t *testing.T) {
	tests := []struct {
		name     string
		identity NetworkIdentity
		required bool
		expected error
	}{
		{name: "unknown ID", identity: NetworkIdentity{}, expected: ErrNetworkIDUnavailable},
		{name: "mainnet zero", identity: NetworkIdentity{NetworkID: uint32Pointer(0)}, required: false},
		{name: "unrestricted boundary", identity: NetworkIdentity{NetworkID: uint32Pointer(1024)}, required: false},
		{name: "restricted missing version", identity: NetworkIdentity{NetworkID: uint32Pointer(1025)}, expected: ErrBuildVersionUnavailable},
		{name: "restricted invalid version", identity: NetworkIdentity{NetworkID: uint32Pointer(1025), BuildVersion: "invalid"}, expected: ErrInvalidBuildVersion},
		{name: "pre 1.11", identity: NetworkIdentity{NetworkID: uint32Pointer(1025), BuildVersion: "1.10.9"}, required: false},
		{name: "1.11 beta", identity: NetworkIdentity{NetworkID: uint32Pointer(1025), BuildVersion: "1.11.0-b1"}, required: false},
		{name: "1.11 release candidate", identity: NetworkIdentity{NetworkID: uint32Pointer(1025), BuildVersion: "1.11.0-rc1"}, required: false},
		{name: "1.11 exact", identity: NetworkIdentity{NetworkID: uint32Pointer(1025), BuildVersion: "1.11.0"}, required: true},
		{name: "post 1.11", identity: NetworkIdentity{NetworkID: uint32Pointer(21337), BuildVersion: "2.2.3"}, required: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			required, err := NetworkIDRequired(tt.identity)
			if tt.expected != nil {
				require.ErrorIs(t, err, tt.expected)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.required, required)
		})
	}
}

func TestApplyNetworkIDPolicy(t *testing.T) {
	restricted := NetworkIdentity{NetworkID: uint32Pointer(21337), BuildVersion: "1.12.0"}

	t.Run("applies to outer and Batch inner transactions", func(t *testing.T) {
		tx := map[string]any{
			"TransactionType": "Batch",
			"RawTransactions": []map[string]any{
				{"RawTransaction": map[string]any{"TransactionType": "Payment"}},
				{"RawTransaction": map[string]any{"TransactionType": "OfferCreate"}},
			},
		}

		require.NoError(t, ApplyNetworkIDPolicy(tx, restricted))
		require.Equal(t, uint32(21337), tx["NetworkID"])
		for _, wrapper := range tx["RawTransactions"].([]map[string]any) {
			inner := wrapper["RawTransaction"].(map[string]any)
			require.Equal(t, uint32(21337), inner["NetworkID"])
		}
	})

	t.Run("matching explicit value is preserved", func(t *testing.T) {
		tx := map[string]any{"NetworkID": uint32(21337)}
		require.NoError(t, ApplyNetworkIDPolicy(tx, restricted))
		require.Equal(t, uint32(21337), tx["NetworkID"])
	})

	t.Run("mismatching explicit value", func(t *testing.T) {
		err := ApplyNetworkIDPolicy(map[string]any{"NetworkID": uint32(21338)}, restricted)
		require.ErrorIs(t, err, ErrNetworkIDFieldMismatch)
	})

	t.Run("invalid explicit type", func(t *testing.T) {
		err := ApplyNetworkIDPolicy(map[string]any{"NetworkID": "21337"}, restricted)
		require.ErrorIs(t, err, ErrNetworkIDFieldIsNotAUint32)
	})

	for _, tt := range []struct {
		name     string
		identity NetworkIdentity
	}{
		{name: "unrestricted network", identity: NetworkIdentity{NetworkID: uint32Pointer(1), BuildVersion: "1.12.0"}},
		{name: "pre 1.11 restricted network", identity: NetworkIdentity{NetworkID: uint32Pointer(21337), BuildVersion: "1.10.0"}},
	} {
		t.Run(tt.name+" omits NetworkID", func(t *testing.T) {
			tx := map[string]any{}
			require.NoError(t, ApplyNetworkIDPolicy(tx, tt.identity))
			require.NotContains(t, tx, "NetworkID")
		})
		t.Run(tt.name+" rejects explicit NetworkID", func(t *testing.T) {
			tx := map[string]any{"NetworkID": *tt.identity.NetworkID}
			err := ApplyNetworkIDPolicy(tx, tt.identity)
			require.ErrorIs(t, err, ErrNetworkIDFieldUnexpected)
		})
	}
}
