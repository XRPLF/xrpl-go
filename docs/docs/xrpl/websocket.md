# WebSocket

Use the WebSocket client for live ledger updates as well as queries and transaction submission. Unlike [JSON-RPC](/docs/xrpl/rpc), it maintains a connection that you must open and close.

[Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/websocket) · [XRPL subscribe API](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/subscription-methods/subscribe)

## Subscribe to ledger updates

After [installing the SDK](/docs/installation), save this as `main.go` and run `go run .`. It prints closed ledger indexes from Testnet for 30 seconds, then disconnects. No wallet or funds are needed.

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

Register handlers before subscribing. Keep handlers short so they do not delay message processing. This bounded example does not implement recovery after a disconnect.

## Connection lifecycle

```text
configure -> Connect -> register handlers -> Subscribe -> receive events -> Disconnect
```

- `Connect` discovers network identity before exposing the connection to requests. Calling it while connected or connecting returns `ErrAlreadyConnected`.
- `Disconnect` succeeds when no connection is active. Pending requests return `ErrDisconnected` after a manual or unexpected disconnect.
- Requests are **not replayed** after reconnection. A failed transaction request may still have reached the server. See [submission recovery](/docs/xrpl/submission#check-the-final-result).
- Automatic reconnect does **not** replay subscriptions. Your application must subscribe again after reconnecting. Use a ledger query to recover any missed state rather than assuming the stream is continuous.
- Serialize calls to `Connect` and `Disconnect`. Do not call `Connect` synchronously from a stream or error handler.

A write or write-deadline failure closes the failed socket. The active read loop can reconnect before a later request. `WithMaxReconnects` limits reconnection attempts, and receiving a message successfully resets that count. `ErrMaxReconnectionAttemptsReached` unwraps the last connection error. It is delivered only to the `OnError` handler, and the client stops reconnecting after it.

## Handler behavior

Register one handler per stream with the appropriate `OnXxx` method. A later registration replaces the previous handler. An event already queued can still use the old handler.

| Subscription | Handler |
| --- | --- |
| `ledger` | `OnLedgerClosed` |
| `transactions`, `transactions_proposed`, `accounts`, `accounts_proposed`, `books` | `OnTransactions` |
| `validations` | `OnValidationReceived` |
| `peer_status` | `OnPeerStatusChange` |
| `consensus` | `OnConsensusPhase` |
| `book_changes` | `OnBookChanges` |

Order-book subscriptions arrive as transaction messages, so handle them with `OnTransactions`. The client does not call `OnOrderBook` for server messages. The `server` and `manifests` streams have no handler, and their messages reach `OnError` as `ErrUnknownStreamType`.

A handler is serialized with itself, even across disconnect/reconnect lifecycles. Different handlers can run concurrently, so protect application state shared between them. Delivery is unbuffered within each stream and applies backpressure.

Typed `bookChanges` updates preserve `validated`, `domain`, `mpt_issuance_id_a`, and `mpt_issuance_id_b`. See the [subscription types](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/subscription/types) for notification fields.

## Configuration

Chain options on `websocket.NewClientConfig()` before calling `NewClient`.

| Option | Purpose |
| --- | --- |
| `WithHost` | Sets the `wss://` endpoint |
| `WithTimeout` | Limits a complete request, including its write and response wait |
| `WithMaxReconnects` | Limits reconnection attempts after read failures |
| `WithMaxResponseSize` | Caps each inbound message, default 16 MiB |
| `WithFaucetProvider` | Enables [test funding](/docs/xrpl/faucet) |
| `WithMaxFeeXRP`, `WithFeeCushion` | Set an exact XRP fee limit string and a fee multiplier |
| `WithMaxRetries`, `WithRetryDelay` | Configure [reliable-submission monitoring](/docs/xrpl/submission#wait-for-validation), not connection retries |

For response limits, `0` disables the cap and a negative value restores the default. Fee calculation returns `ErrInvalidFeeValue` for non-finite, negative, or malformed values, and `ErrFeeHasTooManyDecimals` for XRP fees that cannot be represented as whole drops.

`websocket.SetLogger` replaces the SDK warning logger. Passing nil silences warnings, including remote non-TLS URL warnings.

## Network identity

`Connect` gets the network ID and build version from `server_info`. `client.NetworkIdentity()` returns a snapshot. A nil network ID means it is not known.

`WithNetworkIdentity(networkID, buildVersion)` bypasses discovery only when the build version is non-empty. Both values must come from trusted deployment configuration.

Every reconnect repeats discovery unless an identity override is configured. A reconnect rejects a network ID different from the previously discovered one.

## Queries and transactions

Use the same [request and response types](/docs/xrpl/queries) as JSON-RPC. The clients share [submission and finality](/docs/xrpl/submission) behavior, but WebSocket submission options come from `xrpl/websocket/types`.

For sponsor signing and optional preflight, see [Sponsorship](/docs/xrpl/sponsorship).
