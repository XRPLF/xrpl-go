package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	txrequests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	rpctypes "github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

const (
	currencyCode = "FOO"
)

func main() {
	//
	// Configure client
	//
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)

	//
	// Configure wallets
	//
	fmt.Println("⏳ Setting up wallets...")
	coldWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating cold wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&coldWallet)
	if err != nil {
		fmt.Printf("❌ Error funding cold wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Cold wallet funded!")

	hotWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating hot wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&hotWallet)
	if err != nil {
		fmt.Printf("❌ Error funding hot wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Hot wallet funded!")
	fmt.Println()

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println("💳 Cold wallet:", coldWallet.ClassicAddress)
	fmt.Println("💳 Hot wallet:", hotWallet.ClassicAddress)
	fmt.Println()

	if err := configureColdWallet(client, coldWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := createTrustLine(client, coldWallet, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := issueTokens(client, coldWallet, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := clawBackTokens(client, coldWallet, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}
}

func configureColdWallet(client *rpc.Client, coldWallet wallet.Wallet) error {
	fmt.Println("⏳ Configuring cold address settings...")
	coldWalletAccountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		TickSize:     types.TickSize(5),
		TransferRate: types.TransferRate(0),
		Domain:       types.Domain("6578616D706C652E636F6D"), // example.com
	}

	coldWalletAccountSet.SetAsfAllowTrustLineClawback()
	coldWalletAccountSet.SetDisallowXRP()

	coldWalletAccountSet.SetRequireDestTag()

	response, err := client.SubmitTxAndWait(coldWalletAccountSet.Flatten(), &rpctypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if err := checkResult(response, transactions.TesSUCCESS); err != nil {
		return err
	}

	fmt.Println("✅ Cold address settings configured!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func createTrustLine(client *rpc.Client, coldWallet, hotWallet wallet.Wallet) error {
	fmt.Println("⏳ Creating trust line from hot to cold address...")
	hotColdTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	response, err := client.SubmitTxAndWait(hotColdTrustSet.Flatten(), &rpctypes.SubmitOptions{Autofill: true, Wallet: &hotWallet})
	if err != nil {
		return err
	}

	if err := checkResult(response, transactions.TesSUCCESS); err != nil {
		return err
	}

	fmt.Println("✅ Trust line from hot to cold address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func issueTokens(client *rpc.Client, coldWallet, hotWallet wallet.Wallet) error {
	fmt.Println("⏳ Sending tokens from cold wallet to hot wallet...")
	coldToHotPayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "3800",
		},
		Destination:    types.Address(hotWallet.ClassicAddress),
		DestinationTag: types.DestinationTag(1),
	}

	response, err := client.SubmitTxAndWait(coldToHotPayment.Flatten(), &rpctypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if err := checkResult(response, transactions.TesSUCCESS); err != nil {
		return err
	}

	fmt.Println("✅ Tokens sent from cold wallet to hot wallet!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func clawBackTokens(client *rpc.Client, coldWallet, hotWallet wallet.Wallet) error {
	fmt.Println("⏳ Clawing back tokens from hot wallet...")

	coldWalletClawback := &transactions.Clawback{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(hotWallet.ClassicAddress),
			Value:    "50",
		},
	}

	response, err := client.SubmitTxAndWait(coldWalletClawback.Flatten(), &rpctypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if err := checkResult(response, transactions.TesSUCCESS); err != nil {
		return err
	}

	fmt.Println("✅ Tokens clawed back from customer one!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

//nolint:unparam // This example always expects success. Keep expected so readers can adapt this starting point to other results.
func checkResult(response *txrequests.TxResponse, expected transactions.TxResult) error {
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("transaction failed: validated=%t, result=%s, expected=%s", response.Validated, response.Meta.TransactionResult, expected)
	}
	return nil
}
