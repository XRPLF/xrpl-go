package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

const (
	currencyCode = "FOO"
)

func main() {
	//
	// Configure client
	//
	fmt.Println("⏳ Setting up client...")
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("❌ Error disconnecting:", err)
		}
	}()

	fmt.Println("✅ Client configured!")
	fmt.Println()

	fmt.Println("⏳ Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connection: ", client.IsConnected())

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

func configureColdWallet(client *websocket.Client, coldWallet wallet.Wallet) error {
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

	response, err := client.SubmitTxAndWait(coldWalletAccountSet.Flatten(), &wstypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if !response.Validated {
		return fmt.Errorf("cold wallet unfreezing failed")
	}

	fmt.Println("✅ Cold address settings configured!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func createTrustLine(client *websocket.Client, coldWallet, hotWallet wallet.Wallet) error {
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

	response, err := client.SubmitTxAndWait(hotColdTrustSet.Flatten(), &wstypes.SubmitOptions{Autofill: true, Wallet: &hotWallet})
	if err != nil {
		return err
	}

	if !response.Validated {
		return fmt.Errorf("trust line from hot to cold address creation failed")
	}

	fmt.Println("✅ Trust line from hot to cold address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func issueTokens(client *websocket.Client, coldWallet, hotWallet wallet.Wallet) error {
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

	response, err := client.SubmitTxAndWait(coldToHotPayment.Flatten(), &wstypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if !response.Validated {
		return fmt.Errorf("tokens not sent from cold wallet to hot wallet")
	}

	fmt.Println("✅ Tokens sent from cold wallet to hot wallet!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}

func clawBackTokens(client *websocket.Client, coldWallet, hotWallet wallet.Wallet) error {
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

	response, err := client.SubmitTxAndWait(coldWalletClawback.Flatten(), &wstypes.SubmitOptions{Autofill: true, Wallet: &coldWallet})
	if err != nil {
		return err
	}

	if !response.Validated {
		return fmt.Errorf("tokens not clawed back from customer one")
	}

	fmt.Println("✅ Tokens clawed back from customer one!")
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}
