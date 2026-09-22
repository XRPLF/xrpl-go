package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	fmt.Println("⏳ Connecting to testnet...")
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := rpc.NewClient(cfg)

	// Use fresh accounts and create the liquidity needed by the payment path.
	var source, issuer, receiver wallet.Wallet
	for _, w := range []*wallet.Wallet{&source, &issuer, &receiver} {
		*w, err = wallet.New(crypto.ED25519())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("⏳ Funding wallet:", w.ClassicAddress)
		if err := client.FundWallet(w); err != nil {
			fmt.Println(err)
			return
		}
	}

	accountSet := transaction.AccountSet{
		BaseTx: transaction.BaseTx{Account: issuer.ClassicAddress},
	}
	accountSet.SetAsfDefaultRipple()
	if err := submitAndWait(client, accountSet.Flatten(), issuer); err != nil {
		fmt.Println("❌", err)
		return
	}

	trust := transaction.TrustSet{
		BaseTx: transaction.BaseTx{Account: receiver.ClassicAddress},
		LimitAmount: types.IssuedCurrencyAmount{
			Issuer: issuer.ClassicAddress, Currency: "USD", Value: "100",
		},
	}
	if err := submitAndWait(client, trust.Flatten(), receiver); err != nil {
		fmt.Println("❌", err)
		return
	}

	// The issuer sells 10 USD for 10 XRP. The source can pay XRP while the
	// receiver gets USD, without depending on an existing public order book.
	offer := transaction.OfferCreate{
		BaseTx:    transaction.BaseTx{Account: issuer.ClassicAddress},
		TakerPays: types.XRPCurrencyAmount(10_000_000),
		TakerGets: types.IssuedCurrencyAmount{
			Issuer: issuer.ClassicAddress, Currency: "USD", Value: "10",
		},
	}
	if err := submitAndWait(client, offer.Flatten(), issuer); err != nil {
		fmt.Println("❌", err)
		return
	}

	amount := types.IssuedCurrencyAmount{
		Issuer: issuer.ClassicAddress, Currency: "USD", Value: "1",
	}
	fmt.Println("⏳ Getting paths...")
	res, err := client.GetRipplePathFind(&path.RipplePathFindRequest{
		SourceAccount: source.ClassicAddress,
		SourceCurrencies: []pathtypes.RipplePathFindCurrency{
			{Currency: "XRP"},
		},
		DestinationAccount: receiver.ClassicAddress,
		DestinationAmount:  amount,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("🌐 Computed paths: %d\n", len(res.Alternatives))
	if len(res.Alternatives) == 0 {
		fmt.Println("❌ No alternatives found")
		return
	}

	fmt.Println("⏳ Submitting payment through path:", res.Alternatives[0].PathsComputed)
	payment := transaction.Payment{
		BaseTx:      transaction.BaseTx{Account: source.ClassicAddress},
		Destination: receiver.ClassicAddress,
		Amount:      amount,
		// Spend at most 1 XRP to deliver 1 USD at the offer's exchange rate.
		SendMax: types.XRPCurrencyAmount(1_000_000),
		Paths:   res.Alternatives[0].PathsComputed,
	}
	if err := submitAndWait(client, payment.Flatten(), source); err != nil {
		fmt.Println("❌", err)
		return
	}
}

func submitAndWait(client *rpc.Client, tx transaction.FlatTransaction, signer wallet.Wallet) error {
	response, err := client.SubmitTxAndWait(tx, &rpctypes.SubmitOptions{
		Autofill: true,
		Wallet:   &signer,
	})
	if err != nil {
		return fmt.Errorf("submit %s: %w", tx.TxType(), err)
	}
	if !response.Validated || response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return fmt.Errorf("%s: validated=%t, result=%s", tx.TxType(), response.Validated, response.Meta.TransactionResult)
	}
	fmt.Printf("✅ %s submitted\n", tx.TxType())
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	return nil
}
