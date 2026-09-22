package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	fmt.Println("⏳ Connecting to testnet...")
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	if !client.IsConnected() {
		fmt.Println("❌ Failed to connect to testnet")
		return
	}

	fmt.Println("✅ Connected to testnet")
	fmt.Println()

	// Funding a fresh account gives this query a transaction to retrieve.
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("⏳ Funding wallet...")
	if err := client.FundWallet(&w); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("⏳ Fetching account transactions...")
	txs, err := client.GetAccountTransactions(&account.TransactionsRequest{
		Account:        w.ClassicAddress,
		LedgerIndexMin: -1,
		LedgerIndexMax: -1,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Number of transactions:", len(txs.Transactions))
	for _, tx := range txs.Transactions {
		fmt.Println(tx.Tx)
	}
}
