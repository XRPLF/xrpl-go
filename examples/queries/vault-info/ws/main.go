package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	fmt.Println("⏳ Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Connected to server")

	// Query vault info by VaultID
	vaultID := "9E48171960CD9F62C3A7B6559315A510AE544C3F51E02947B5D4DAC8AA66C3BA"
	fmt.Println("⏳ Fetching vault info...")
	res, err := client.GetVaultInfo(&vault.InfoRequest{
		VaultID: vaultID,
	})
	if err != nil {
		fmt.Printf("❌ Error querying vault info: %s\n", err)
		return
	}

	fmt.Println("✅ Vault info retrieved:")
	fmt.Printf("  Owner: %s\n", res.Vault.Owner)
	fmt.Printf("  Account: %s\n", res.Vault.Account)
	fmt.Printf("  AssetsTotal: %s\n", res.Vault.AssetsTotal)
	fmt.Printf("  AssetsAvailable: %s\n", res.Vault.AssetsAvailable)
	fmt.Printf("🌐 Validated: %t\n", res.Validated)
	fmt.Println()
}
