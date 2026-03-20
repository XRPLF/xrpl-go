package faucet

import (
	"testing"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// Note: This test interacts with the actual Devnet faucet.
// The result and behavior may vary based on the faucet's current state and rate limits.
// Manual verification of the printed result is recommended.
func TestDevnetFaucetProvider_FundWallet(t *testing.T) {
	// Create a new DevnetFaucetProvider
	provider := NewDevnetFaucetProvider()

	// Test wallet address
	testWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		t.Errorf("Wallet creation error: %v", err)
	}

	// Call FundWallet
	err = provider.FundWallet(testWallet.ClassicAddress)
	// Check for errors
	if err != nil {
		t.Errorf("FundWallet returned an error: %v", err)
	}
}
