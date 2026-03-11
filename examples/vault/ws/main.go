package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
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

	fmt.Println("Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Setting up wallet...")
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("Error creating wallet: %s\n", err)
		return
	}
	if err = client.FundWallet(&w); err != nil {
		fmt.Printf("Error funding wallet: %s\n", err)
		return
	}
	fmt.Println("Wallet funded!")
	fmt.Println("Wallet:", w.ClassicAddress)
	fmt.Println()

	submitOpts := &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &w,
	}

	//
	// Create an XRP vault
	//
	fmt.Println("Creating XRP vault...")
	vaultCreate := &transactions.VaultCreate{
		BaseTx: transactions.BaseTx{
			Account: w.GetAddress(),
		},
		Asset: ledger.Asset{
			Currency: "XRP",
		},
	}

	response, err := client.SubmitTxAndWait(vaultCreate.Flatten(), submitOpts)
	if err != nil {
		fmt.Printf("Error submitting VaultCreate: %s\n", err)
		return
	}

	if !response.Validated {
		fmt.Printf("VaultCreate failed! Response: %+v\n", response)
		return
	}

	fmt.Println("Vault created!")
	fmt.Printf("Hash: %s\n", response.Hash.String())
	fmt.Println()

	// Extract VaultID from metadata
	meta := response.Meta.AsTxObjMeta()
	var vaultID string
	for _, node := range meta.AffectedNodes {
		if node.CreatedNode != nil && node.CreatedNode.LedgerEntryType == ledger.VaultEntry {
			vaultID = node.CreatedNode.LedgerIndex
			break
		}
	}
	if vaultID == "" {
		fmt.Println("Vault ID not found in metadata")
		return
	}
	fmt.Printf("VaultID: %s\n", vaultID)
	fmt.Println()

	//
	// Deposit into the vault
	//
	fmt.Println("Depositing 1000000 drops into vault...")
	vaultDeposit := &transactions.VaultDeposit{
		BaseTx: transactions.BaseTx{
			Account: w.GetAddress(),
		},
		VaultID: types.Hash256(vaultID),
		Amount:  types.XRPCurrencyAmount(1000000),
	}

	response, err = client.SubmitTxAndWait(vaultDeposit.Flatten(), submitOpts)
	if err != nil {
		fmt.Printf("Error submitting VaultDeposit: %s\n", err)
		return
	}

	if !response.Validated {
		fmt.Printf("VaultDeposit failed! Response: %+v\n", response)
		return
	}

	fmt.Println("Deposit successful!")
	fmt.Printf("Hash: %s\n", response.Hash.String())
	fmt.Println()

	//
	// Update vault settings
	//
	fmt.Println("Updating vault settings...")
	data := types.Data("DEADBEEF")
	vaultSet := &transactions.VaultSet{
		BaseTx: transactions.BaseTx{
			Account: w.GetAddress(),
		},
		VaultID: types.Hash256(vaultID),
		Data:    &data,
	}

	response, err = client.SubmitTxAndWait(vaultSet.Flatten(), submitOpts)
	if err != nil {
		fmt.Printf("Error submitting VaultSet: %s\n", err)
		return
	}

	if !response.Validated {
		fmt.Printf("VaultSet failed! Response: %+v\n", response)
		return
	}

	fmt.Println("Vault settings updated!")
	fmt.Printf("Hash: %s\n", response.Hash.String())
	fmt.Println()

	//
	// Withdraw from the vault
	//
	fmt.Println("Withdrawing 1000000 drops from vault...")
	vaultWithdraw := &transactions.VaultWithdraw{
		BaseTx: transactions.BaseTx{
			Account: w.GetAddress(),
		},
		VaultID: types.Hash256(vaultID),
		Amount:  types.XRPCurrencyAmount(1000000),
	}

	response, err = client.SubmitTxAndWait(vaultWithdraw.Flatten(), submitOpts)
	if err != nil {
		fmt.Printf("Error submitting VaultWithdraw: %s\n", err)
		return
	}

	if !response.Validated {
		fmt.Printf("VaultWithdraw failed! Response: %+v\n", response)
		return
	}

	fmt.Println("Withdrawal successful!")
	fmt.Printf("Hash: %s\n", response.Hash.String())
	fmt.Println()

	//
	// Delete the vault
	//
	fmt.Println("Deleting vault...")
	vaultDelete := &transactions.VaultDelete{
		BaseTx: transactions.BaseTx{
			Account: w.GetAddress(),
		},
		VaultID: types.Hash256(vaultID),
	}

	response, err = client.SubmitTxAndWait(vaultDelete.Flatten(), submitOpts)
	if err != nil {
		fmt.Printf("Error submitting VaultDelete: %s\n", err)
		return
	}

	if !response.Validated {
		fmt.Printf("VaultDelete failed! Response: %+v\n", response)
		return
	}

	fmt.Println("Vault deleted!")
	fmt.Printf("Hash: %s\n", response.Hash.String())
	fmt.Println()
}
