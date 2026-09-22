package main

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path"
	"github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	txrequests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	txntypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

// stringToHex converts a string to its hex representation
func stringToHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

func main() {
	fmt.Println("🚀 Starting PermissionedDEX Example with WebSocket on Devnet")
	fmt.Println()

	// Setup client
	fmt.Println("⏳ Setting up devnet WebSocket client...")
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	defer func() {
		if err := client.Disconnect(); err != nil {
			fmt.Printf("Error disconnecting: %s\n", err)
		}
	}()

	if err := client.Connect(); err != nil {
		fmt.Printf("❌ Error connecting to devnet: %s\n", err)
		return
	}

	if !client.IsConnected() {
		fmt.Println("❌ Failed to connect to devnet")
		return
	}

	fmt.Println("✅ Connected to devnet")
	fmt.Println()

	// Setup wallets
	fmt.Println("⏳ Setting up wallets...")

	// Issuer wallet (testContext.wallet equivalent)
	issuerWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating issuer wallet: %s\n", err)
		return
	}

	err = client.FundWallet(&issuerWallet)
	if err != nil {
		fmt.Printf("❌ Error funding issuer wallet: %s\n", err)
		return
	}
	fmt.Printf("✅ Issuer wallet funded: %s\n", issuerWallet.ClassicAddress)

	// Wallet1
	wallet1, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating wallet1: %s\n", err)
		return
	}

	err = client.FundWallet(&wallet1)
	if err != nil {
		fmt.Printf("❌ Error funding wallet1: %s\n", err)
		return
	}
	fmt.Printf("✅ Wallet1 funded: %s\n", wallet1.ClassicAddress)

	// Wallet2
	wallet2, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating wallet2: %s\n", err)
		return
	}

	err = client.FundWallet(&wallet2)
	if err != nil {
		fmt.Printf("❌ Error funding wallet2: %s\n", err)
		return
	}
	fmt.Printf("✅ Wallet2 funded: %s\n", wallet2.ClassicAddress)
	fmt.Println()

	domainID, err := configureDomain(client, issuerWallet, wallet1, wallet2)
	if err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := fundDomainTokens(client, issuerWallet, wallet1, wallet2); err != nil {
		fmt.Println("❌", err)
		return
	}
	offerResponse, err := createHybridOffer(client, issuerWallet, wallet1, domainID)
	if err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := inspectDomainOffer(client, issuerWallet, wallet2, domainID, offerResponse); err != nil {
		fmt.Println("❌", err)
		return
	}
	if err := crossDomainOffers(client, issuerWallet, wallet1, wallet2, domainID); err != nil {
		fmt.Println("❌", err)
		return
	}
}

