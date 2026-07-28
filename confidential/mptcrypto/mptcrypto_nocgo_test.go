//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package mptcrypto_test

import (
	"math"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestDecryptAmountWithoutCgo(t *testing.T) {
	_, err := mptcrypto.DecryptAmount(mptcrypto.Ciphertext{}, mptcrypto.PrivateKey{}, 0, 0)
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}

func TestDecryptAmountValidatesRangeWithoutCgo(t *testing.T) {
	tests := []struct {
		name      string
		rangeLow  uint64
		rangeHigh uint64
	}{
		{name: "low exceeds high", rangeLow: 2, rangeHigh: 1},
		{name: "high is max uint64", rangeLow: 0, rangeHigh: math.MaxUint64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mptcrypto.DecryptAmount(mptcrypto.Ciphertext{}, mptcrypto.PrivateKey{}, tt.rangeLow, tt.rangeHigh)
			require.ErrorIs(t, err, mptcrypto.ErrInvalidAmountRange)
			require.NotErrorIs(t, err, mptcrypto.ErrCgoRequired)
		})
	}
}
