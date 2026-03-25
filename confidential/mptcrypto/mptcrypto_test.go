//go:build cgo

package mptcrypto_test

import (
	"testing"

	"github.com/Peersyst/xrpl-go/confidential/mptcrypto"
)

func TestGenerateKeypair(t *testing.T) {
	priv, pub, err := mptcrypto.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	if priv == [mptcrypto.PrivKeySize]byte{} {
		t.Fatal("privkey is all zeros")
	}
	// compressed secp256k1 pubkey starts with 0x02 or 0x03
	if pub[0] != 0x02 && pub[0] != 0x03 {
		t.Fatalf("unexpected pubkey prefix: 0x%02x", pub[0])
	}
}
