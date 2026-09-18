# XRPL clients, wallets, and transactions

Use these packages to query the XRP Ledger, subscribe to ledger events, and build, sign, and submit transactions.

[Installation and quick start](../README.md#quick-start) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl) · [Examples](../examples)

## Choose a client

Both clients support ledger queries, transaction autofill, and submission. Choose based on how your application uses the network:

| Client | Use it when |
| --- | --- |
| [`rpc`](rpc) | You need HTTP requests for queries or transaction submission. |
| [`websocket`](websocket) | You also need a persistent connection and event subscriptions. |

Start with the [JSON-RPC example](../README.md#quick-start) for a single request. For live ledger updates, use the WebSocket example below.

## Read and write paths

**Ledger entries describe stored state. Transactions request changes to that state.** Use either client for both paths.

```text
READ: typed account query
  client.GetAccountInfo() -> ledger -> response.AccountData (AccountRoot)

READ: generic ledger entry query
  queries/ledger.EntryRequest -> client.GetLedgerEntry() -> ledger
    response -> EntryResponse
      Node (JSON fields) -----------------------> map
      NodeBinary (hex) -> binarycodec.Decode() --> map
    map -> JSON marshal/unmarshal -> ledger-entry-types struct

WRITE: submit a transaction
  transaction struct -> Flatten() -> client.Autofill()
    -> wallet.Sign() [signing payload -> signature -> encoded blob]
    -> client.SubmitTxBlobAndWait() -> validated result + metadata
    -> query the ledger again to read the resulting state
```

**Typed queries decode for you.** For example, `GetAccountInfo()` returns `AccountData` as a [`ledger.AccountRoot`](ledger-entry-types/account_root.go). No manual conversion is needed.

**`GetLedgerEntry()` is generic.** Its `EntryResponse.Node` is a `FlatLedgerObject` map, even when you request an AccountRoot. You can use the map directly or select the matching [`ledger-entry-types`](ledger-entry-types) struct from `LedgerEntryType`, then marshal the map and unmarshal into it. With `Binary: true`, decode `NodeBinary` with `binarycodec.Decode()` first to obtain the map.

Set the request's `LedgerIndex` to `"validated"` when you need validated state.

For writes, `wallet.Sign()` uses `EncodeForSigning()` to prepare the signing payload and `Encode()` to produce the signed blob. Check the transaction result, not just whether it was validated: a validated transaction can have failed.

## Subscribe to ledger updates

This example connects to Testnet and prints closed ledger indexes for 30 seconds, then disconnects. It needs network access, but no wallet or funds.

After [installing the SDK](../README.md#quick-start), save this as `main.go` in your application and run `go run .`:

```go
package main

import (
	"fmt"
	"log"
	"time"

	subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233"),
	)
	if err := client.Connect(); err != nil {
		return err
	}
	defer func() {
		if err := client.Disconnect(); err != nil {
			log.Printf("disconnect: %v", err)
		}
	}()

	client.OnError(func(err error) {
		log.Printf("stream: %v", err)
	})
	client.OnLedgerClosed(func(ledger *streamtypes.LedgerStream) {
		fmt.Printf("Ledger closed: %d\n", ledger.LedgerIndex)
	})

	if _, err := client.Subscribe(&subscribe.Request{
		Streams: []string{"ledger"},
	}); err != nil {
		return err
	}

	time.Sleep(30 * time.Second)
	return nil
}
```

Register handlers before subscribing so they are ready for incoming events. Keep handlers short: a slow handler can delay other messages on the connection.

For more subscription options, see the [WebSocket guide](https://xrplf.github.io/xrpl-go/docs/xrpl/websocket) and [subscription example](../examples/subscription/ws).

## Build and send transactions

Use [`transaction`](transaction) for typed transaction models and [`wallet`](wallet) for local signing, following the [write path above](#read-and-write-paths). Autofill sets network fields such as the fee and sequence. Signing returns the signed blob and transaction hash.

Both clients also provide `SubmitTxAndWait()` to autofill, sign, submit, and wait in one call. For offline signing, prepare the network fields first. Signing itself does not need a network connection.

Start with a complete payment example for [JSON-RPC](../examples/send-xrp/rpc) or [WebSocket](../examples/send-xrp/ws). See the [transaction guide](https://xrplf.github.io/xrpl-go/docs/xrpl/transaction) for more transaction types.

## Wallets and multisigning

The [`wallet`](wallet) package creates wallets and derives them from seeds or mnemonics. Use the [wallet example](../examples/wallet) and [wallet guide](https://xrplf.github.io/xrpl-go/docs/xrpl/wallet) to get started.

For multisigning, each wallet produces a signed blob with `Multisign()`. The top-level `xrpl.Multisign()` function combines those blobs. See the complete [multisigning example](../examples/multisigning/rpc).

**Protect wallet secrets.** Never print, log, or commit real seeds, private keys, or mnemonics. Test with non-production funds and read the [security and audit notice](../README.md#security-and-audits) before production use.

## Find a package

| Package | Purpose |
| --- | --- |
| [`queries`](queries) | Request and response types for ledger APIs |
| [`ledger-entry-types`](ledger-entry-types) | Typed ledger objects |
| [`transaction/types`](transaction/types) | Amounts, addresses, and other transaction field types |
| [`currency`](currency) | Convert XRP and drops |
| [`hash`](hash) | Compute XRPL hashes |
| [`time`](time) | Convert XRPL timestamps |

For low-level encoding and key operations, see the [address codec](../address-codec), [binary codec](../binary-codec), and [keypairs](../keypairs) packages.