func configureDomain(client *websocket.Client, issuerWallet, wallet1, wallet2 wallet.Wallet) (string, error) {
	// Set the default ripple flag on the issuer's wallet
	fmt.Println("⏳ Setting default ripple flag on issuer wallet...")
	accountSetTx := &transaction.AccountSet{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
	}
	accountSetTx.SetAsfDefaultRipple()

	if _, err := submitAndWait(client, accountSetTx.Flatten(), issuerWallet); err != nil {
		return "", fmt.Errorf("set default ripple flag: %w", err)
	}
	fmt.Println("✅ Default ripple flag set")

	// Create credentials from issuer to wallet1 and wallet2
	credentialType := txntypes.CredentialType(stringToHex("Passport"))

	fmt.Println("⏳ Creating credential for wallet1...")
	credentialCreateTx1 := &transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
		Subject:        txntypes.Address(wallet1.ClassicAddress),
		CredentialType: credentialType,
	}

	if _, err := submitAndWait(client, credentialCreateTx1.Flatten(), issuerWallet); err != nil {
		return "", fmt.Errorf("create credential for wallet1: %w", err)
	}
	fmt.Println("✅ Credential created for wallet1")

	fmt.Println("⏳ Creating credential for wallet2...")
	credentialCreateTx2 := &transaction.CredentialCreate{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
		Subject:        txntypes.Address(wallet2.ClassicAddress),
		CredentialType: credentialType,
	}

	if _, err := submitAndWait(client, credentialCreateTx2.Flatten(), issuerWallet); err != nil {
		return "", fmt.Errorf("create credential for wallet2: %w", err)
	}
	fmt.Println("✅ Credential created for wallet2")

	// Create a Permissioned Domain ledger object
	fmt.Println("⏳ Creating PermissionedDomain...")
	permissionedDomainTx := &transaction.PermissionedDomainSet{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
		AcceptedCredentials: txntypes.AuthorizeCredentialList{
			{
				Credential: txntypes.Credential{
					CredentialType: credentialType,
					Issuer:         txntypes.Address(issuerWallet.ClassicAddress),
				},
			},
		},
	}

	if _, err := submitAndWait(client, permissionedDomainTx.Flatten(), issuerWallet); err != nil {
		return "", fmt.Errorf("create PermissionedDomain: %w", err)
	}
	fmt.Println("✅ PermissionedDomain created")

	// Accept credentials from wallet1 and wallet2
	fmt.Println("⏳ Accepting credential from wallet1...")
	credentialAcceptTx1 := &transaction.CredentialAccept{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet1.ClassicAddress),
		},
		Issuer:         txntypes.Address(issuerWallet.ClassicAddress),
		CredentialType: credentialType,
	}

	if _, err := submitAndWait(client, credentialAcceptTx1.Flatten(), wallet1); err != nil {
		return "", fmt.Errorf("accept credential from wallet1: %w", err)
	}
	fmt.Println("✅ Credential accepted by wallet1")

	fmt.Println("⏳ Accepting credential from wallet2...")
	credentialAcceptTx2 := &transaction.CredentialAccept{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet2.ClassicAddress),
		},
		Issuer:         txntypes.Address(issuerWallet.ClassicAddress),
		CredentialType: credentialType,
	}

	if _, err := submitAndWait(client, credentialAcceptTx2.Flatten(), wallet2); err != nil {
		return "", fmt.Errorf("accept credential from wallet2: %w", err)
	}
	fmt.Println("✅ Credential accepted by wallet2")

	// Fetch the domainID from the PermissionedDomain ledger object
	fmt.Println("⏳ Fetching PermissionedDomain details...")
	objectsReq := &account.ObjectsRequest{
		Account: txntypes.Address(issuerWallet.ClassicAddress),
		Type:    "permissioned_domain",
	}

	objectsResp, err := client.GetAccountObjects(objectsReq)
	if err != nil {
		return "", fmt.Errorf("error fetching account objects: %w", err)
	}

	if len(objectsResp.AccountObjects) == 0 {
		return "", fmt.Errorf("no PermissionedDomain object found")
	}

	permDomainObject := objectsResp.AccountObjects[0]
	domainID, ok := permDomainObject["index"].(string)
	if !ok {
		return "", fmt.Errorf("could not extract domain ID")
	}
	fmt.Printf("✅ PermissionedDomain ID: %s\n", domainID)
	return domainID, nil
}

func fundDomainTokens(client *websocket.Client, issuerWallet, wallet1, wallet2 wallet.Wallet) error {
	// Establish trust lines for USD IOU Token
	fmt.Println("⏳ Creating trust lines for USD...")

	// Wallet1 trust line
	trustSetTx1 := &transaction.TrustSet{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet1.ClassicAddress),
		},
		LimitAmount: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10000",
		},
	}

	if _, err := submitAndWait(client, trustSetTx1.Flatten(), wallet1); err != nil {
		return fmt.Errorf("create trust line for wallet1: %w", err)
	}
	fmt.Println("✅ Trust line created for wallet1")

	// Wallet2 trust line
	trustSetTx2 := &transaction.TrustSet{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet2.ClassicAddress),
		},
		LimitAmount: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10000",
		},
	}

	if _, err := submitAndWait(client, trustSetTx2.Flatten(), wallet2); err != nil {
		return fmt.Errorf("create trust line for wallet2: %w", err)
	}
	fmt.Println("✅ Trust line created for wallet2")

	// Send USD tokens to wallet1 and wallet2
	fmt.Println("⏳ Funding wallets with USD tokens...")

	// Payment to wallet1
	paymentTx1 := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
		Amount: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10000",
		},
		Destination: txntypes.Address(wallet1.ClassicAddress),
	}

	if _, err := submitAndWait(client, paymentTx1.Flatten(), issuerWallet); err != nil {
		return fmt.Errorf("send USD to wallet1: %w", err)
	}
	fmt.Println("✅ USD sent to wallet1")

	// Payment to wallet2
	paymentTx2 := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(issuerWallet.ClassicAddress),
		},
		Amount: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10000",
		},
		Destination: txntypes.Address(wallet2.ClassicAddress),
	}

	if _, err := submitAndWait(client, paymentTx2.Flatten(), issuerWallet); err != nil {
		return fmt.Errorf("send USD to wallet2: %w", err)
	}
	fmt.Println("✅ USD sent to wallet2")
	return nil
}

