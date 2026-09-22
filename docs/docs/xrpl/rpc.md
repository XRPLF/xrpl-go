# JSON-RPC

Use the JSON-RPC client for HTTP queries and transaction submission. Choose [WebSocket](/docs/xrpl/websocket) if you also need live subscriptions.

[Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/rpc) · [Server API reference](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods)

## Read an account

After [installing the SDK](/docs/installation), save this as `main.go` and run `go run .`. It reads a Testnet account from a validated ledger. No wallet or signing key is needed.

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Peersyst/xrpl-go/xrpl/queries/account"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func main() {
	cfg, err := rpc.NewClientConfig(
		"https://s.altnet.rippletest.net:51234/",
		rpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClient(cfg)

	response, err := client.GetAccountInfo(&account.InfoRequest{
		Account:     "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		LedgerIndex: common.Validated,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Balance in drops:", response.AccountData.Balance)
	fmt.Println("Validated:", response.Validated)
}
```

Public test networks can reset. An account lookup can fail if the sample account does not exist. Substitute a funded Testnet address, or use the server-info query in [Getting started](/docs/intro), which does not depend on an account.

The client decodes the response into Go types. You do not need to decode JSON manually. See [Request and response types](/docs/xrpl/queries) for ledger selection, pagination, and generic requests.

## Configure the client

Pass options to `NewClientConfig`, then pass the configuration to `NewClient`.

| Option | Purpose |
| --- | --- |
| `WithTimeout` | Sets the timeout for one request attempt and the `*http.Client` timeout. A custom `*http.Client` with its own non-zero timeout takes precedence only when `WithTimeout` comes before `WithHTTPClient`. Otherwise, `WithTimeout` overwrites the custom client's timeout |
| `WithHTTPClient` | Supplies a custom `HTTPClient` implementation |
| `WithFaucetProvider` | Enables [test funding](/docs/xrpl/faucet) through the client |
| `WithMaxResponseSize` | Caps HTTP response bodies, default 64 MiB |
| `WithMaxFeeXRP` | Sets an exact maximum fee as an XRP decimal string, such as `"2"` |
| `WithFeeCushion` | Sets the fee multiplier, such as `1.2` |
| `WithMaxRetries`, `WithRetryDelay` | Configure [reliable-submission monitoring](/docs/xrpl/submission#wait-for-validation). `WithRetryDelay` also sets the first HTTP 503 backoff |

For response limits, `0` disables the cap and a negative value restores the default. Do not disable the cap unless your application needs it.

The client retries a request that receives HTTP 503 up to 3 times. The backoff starts at `WithRetryDelay` and doubles after each attempt. `WithMaxRetries` does not change this count, and `WithTimeout` bounds each attempt rather than the whole retry window. After the last 503, the request returns a `ClientError` reporting an overloaded server.

Fee calculation returns `ErrInvalidFeeValue` for a non-finite, negative, or malformed fee value, and `ErrFeeHasTooManyDecimals` when an XRP fee cannot be represented as whole drops.

Use HTTPS for remote endpoints. The client rejects authorization over plaintext transport and authenticated HTTPS-to-HTTP redirects. A nil custom HTTP client returns `ErrNilHTTPClient`. A custom `*http.Client` keeps the redirect check. Any other `HTTPClient` implementation controls its own redirects and credentials.

`rpc.SetLogger` replaces the logger for SDK warnings, including remote non-TLS URL warnings. Passing nil silences those warnings.

## Network identity

Before autofill or client-side signing, the client discovers the network ID and server build version through `server_info`. Read the current result with `client.NetworkIdentity()`. A nil network ID means discovery is incomplete.

For a trusted deployment, `rpc.WithNetworkIdentity(networkID, buildVersion)` can supply the identity. Discovery is bypassed only when the build version is non-empty. Do not use guessed values: the SDK uses them to apply the transaction `NetworkID` policy.

## Send a transaction

Start with [Transactions](/docs/xrpl/transaction) for a complete Testnet payment. [Submission and finality](/docs/xrpl/submission) documents autofill, signing, simulation, cancellation, and result handling for both clients.

Use [Sponsorship](/docs/xrpl/sponsorship) for sponsor signing and the optional `ValidateSponsorship` preflight.
