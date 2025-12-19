package main

import (
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	receiverWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

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

	fmt.Println("⏳ Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Connected to server")
	fmt.Println()

	balance, err := client.GetXrpBalance(w.GetAddress())

	if err != nil || balance == "0" {
		fmt.Println("⏳ Funding wallet...")
		err = client.FundWallet(&w)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("💸 Wallet funded")
	}

	balance, _ = client.GetXrpBalance(w.GetAddress())

	fmt.Printf("💸 Balance: %s\n", balance)
	fmt.Println(w.GetAddress())

	amount, err := currency.XrpToDrops("1")
	if err != nil {
		fmt.Println(err)
		return
	}

	amountUint, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("⏳ Sending payment...")
	payment := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(w.GetAddress()),
		},
		Destination: types.Address(receiverWallet.GetAddress()),
		Amount:      types.XRPCurrencyAmount(amountUint),
		DeliverMax:  types.XRPCurrencyAmount(amountUint),
	}

	flatTx := payment.Flatten()

	err = client.Autofill(&flatTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	txBlob, _, err := w.Sign(flatTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	response, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Payment submitted")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Printf("🌐 Validated: %t\n", response.Validated)
}
