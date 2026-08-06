# websocket

## Overview

The `websocket` package provides a WebSocket client for interacting with the XRPL network via its WebSocket API. This client handles the communication with XRPL nodes, allowing you to:

- Send requests to query the ledger state.
- Submit transactions to the network.
- Receive responses and handle errors.
- Manage the connections configuration.

## Config

The `websocket` package provides a `ClientConfig` struct that allows you to configure the WebSocket client. Every time you create a new `Client`, you need to pass a `ClientConfig` value as an argument. You can initialize it with the `NewClientConfig` function.

`ClientConfig` follows the options pattern, so you can apply different options before you create the client:

### Host

The `WithHost` option allows you to set the host of the WebSocket client.

```go
func (wc ClientConfig) WithHost(host string) ClientConfig
```

### FaucetProvider

The `WithFaucetProvider` option sets the faucet provider of the WebSocket client. The predefined providers are `TestnetFaucetProvider` and `DevnetFaucetProvider`. You can also implement the `FaucetProvider` interface.

```go
func (wc ClientConfig) WithFaucetProvider(fp common.FaucetProvider) ClientConfig
```

### MaxRetries

`WithMaxRetries` limits consecutive incomplete reliable-submission monitoring rounds caused by query or transport errors. A complete round resets the count. It does not limit successful finality polling.

```go
func (wc ClientConfig) WithMaxRetries(maxRetries int) ClientConfig
```

### MaxReconnects

`WithMaxReconnects` limits reconnection attempts after a WebSocket read error. The reconnect count resets after the client receives a message successfully.

```go
func (wc ClientConfig) WithMaxReconnects(maxReconnects int) ClientConfig
```

### RetryDelay

The `WithRetryDelay` option sets the delay between reliable-submission polling rounds. The delay can be zero, but it must not be negative. A reliable-submission method returns `InvalidPollIntervalError` before it sends the transaction when the delay is negative.

```go
func (wc ClientConfig) WithRetryDelay(retryDelay time.Duration) ClientConfig
```

### FeeCushion

The `WithFeeCushion` option allows you to set the fee cushion for a transaction.

```go
func (wc ClientConfig) WithFeeCushion(feeCushion float64) ClientConfig
```

### MaxFeeXRP

The `WithMaxFeeXRP` option allows you to set the maximum fee in XRP that the WebSocket client will use. Use a decimal string to preserve the exact limit.

```go
func (wc ClientConfig) WithMaxFeeXRP(maxFeeXRP string) ClientConfig
```

Fee calculation returns `ErrInvalidFeeValue` for a non-finite, negative, or malformed fee value. It returns `ErrFeeHasTooManyDecimals` when an XRP fee cannot be represented as a whole number of drops.

### MaxResponseSize

The `WithMaxResponseSize` option caps inbound WebSocket messages. The default is 16 MiB per message. Set it to `0` to disable the limit. A negative value restores the default.

```go
func (wc ClientConfig) WithMaxResponseSize(maxResponseSize int64) ClientConfig
```

### Logger

The `SetLogger` function overrides the logger used for SDK warnings, such as remote non-TLS URL warnings. Pass `nil` to silence these warnings.

```go
func SetLogger(l *log.Logger)
```

### Network identity

By default, `Connect` gets the network ID and rippled build version from `server_info` before it makes the connection available to requests. The result is available in `Client.NetworkID` and `Client.BuildVersion`. A nil `NetworkID` means that the identity is not known. The client uses this identity to apply the correct `NetworkID` transaction policy.

`WithNetworkIdentity` bypasses discovery. Use it only when both values come from trusted deployment configuration.

```go
func (wc ClientConfig) WithNetworkIdentity(networkID uint32, buildVersion string) ClientConfig
```

### Timeout

The `WithTimeout` option sets the timeout for a complete request, including the write and the response wait.

```go
func (wc ClientConfig) WithTimeout(timeout time.Duration) ClientConfig
```

## Connection

As the `websocket` package is a WebSocket client, it needs to be connected to a WebSocket server. Pending requests return `ErrDisconnected` as soon as a manual or unexpected disconnect occurs. The client does not replay these requests after reconnection. Calling `Disconnect` when no connection is active succeeds without an error.

