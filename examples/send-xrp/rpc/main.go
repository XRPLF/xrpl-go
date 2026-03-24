package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"

	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithMaxFeeXRP(5.0),
		rpc.WithFeeCushion(1.5),
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

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

	xrpAmount, err := currency.XrpToDrops("1")
	if err != nil {
		fmt.Println(err)
		return
	}

	xrpAmountInt, err := strconv.ParseInt(xrpAmount, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	if xrpAmountInt < 0 {
		fmt.Printf("❌ XRP amount %d cannot be negative\n", xrpAmountInt)
		return
	}

	fmt.Println("⏳ Sending 1 XRP to rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe...")
	p := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: w.GetAddress(),
		},
		Destination: "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		Amount:      types.XRPCurrencyAmount(xrpAmountInt),
		DeliverMax:  types.XRPCurrencyAmount(xrpAmountInt),
	}

	flattenedTx := p.Flatten()

	if err := client.Autofill(&flattenedTx); err != nil {
		fmt.Println(err)
		return
	}

	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Now()
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	metadata := res.Meta.AsPaymentMetadata()

	fmt.Println("✅ Payment submitted")
	fmt.Printf("🌐 Hash: %s\n", res.Hash)
	fmt.Printf("🌐 Validated: %t\n", res.Validated)
	fmt.Printf("🌐 DeliveredAmount (drops): %s\n", metadata.DeliveredAmount)
	fmt.Printf("⏱️  Took: %s\n", time.Since(start))
	fmt.Println()
	fmt.Println("⏳ Using SubmitTxAndWait with wallet")
	fmt.Println()

	flattenedTx2 := p.Flatten()
	start = time.Now()
	resp, err := client.SubmitTxAndWait(flattenedTx2, &rpctypes.SubmitOptions{
		Autofill: true,
		Wallet:   &w,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	metadata = resp.Meta.AsPaymentMetadata()

	fmt.Println("✅ Payment submitted via SubmitTxAndWait")
	fmt.Printf("🌐 Hash: %s\n", resp.Hash)
	fmt.Printf("🌐 Validated: %t\n", resp.Validated)
	fmt.Printf("🌐 DeliveredAmount (drops): %s\n", metadata.DeliveredAmount)
	fmt.Printf("⏱️  Took: %s\n", time.Since(start))
}
