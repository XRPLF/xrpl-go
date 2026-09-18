package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerqueries "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

	fmt.Println("⏳ Fetching ledger...")
	ledger, err := client.GetLedger(&ledgerqueries.Request{
		LedgerIndex: common.LedgerIndex(5115183),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ Ledger retrieved")
	fmt.Println("🌐 Hash:", ledger.Ledger.LedgerHash)
	fmt.Println("🌐 Index:", ledger.Ledger.LedgerIndex)
}
