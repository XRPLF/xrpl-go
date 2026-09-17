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

`WithMaxRetries` limits consecutive incomplete monitoring rounds caused by query or transport errors. A complete round resets the count. It does not limit successful finality polling. The value must be positive. A reliable-submission method returns `ErrInvalidMaxRetries` before it sends the transaction when the value is zero or negative. `WithRetryDelay` sets the interval between polling rounds. The delay can be zero, but it must not be negative. A reliable-submission method returns `ErrInvalidPollInterval` before it sends the transaction when the delay is negative.

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

By default, the client gets the network ID and rippled build version from
`server_info` before it autofills or signs an unsigned transaction. Use
`Client.NetworkIdentity()` to read the result. A nil returned network ID means
that the identity is not known. The client uses this identity to apply the
correct `NetworkID` transaction policy.

```go
networkID, buildVersion := client.NetworkIdentity()
```

`WithNetworkIdentity` bypasses discovery when `buildVersion` is non-empty. An
empty build version leaves the identity incomplete, so the client performs
discovery. Use trusted deployment configuration for both values.

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

Both methods support `Batch` transactions and fill the inner `RawTransactions` and the outer `Batch` transaction. They convert X-addresses in `Account`, `Destination`, `Authorize`, `Unauthorize`, `Owner`, `RegularKey`, `Delegate`, `NFTokenMinter`, `Subject`, `Issuer`, and `Holder` to classic addresses. Embedded tags in `Account` and `Destination` populate `SourceTag` and `DestinationTag`. A conflicting explicit tag returns `ErrMismatchedTag`.

```go
func (c *Client) Autofill(tx *transaction.FlatTransaction) error
func (c *Client) AutofillMultisigned(tx *transaction.FlatTransaction, nSigners uint64) error
```

Autofill sets `Fee` only when the transaction has none. Transaction types with a special cost carry it automatically, including the ten base fees a confidential MPT transaction owes, which also applies to a confidential transaction nested inside a `Batch`. A hand-set `Fee` is submitted unchanged, so a value sized for an ordinary transaction underpays and the submission fails with `telINSUF_FEE_P`. See the [confidential guide](/docs/confidential) for the full cost breakdown.

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

A missing `engine_result` or a preliminary `tem` result returns `ErrPreliminaryResult` immediately. The error message includes the engine result and its message. The client monitors `tes`, `ter`, `tec`, `tef`, `tel`, and non-empty unknown preliminary results. An exact `txnNotFound` response is inconclusive and the client retries it.

Each polling round waits for the configured interval, requests the latest validated ledger, and then looks up the transaction. The transaction expires only when the validated ledger is strictly greater than `LastLedgerSequence` and the final transaction lookup does not return a validated result. Validation exactly at `LastLedgerSequence` is accepted. The final lookup reduces a race with lagging read backends, but without `searched_all` it does not prove absence from history that the endpoint does not provide.

```go
func (c *Client) SubmitTxAndWait(tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxAndWaitContext(ctx context.Context, tx transaction.FlatTransaction, opts *rpctypes.SubmitOptions) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWait(txBlob string, failHard bool) (*requests.TxResponse, error)
func (c *Client) SubmitTxBlobAndWaitContext(ctx context.Context, txBlob string, failHard bool) (*requests.TxResponse, error)
```

Every validated transaction response returns with a nil error, including validated `tec` results. Inspect `TxResponse.Meta.TransactionResult` to determine the validated engine result. `ErrTransactionExpired` reports the preliminary engine result and ledger expiry details. `ErrFinalityTransport` reports repeated query or transport failure and wraps the last failure. Context-aware methods propagate caller cancellation through transaction preparation queries, submission, and finality monitoring, and return `ctx.Err()` directly on cancellation or deadline.

The client verifies that each validated-ledger response is marked as validated and contains a ledger index. A negative polling interval returns `ErrInvalidPollInterval` before submission. A zero or negative maximum retry value returns `ErrInvalidMaxRetries` before submission. A zero `LastLedgerSequence` returns `ErrInvalidLastLedgerSequence` before submission.

### Simulate

