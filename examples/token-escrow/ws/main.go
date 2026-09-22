package main

import (
	"fmt"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/typecheck"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	txrequests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	rippleTime "github.com/Peersyst/xrpl-go/xrpl/time"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	txnTypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	wstypes "github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

func main() {
	//
	// Configure client
	//
	fmt.Println("⏳ Setting up client...")
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.devnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
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

	// Configure wallets
	issuerWallet, holderWallet, holderWallet2 := createWallets(client)

	// Configure issuer wallet to allow trust line locking
	if err := configureIssuerWallet(client, issuerWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	// Create trust line from holder to issuer
	if err := createTrustLine(client, issuerWallet, holderWallet, holderWallet2); err != nil {
		fmt.Println("❌", err)
		return
	}

	// Mint token from issuer to holder
	if err := mintToken(client, issuerWallet, holderWallet); err != nil {
		fmt.Println("❌", err)
		return
	}

	// Create escrow, the holder will escrow 100 tokens to holder 2.
	offerSequence, finishAfter, err := createEscrow(client, issuerWallet, holderWallet, holderWallet2)
	if err != nil {
		fmt.Println("❌", err)
		return
	}

	// Finish escrow after the ledger's close time passes FinishAfter.
	if err := finishEscrow(client, holderWallet, holderWallet2, offerSequence, finishAfter); err != nil {
		fmt.Println("❌", err)
		return
	}
}

// createWallets configures the issuer and holder wallets.
func createWallets(client *websocket.Client) (issuerWallet, holderWallet, holderWallet2 wallet.Wallet) {
	fmt.Println("⏳ Setting up wallets...")
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
	fmt.Println("💸 Issuer wallet funded!")

	// Holder wallet
	holderWallet, err = wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating holder wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&holderWallet)
	if err != nil {
		fmt.Printf("❌ Error funding holder wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Holder wallet funded!")

	// Holder wallet 2
	holderWallet2, err = wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Printf("❌ Error creating holder wallet 2: %s\n", err)
		return
	}
	err = client.FundWallet(&holderWallet2)
	if err != nil {
		fmt.Printf("❌ Error funding holder wallet 2: %s\n", err)
		return
	}
	fmt.Println("💸 Holder wallet 2 funded!")

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println("💳 Issuer wallet:", issuerWallet.ClassicAddress)
	fmt.Println("💳 Holder wallet:", holderWallet.ClassicAddress)
	fmt.Println("💳 Holder wallet 2:", holderWallet2.ClassicAddress)
	fmt.Println()

	return issuerWallet, holderWallet, holderWallet2
}

// configureIssuerWallet configures the issuer wallet to allow trust line locking.
func configureIssuerWallet(client *websocket.Client, issuerWallet wallet.Wallet) error {
	fmt.Println("⏳ Configuring issuer wallet...")
	accountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: issuerWallet.ClassicAddress,
		},
	}
	accountSet.SetAsfAllowTrustLineLocking()
	accountSetResponse, err := client.SubmitTxAndWait(accountSet.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &issuerWallet,
	})
	if err != nil {
		return err
	}
	if err := checkResult(accountSetResponse, transactions.TesSUCCESS); err != nil {
		return err
	}
	fmt.Println("✅ Issuer wallet configured!")
	fmt.Printf("🌐 Hash: %s\n", accountSetResponse.Hash.String())
	fmt.Println()
	return nil
}

// createTrustLine creates a trust line for the holder wallet.
func createTrustLine(client *websocket.Client, issuerWallet, holderWallet, holderWallet2 wallet.Wallet) error {
	fmt.Println("⏳ Creating trust line for holder wallet...")
	trustLine := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: holderWallet.ClassicAddress,
		},
		LimitAmount: txnTypes.IssuedCurrencyAmount{
			Issuer:   issuerWallet.ClassicAddress,
			Currency: currency.ConvertStringToHex("ABCD"),
			Value:    "1000000",
		},
	}
	trustLine.SetSetNoRippleFlag()
	trustLineResponse, err := client.SubmitTxAndWait(trustLine.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet,
	})
	if err != nil {
		return err
	}
	if err := checkResult(trustLineResponse, transactions.TesSUCCESS); err != nil {
		return err
	}
	fmt.Println("✅ Trust line created for holder wallet!")
	fmt.Printf("🌐 Hash: %s\n", trustLineResponse.Hash.String())
	fmt.Println()

	fmt.Println("⏳ Creating trust line for holder wallet 2...")
	trustLine = &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: holderWallet2.ClassicAddress,
		},
		LimitAmount: txnTypes.IssuedCurrencyAmount{
			Issuer:   issuerWallet.ClassicAddress,
			Currency: currency.ConvertStringToHex("ABCD"),
			Value:    "1000000",
		},
	}
	trustLine.SetSetNoRippleFlag()
	trustLineResponse, err = client.SubmitTxAndWait(trustLine.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet2,
	})
	if err != nil {
		return err
	}
	if err := checkResult(trustLineResponse, transactions.TesSUCCESS); err != nil {
		return err
	}
	fmt.Println("✅ Trust line created for holder wallet 2!")
	fmt.Printf("🌐 Hash: %s\n", trustLineResponse.Hash.String())
	fmt.Println()
	return nil
}

