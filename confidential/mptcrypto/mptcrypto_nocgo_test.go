//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package mptcrypto_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestDecryptAmountWithoutCgo(t *testing.T) {
	_, err := mptcrypto.DecryptAmount(mptcrypto.Ciphertext{}, mptcrypto.PrivateKey{}, 2, 1)
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}
