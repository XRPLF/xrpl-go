package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
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

	fmt.Println("💸 Wallet funded")
	fmt.Println()

	info, err := client.GetAccountInfo(&account.InfoRequest{
		Account: w.GetAddress(),
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("🌐 Current wallet sequence:", info.AccountData.Sequence)
	fmt.Println()

	fmt.Println("⏳ Submitting TicketCreate transaction...")
	tc := &transaction.TicketCreate{
		BaseTx: transaction.BaseTx{
			Account:  w.GetAddress(),
			Sequence: info.AccountData.Sequence,
		},
		TicketCount: 10,
	}

	flatTc := tc.Flatten()

	if err := client.Autofill(&flatTc); err != nil {
		fmt.Println(err)
		return
	}

	blob, _, err := w.Sign(flatTc)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.SubmitTxBlobAndWait(blob, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ TicketCreate transaction submitted")
	fmt.Printf("🌐 Hash: %s\n", res.Hash)
	fmt.Printf("🌐 Validated: %t\n", res.Validated)
	fmt.Println()

	objects, err := client.GetAccountObjects(&account.ObjectsRequest{
		Account: w.GetAddress(),
		Type:    "ticket",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(objects.AccountObjects) == 0 {
		fmt.Println("❌ No tickets found")
		return
	}
	seq, ok := typecheck.ToUint32(objects.AccountObjects[0]["TicketSequence"])
	if !ok {
		fmt.Println("❌ Invalid ticket sequence")
		return
	}
	fmt.Println("🌐 Ticket sequence:", seq)

	fmt.Println("⏳ Submitting AccountSet transaction...")
	as := &transaction.AccountSet{
		BaseTx: transaction.BaseTx{
			Account:        w.GetAddress(),
			Sequence:       0,
			TicketSequence: seq,
		},
	}

	flatAs := as.Flatten()

	if err := client.Autofill(&flatAs); err != nil {
		fmt.Println(err)
		return
	}

	flatAs["Sequence"] = uint32(0)

	blob, _, err = w.Sign(flatAs)
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err = client.SubmitTxBlobAndWait(blob, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ AccountSet transaction submitted")
	fmt.Printf("🌐 Hash: %s\n", res.Hash)
	fmt.Printf("🌐 Validated: %t\n", res.Validated)
}
