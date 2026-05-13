package main

import (
	"encoding/hex"
	"fmt"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

const (
	destinationAddress = "rDwvihpE48E48F8rvNrqTb2UGWv62xqYTg"
	currency           = "USD"
	value              = "100"
	issuer             = "rDwvihpE48E48F8rvNrqTb2UGWv62xqYTg"
)

func main() {
	// WARNING: Do not print or log real wallet secrets. Private keys, seeds, and
	// mnemonics are credentials; exposing them leaks control of the account.
	mnemonicWallet, err := wallet.FromMnemonic("monster march exile fee forget response seven push dragon oil clinic attack black miss craft surface patient stomach tank float cabbage visual image resource")
	if err != nil {
		panic(err)
	}

	fmt.Println("Wallet generated from mnemonic")

	fmt.Printf("Private key: %s\n", mnemonicWallet.PrivateKey)
	fmt.Printf("Public 	key: %s\n", mnemonicWallet.PublicKey)
	fmt.Printf("Classic address: %s\n", mnemonicWallet.ClassicAddress)
	fmt.Printf("Seed: %s\n", mnemonicWallet.Seed)

	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		panic(err)
	}
	fmt.Println("Wallet generated from random seed")

	fmt.Printf("Private key: %s\n", w.PrivateKey)
	fmt.Printf("Public 	key: %s\n", w.PublicKey)
	fmt.Printf("Classic address: %s\n", w.ClassicAddress)
	fmt.Printf("Seed: %s\n", w.Seed)

	walletFromSeed, _ := wallet.FromSeed(w.Seed, "")

	fmt.Println("\nWallet generated from previous seed")

	fmt.Printf("Private key: %s\n", walletFromSeed.PrivateKey)
	fmt.Printf("Public 	key: %s\n", walletFromSeed.PublicKey)
	fmt.Printf("Classic address: %s\n", walletFromSeed.ClassicAddress)
	fmt.Printf("Seed: %s\n", walletFromSeed.Seed)

	walletFromSecret, _ := wallet.FromSecret(w.Seed)

	fmt.Println("\nWallet generated from previous seed")

	fmt.Printf("Private key: %s\n", walletFromSecret.PrivateKey)
	fmt.Printf("Public 	key: %s\n", walletFromSecret.PublicKey)
	fmt.Printf("Classic address: %s\n", walletFromSecret.ClassicAddress)
	fmt.Printf("Seed: %s\n", walletFromSecret.Seed)

	tx := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(w.ClassicAddress),
			Memos: []types.MemoWrapper{
				{
					Memo: types.Memo{
						MemoData:   hex.EncodeToString([]byte("Hello, World!")),
						MemoFormat: hex.EncodeToString([]byte("text/plain")),
						MemoType:   hex.EncodeToString([]byte("message")),
					},
				},
				{
					Memo: types.Memo{
						MemoData:   hex.EncodeToString([]byte("Hello, World 2!")),
						MemoFormat: hex.EncodeToString([]byte("text/plain")),
						MemoType:   hex.EncodeToString([]byte("message2")),
					},
				},
			},
		},
		Amount: types.IssuedCurrencyAmount{
			Issuer:   issuer,
			Currency: currency,
			Value:    value,
		},
		Destination: types.Address(destinationAddress),
	}

	fmt.Println(tx.Flatten())

	fmt.Println("\nSigning a transaction with wallet generated from seed")

	txBlob, hash, err := w.Sign(tx.Flatten())
	if err != nil {
		panic(err)
	}

	fmt.Printf("txBlob: %s\n", txBlob)
	fmt.Printf("hash: %s\n", hash)

	fmt.Println("\nSigning a transaction with wallet generated from mnemonic")

	mnemonicTx := transactions.Payment{
		BaseTx: transactions.BaseTx{
			TransactionType: "Payment",
			Account:         types.Address(mnemonicWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Issuer:   issuer,
			Currency: currency,
			Value:    value,
		},
		Destination: types.Address(destinationAddress),
	}

	mnemonicTxBlob, mnemonicHash, err := mnemonicWallet.Sign(mnemonicTx.Flatten())
	if err != nil {
		panic(err)
	}

	fmt.Printf("txBlob: %s\n", mnemonicTxBlob)
	fmt.Printf("hash: %s\n", mnemonicHash)
}
