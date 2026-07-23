//go:build cgo && !js && !wasip1 && !tinygo && !gofuzz && (linux || darwin) && (amd64 || arm64)

package commitment_test

import (
	"math"
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/commitment"
	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	tests := []struct {
		name   string
		amount uint64
	}{
		{"pass - zero amount", 0},
		{"pass - small amount", 42},
		{"pass - one million", 1_000_000},
		{"pass - max uint64", math.MaxUint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := commitment.Create(tt.amount, bf)
			require.NoError(t, err)
			require.Len(t, c, mptcrypto.CommitmentSize*2)

			prefix := c[:2]
			require.Contains(t, []string{"02", "03"}, prefix, "commitment prefix: got %q", prefix)
		})
	}
}

func TestCreateDeterministic(t *testing.T) {
	bf, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	c1, err := commitment.Create(100, bf)
	require.NoError(t, err)

	c2, err := commitment.Create(100, bf)
	require.NoError(t, err)

	require.Equal(t, c1, c2, "same inputs should produce the same commitment")
}

func TestCreateDifferentInputs(t *testing.T) {
	bf1, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)
	bf2, err := elgamal.GenerateBlindingFactor()
	require.NoError(t, err)

	t.Run("pass - different amounts, same blinding factor", func(t *testing.T) {
		c1, err := commitment.Create(100, bf1)
		require.NoError(t, err)
		c2, err := commitment.Create(200, bf1)
		require.NoError(t, err)
		require.NotEqual(t, c1, c2)
	})

	t.Run("pass - same amount, different blinding factors", func(t *testing.T) {
		c1, err := commitment.Create(100, bf1)
		require.NoError(t, err)
		c2, err := commitment.Create(100, bf2)
		require.NoError(t, err)
		require.NotEqual(t, c1, c2)
	})
}

func TestCreateInvalidBlindingFactor(t *testing.T) {
	tests := []struct {
		name string
		bf   string
	}{
		{"fail - bad hex", "not-valid-hex"},
		{"fail - too short", "0102"},
		{"fail - empty", ""},
		{"fail - too long", strings.Repeat("ab", 33)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := commitment.Create(42, tt.bf)
			require.ErrorIs(t, err, commitment.ErrInvalidBlindingFactor)
		})
	}
}