Callers must serialize concurrent `Connect` and `Disconnect` calls. Do not call `Connect` synchronously from a stream or error handler. The `Client` type exposes the following connection methods:

```go
// Connection methods
func (c *Client) Connect() error
func (c *Client) Disconnect() error

// Connection status
func (c *Client) IsConnected() bool
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
func (c *Client) Request(reqParams interfaces.Request) (*ClientResponse, error)
```

### Autofill/AutofillMultisigned

The `Autofill` method is used to autofill fields in a flat transaction. This method adds dynamic fields such as `LastLedgerSequence` and `Fee`, and it applies the network `NetworkID` policy. It returns an error if the transaction is not valid or an internal request fails. The `AutofillMultisigned` method provides the same behavior for multisigned transactions.

Both methods support `Batch` transactions and fill the inner `RawTransactions` and the outer `Batch` transaction. They convert X-addresses in `Account`, `Destination`, `Authorize`, `Unauthorize`, `Owner`, and `RegularKey` to classic addresses. Embedded tags in `Account` and `Destination` populate `SourceTag` and `DestinationTag`. A conflicting explicit tag returns `ErrMismatchedTag`.

```go
func (c *Client) Autofill(tx *transaction.FlatTransaction) error
func (c *Client) AutofillMultisigned(tx *transaction.FlatTransaction, nSigners uint64) error
```

### Submit

The `SubmitTx` and `SubmitTxBlob` methods submit a transaction to the XRPL network. They return a `SubmitResponse` with the immediate submission result. `SubmitTxBlob` requires a signed transaction blob. `SubmitTx` accepts a signed flat transaction, or it can sign an unsigned transaction when `SubmitOptions.Wallet` is set. It enables autofill only when `SubmitOptions.Autofill` is true.

Submission rejects incomplete, empty, or mixed signing fields before it sends the request. `SubmitMultisigned` requires a structurally complete multisigned transaction blob.

```go
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.SubmitResponse, error)
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error)
func (c *Client) SubmitMultisigned(txBlob string, failHard bool) (*requests.SubmitMultisignedResponse, error)
```

### SubmitTxAndWait/SubmitTxBlobAndWait

The reliable-submission methods require `LastLedgerSequence` before they send the `submit` request. The Go SDK does not enable autofill by default. Provide `LastLedgerSequence` directly or set `Autofill: true` when you submit a transaction that the client can sign.

Only a preliminary `tem` result returns `PreliminaryResultError` immediately. The error includes the engine result and its message. The client monitors `tes`, `ter`, `tec`, `tef`, `tel`, and unknown preliminary results. An exact `txnNotFound` response is inconclusive and the client retries it.

Each polling round waits for the configured interval, requests the latest validated ledger, checks expiry, and then looks up the transaction. The transaction expires only when the validated ledger is strictly greater than `LastLedgerSequence`. Validation exactly at `LastLedgerSequence` is accepted. The expiry decision does not use bounded `tx` searches or `searched_all`.

```go
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxAndWaitContext(ctx context.Context, tx transaction.FlatTransaction, opts *wstypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWaitContext(ctx context.Context, txBlob string, failHard bool) (*requests.TxResponse, error)
```

Every validated transaction response returns with a nil error, including validated `tec` results. Inspect `TxResponse.Meta.TransactionResult` to determine the validated engine result. `TransactionExpiredError` retains the preliminary engine result and ledger expiry details. `FinalityTransportError` reports repeated query or transport failure. Context-aware methods return `ctx.Err()` directly on cancellation or deadline.

The client verifies that each validated-ledger response is marked as validated and contains a ledger index. A negative polling interval returns `InvalidPollIntervalError` before submission.

A WebSocket write or write-deadline failure invalidates and closes the failed socket. The active client read loop can then reconnect before a later request.

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

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
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

	// Submit the transaction to the network and wait for it to be included in a ledger
	res, err := client.SubmitTxBlobAndWait(txBlob, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}
```
