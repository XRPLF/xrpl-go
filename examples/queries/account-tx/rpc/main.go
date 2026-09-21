package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithMaxFeeXRP("5"),
		rpc.WithFeeCushion(1.5),
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

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

	fmt.Println("✅ Number of transactions:", len(txs.Transactions))
	for _, tx := range txs.Transactions {
		fmt.Println(tx.Tx)
	}
}