`Simulate` runs an XLS-69 dry run against the current open-ledger state. It sends JSON transaction input or a blob without local request preflight or network-identity discovery. The server validates the input. Nil-request protection and response validation remain enabled. It returns either decoded or binary transaction and metadata output. Supply an unsigned transaction, and do not send signed transactions to an untrusted node. A simulation does not guarantee the result of a later submission.

```go
func (c *Client) Simulate(req *transactions.SimulateRequest) (*transactions.SimulateResponse, error)
```

### Sponsorship preflight

`ValidateSponsorship` checks a sponsored transaction (XLS-68) against the `Sponsorship` ledger entry between its sponsor and its sponsee, reading the current ledger. The check is explicit and optional: autofill and submission never query sponsorship on their own.

The transaction must carry `Sponsor` and a valid `SponsorFlags`, and either a `Fee` or a nonempty `estimatedFee` in drops. Before any lookup, the sponsorship fields are checked with the same rules `BaseTx.Validate` applies, and a failure is returned as the matching `xrpl/transaction` error: for example `ErrSponsorFieldsMissing`, `ErrInvalidSponsorFlags`, `ErrSponsorAccountConflict`, `ErrSponsorDelegateConflict` for reserve sponsorship on a delegated transaction, `ErrReserveSponsorshipNotAllowed`, `ErrInnerBatchFeeSponsorship`, or `ErrInvalidSponsorSignature`. A present `SponsorSignature` must hold `SigningPubKey` with `TxnSignature` or a `Signers` array, except on an inner Batch transaction (`tfInnerBatchTxn`), where it may hold at most an empty `SigningPubKey` because the sponsor signs the outer `Batch`.

The sponsee is the transaction's `Delegate` when present and its `Account` otherwise. Only the sponsor and resolved sponsee addresses are converted to classic addresses for the lookup, because xrpld accepts only classic addresses in the `sponsorship` selector. Unrelated addresses, explicit tags, and Batch inner transactions are outside this lookup's validation scope. The caller's transaction is not modified.

Without a `Sponsorship` entry, only a sponsor co-signature (`SponsorSignature`) authorizes the sponsorship. With an entry, its budget always applies, even to a co-signed transaction, because rippled prefers the pre-funded fee payer whenever the entry exists: a sponsored fee must fit within `FeeAmount` and any `MaxFee` cap. Pre-funded use is additionally rejected when the entry sets `lsfSponsorshipRequireSignForFee` or `lsfSponsorshipRequireSignForReserve` for the requested sponsorship type. A zero fee draws nothing from the entry.

A nil error means the preflight completed; read `SponsorshipValidation.Valid` and `SponsorshipValidation.Reason`, which wraps an `ErrSponsorship*` sentinel. A non-nil error means the preflight could not run, because the transaction inputs were unusable or the `ledger_entry` lookup failed. Only `entryNotFound` counts as an absent entry; transport, permission, and decoding failures, and an entry whose `Owner` or `Sponsee` differs from the requested pair (`ErrSponsorshipEntryMismatch`), are returned as errors.

The check does not confirm that the sponsor account exists, which rippled requires even for a co-signed transaction, or that a co-signing sponsor's balance minus its reserve covers the fee when no entry pays it. It does not check that the fee includes the extra base fee rippled charges for each `SponsorSignature.Signers` entry, nor, for a co-signed inner Batch transaction, that the sponsor also signs the outer `Batch` through `BatchSigners`, which the inner transaction alone cannot show. It does not check `RemainingOwnerCount` either, because the client cannot know how many reserved objects the transaction creates; the entry is returned in `SponsorshipValidation.Sponsorship` for callers that do. rippled remains authoritative.

```go
func (c *Client) ValidateSponsorship(tx transaction.FlatTransaction, estimatedFee string) (SponsorshipValidation, error)
func (c *Client) ValidateSponsorshipContext(ctx context.Context, tx transaction.FlatTransaction, estimatedFee string) (SponsorshipValidation, error)
```

### Server definitions

`GetServerDefinitions` retrieves the server protocol definitions. Set `DefinitionsRequest.Hash` to a cached hash to allow a hash-only unchanged response.

```go
func (c *Client) GetServerDefinitions(req *server.DefinitionsRequest) (*server.DefinitionsResponse, error)
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
