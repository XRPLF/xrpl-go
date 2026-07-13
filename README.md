# XRPL-GO

[![Go Reference](https://pkg.go.dev/badge/github.com/Peersyst/xrpl-go.svg)](https://pkg.go.dev/github.com/Peersyst/xrpl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Peersyst/xrpl-go)](https://goreportcard.com/report/github.com/Peersyst/xrpl-go)
[![Release Card](https://img.shields.io/github/v/release/XRPLF/xrpl-go?include_prereleases)](https://github.com/XRPLF/xrpl-go/releases)

`xrpl-go` is a Go SDK for interacting with the [XRP Ledger](https://xrpl.org/). It provides address codecs, key management, binary serialization, typed transaction models, RPC and WebSocket clients, wallet helpers, local transaction signing, and multisigning.

## Reference documentation

See the [xrpl-go documentation](https://xrplf.github.io/xrpl-go/docs/installation) for guides and package docs, or browse the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go).

## Installation

`xrpl-go` requires Go `1.25.12` or later.

```bash
go get github.com/Peersyst/xrpl-go
```

## Quickstart

This example creates and funds a Testnet wallet, builds a payment, autofills the network fields, signs it locally, and submits it to a validated ledger.

```go
package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClient(cfg)

	sender, err := wallet.New(crypto.ED25519())
	if err != nil {
		log.Fatal(err)
	}

	if err := client.FundWallet(&sender); err != nil {
		log.Fatal(err)
	}

	drops, err := currency.XrpToDrops("1")
	if err != nil {
		log.Fatal(err)
	}
	amount, err := strconv.ParseUint(drops, 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	payment := transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: sender.ClassicAddress,
		},
		Destination: types.Address("rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe"),
		Amount:      types.XRPCurrencyAmount(amount),
		DeliverMax:  types.XRPCurrencyAmount(amount),
	}

	flatTx := payment.Flatten()
	if err := client.Autofill(&flatTx); err != nil {
		log.Fatal(err)
	}

	txBlob, txHash, err := sender.Sign(flatTx)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("submitted hash: %s\n", txHash)
	fmt.Printf("validated: %t, ledger: %d\n", res.Validated, res.LedgerIndex)
}
```

Never print, log, commit, or send real seeds, private keys, or mnemonics to telemetry. Anyone with those values can control the account.

## Transaction lifecycle

The usual write path is:

1. Build a typed transaction, such as `transaction.Payment`.
2. Call `Flatten()` to get a `transaction.FlatTransaction`.
3. Call `client.Autofill()` to add network fields like `Fee`, `Sequence`, and `LastLedgerSequence`.
4. Sign locally with `wallet.Sign()`.
5. Submit with `client.SubmitTxBlobAndWait()`, or use `client.SubmitTxAndWait()` to autofill, sign, submit, and wait in one call.

A transaction is a command you submit. Ledger entries are durable state you query after validation. The validated transaction response includes metadata that explains which ledger entries were created, modified, or deleted.

For offline signing, keep the `Autofill()` and `Sign()` boundary explicit: autofill needs network access, signing only needs the transaction data and wallet credentials.

See [`examples/send-xrp/rpc`](examples/send-xrp/rpc) for a longer payment example.

## Packages

| Package | Use it for |
| --- | --- |
| `address-codec` | Encode and decode XRPL classic addresses and X-addresses |
| `binary-codec` | Encode and decode XRPL objects and transactions in canonical binary format |
| `keypairs` | Generate seeds, derive keypairs, sign payloads, and verify signatures |
| `xrpl/rpc` | Send JSON-RPC requests, autofill transactions, submit transactions, and fund Testnet or Devnet wallets |
| `xrpl/websocket` | Connect to WebSocket servers, make requests, submit transactions, and subscribe to ledger streams |
| `xrpl/transaction` | Build typed XRPL transaction models |
| `xrpl/wallet` | Create wallets, derive wallets from seeds or mnemonics, sign transactions, multisign transactions, and authorize payment channels |

## Guides and resources

- [Create wallets and sign transactions](https://xrplf.github.io/xrpl-go/docs/xrpl/wallet)
- [Use the RPC client](https://xrplf.github.io/xrpl-go/docs/xrpl/rpc)
- [Use the WebSocket client](https://xrplf.github.io/xrpl-go/docs/xrpl/websocket)
- [Build transactions](https://xrplf.github.io/xrpl-go/docs/xrpl/transaction)
- [Learn XRPL concepts and protocol rules](https://xrpl.org/docs)

## Security and audits

The signing functionality in this repository has not been independently audited. Treat local signing as security-sensitive code: protect wallet secrets, test on Testnet or Devnet first, and review signing behavior before using it to control production funds.

## Contributing

Development setup, test commands, docs-site commands, and pull request guidance are in [CONTRIBUTING.md](CONTRIBUTING.md).

## Report an issue

If you find a bug or documentation gap, please open an issue in the [XRPL-GO GitHub repository](https://github.com/XRPLF/xrpl-go/issues).

## License

The `xrpl-go` library is licensed under the MIT License.
