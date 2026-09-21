package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/vault"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

func main() {
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
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

	// Create an empty XRP vault so the query does not depend on a saved ID.
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating wallet:", err)
		return
	}
	fmt.Println("⏳ Funding wallet...")
	if err := client.FundWallet(&w); err != nil {
		fmt.Println("❌ Error funding wallet:", err)
		return
	}

	fmt.Println("⏳ Creating XRP vault...")
	create := transaction.VaultCreate{
		BaseTx: transaction.BaseTx{Account: w.ClassicAddress},
		Asset:  ledger.Asset{Currency: "XRP"},
	}
	response, err := client.SubmitTxAndWait(create.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &w,
	})
	if err != nil {
		fmt.Println("❌ Error creating vault:", err)
		return
	}
	if !response.Validated || response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		fmt.Println("❌ Vault creation failed:", response.Meta.TransactionResult)
		return
	}
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())

	var vaultID string
	for _, node := range response.Meta.AsTxObjMeta().AffectedNodes {
		if node.CreatedNode != nil && node.CreatedNode.LedgerEntryType == "Vault" {
			vaultID = node.CreatedNode.LedgerIndex
			break
		}
	}
	if vaultID == "" {
		fmt.Println("❌ Vault ID not found in metadata")
		return
	}

	// Query the vault created by this run.
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
