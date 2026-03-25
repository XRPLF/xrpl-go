package main

import (
	"fmt"
	"time"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	rippleTime "github.com/Peersyst/xrpl-go/xrpl/time"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	txnTypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

// safeInt64ToUint32 safely converts int64 to uint32 with bounds checking
func safeInt64ToUint32(value int64) uint32 {
	if value < 0 {
		return 0
	}
	if value > int64(^uint32(0)) {
		return ^uint32(0) // max uint32 value
	}
	return uint32(value)
}

func main() {
	//
	// Configure client
	//
	fmt.Println("⏳ Setting up client...")
	cfg, err := rpc.NewClientConfig(
		"https://s.devnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	client := rpc.NewClient(cfg)
	fmt.Println("✅ Client configured!")
	fmt.Println()

	// Configure wallets
	issuerWallet, holderWallet, holderWallet2 := createWallets(client)

	// Configure issuer wallet to allow trust line locking
	configureIssuerWallet(client, issuerWallet)

	// Create trust line from holder to issuer
	createTrustLine(client, issuerWallet, holderWallet, holderWallet2)

	// Mint token from issuer to holder
	mintToken(client, issuerWallet, holderWallet)

	// Create escrow, the holder will escrow 100 tokens to the issuer
	offerSequence := createEscrow(client, issuerWallet, holderWallet, holderWallet2)

	// Finish escrow
	finishEscrow(client, holderWallet, holderWallet2, offerSequence)
}

// createWallets configures the issuer and holder wallets.
func createWallets(client *rpc.Client) (issuerWallet, holderWallet, holderWallet2 wallet.Wallet) {
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
func configureIssuerWallet(client *rpc.Client, issuerWallet wallet.Wallet) {
	fmt.Println("⏳ Configuring issuer wallet...")
	accountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: issuerWallet.ClassicAddress,
		},
	}
	accountSet.SetAsfAllowTrustLineLocking()
	accountSetResponse, err := client.SubmitTxAndWait(accountSet.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &issuerWallet,
	})
	if err != nil {
		fmt.Printf("❌ Error configuring issuer wallet: %s\n", err)
		return
	}
	fmt.Println("✅ Issuer wallet configured!")
	fmt.Printf("🌐 Hash: %s\n", accountSetResponse.Hash.String())
	fmt.Println()
}

// createTrustLine creates a trust line for the holder wallet.
func createTrustLine(client *rpc.Client, issuerWallet, holderWallet, holderWallet2 wallet.Wallet) {
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
	trustLineResponse, err := client.SubmitTxAndWait(trustLine.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet,
	})
	if err != nil {
		fmt.Printf("❌ Error creating trust line: %s\n", err)
		return
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
	trustLineResponse, err = client.SubmitTxAndWait(trustLine.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet2,
	})
	if err != nil {
		fmt.Printf("❌ Error creating trust line: %s\n", err)
		return
	}
	fmt.Println("✅ Trust line created for holder wallet 2!")
	fmt.Printf("🌐 Hash: %s\n", trustLineResponse.Hash.String())
	fmt.Println()
}

// mintToken mints a token for the holder wallet.
func mintToken(client *rpc.Client, issuerWallet, holderWallet wallet.Wallet) {
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
	tokenResponse, err := client.SubmitTxAndWait(token.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &issuerWallet,
	})
	if err != nil {
		fmt.Printf("❌ Error minting token: %s\n", err)
		return
	}
	fmt.Println("✅ Token minted!")
	fmt.Printf("🌐 Hash: %s\n", tokenResponse.Hash.String())
	fmt.Println()
}

// createEscrow creates an escrow for the holder wallet.
func createEscrow(client *rpc.Client, issuerWallet, holderWallet, holderWallet2 wallet.Wallet) (offerSequence uint32) {
	fmt.Println("⏳ Creating escrow...")
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
		CancelAfter: safeInt64ToUint32(rippleTime.UnixTimeToRippleTime(time.Now().Unix()) + 4000),
		FinishAfter: safeInt64ToUint32(rippleTime.UnixTimeToRippleTime(time.Now().Unix() + 5)),
	}
	escrowResponse, err := client.SubmitTxAndWait(escrow.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet,
	})
	if err != nil {
		fmt.Printf("❌ Error creating escrow: %s\n", err)
		return
	}
	fmt.Println("✅ Escrow created!")
	fmt.Printf("🌐 Hash: %s\n", escrowResponse.Hash.String())
	fmt.Printf("🌐 Sequence: %d\n", escrowResponse.TxJSON.Sequence())
	fmt.Println()

	return escrowResponse.TxJSON.Sequence()
}

// finishEscrow finishes the escrow for the holder wallet 2.
func finishEscrow(client *rpc.Client, holderWallet, holderWallet2 wallet.Wallet, offerSequence uint32) {
	fmt.Println("⏳ Finishing escrow...")
	escrow := &transactions.EscrowFinish{
		BaseTx: transactions.BaseTx{
			Account: holderWallet2.ClassicAddress,
		},
		Owner:         holderWallet.ClassicAddress,
		OfferSequence: offerSequence,
	}
	escrowResponse, err := client.SubmitTxAndWait(escrow.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &holderWallet2,
	})
	if err != nil {
		fmt.Printf("❌ Error finishing escrow: %s\n", err)
		return
	}
	fmt.Println("✅ Escrow finished!")
	fmt.Printf("🌐 Hash: %s\n", escrowResponse.Hash.String())
	fmt.Println()
}
