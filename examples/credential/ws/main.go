package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	rippleTime "github.com/Peersyst/xrpl-go/xrpl/time"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

func main() {
	fmt.Println("⏳ Setting up client...")

	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Println("❌ Error disconnecting:", err)
		}
	}()
	fmt.Println("Connecting to server...")
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
	fmt.Println("⏳ Setting up credential issuer wallet...")
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

	// Subject (destination)
	fmt.Println("⏳ Setting up Subject wallet...")
	subjectWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating subject wallet: %s\n", err)
		return
	}

	err = client.FundWallet(&subjectWallet)
	if err != nil {
		fmt.Printf("❌ Error funding subject wallet: %s\n", err)
		return
	}
	fmt.Printf("✅ Subject wallet funded: %s\n", subjectWallet.ClassicAddress)

	// -----------------------------------------------------

	// Creating the CredentialCreate transaction
	fmt.Println("⏳ Creating CredentialCreate transaction...")

	expiration, err := rippleTime.IsoTimeToRippleTime(time.Now().Add(time.Hour * 24).Format(time.RFC3339))
	credentialType := types.CredentialType("6D795F63726564656E7469616C")

	if err != nil {
		fmt.Printf("❌ Error converting expiration to ripple time: %s\n", err)
		return
	}

	expirationUint32, ok := typecheck.ToUint32(expiration)
	if !ok {
		fmt.Printf("❌ Expiration time %d is out of uint32 range\n", expiration)
		return
	}

	txn := &transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		CredentialType: credentialType,
		Subject:        types.Address(subjectWallet.ClassicAddress),
		Expiration:     expirationUint32,
		URI:            hex.EncodeToString([]byte("https://example.com")),
	}

	if err := submitAndWait(client, txn.Flatten(), issuer); err != nil {
		fmt.Println("❌", err)
		return
	}

	// -----------------------------------------------------

	// Creating the CredentialAccept transaction
	fmt.Println("⏳ Creating CredentialAccept transaction...")

	acceptTxn := &transaction.CredentialAccept{
		BaseTx: transaction.BaseTx{
			Account: types.Address(subjectWallet.ClassicAddress),
		},
		CredentialType: credentialType,
		Issuer:         types.Address(issuer.ClassicAddress),
	}

	if err := submitAndWait(client, acceptTxn.Flatten(), subjectWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	// -----------------------------------------------------

	// Creating the CredentialDelete transaction
	fmt.Println("⏳ Creating CredentialDelete transaction...")

	deleteTxn := &transaction.CredentialDelete{
		BaseTx: transaction.BaseTx{
			Account: types.Address(issuer.ClassicAddress),
		},
		CredentialType: credentialType,
		Issuer:         types.Address(issuer.ClassicAddress),
		Subject:        types.Address(subjectWallet.ClassicAddress),
	}

	if err := submitAndWait(client, deleteTxn.Flatten(), issuer); err != nil {
		fmt.Println("❌", err)
		return
	}
}

// submitAndWait autofills, signs, and submits one transaction.
func submitAndWait(client *websocket.Client, tx transaction.FlatTransaction, signer wallet.Wallet) error {
	fmt.Printf("⏳ Submitting %s transaction...\n", tx["TransactionType"])
	response, err := client.SubmitTxAndWait(tx, &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &signer,
	})
	if err != nil {
		return fmt.Errorf("submit %s: %w", tx["TransactionType"], err)
	}
	if !response.Validated || response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return fmt.Errorf("%s: validated=%t, result=%s", tx["TransactionType"], response.Validated, response.Meta.TransactionResult)
	}
	fmt.Printf("✅ %s transaction submitted\n", tx["TransactionType"])
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return nil
}
