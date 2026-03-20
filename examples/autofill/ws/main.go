package main

import (
	"encoding/hex"
	"fmt"

	transactions "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	wsClient := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233"),
	)

	defer func() {
		if err := wsClient.Disconnect(); err != nil {
			fmt.Println("Error disconnecting:", err)
		}
	}()

	fmt.Println("Connecting to server...")
	if err := wsClient.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connection: ", wsClient.IsConnected())

	w, err := wallet.FromSeed("sEdSMVV4dJ1JbdBxmakRR4Puu3XVZz2", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	receiverWallet, err := wallet.FromSeed("sEd7d8Ci9nevdLCeUMctF3uGXp9WQqJ", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	payment := transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(w.GetAddress()),
			Signers: []types.Signer{
				{
					SignerData: types.SignerData{
						Account:       types.Address(w.GetAddress()),
						SigningPubKey: w.PublicKey,
						TxnSignature:  "",
					},
				},
				{
					SignerData: types.SignerData{
						Account:       types.Address(w.GetAddress()),
						SigningPubKey: w.PublicKey,
						TxnSignature:  "",
					},
				},
			},
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
						MemoType:   hex.EncodeToString([]byte("message 2")),
					},
				},
			},
		},
		Destination: types.Address(receiverWallet.GetAddress()),
		Amount:      types.XRPCurrencyAmount(100000000),
	}

	tx := payment.Flatten()

	fmt.Println("Transaction before autofill", tx)

	err = wsClient.Autofill(&tx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Transaction after autofill", tx)
}
