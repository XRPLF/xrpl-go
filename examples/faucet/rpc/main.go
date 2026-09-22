package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	fmt.Println("⏳ Funding wallet on testnet...")

	// For devnet, use https://s.devnet.rippletest.net:51234/ and
	// faucet.NewDevnetFaucetProvider() instead.
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

	wallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err := client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		balance = "0"
	}

	fmt.Println("💳 Balance", balance)
	fmt.Println()

	fmt.Println("⏳ Funding wallet...")
	err = client.FundWallet(&wallet)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance, err = client.GetXrpBalance(wallet.ClassicAddress)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("💳 Balance", balance)
	fmt.Println()
}
