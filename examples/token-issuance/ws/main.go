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
			WithHost("wss://s.altnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)

	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	fmt.Println("✅ Client configured!")
	fmt.Println()

	fmt.Println("Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connection: ", client.IsConnected())
	fmt.Println()

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

	customerOneWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating token wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&customerOneWallet)
	if err != nil {
		fmt.Printf("❌ Error funding customer one wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Customer one wallet funded!")
	fmt.Println()

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println("💳 Cold wallet:", coldWallet.ClassicAddress)
	fmt.Println("💳 Hot wallet:", hotWallet.ClassicAddress)
	fmt.Println("💳 Customer one wallet:", customerOneWallet.ClassicAddress)
	fmt.Println()

	if err := configureColdWallet(client, coldWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := configureHotWallet(client, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := createHotTrustLine(client, coldWallet, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := createCustomerTrustLine(client, coldWallet, customerOneWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := issueToHotWallet(client, coldWallet, hotWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := issueToCustomer(client, coldWallet, customerOneWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := freezeColdWallet(client, coldWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := tryFrozenPayment(client, coldWallet, hotWallet, customerOneWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := unfreezeColdWallet(client, coldWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	if err := sendAfterUnfreeze(client, coldWallet, hotWallet, customerOneWallet); err != nil {
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

	coldWalletAccountSet.SetAsfDefaultRipple()
	coldWalletAccountSet.SetDisallowXRP()

	coldWalletAccountSet.SetRequireDestTag()

	return submitAndWait(client, coldWalletAccountSet.Flatten(), coldWallet, transactions.TesSUCCESS)
}

func configureHotWallet(client *websocket.Client, hotWallet wallet.Wallet) error {
	fmt.Println("⏳ Configuring hot address settings...")
	hotWalletAccountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Domain: types.Domain("6578616D706C652E636F6D"), // example.com
	}

	hotWalletAccountSet.SetAsfRequireAuth()
	hotWalletAccountSet.SetDisallowXRP()
	hotWalletAccountSet.SetRequireDestTag()

	return submitAndWait(client, hotWalletAccountSet.Flatten(), hotWallet, transactions.TesSUCCESS)
}

func createHotTrustLine(client *websocket.Client, coldWallet, hotWallet wallet.Wallet) error {
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

	return submitAndWait(client, hotColdTrustSet.Flatten(), hotWallet, transactions.TesSUCCESS)
}

func createCustomerTrustLine(client *websocket.Client, coldWallet, customerOneWallet wallet.Wallet) error {
	fmt.Println("⏳ Creating trust line from customer one to cold address...")
	customerOneColdTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerOneWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	return submitAndWait(client, customerOneColdTrustSet.Flatten(), customerOneWallet, transactions.TesSUCCESS)
}

func issueToHotWallet(client *websocket.Client, coldWallet, hotWallet wallet.Wallet) error {
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

	return submitAndWait(client, coldToHotPayment.Flatten(), coldWallet, transactions.TesSUCCESS)
}

func issueToCustomer(client *websocket.Client, coldWallet, customerOneWallet wallet.Wallet) error {
	fmt.Println("⏳ Sending tokens from cold wallet to customer one...")
	coldToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	return submitAndWait(client, coldToCustomerOnePayment.Flatten(), coldWallet, transactions.TesSUCCESS)
}

func freezeColdWallet(client *websocket.Client, coldWallet wallet.Wallet) error {
	fmt.Println("⏳ Freezing cold wallet...")
	freezeColdWallet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
	}

	freezeColdWallet.SetAsfGlobalFreeze()

	return submitAndWait(client, freezeColdWallet.Flatten(), coldWallet, transactions.TesSUCCESS)
}

func tryFrozenPayment(client *websocket.Client, coldWallet, hotWallet, customerOneWallet wallet.Wallet) error {
	fmt.Println("⏳ Trying to send tokens from hot wallet to customer one...")
	hotToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	// Global freeze blocks holder-to-holder payments with tecPATH_DRY.
	return submitAndWait(client, hotToCustomerOnePayment.Flatten(), hotWallet, transactions.TecPATH_DRY)
}

func unfreezeColdWallet(client *websocket.Client, coldWallet wallet.Wallet) error {
	fmt.Println("⏳ Unfreezing cold wallet...")
	unfreezeColdWallet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
	}

	unfreezeColdWallet.ClearAsfGlobalFreeze()

	return submitAndWait(client, unfreezeColdWallet.Flatten(), coldWallet, transactions.TesSUCCESS)
}

func sendAfterUnfreeze(client *websocket.Client, coldWallet, hotWallet, customerOneWallet wallet.Wallet) error {
	fmt.Println("⏳ Trying to send tokens from hot wallet to customer one...")
	hotToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	return submitAndWait(client, hotToCustomerOnePayment.Flatten(), hotWallet, transactions.TesSUCCESS)
}

// submitAndWait requires the expected result in a validated ledger.
func submitAndWait(client *websocket.Client, tx transactions.FlatTransaction, signer wallet.Wallet, expected transactions.TxResult) error {
	fmt.Printf("⏳ Submitting %s transaction...\n", tx["TransactionType"])
	response, err := client.SubmitTxAndWait(tx, &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &signer,
	})
	if err != nil {
		return fmt.Errorf("submit %s: %w", tx["TransactionType"], err)
	}
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("%s: validated=%t, result=%s, expected=%s", tx["TransactionType"], response.Validated, response.Meta.TransactionResult, expected)
	}
	fmt.Printf("✅ %s validated with expected result %s\n", tx["TransactionType"], expected)
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}
