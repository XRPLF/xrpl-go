package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.devnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

	// Query vault info by VaultID
	vaultID := "9E48171960CD9F62C3A7B6559315A510AE544C3F51E02947B5D4DAC8AA66C3BA"
	res, err := client.GetVaultInfo(&vault.InfoRequest{
		VaultID: vaultID,
	})
	if err != nil {
		fmt.Printf("Error querying vault info: %s\n", err)
		return
	}

	fmt.Println("Vault Info:")
	fmt.Printf("  Owner: %s\n", res.Vault.Owner)
	fmt.Printf("  Account: %s\n", res.Vault.Account)
	fmt.Printf("  AssetsTotal: %s\n", res.Vault.AssetsTotal)
	fmt.Printf("  AssetsAvailable: %s\n", res.Vault.AssetsAvailable)
	fmt.Printf("  Validated: %t\n", res.Validated)
	fmt.Println()
}
