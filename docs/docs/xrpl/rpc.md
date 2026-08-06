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

### HTTP client

The `WithHTTPClient` option sets the HTTP client that sends RPC requests.

```go
func WithHTTPClient(cl HTTPClient) ConfigOpt
```

### FaucetProvider

The `WithFaucetProvider` option sets the faucet provider of the RPC client. The predefined providers are `TestnetFaucetProvider` and `DevnetFaucetProvider`. You can also implement the `FaucetProvider` interface.

```go
func WithFaucetProvider(fp common.FaucetProvider) ConfigOpt
```

### Reliable-submission polling

`WithMaxRetries` limits consecutive incomplete monitoring rounds caused by query or transport errors. A complete round resets the count. It does not limit successful finality polling. `WithRetryDelay` sets the interval between polling rounds. The delay can be zero, but it must not be negative. A reliable-submission method returns `InvalidPollIntervalError` before it sends the transaction when the delay is negative.

```go
func WithMaxRetries(maxRetries int) ConfigOpt
func WithRetryDelay(retryDelay time.Duration) ConfigOpt
```

### MaxFeeXRP

The `WithMaxFeeXRP` option allows you to set the maximum fee in XRP that the client will use. Use a decimal string to preserve the exact limit.

```go
func WithMaxFeeXRP(maxFeeXRP string) ConfigOpt
```

### FeeCushion

The `WithFeeCushion` option allows you to set the fee cushion for a transaction.

```go
func WithFeeCushion(feeCushion float64) ConfigOpt
```

Fee calculation returns `ErrInvalidFeeValue` for a non-finite, negative, or malformed fee value. It returns `ErrFeeHasTooManyDecimals` when an XRP fee cannot be represented as a whole number of drops.

### MaxResponseSize

The `WithMaxResponseSize` option caps HTTP response bodies. The default is 64 MiB. Set it to `0` to disable the limit. A negative value restores the default.

```go
func WithMaxResponseSize(maxResponseSize int64) ConfigOpt
```

### Logger

The `SetLogger` function overrides the logger used for SDK warnings, such as remote non-TLS URL warnings. Pass `nil` to silence these warnings.

```go
func SetLogger(l *log.Logger)
```

### Network identity

By default, the client gets the network ID and rippled build version from `server_info` before it autofills or signs an unsigned transaction. The result is available in `Client.NetworkID` and `Client.BuildVersion`. A nil `NetworkID` means that the identity is not known. The client uses this identity to apply the correct `NetworkID` transaction policy.

`WithNetworkIdentity` bypasses discovery. Use it only when both values come from trusted deployment configuration.

```go
func WithNetworkIdentity(networkID uint32, buildVersion string) ConfigOpt
```

### Timeout

The `WithTimeout` option sets the RPC request timeout. It also updates the timeout of the default HTTP client.

```go
func WithTimeout(timeout time.Duration) ConfigOpt
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
func (c *Client) SubmitTx(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.SubmitResponse, error)
func (c *Client) SubmitTxBlob(txBlob string, failHard bool) (*requests.SubmitResponse, error)
func (c *Client) SubmitMultisigned(txBlob string, failHard bool) (*requests.SubmitMultisignedResponse, error)
```

### SubmitTxAndWait/SubmitTxBlobAndWait

The reliable-submission methods require `LastLedgerSequence` before they send the `submit` request. The Go SDK does not enable autofill by default. Provide `LastLedgerSequence` directly or set `Autofill: true` when you submit a transaction that the client can sign.

Only a preliminary `tem` result returns `PreliminaryResultError` immediately. The error includes the engine result and its message. The client monitors `tes`, `ter`, `tec`, `tef`, `tel`, and unknown preliminary results. An exact `txnNotFound` response is inconclusive and the client retries it.

Each polling round waits for the configured interval, requests the latest validated ledger, checks expiry, and then looks up the transaction. The transaction expires only when the validated ledger is strictly greater than `LastLedgerSequence`. Validation exactly at `LastLedgerSequence` is accepted. The expiry decision does not use bounded `tx` searches or `searched_all`.

```go
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxAndWaitContext(ctx context.Context, tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWaitContext(ctx context.Context, txBlob string, failHard bool) (*requests.TxResponse, error)
```

Every validated transaction response returns with a nil error, including validated `tec` results. Inspect `TxResponse.Meta.TransactionResult` to determine the validated engine result. `TransactionExpiredError` retains the preliminary engine result and ledger expiry details. `FinalityTransportError` reports repeated query or transport failure. Context-aware methods return `ctx.Err()` directly on cancellation or deadline.

The client verifies that each validated-ledger response is marked as validated and contains a ledger index. A negative polling interval returns `InvalidPollIntervalError` before submission.

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

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/currency"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {

	// Create a new rpc client config with a testnet faucet provider
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithMaxFeeXRP("5.0"),
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
	fmt.Println(res)
}
```