func createHybridOffer(client *websocket.Client, issuerWallet, wallet1 wallet.Wallet, domainID string) (*txrequests.TxResponse, error) {
	// Create hybrid offer
	fmt.Println("⏳ Creating hybrid offer...")
	offerCreateTx := &transaction.OfferCreate{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet1.ClassicAddress),
		},
		TakerGets: txntypes.XRPCurrencyAmount(1000),
		TakerPays: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10",
		},
		DomainID: &domainID,
	}
	offerCreateTx.SetHybridFlag()

	offerResponse, err := submitAndWait(client, offerCreateTx.Flatten(), wallet1)
	if err != nil {
		return nil, fmt.Errorf("create hybrid offer: %w", err)
	}
	fmt.Println("✅ Hybrid offer created")
	return offerResponse, nil
}

func inspectDomainOffer(client *websocket.Client, issuerWallet, wallet2 wallet.Wallet, domainID string, offerResponse *txrequests.TxResponse) error {
	// Validate offer characteristics
	fmt.Println("⏳ Validating offer characteristics...")

	// Get the transaction details instead of ledger entry (simplified for example)
	txReq := &txrequests.TxRequest{
		Transaction: offerResponse.Hash.String(),
	}

	txResp, err := client.Request(txReq)
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	var txResponse txrequests.TxResponse
	err = txResp.GetResult(&txResponse)
	if err != nil {
		return fmt.Errorf("error parsing transaction response: %w", err)
	}

	offerNode := txResponse.TxJSON
	fmt.Printf("✅ Offer ledger object retrieved\n")
	fmt.Printf("   📊 LedgerEntryType: %v\n", offerNode["LedgerEntryType"])
	fmt.Printf("   🏷️  DomainID: %v\n", offerNode["DomainID"])
	fmt.Printf("   👤 Account: %v\n", offerNode["Account"])

	// Validate AdditionalBooks field if present
	if additionalBooks, exists := offerNode["AdditionalBooks"]; exists {
		fmt.Printf("   📚 AdditionalBooks found: %v\n", additionalBooks)
	}

	// Validate book offers
	fmt.Println("⏳ Testing book_offers with domain...")
	bookOffersReq := &path.BookOffersRequest{
		TakerGets: types.BookOfferCurrency{
			Currency: "XRP",
		},
		TakerPays: types.BookOfferCurrency{
			Currency: "USD",
			Issuer:   string(issuerWallet.ClassicAddress),
		},
		Taker:  txntypes.Address(wallet2.ClassicAddress),
		Domain: &domainID,
	}

	bookOffersResp, err := client.GetBookOffers(bookOffersReq)
	if err != nil {
		return fmt.Errorf("error getting book offers: %w", err)
	}

	fmt.Printf("✅ Book offers retrieved: %d offers found\n", len(bookOffersResp.Offers))
	if len(bookOffersResp.Offers) > 0 {
		offer := bookOffersResp.Offers[0]
		fmt.Printf("   💰 TakerGets: %v\n", offer.TakerGets)
		fmt.Printf("   💵 TakerPays: %v\n", offer.TakerPays)
		fmt.Printf("   👤 Account: %v\n", offer.Account)
	}

	// Test subscribe command
	fmt.Println("⏳ Testing subscribe command with domain...")
	subscribeReq := &subscribe.Request{
		Streams: []string{"ledger"},
	}

	subscribeResp, err := client.Subscribe(subscribeReq)
	if err != nil {
		return fmt.Errorf("error subscribing: %w", err)
	}
	fmt.Printf("✅ Subscribe request successful\n")
	fmt.Printf("   📊 Server status: %s\n", subscribeResp.ServerStatus)
	fmt.Printf("   🔔 Note: Domain-specific book subscriptions would be tested here\n")
	return nil
}

