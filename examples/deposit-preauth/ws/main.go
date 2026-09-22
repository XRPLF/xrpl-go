package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	rippletime "github.com/Peersyst/xrpl-go/xrpl/time"
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
		fmt.Printf("❌ Error creating credential issuer wallet: %s\n", err)
		return
	}

	err = client.FundWallet(&issuer)
	if err != nil {
		fmt.Printf("❌ Error funding credential issuer wallet: %s\n", err)
		return
	}
	fmt.Printf("✅ Credential issuer wallet funded: %s\n", issuer.ClassicAddress)

	// -----------------------------------------------------

	// Holder 1
	fmt.Println("⏳ Setting up holder 1 wallet...")
	holderWallet1, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating holder 1 wallet: %s\n", err)
		return
	}

	err = client.FundWallet(&holderWallet1)
	if err != nil {
		fmt.Printf("❌ Error funding holder 1 wallet: %s\n", err)
		return
	}
	fmt.Printf("✅ Holder 1 wallet funded: %s\n", holderWallet1.ClassicAddress)

	// -----------------------------------------------------

	credentialType := types.CredentialType("6D795F63726564656E7469616C") // my_credential
	if err := configureCredentialAuthorization(client, issuer, holderWallet1, credentialType); err != nil {
		fmt.Println("❌", err)
		return
	}
	credentialID, err := getCredentialID(client, holderWallet1)
	if err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := sendPaymentsAndRevokeAuthorization(client, issuer, holderWallet1, credentialType, credentialID); err != nil {
		fmt.Println("❌", err)
		return
	}
}

func configureCredentialAuthorization(client *websocket.Client, issuer, holderWallet1 wallet.Wallet, credentialType types.CredentialType) error {
	// Enabling DepositAuth on the issuer account with an AccountSet transaction
	fmt.Println("⏳ Enabling DepositAuth on the issuer account...")
	accountSetTx := &transaction.AccountSet{
		BaseTx: transaction.BaseTx{
			Account:         issuer.ClassicAddress,
			TransactionType: transaction.AccountSetTx,
		},
	}
	accountSetTx.SetAsfDepositAuth()

	if err := submitAndWait(client, accountSetTx.Flatten(), issuer, transaction.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Creating the CredentialCreate transaction
	fmt.Println("⏳ Creating the CredentialCreate transaction...")

	expiration, err := rippletime.IsoTimeToRippleTime(time.Now().Add(time.Hour * 24).Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("error converting expiration to ripple time: %w", err)
	}
	expirationUint32, ok := typecheck.ToUint32(expiration)
	if !ok {
		return fmt.Errorf("expiration time %d is out of uint32 range", expiration)
	}

	credentialCreateTx := &transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{
			Account:         issuer.ClassicAddress,
			TransactionType: transaction.CredentialCreateTx,
		},
		Expiration:     expirationUint32,
		CredentialType: credentialType,
		Subject:        types.Address(holderWallet1.ClassicAddress),
		URI:            hex.EncodeToString([]byte("https://example.com")),
	}

	if err := submitAndWait(client, credentialCreateTx.Flatten(), issuer, transaction.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Creating the CredentialAccept transaction
	fmt.Println("⏳ Creating the CredentialAccept transaction...")

	credentialAcceptTx := &transaction.CredentialAccept{
		BaseTx: transaction.BaseTx{
			Account:         holderWallet1.ClassicAddress,
			TransactionType: transaction.CredentialAcceptTx,
		},
		CredentialType: credentialType,
		Issuer:         types.Address(issuer.ClassicAddress),
	}

	if err := submitAndWait(client, credentialAcceptTx.Flatten(), holderWallet1, transaction.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Creating the DepositPreauth transaction
	fmt.Println("⏳ Creating the DepositPreauth transaction using AuthorizeCredentials...")

	depositPreauthTx := &transaction.DepositPreauth{
		BaseTx: transaction.BaseTx{
			Account:         issuer.ClassicAddress,
			TransactionType: transaction.DepositPreauthTx,
		},
		AuthorizeCredentials: []types.AuthorizeCredentialsWrapper{
			{
				Credential: types.AuthorizeCredentials{
					Issuer:         issuer.ClassicAddress,
					CredentialType: credentialType,
				},
			},
		},
	}

	return submitAndWait(client, depositPreauthTx.Flatten(), issuer, transaction.TesSUCCESS)
}

func getCredentialID(client *websocket.Client, holderWallet1 wallet.Wallet) (string, error) {
	// Get the credential ID
	fmt.Println("⏳ Getting the credential ID from the holder 1 account...")

	objectsRequest := &account.ObjectsRequest{
		Account:     holderWallet1.ClassicAddress,
		Type:        account.CredentialObject,
		LedgerIndex: common.Validated,
	}

	objectsResponse, err := client.GetAccountObjects(objectsRequest)
	if err != nil {
		return "", fmt.Errorf("error getting the credential ID: %w", err)
	}

	// Check if we have any credential objects
	if len(objectsResponse.AccountObjects) == 0 {
		return "", fmt.Errorf("no credential objects found")
	}

	// Extract the credential ID
	credentialID, ok := objectsResponse.AccountObjects[0]["index"].(string)
	if !ok {
		return "", fmt.Errorf("could not extract credential ID from response")
	}

	fmt.Printf("✅ Credential ID: %s\n", credentialID)
	fmt.Println()
	return credentialID, nil
}

func sendPaymentsAndRevokeAuthorization(client *websocket.Client, issuer, holderWallet1 wallet.Wallet, credentialType types.CredentialType, credentialID string) error {
	// Sending XRP to the holder 1 account
	fmt.Println("⏳ Sending XRP to the issuer account, should succeed...")

	sendTx := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:         holderWallet1.ClassicAddress,
			TransactionType: transaction.PaymentTx,
		},
		Amount:        types.XRPCurrencyAmount(1000000),
		Destination:   issuer.ClassicAddress,
		CredentialIDs: types.CredentialIDs{credentialID},
	}

	if err := submitAndWait(client, sendTx.Flatten(), holderWallet1, transaction.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Unauthorizing the holder 1 account
	fmt.Println("⏳ Unauthorizing the holder 1 account with the DepositPreauth transaction and the UnauthorizeCredentials field...")

	unauthorizeTx := &transaction.DepositPreauth{
		BaseTx: transaction.BaseTx{
			Account:         issuer.ClassicAddress,
			TransactionType: transaction.DepositPreauthTx,
		},
		UnauthorizeCredentials: []types.AuthorizeCredentialsWrapper{
			{
				Credential: types.AuthorizeCredentials{
					Issuer:         issuer.ClassicAddress,
					CredentialType: credentialType,
				},
			},
		},
	}

	if err := submitAndWait(client, unauthorizeTx.Flatten(), issuer, transaction.TesSUCCESS); err != nil {
		return err
	}

	// -----------------------------------------------------

	// Sending XRP to the holder 1 account again (which should fail)
	fmt.Println("⏳ Sending XRP to the issuer account again (which should fail)...")

	sendTx2 := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:         holderWallet1.ClassicAddress,
			TransactionType: transaction.PaymentTx,
		},
		Amount:        types.XRPCurrencyAmount(1000000),
		Destination:   issuer.ClassicAddress,
		CredentialIDs: types.CredentialIDs{credentialID},
	}

	return submitAndWait(client, sendTx2.Flatten(), holderWallet1, transaction.TecNO_PERMISSION)
}

// submitAndWait requires the expected result in a validated ledger.
func submitAndWait(client *websocket.Client, tx transaction.FlatTransaction, signer wallet.Wallet, expected transaction.TxResult) error {
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
