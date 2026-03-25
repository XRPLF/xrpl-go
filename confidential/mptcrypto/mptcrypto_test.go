//go:build cgo

package mptcrypto_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
	"github.com/stretchr/testify/require"
)

func TestGenerateKeypair(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	require.NoError(t, err)
	require.NotEqual(t, [mptcrypto.PrivKeySize]byte{}, priv, "privkey is all zeros")
	// compressed secp256k1 pubkey starts with 0x02 or 0x03
	require.Contains(t, []byte{0x02, 0x03}, pub[0], "unexpected pubkey prefix: 0x%02x", pub[0])
}
