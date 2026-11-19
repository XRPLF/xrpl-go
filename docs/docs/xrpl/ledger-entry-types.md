# Ledger data

Ledger entries describe stored state. Transactions request changes to that state. The `xrpl/ledger-entry-types` package provides Go structs for ledger entries. Its Go package name is `ledger`, so use an import alias when also importing ledger query types.

[Go ledger types](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types) · [XRPL ledger-entry reference](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types)

## Typed and generic reads

| Query | Result | What you do |
| --- | --- | --- |
| `GetAccountInfo` | `AccountData` is a `ledger.AccountRoot` | Access typed fields directly |
| `GetLedgerEntry`, JSON | `Node` is a `FlatLedgerObject` map | Read the map, or decode it into the appropriate struct |
| `GetLedgerEntry`, binary | `NodeBinary` is a hex string | Call `binarycodec.Decode`, then use the resulting map |

```text
specific query -> typed response
ledger_entry JSON -> map -> optional struct conversion
ledger_entry binary -> binarycodec.Decode -> map -> optional struct conversion
```

Do not manually convert a response that is already typed. See the [account query example](/docs/xrpl/rpc#read-an-account).

## Read a generic ledger entry

This complete Testnet example requests an AccountRoot and converts the generic response to the corresponding struct. Save it as `main.go` and run `go run .`. Use a funded Testnet address if the example account is unavailable after a network reset.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	ledger "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgerquery "github.com/Peersyst/xrpl-go/xrpl/queries/ledger"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func main() {
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClient(cfg)
	response, err := client.GetLedgerEntry(&ledgerquery.EntryRequest{
		AccountRoot: "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		LedgerIndex: common.Validated,
	})
	if err != nil {
		log.Fatal(err)
	}
	if response.Node == nil {
		log.Fatal("response contains no JSON ledger entry")
	}
	data, err := json.Marshal(response.Node)
	if err != nil {
		log.Fatal(err)
	}
	var account ledger.AccountRoot
	if err := json.Unmarshal(data, &account); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Balance in drops:", account.Balance)
}
```

The request selects an AccountRoot, so the example knows which struct to use. For an arbitrary index, inspect `LedgerEntryType` before choosing a struct. Do not decode every node into `AccountRoot`.

## Select an entry

`EntryRequest` requires exactly one top-level selector. Use `Index` for a known ledger-entry index or a supported typed selector such as `AccountRoot` or `MPToken`. Zero or multiple selectors and invalid object-selector forms fail validation.

See [ledger_entry](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/ledger-methods/ledger_entry) for protocol selectors and the [Go EntryRequest](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/queries/ledger#EntryRequest) for their SDK forms. [Hash helpers](/docs/xrpl/hash) can derive some indexes offline.

## Binary and deleted entries

With `Binary: true`, decode `response.NodeBinary` using `binarycodec.Decode`. Its returned map can be converted through JSON in the same way as `Node`. See [Binary codec](/docs/binary-codec).

Clio deleted-entry responses can also include `DeletedLedgerIndex` and `LedgerHash`. JSON and binary response forms remain separate. Do not treat the absence of `Node` alone as proof that an entry does not exist.

## Field representation

- MPT ledger amounts are quoted base-10 strings, and `OwnerNode` fields are hexadecimal strings. See [MPT ledger values](/docs/xrpl/mpt#ledger-values).
- `PriceData.AssetPrice` is a pointer so absent and explicit zero prices remain distinct. Use `ledger.AssetPrice(value)` to construct it. Oracle `Scale` supports `0` through `20`. `Flatten` omits `Scale` when the price is absent, and `Validate` rejects a nonzero scale without a price.
- `MPTokenIssuance` holds issuer/auditor encryption keys and `ConfidentialOutstandingAmount`. `MPToken` holds the holder key, spending/inbox ciphertexts, issuer/auditor mirror ciphertexts, and `ConfidentialBalanceVersion`. Use [confidential builders](/docs/confidential/builders) to work with this state.
- `LsfMPTCanHoldConfidentialBalance` describes the issuance capability. `LsifMPTCanHoldConfidentialBalance` on `ImmutableFlags` describes its immutability restriction.

Keep protocol field descriptions in the XRPL reference. The Go API reference documents available structs and their field types.
