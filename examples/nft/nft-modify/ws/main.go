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
		fmt.Println(err)
		return
	}

	if !client.IsConnected() {
		fmt.Println("❌ Failed to connect to devnet")
		return
	}

	fmt.Println("✅ Connected to devnet")
	fmt.Println()

	// Create and fund the nft wallet
	fmt.Println("⏳ Funding wallet...")
	nftWallet, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println("❌ Error creating nft wallet:", err)
		return
	}
	if err := client.FundWallet(&nftWallet); err != nil {
		fmt.Println("❌ Error funding nft wallet:", err)
		return
	}
	fmt.Println("💸 NFT wallet funded!")
	fmt.Println()

	// Mint NFT
	nftMint := transaction.NFTokenMint{
		BaseTx: transaction.BaseTx{
			Account:         nftWallet.ClassicAddress,
			TransactionType: transaction.NFTokenMintTx,
		},
		NFTokenTaxon: 0,
		URI:          txnTypes.NFTokenURI("68747470733A2F2F676F6F676C652E636F6D"), // https://google.com
	}
	nftMint.SetMutableFlag()
	nftMint.SetTransferableFlag()

	responseMint, err := client.SubmitTxAndWait(nftMint.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftWallet,
	})
	if err != nil {
		fmt.Println("❌ Error minting NFT:", err)
		return
	}
	if err := checkResult(responseMint, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ NFT minted successfully! - 🌎 Hash: ", responseMint.Hash)
	fmt.Println()

	metaMap := responseMint.Meta.AsNFTokenMintMetadata()

	if metaMap.NFTokenID == nil {
		fmt.Println("❌ nftoken_id not found or not a string")
		return
	}

	fmt.Println("🌎 nftoken_id:", metaMap.NFTokenID.String())
	fmt.Println()

	// Update NFT
	nftModify := transaction.NFTokenModify{
		BaseTx: transaction.BaseTx{
			Account:         nftWallet.ClassicAddress,
			TransactionType: transaction.NFTokenModifyTx,
		},
		URI:       "68747470733A2F2F7961686F6F2E636F6D", // https://yahoo.com
		NFTokenID: txnTypes.NFTokenID(metaMap.NFTokenID.String()),
	}

	responseModify, err := client.SubmitTxAndWait(nftModify.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftWallet,
	})
	if err != nil {
		fmt.Println("❌ Error modifying NFT:", err)
		return
	}
	if err := checkResult(responseModify, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ NFT URI modified successfully! - 🌎 Hash: ", responseModify.Hash)
}

func checkResult(response *txrequests.TxResponse, expected transaction.TxResult) error {
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("transaction failed: validated=%t, result=%s, expected=%s", response.Validated, response.Meta.TransactionResult, expected)
	}
	return nil
}
