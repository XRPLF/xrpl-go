package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

const (
	currencyCode = "USDA"
)

func main() {
	fmt.Println("⏳ Setting up client...")

	client := getClient()
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("❌ Error disconnecting:", err)
		}
	}()

	fmt.Println("⏳ Connecting to server...")
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("✅ Client configured!")
	fmt.Println()

	fmt.Printf("Connection: %t", client.IsConnected())
	fmt.Println()

	// Configure wallets

	// Issuer
	fmt.Println("⏳ Setting up issuer wallet...")
	issuer, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating issuer wallet: %s\n", err)
		return
	}

	err = client.FundWallet(&issuer)
	if err != nil {
		fmt.Printf("❌ Error funding issuer wallet: %s\n", err)
		return
	}
	fmt.Printf("✅ Issuer wallet funded: %s\n", issuer.ClassicAddress)

	// -----------------------------------------------------

	// Holder 1
	fmt.Println("⏳ Setting up holder 1 wallet...")
	holderWallet1, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating holder wallet 1: %s\n", err)
		return
	}

	err = client.FundWallet(&holderWallet1)
	if err != nil {
		fmt.Printf("❌ Error funding holder wallet 1: %s\n", err)
		return
	}
	fmt.Printf("✅ Holder wallet 1 funded: %s\n", holderWallet1.ClassicAddress)

	// -----------------------------------------------------

	// Holder 2
	fmt.Println("⏳ Setting up holder 2 wallet...")
	holderWallet2, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating holder wallet 2: %s\n", err)
		return
	}

	err = client.FundWallet(&holderWallet2)
	if err != nil {
		fmt.Printf("❌ Error funding holder wallet 2: %s\n", err)
		return
	}
	fmt.Printf("✅ Holder wallet 2 funded: %s\n", holderWallet2.ClassicAddress)
	fmt.Println()

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println()

	// -----------------------------------------------------

	if err := configureTrustLines(client, issuer, holderWallet1, holderWallet2); err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := mintAndTransferTokens(client, issuer, holderWallet1, holderWallet2); err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := freezeAndTryTransfers(client, issuer, holderWallet1, holderWallet2); err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := unfreezeAndTransferTokens(client, issuer, holderWallet1, holderWallet2); err != nil {
		fmt.Println("❌", err)
		return
	}
}

// getClient returns a new websocket client
func getClient() *websocket.Client {
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)

	return client
}

func configureTrustLines(client *websocket.Client, issuer, holderWallet1, holderWallet2 wallet.Wallet) error {
	// Configuring Issuing account
	fmt.Println("⏳ Configuring issuer address settings...")
	accountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		Domain: types.Domain("697373756572"), // issuer
	}

	accountSet.SetAsfDefaultRipple()
	if err := submitAndWait(client, accountSet.Flatten(), issuer, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Trustline from the holder 1 to the issuer
	fmt.Println("⏳ Setting up trustline from holder 1 to the issuer...")
	trustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet1.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "1000000000",
		},
	}
	trustSet.SetSetNoRippleFlag()
	if err := submitAndWait(client, trustSet.Flatten(), holderWallet1, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Trustline from the holder 2 to the issuer
	fmt.Println("⏳ Setting up trustline from holder 2 to the issuer...")
	trustSet = &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet2.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "1000000000",
		},
	}
	trustSet.SetSetNoRippleFlag()
	return submitAndWait(client, trustSet.Flatten(), holderWallet2, transactions.TesSUCCESS)
}

func mintAndTransferTokens(client *websocket.Client, issuer, holderWallet1, holderWallet2 wallet.Wallet) error {
	// Minting to Holder 1
	fmt.Println("⏳ Minting to Holder 1...")
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		Destination: types.Address(holderWallet1.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "50000",
		},
	}
	if err := submitAndWait(client, payment.Flatten(), issuer, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Minting to Holder 2
	fmt.Println("⏳ Minting to Holder 2...")
	payment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		Destination: types.Address(holderWallet2.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "40000",
		},
	}
	if err := submitAndWait(client, payment.Flatten(), issuer, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Sending payment from Holder 1 to Holder 2
	fmt.Println("⏳ Sending payment from Holder 1 to Holder 2...")
	payment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet1.ClassicAddress),
		},
		Destination: types.Address(holderWallet2.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "20",
		},
	}
	return submitAndWait(client, payment.Flatten(), holderWallet1, transactions.TesSUCCESS)
}

