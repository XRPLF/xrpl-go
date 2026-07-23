//go:build !cgo || js || wasip1 || tinygo || gofuzz || !(linux || darwin) || !(amd64 || arm64)

package elgamal_test

import (
	"strings"
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/elgamal"
	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestDecryptWithoutCgo(t *testing.T) {
	ciphertext := strings.Repeat("00", mptcrypto.CiphertextSize)
	privateKey := strings.Repeat("00", mptcrypto.PrivKeySize)

	_, err := elgamal.Decrypt(ciphertext, privateKey, elgamal.AmountRange{Low: 0, High: 0})
	require.ErrorIs(t, err, elgamal.ErrDecryptFailed)
	require.ErrorIs(t, err, mptcrypto.ErrCgoRequired)
}
