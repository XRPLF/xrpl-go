# websocket

## Overview

The `websocket` package provides a WebSocket client for interacting with the XRPL network via its WebSocket API. This client handles the communication with XRPL nodes, allowing you to:

- Send requests to query the ledger state.
- Submit transactions to the network.
- Receive responses and handle errors.
- Manage the connections configuration.

## Config

The `websocket` package provides a `Config` struct that allows you to configure the WebSocket client. Every time you create a new `Client`, you need to pass a `Config` struct as an argument. You can initialize a `Config` struct using the `NewClientConfig` function.

`Config` struct follows the options pattern, so you can pass different options to the `NewClientConfig` function:

### Host

The `WithHost` option allows you to set the host of the WebSocket client.

```go
func (wc ClientConfig) WithHost(host string) ClientConfig
```

### FaucetProvider

The `FaucetProvider` option allows you to set the faucet provider of the WebSocket client. There're two predefined faucet providers: `TestnetFaucetProvider` and `DevnetFaucetProvider`. You can also implement your own faucet provider by implementing the `FaucetProvider` interface.

```go
func (wc ClientConfig) WithFaucetProvider(fp common.FaucetProvider) ClientConfig
```

### MaxRetries

The `WithMaxRetries` option allows you to set the maximum number of retries for a transaction.

```go
func (wc ClientConfig) WithMaxRetries(maxRetries int) ClientConfig
```

### RetryDelay

The `WithRetryDelay` option allows you to set the delay between retries for a transaction.

```go
func (wc ClientConfig) WithRetryDelay(retryDelay time.Duration) ClientConfig
```

### FeeCushion

The `WithFeeCushion` option allows you to set the fee cushion for a transaction.

```go
func (wc ClientConfig) WithFeeCushion(feeCushion float32) ClientConfig
```

### MaxFeeXRP

The `WithMaxFeeXRP` option allows you to set the maximum fee in XRP that the WebSocket client will use.

```go
func (wc ClientConfig) WithMaxFeeXRP(maxFeeXrp float32) ClientConfig
```

### MaxResponseSize

The `WithMaxResponseSize` option caps inbound WebSocket messages. The default is 16 MiB per message. Set it to `0` to disable the limit.

```go
func (wc ClientConfig) WithMaxResponseSize(maxResponseSize int64) ClientConfig
```

### Logger

The `SetLogger` function overrides the logger used for SDK warnings, such as remote non-TLS URL warnings. Pass `nil` to silence these warnings.

```go
func SetLogger(l *log.Logger)
```

## Connection

As the `websocket` package is a WebSocket client, it needs to be connected to a WebSocket server. The `Client` type exposes the following methods to connect to a WebSocket server:

```go
// Connection methods
func (c *Client) Connect() error
func (c *Client) Disconnect() error

// Connection status
func (c *Client) IsConnected() bool

// Connection
func (c *Client) Conn() *websocket.Conn
```

So, for example, if you want to connect to the `devnet` ledger, you can do it this way:

```go
client := websocket.NewClient(websocket.NewClientConfig().WithHost("wss://s.altnet.rippletest.net:51233"))
defer client.Disconnect()

err := client.Connect()
if err != nil {
    // ...
}

if !client.IsConnected() {
    // ...
}
```

## Methods

The `Client` type exposes the following methods to interact with the XRPL network:

### Request

The `Request` method is used to send a request to the server and returns the response. This method is mostly used to send client [`queries`](/docs/xrpl/queries) to the server.

```go
func (c *Client) Request(reqParams XRPLRequest) (*ClientResponse, error)
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

The `websocket` package provides query wrappers that allows you to send client [`queries`](/docs/xrpl/queries) to the server.

## Examples

### How to send a payment transaction

This example shows how to send a payment transaction to the XRPL testnet with the `websocket` package.

```go
package main

import (
	"fmt"
	"strconv"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {

	// Create a new websocket client with a testnet faucet provider
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer client.Disconnect()

	// Connect to the testnet
	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	// Check if the client is connected
	if !client.IsConnected() {
		return
	}

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

	// Submit the transaction to the network and wait for it to be included in a ledge
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println(err)
		return
	}
}

```