func freezeAndTryTransfers(client *websocket.Client, issuer, holderWallet1, holderWallet2 wallet.Wallet) error {
	// Freezing and Deep Freezing holder1
	fmt.Println("⏳ Freezing and Deep Freezing holder 1 trustline...")
	trustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(holderWallet1.ClassicAddress),
			Value:    "0",
		},
	}
	trustSet.SetSetFreezeFlag()
	trustSet.SetSetDeepFreezeFlag()

	if err := submitAndWait(client, trustSet.Flatten(), issuer, transactions.TesSUCCESS); err != nil {
		return err
	}

	// ------------------- SHOULD FAIL ⬇️ ------------------

	// Sending payment from Holder 1 to Holder 2 (which should fail), Holder 1 can't decrease its balance
	fmt.Println("⏳ Sending payment from Holder 1 to Holder 2 (which should fail). Holder 1 can't decrease its balance...")
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet1.ClassicAddress),
		},
		Destination: types.Address(holderWallet2.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "10",
		},
	}
	if err := submitAndWait(client, payment.Flatten(), holderWallet1, transactions.TecPATH_DRY); err != nil {
		return err
	}

	// ------------------- SHOULD FAIL ⬇️ ------------------

	// Sending payment from Holder 2 to Holder 1 (which should fail), Holder 1 can't increase its balance
	fmt.Println("⏳ Sending payment from Holder 2 to Holder 1 (which should fail). Holder 1 can't increase its balance...")
	payment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet2.ClassicAddress),
		},
		Destination: types.Address(holderWallet1.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "10",
		},
	}
	if err := submitAndWait(client, payment.Flatten(), holderWallet2, transactions.TecPATH_DRY); err != nil {
		return err
	}

	// ------------------- SHOULD FAIL ⬇️ ------------------

	// Creating OfferCreate transaction (which should fail), Holder 1 can't create an offer
	fmt.Println("⏳ Creating OfferCreate transaction (which should fail). Holder 1 can't create an offer...")
	offerCreate := &transactions.OfferCreate{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet1.ClassicAddress),
		},
		TakerPays: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "10",
		},
		TakerGets: types.XRPCurrencyAmount(10),
	}
	return submitAndWait(client, offerCreate.Flatten(), holderWallet1, transactions.TecFROZEN)
}

func unfreezeAndTransferTokens(client *websocket.Client, issuer, holderWallet1, holderWallet2 wallet.Wallet) error {
	// Unfreezing and Deep Unfreezing holder 1
	fmt.Println("⏳ Unfreezing and Deep Unfreezing holder 1 trustline...")
	trustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(holderWallet1.ClassicAddress),
			Value:    "0",
		},
	}
	trustSet.SetClearFreezeFlag()
	trustSet.SetClearDeepFreezeFlag()
	if err := submitAndWait(client, trustSet.Flatten(), issuer, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Sending payment from Holder 1 to Holder 2 (which should succeed), Holder 1 can decrease its balance
	fmt.Println("⏳ Sending payment from Holder 1 to Holder 2 (which should succeed). Holder 1 can decrease its balance...")
	payment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet1.ClassicAddress),
		},
		Destination: types.Address(holderWallet2.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "10",
		},
	}
	if err := submitAndWait(client, payment.Flatten(), holderWallet1, transactions.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Sending payment from Holder 2 to Holder 1 (which should succeed), Holder 1 can increase its balance
	fmt.Println("⏳ Sending payment from Holder 2 to Holder 1 (which should succeed). Holder 1 can increase its balance...")
	payment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(holderWallet2.ClassicAddress),
		},
		Destination: types.Address(holderWallet1.ClassicAddress),
		Amount: types.IssuedCurrencyAmount{
			Currency: currency.ConvertStringToHex(currencyCode),
			Issuer:   types.Address(issuer.ClassicAddress),
			Value:    "10",
		},
	}
	return submitAndWait(client, payment.Flatten(), holderWallet2, transactions.TesSUCCESS)
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
