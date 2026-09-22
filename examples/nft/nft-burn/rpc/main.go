package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	txrequests "github.com/Peersyst/xrpl-go/xrpl/queries/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/rpc/types"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	txnTypes "github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	// Initialize the RPC client configuration
	cfg, err := rpc.NewClientConfig(
		"https://s.devnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewDevnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	// Create the RPC client
	client := rpc.NewClient(cfg)

	// Step 1: Fund wallets
	fmt.Println("⏳ Funding wallets...")

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
	fmt.Println()

	// Step 2: Mint an NFT
	fmt.Println("⏳ Minting NFT...")

	nftMint := transaction.NFTokenMint{
		BaseTx: transaction.BaseTx{
			Account:         nftMinter.ClassicAddress,
			TransactionType: transaction.NFTokenMintTx,
		},
		NFTokenTaxon: 0,
		URI:          txnTypes.NFTokenURI("68747470733A2F2F676F6F676C652E636F6D"), // https://google.com
	}
	nftMint.SetTransferableFlag()

	responseMint, err := client.SubmitTxAndWait(nftMint.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftMinter,
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

	// Step 3: Retrieve the token ID
	fmt.Println("⏳ Retrieving NFT ID...")

	metadata := responseMint.Meta.AsNFTokenMintMetadata()

	if metadata.NFTokenID == nil {
		fmt.Println("❌ nftoken_id not found or not a string")
		return
	}

	fmt.Println("🌎 nftoken_id:", metadata.NFTokenID.String())
	fmt.Println()

	// Step 4: Burn the NFT
	fmt.Println("⏳ Burn the NFT...")

	nftBurn := transaction.NFTokenBurn{
		BaseTx: transaction.BaseTx{
			Account:         nftMinter.ClassicAddress,
			TransactionType: transaction.NFTokenBurnTx,
		},
		NFTokenID: txnTypes.NFTokenID(metadata.NFTokenID.String()),
	}

	responseBurn, err := client.SubmitTxAndWait(nftBurn.Flatten(), &types.SubmitOptions{
		Autofill: true,
		Wallet:   &nftMinter,
	})
	if err != nil {
		fmt.Println("❌ Error burning NFT:", err)
		return
	}
	if err := checkResult(responseBurn, transaction.TesSUCCESS); err != nil {
		fmt.Println("❌", err)
		return
	}
	fmt.Println("✅ NFT burned successfully! - 🌎 Hash: ", responseBurn.Hash)
}

func checkResult(response *txrequests.TxResponse, expected transaction.TxResult) error {
	if !response.Validated || response.Meta.TransactionResult != expected.String() {
		return fmt.Errorf("transaction failed: validated=%t, result=%s, expected=%s", response.Validated, response.Meta.TransactionResult, expected)
	}
	return nil
}
