package client

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveNetworkIdentity(t *testing.T) {
	tests := []struct {
		name             string
		override         *uint32
		discovered       NetworkIdentity
		expectedID       *uint32
		expectedBuild    string
		expectedErr      error
		preserveOverride bool
	}{
		{
			name:          "missing network ID remains unknown",
			discovered:    NetworkIdentity{BuildVersion: "1.12.0"},
			expectedBuild: "1.12.0",
		},
		{
			name:        "missing network ID cannot verify override",
			override:    uint32Pointer(21337),
			discovered:  NetworkIdentity{BuildVersion: "1.12.0"},
			expectedErr: ErrNetworkIDOverrideUnverified,
		},
		{
			name: "valid zero",
			discovered: NetworkIdentity{
				NetworkID:    uint32Pointer(0),
				BuildVersion: "1.12.0",
			},
			expectedID:    uint32Pointer(0),
			expectedBuild: "1.12.0",
		},
		{
			name:     "matching override is preserved",
			override: uint32Pointer(21337),
			discovered: NetworkIdentity{
				NetworkID:    uint32Pointer(21337),
				BuildVersion: "1.12.0",
			},
			expectedID:       uint32Pointer(21337),
			expectedBuild:    "1.12.0",
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
			if tt.expectedID == nil {
				require.Nil(t, resolved.NetworkID)
			} else {
				require.NotNil(t, resolved.NetworkID)
				require.Equal(t, *tt.expectedID, *resolved.NetworkID)
			}
			require.Equal(t, tt.expectedBuild, resolved.BuildVersion)
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
		{name: "unknown ID", identity: NetworkIdentity{}, required: false},
		{name: "mainnet zero", identity: NetworkIdentity{NetworkID: uint32Pointer(0)}, required: false},
		{name: "unrestricted boundary", identity: NetworkIdentity{NetworkID: uint32Pointer(1024)}, required: false},
		{name: "restricted missing version", identity: NetworkIdentity{NetworkID: uint32Pointer(1025)}, required: false},
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

func TestCompareRippledVersions(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		expected int
	}{
		{name: "beta ten after beta two", left: "1.11.0-b10", right: "1.11.0-b2", expected: 1},
		{name: "beta two before beta ten", left: "1.11.0-b2", right: "1.11.0-b10", expected: -1},
		{name: "release candidate ten after release candidate two", left: "1.11.0-rc10", right: "1.11.0-rc2", expected: 1},
		{name: "numeric tail can exceed uint64", left: "1.11.0-b18446744073709551616", right: "1.11.0-b9", expected: 1},
		{name: "leading zeroes do not change numeric tail", left: "1.11.0-b002", right: "1.11.0-b2", expected: 0},
		{name: "beta before release candidate", left: "1.11.0-b10", right: "1.11.0-rc2", expected: -1},
		{name: "missing numeric tail uses text order", left: "1.11.0-b", right: "1.11.0-b2", expected: -1},
		{name: "release after prerelease", left: "1.11.0", right: "1.11.0-rc10", expected: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comparison, err := compareRippledVersions(tt.left, tt.right)
			require.NoError(t, err)
			require.Equal(t, tt.expected, comparison)
		})
	}
}

func TestCompareRippledVersionsPreservesParseError(t *testing.T) {
	_, err := compareRippledVersions("1.invalid.0", "1.0.0")

	var numberError *strconv.NumError
	require.ErrorAs(t, err, &numberError)
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

	t.Run("unknown identity omits NetworkID", func(t *testing.T) {
		tx := map[string]any{}
		require.NoError(t, ApplyNetworkIDPolicy(tx, NetworkIdentity{}))
		require.NotContains(t, tx, "NetworkID")
	})

	t.Run("missing build version omits NetworkID", func(t *testing.T) {
		tx := map[string]any{}
		require.NoError(t, ApplyNetworkIDPolicy(tx, NetworkIdentity{NetworkID: uint32Pointer(21337)}))
		require.NotContains(t, tx, "NetworkID")
	})

	t.Run("unknown identity preserves explicit NetworkID", func(t *testing.T) {
		tx := map[string]any{"NetworkID": uint32(21337)}
		require.NoError(t, ApplyNetworkIDPolicy(tx, NetworkIdentity{}))
		require.Equal(t, uint32(21337), tx["NetworkID"])
	})

	t.Run("missing build version preserves matching explicit NetworkID", func(t *testing.T) {
		tx := map[string]any{"NetworkID": uint32(21337)}
		identity := NetworkIdentity{NetworkID: uint32Pointer(21337)}
		require.NoError(t, ApplyNetworkIDPolicy(tx, identity))
		require.Equal(t, uint32(21337), tx["NetworkID"])
	})

	t.Run("missing build version rejects mismatching explicit NetworkID", func(t *testing.T) {
		tx := map[string]any{"NetworkID": uint32(9999)}
		identity := NetworkIdentity{NetworkID: uint32Pointer(21337)}
		err := ApplyNetworkIDPolicy(tx, identity)
		require.ErrorIs(t, err, ErrNetworkIDFieldMismatch)
		require.Equal(t, uint32(9999), tx["NetworkID"])
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

func uint32Pointer(value uint32) *uint32 {
	return &value
}
