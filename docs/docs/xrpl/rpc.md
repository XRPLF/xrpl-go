# rpc

## Overview

The `rpc` package provides the RPC client for interacting with the XRPL network via its RPC API. This client handles the communication with XRPL nodes, allowing you to:

- Send requests to query the ledger state.
- Submit transactions to the network.
- Receive responses and handle errors.
- Manage the connections configuration.

## Client

The `rpc` package provides a `Client` type of communication with XRPL nodes. This client is configurable and let the user submit transactions and make queries.

In order to create a new `Client`, you can use the `NewClient` function:

```go
cfg, err := rpc.NewClientConfig("<url>")
if err != nil {
	// ...
}
client := rpc.NewClient(cfg)
```

Every time you create a new `Client`, you need to provide a `Config` struct as an argument. You can initialize a `Config` struct using the `NewClientConfig` function.

`Config` struct follows the options pattern, so you can pass different options to the `NewClientConfig` function.

### FaucetProvider

The `WithFaucetProvider` option allows you to set the faucet provider of the WebSocket client. There're two predefined faucet providers: `TestnetFaucetProvider` and `DevnetFaucetProvider`. You can also implement your own faucet provider by implementing the `FaucetProvider` interface.

```go
func (wc ClientConfig) WithFaucetProvider(fp common.FaucetProvider) ClientConfig
```

### MaxFeeXRP

The `WithMaxFeeXRP` option allows you to set the maximum fee in XRP that the client will use.

```go
func (wc ClientConfig) WithMaxFeeXRP(maxFeeXRP float32) ClientConfig
```

### FeeCushion

The `WithFeeCushion` option allows you to set the fee cushion for a transaction.

```go
func (wc ClientConfig) WithFeeCushion(feeCushion float32) ClientConfig
```

### MaxResponseSize

The `WithMaxResponseSize` option caps HTTP response bodies. The default is 64 MiB. Set it to `0` to disable the limit.

```go
func WithMaxResponseSize(maxResponseSize int64) ConfigOpt
```

### Logger

The `SetLogger` function overrides the logger used for SDK warnings, such as remote non-TLS URL warnings. Pass `nil` to silence these warnings.

```go
func SetLogger(l *log.Logger)
```

So, for example, if you want to set a custom `FaucetProvider` and `FeeCushion`, you can do it this way:

```go
cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/",
	rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	rpc.WithFeeCushion(1.5),
)
if err != nil {
	// ...
}
client := rpc.NewClient(cfg)
```

## Methods

`Client` offers different methods to interact with the XRPL network.

### Request

The `Request` method is used to make queries to the XRPL network. It returns a `XRPLResponse` interface. This method is used in the client's queries requests.

```go
// Client methods
func (c *Client) Request(reqParams XRPLRequest) (XRPLResponse, error)
```

### Autofill/AutofillMultisigned

The `Autofill` method is used to autofill some fields in a flat transaction. This method is useful for adding dynamic fields like `LastLedgerSequence` or `Fee`. It returns an error if the transaction is not valid or some internal call fails. There's also a `AutofillMultisigned` method that works the same way but for multisigned transactions.
Both methods support `Batch` transactions, filling in both the inner `RawTransactions` and the outer `Batch` transaction.

```go
func (c *Client) Autofill(tx *transaction.FlatTransaction) error
func (c *Client) AutofillMultisigned(tx *transaction.FlatTransaction, nSigners uint64) error
```

### Submit

The `SubmitTx` and `SubmitTxBlob` methods are used to submit a transaction to the XRPL network. They return a `SubmitResponse` struct containing the immediate submission result and status for the submitted blob or flattened transaction. The inputted transaction must be signed. There's also a `SubmitMultisigned` method that works the same way but for multisigned transactions.

```go
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.SubmitResponse, error)
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error)
func (c *Client) SubmitMultisigned(txBlob string, failHard bool) (*requests.SubmitMultisignedResponse, error)
```

### SubmitTxAndWait/SubmitTxBlobAndWait

The `SubmitTxAndWait` and `SubmitTxBlobAndWait` methods are used to submit a transaction to the XRPL network and wait for it to be included in a ledger. They return a `TxResponse` struct containing the finalized ledger transaction result for the flattened transaction or blob submitted.

```go
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error)
```

## Queries

`Client` also exposes methods to make queries to the XRPL network. These methods are wrappers of the queries requests exposed by the [`queries`](/docs/xrpl/queries) package.

## Usage

To use the `rpc` package, you need to import it in your project:

```go
import "github.com/Peersyst/xrpl-go/xrpl/rpc"
```

## Examples

### How to send a payment transaction

This example shows how to send a payment transaction to the XRPL testnet with the `rpc` package.

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {

	// Create a new rpc client config with a testnet faucet provider
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithMaxFeeXRP(5.0),
		rpc.WithFeeCushion(1.5),
		rpc.WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	if err != nil {
		panic(err)
	}

	// Create a new rpc client with the config
	client := rpc.NewClient(cfg)

	// Create a new wallet with the ed25519 algorithm
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fund the wallet with the testnet faucet
	if err := client.FundWallet(&w); err != nil {
		fmt.Println(err)
		return
	}

	// Convert the amount to drops
	xrpAmount, err := currency.XrpToDrops("1")
	if err != nil {
		fmt.Println(err)
		return
	}

	xrpAmountInt, err := strconv.ParseInt(xrpAmount, 10, 64)
	if err != nil {
		fmt.Println(err)
		return
	}

	p := &transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account: types.Address(w.GetAddress()),
		},
		Destination: "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		Amount:      types.XRPCurrencyAmount(xrpAmountInt),
		DeliverMax:  types.XRPCurrencyAmount(xrpAmountInt),
	}

	flattenedTx := p.Flatten()

	// Autofill the transaction with the client's config
	if err := client.Autofill(&flattenedTx); err != nil {
		fmt.Println(err)
		return
	}

	// Sign the transaction with the wallet
	txBlob, _, err := w.Sign(flattenedTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Submit the transaction to the network and wait for it to be included in a ledger
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println(err)
		return
	}
}
```
