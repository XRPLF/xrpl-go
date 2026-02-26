package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
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
		panic("Failed to connect to testnet")
	}

	fmt.Println("✅ Connected to testnet")
	fmt.Println()

	fmt.Println("⏳ Funding wallets...")
	// Create and fund the cold wallet (issuer)
	issuerWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating cold wallet:", err)
		return
	}
	if err := client.FundWallet(&issuerWallet); err != nil {
		fmt.Println("❌ Error funding cold wallet:", err)
		return
	}
	fmt.Println("💸 Cold wallet funded!")

	// Create and fund the hot wallet (holder)
	hotWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating hot wallet:", err)
		return
	}
	if err := client.FundWallet(&hotWallet); err != nil {
		fmt.Println("❌ Error funding hot wallet:", err)
		return
	}
	fmt.Println("💸 Hot wallet funded!")

	// Create and fund a customer wallet
	customerWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating customer wallet:", err)
		return
	}
	if err := client.FundWallet(&customerWallet); err != nil {
		fmt.Println("❌ Error funding customer wallet:", err)
		return
	}
	fmt.Println("💸 Customer wallet funded!")
	fmt.Println()

	amount := types.XRPCurrencyAmount(10000)

	fmt.Println("⏳ Issuing MPToken...")
	// Create the MPTokenIssuanceCreate transaction.
	issuanceTx := &transactions.MPTokenIssuanceCreate{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
		AssetScale:      types.AssetScale(2),
		TransferFee:     types.TransferFee(314),
		MaximumAmount:   &amount,
		MPTokenMetadata: types.MPTokenMetadata("464F4F"), // "FOO" in hex
	}
	// Since TransferFee is provided, enable the TfMPTCanTransfer flag.
	issuanceTx.SetMPTCanTransferFlag()

	// Flatten, autofill, sign, and submit the transaction.
	flattenedTx := issuanceTx.Flatten()
	if err := client.Autofill(&flattenedTx); err != nil {
		fmt.Println("❌ Error autofilling issuance transaction:", err)
		return
	}

	txBlob, _, err := issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Println("❌ Error signing issuance transaction:", err)
		return
	}

	response, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println("❌ Error submitting issuance transaction:", err)
		return
	}

	if !response.Validated {
		fmt.Println("❌ MPToken issuance transaction failed to validate!")
		return
	}

	fmt.Println("✅ MPToken Issuance Create transaction validated!")
	fmt.Println("🌐 Transaction Hash:", response.Hash.String())
}
