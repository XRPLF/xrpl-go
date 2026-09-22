# Request and response types

The Go package is named `xrpl/queries`, but it does not send network requests itself. It defines the types you pass to client methods and the responses those methods return.

```text
request struct -> client method -> XRPL server -> response struct
```

Use [JSON-RPC](/docs/xrpl/rpc) or [WebSocket](/docs/xrpl/websocket) for the connection. The [XRPL API reference](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods) defines server methods, fields, and errors. This guide explains their Go usage.

## Make a typed query

With a connected or configured `client`, in a function that returns an error:

```go
response, err := client.GetAccountInfo(&account.InfoRequest{
	Account:     address,
	LedgerIndex: common.Validated,
})
if err != nil {
	return err
}
fmt.Println(response.AccountData.Balance)
```

Import `account` from `xrpl/queries/account` and `common` from `xrpl/queries/common`. `AccountData` is already an `AccountRoot` from [`xrpl/ledger-entry-types`](/docs/xrpl/ledger-entry-types). For a complete program, see [Read an account](/docs/xrpl/rpc#read-an-account).

Use [Ledger data](/docs/xrpl/ledger-entry-types) when you need generic `GetLedgerEntry()` results or binary responses instead of a dedicated typed query.

## Choose a ledger

Use `common.Validated` for validated state. When several requests must describe the same snapshot, take the ledger hash or index from the first response, where the response includes it, and pass it on subsequent requests. Do not keep asking for the latest validated ledger if the results must agree on one snapshot.

Ledger selectors vary by request type. See [ledger index](https://xrpl.org/docs/references/protocol/data-types/basic-data-types#ledger-index) and the relevant method's XRPL reference for server behavior.

## Read multiple pages

A `Marker` is an opaque continuation value. Pass it back unchanged with the same query parameters. Stop when the response has no marker.

This fragment reads trust lines through either client. It pins the ledger after the first response. Supply `address` as a `types.Address` and call it inside a function returning an error:

```go
request := account.LinesRequest{
	Account:     address,
	LedgerIndex: common.Validated,
}
for {
	response, err := client.GetAccountLines(&request)
	if err != nil {
		return err
	}
	for _, line := range response.Lines {
		fmt.Println(line.Currency, line.Balance)
	}
	if response.Marker == nil {
		break
	}
	if response.LedgerHash == "" {
		return fmt.Errorf("response omitted the ledger hash needed for pagination")
	}
	request.LedgerHash = response.LedgerHash
	request.LedgerIndex = nil
	request.Marker = response.Marker
}
```

The example follows [account_lines](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/account-methods/account_lines). Other methods can use different marker types or pagination rules. Do not decode or invent a marker.

## Find request types

The links below lead to the Go API, not a second copy of the server method catalogue.

| Package | Use it for |
| --- | --- |
| [account](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/account) | Account data, trust lines, objects, offers, and transaction history |
| [ledger](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/ledger) | Ledgers and individual ledger entries |
| [transactions](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/transactions) | Submission, transaction lookup, and simulation |
| [subscription](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/subscription) | WebSocket subscriptions |
| [path](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/path) | Paths, order books, and deposit authorization |
| [channel](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/channel) | Payment-channel signature verification |
| [nft](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/nft) | NFT offers |
| [amm](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/amm) | AMM information |
| [oracle](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/oracle) | Aggregated prices |
| [vault](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/vault) | Vault information |
| [clio](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/clio) | Clio-specific queries |
| [server](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/server) | Server state, fees, features, and definitions |
| [utility](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/utility) | Ping and random values |

## API versions and generic requests

The main request packages use API v2. Some packages also provide explicit v1 types, such as `xrpl/queries/account/v1`. The `amm`, `oracle`, `server`, and `vault` packages have no v1 types. Some requests have no v1 type even in packages with a `v1` directory, such as `account.GatewayBalancesRequest`, `ledger.EntryRequest`, and `transactions.SimulateRequest`. A v1 Go type is not interchangeable with the v2 type accepted by `GetAccountInfo`.

Use the client's lower-level `Request` method when you need an explicit version not exposed by its typed wrappers, or a request type without a typed wrapper, such as `transactions.TxRequest`. Its response form differs by transport. See the [RPC](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/rpc#Client.Request) or [WebSocket](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/websocket#Client.Request) API before decoding it.

SDK type availability, server API version support, and amendment activation are separate questions. A Go struct does not establish that a target network supports the operation.

## Simulate before submitting

Use `transactions.SimulateRequest` with exactly one of `TxJSON` or `TxBlob`. `Binary` requests hex transaction and metadata output. Supply an unsigned transaction. See [simulation](/docs/xrpl/submission#simulate) for the SDK's validation boundary and the server reference.

In a function returning an error, with `transactions` imported from `xrpl/queries/transactions` and `transaction` from `xrpl/transaction`:

```go
request := transactions.SimulateRequest{
	TxJSON: transaction.FlatTransaction{
		"TransactionType": "Payment",
		"Account":         sender,
		"Destination":     destination,
		"Amount":          "1000000",
	},
}

response, err := client.Simulate(&request)
if err != nil {
	return err
}
fmt.Println(response.EngineResult, response.EngineResultMessage)
```

`SimulateRequest.ValidateNetworkID` is deprecated and only performs the same nil-request check as `Validate`. Clients do not strip signatures, so never send signed input to an untrusted simulation endpoint.

## Cache server definitions

`GetServerDefinitions` accepts `&server.DefinitionsRequest{Hash: cachedHash}`. A matching hash allows a hash-only response. Retain the cached definitions when the response has no `Fields`.

A full response contains the five core sections. Servers implementing XLS-97 can also return transaction formats, ledger formats, and flag maps. The SDK rejects incomplete section groups and mismatched hash-only responses. See [server_definitions](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/server-info-methods/server_definitions) for the protocol response.

Fetching definitions does not replace the binary codec's embedded definitions automatically. See [Binary codec](/docs/binary-codec#protocol-definitions).