func crossDomainOffers(client *websocket.Client, issuerWallet, wallet1, wallet2 wallet.Wallet, domainID string) error {
	// Test offer crossing within domain
	fmt.Println("⏳ Testing offer crossing within domain...")
	crossingOfferTx := &transaction.OfferCreate{
		BaseTx: transaction.BaseTx{
			Account: txntypes.Address(wallet2.ClassicAddress),
		},
		TakerPays: txntypes.XRPCurrencyAmount(1000),
		TakerGets: txntypes.IssuedCurrencyAmount{
			Currency: "USD",
			Issuer:   txntypes.Address(issuerWallet.ClassicAddress),
			Value:    "10",
		},
		DomainID: &domainID,
	}

	if _, err := submitAndWait(client, crossingOfferTx.Flatten(), wallet2); err != nil {
		return fmt.Errorf("create crossing offer: %w", err)
	}
	fmt.Println("✅ Crossing offer created")

	// Validate that offers are consumed
	fmt.Println("⏳ Validating offer consumption...")

	// Check wallet1 offers
	wallet1ObjectsReq := &account.ObjectsRequest{
		Account: txntypes.Address(wallet1.ClassicAddress),
		Type:    account.OfferObject,
	}

	wallet1ObjectsResp, err := client.GetAccountObjects(wallet1ObjectsReq)
	if err != nil {
		return fmt.Errorf("error getting wallet1 objects: %w", err)
	}

	fmt.Printf("✅ Wallet1 offers remaining: %d\n", len(wallet1ObjectsResp.AccountObjects))

	// Check wallet2 offers
	wallet2ObjectsReq := &account.ObjectsRequest{
		Account: txntypes.Address(wallet2.ClassicAddress),
		Type:    account.OfferObject,
	}

	wallet2ObjectsResp, err := client.GetAccountObjects(wallet2ObjectsReq)
	if err != nil {
		return fmt.Errorf("error getting wallet2 objects: %w", err)
	}

	fmt.Printf("✅ Wallet2 offers remaining: %d\n", len(wallet2ObjectsResp.AccountObjects))

	if len(wallet1ObjectsResp.AccountObjects) == 0 && len(wallet2ObjectsResp.AccountObjects) == 0 {
		fmt.Println("🎉 Success! Offers were successfully crossed and consumed within the PermissionedDEX domain")
	} else {
		fmt.Println("⚠️  Note: Some offers remain unconsumed")
	}

	fmt.Println()
	fmt.Println("🏁 PermissionedDEX example completed successfully!")
	return nil
}

// submitAndWait autofills, signs, and submits one transaction.
func submitAndWait(client *websocket.Client, tx transaction.FlatTransaction, signer wallet.Wallet) (*txrequests.TxResponse, error) {
	fmt.Printf("⏳ Submitting %s transaction...\n", tx["TransactionType"])
	response, err := client.SubmitTxAndWait(tx, &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &signer,
	})
	if err != nil {
		return nil, fmt.Errorf("submit %s: %w", tx["TransactionType"], err)
	}
	if !response.Validated || response.Meta.TransactionResult != transaction.TesSUCCESS.String() {
		return nil, fmt.Errorf("%s: validated=%t, result=%s", tx["TransactionType"], response.Validated, response.Meta.TransactionResult)
	}
	fmt.Printf("✅ %s transaction submitted\n", tx["TransactionType"])
	fmt.Printf("🌐 Hash: %s\n", response.Hash.String())
	fmt.Println()
	return response, nil
}
