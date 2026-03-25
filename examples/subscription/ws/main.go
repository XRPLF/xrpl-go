package main

import (
	"fmt"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
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

	_, err := client.Subscribe(&subscribe.Request{
		Streams: []string{"ledger", "transactions"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	client.OnError(func(err error) {
		fmt.Println("Error: ", err)
	})

	client.OnLedgerClosed(func(ledger *streamtypes.LedgerStream) {
		fmt.Println("Ledger closed: ", ledger.LedgerIndex)
	})

	client.OnTransactions(func(transactions *streamtypes.TransactionStream) {
		fmt.Println("Transactions received: ", transactions.Hash)
	})

	for range 3 {
		res, err := client.GetAccountTransactions(&account.TransactionsRequest{
			Account: "rnPWcg6oixrHX9RSPMYJmaXRb7csfECE5T",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Transactions: ", len(res.Transactions))

		time.Sleep(5 * time.Second)
	}

	_, err = client.Unsubscribe(&subscribe.UnsubscribeRequest{
		Streams: []string{"ledger", "transactions"},
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Unsubscribed from streams: ledger, transactions")

	time.Sleep(10 * time.Second)
}
