package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	txrequests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	txnTypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
	"github.com/Peersyst/xrpl-go/xrpl/websocket/types"
)

func main() {
	// Connect to the XRPL devnet
	fmt.Println("⏳ Connecting to devnet...")
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

	if err := client.Connect(); err != nil {
		fmt.Println("❌ Error connecting to devnet:", err)
		return
	}

	if !client.IsConnected() {
		fmt.Println("❌ Failed to connect to devnet")
		return
	}
	fmt.Println("✅ Connected to devnet")
	fmt.Println()

	// Step 1: Fund wallet
	fmt.Println("⏳ Funding wallet...")

	// Create and fund the NFT minter wallet
	nftMinter, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating NFT minter wallet:", err)
		return
	}
	if err := client.FundWallet(&nftMinter); err != nil {
		fmt.Println("❌ Error funding NFT minter wallet:", err)
		return
	}
	fmt.Println("💸 NFT minter wallet funded!")

	// Step 2: Mint two NFTs with sell offers.
	fmt.Println("⏳ Minting first NFT...")

	nftMint := transaction.NFTokenMint{
		BaseTx: transaction.BaseTx{
			Account:         nftMinter.ClassicAddress,
			TransactionType: transaction.NFTokenMintTx,
		},
		NFTokenTaxon: 0,
		URI:          txnTypes.NFTokenURI("68747470733A2F2F676F6F676C652E636F6D"), // https://google.com
		Amount:       txnTypes.XRPCurrencyAmount(1000000),
	}
	nftMint.SetTransferableFlag()

	responseMint, err := client.SubmitTxAndWait(nftMint.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftMinter,
	})
	if err != nil {
		fmt.Println("❌ Error minting first NFT:", err)
		return
	}
	if err := checkResult(responseMint, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ First NFT minted successfully! - 🌎 Hash: ", responseMint.Hash)
	fmt.Println()

	// Retrieve the offer ID, not the NFT ID.
	fmt.Println("⏳ Retrieving first NFT offer ID...")

	metaMap := responseMint.Meta.AsNFTokenMintMetadata()
	if metaMap.OfferID == nil {
		fmt.Println("❌ offer_id not found")
		return
	}

	fmt.Println("🌎 offer_id:", metaMap.OfferID.String())
	fmt.Println()

	// ------

	fmt.Println("⏳ Minting second NFT...")

	nftMint2 := transaction.NFTokenMint{
		BaseTx: transaction.BaseTx{
			Account:         nftMinter.ClassicAddress,
			TransactionType: transaction.NFTokenMintTx,
		},
		NFTokenTaxon: 0,
		URI:          txnTypes.NFTokenURI("68747470733A2F2F676F6F676C652E636F6D"), // https://google.com
		Amount:       txnTypes.XRPCurrencyAmount(1000000),
	}
	nftMint2.SetTransferableFlag()

	responseMint2, err := client.SubmitTxAndWait(nftMint2.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftMinter,
	})
	if err != nil {
		fmt.Println("❌ Error minting second NFT:", err)
		return
	}
	if err := checkResult(responseMint2, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ Second NFT minted successfully! - 🌎 Hash: ", responseMint2.Hash)
	fmt.Println()

	// Retrieve the second offer ID.
	fmt.Println("⏳ Retrieving second NFT offer ID...")

	metaMap2 := responseMint2.Meta.AsNFTokenMintMetadata()
	if metaMap2.OfferID == nil {
		fmt.Println("❌ offer_id not found")
		return
	}

	fmt.Println("🌎 offer_id:", metaMap2.OfferID.String())
	fmt.Println()

	// Step 4: Cancel the NFT offers
	fmt.Println("⏳ Canceling NFT offers...")

	nftCancel := transaction.NFTokenCancelOffer{
		BaseTx: transaction.BaseTx{
			Account:         nftMinter.ClassicAddress,
			TransactionType: transaction.NFTokenCancelOfferTx,
		},
		NFTokenOffers: []txnTypes.NFTokenID{
			txnTypes.NFTokenID(metaMap.OfferID.String()),
			txnTypes.NFTokenID(metaMap2.OfferID.String()),
		},
	}

	response, err := client.SubmitTxAndWait(nftCancel.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftMinter,
	})
	if err != nil {
		fmt.Println("❌ Error canceling NFT offers:", err)
		return
	}
	if err := checkResult(response, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ NFT offers canceled successfully! - 🌎 Hash: ", response.Hash)
}

func checkResult(response *txrequests.TxResponse, expected transaction.TxResult) error {
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("transaction failed: validated=%t, result=%s, expected=%s", response.Validated, response.Meta.TransactionResult, expected)
	}
	return nil
}