// mintToken mints a token for the holder wallet.
func mintToken(client *websocket.Client, issuerWallet, holderWallet wallet.Wallet) error {
	fmt.Println("⏳ Minting token to holder wallet...")
	token := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: issuerWallet.ClassicAddress,
		},
		Destination: holderWallet.ClassicAddress,
		Amount: txnTypes.IssuedCurrencyAmount{
			Issuer:   issuerWallet.ClassicAddress,
			Currency: currency.ConvertStringToHex("ABCD"),
			Value:    "10000",
		},
	}
	tokenResponse, err := client.SubmitTxAndWait(token.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &issuerWallet,
	})
	if err != nil {
		return err
	}
	if err := checkResult(tokenResponse, transactions.TesSUCCESS); err != nil {
		return err
	}
	fmt.Println("✅ Token minted!")
	fmt.Printf("🌐 Hash: %s\n", tokenResponse.Hash.String())
	fmt.Println()
	return nil
}

// createEscrow creates an escrow for the holder wallet.
func createEscrow(client *websocket.Client, issuerWallet, holderWallet, holderWallet2 wallet.Wallet) (uint32, uint32, error) {
	fmt.Println("⏳ Creating escrow...")
	cancelAfter, ok := typecheck.ToUint32(rippleTime.UnixTimeToRippleTime(time.Now().Unix()) + 4000)
	if !ok {
		return 0, 0, fmt.Errorf("CancelAfter is out of uint32 range")
	}
	finishAfter, ok := typecheck.ToUint32(rippleTime.UnixTimeToRippleTime(time.Now().Unix() + 5))
	if !ok {
		return 0, 0, fmt.Errorf("FinishAfter is out of uint32 range")
	}
	escrow := &transactions.EscrowCreate{
		BaseTx: transactions.BaseTx{
			Account: holderWallet.ClassicAddress,
		},
		Amount: txnTypes.IssuedCurrencyAmount{
			Issuer:   issuerWallet.ClassicAddress,
			Currency: currency.ConvertStringToHex("ABCD"),
			Value:    "100",
		},
		Destination: holderWallet2.ClassicAddress,
		CancelAfter: cancelAfter,
		FinishAfter: finishAfter,
	}
	escrowResponse, err := client.SubmitTxAndWait(escrow.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("create escrow: %w", err)
	}
	if err := checkResult(escrowResponse, transactions.TesSUCCESS); err != nil {
		return 0, 0, err
	}
	fmt.Println("✅ Escrow created!")
	fmt.Printf("🌐 Hash: %s\n", escrowResponse.Hash.String())
	fmt.Printf("🌐 Sequence: %d\n", escrowResponse.TxJSON.Sequence())
	fmt.Println()

	return escrowResponse.TxJSON.Sequence(), finishAfter, nil
}

// finishEscrow finishes the escrow for the holder wallet 2.
func finishEscrow(client *websocket.Client, holderWallet, holderWallet2 wallet.Wallet, offerSequence, finishAfter uint32) error {
	fmt.Println("⏳ Waiting for the validated ledger to pass FinishAfter...")
	deadline := time.Now().Add(time.Minute)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for FinishAfter")
		}
		response, err := client.GetLedger(&ledger.Request{LedgerIndex: common.Validated})
		if err != nil {
			return fmt.Errorf("get validated ledger: %w", err)
		}
		if int64(response.Ledger.CloseTime) > int64(finishAfter) {
			break
		}
		time.Sleep(time.Second)
	}

	fmt.Println("⏳ Finishing escrow...")
	escrow := &transactions.EscrowFinish{
		BaseTx: transactions.BaseTx{
			Account: holderWallet2.ClassicAddress,
		},
		Owner:         holderWallet.ClassicAddress,
		OfferSequence: offerSequence,
	}
	escrowResponse, err := client.SubmitTxAndWait(escrow.Flatten(), &wstypes.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet2,
	})
	if err != nil {
		return fmt.Errorf("finish escrow: %w", err)
	}
	if err := checkResult(escrowResponse, transactions.TesSUCCESS); err != nil {
		return err
	}
	fmt.Println("✅ Escrow finished!")
	fmt.Printf("🌐 Hash: %s\n", escrowResponse.Hash.String())
	fmt.Println()
	return nil
}

//nolint:unparam // Keep the expected result explicit so readers can adapt the example.
func checkResult(response *txrequests.TxResponse, expected transactions.TxResult) error {
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("transaction failed: validated=%t, result=%s, expected=%s", response.Validated, response.Meta.TransactionResult, expected)
	}
	return nil
}
