package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
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

	txs, err := client.GetAccountTransactions(&account.TransactionsRequest{
		Account:     "rMCcNuTcajgw7YTgBy1sys3b89QqjUrMpH",
		LedgerIndex: common.LedgerIndex(4976692),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Number of transactions:", len(txs.Transactions))
	fmt.Println(txs.Transactions[0].Tx)
}
