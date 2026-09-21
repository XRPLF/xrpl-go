# Getting started

Build Go applications on the [XRP Ledger](https://xrpl.org/). Query the ledger, manage wallets, and sign and submit transactions with typed Go models.

## Quick start

Requires **Go 1.25.13 or later**. The core SDK is pure Go and does not require a C/C++ toolchain.

From your application's Go module, install the SDK:

```bash
go get github.com/Peersyst/xrpl-go@latest
```

For a new application, run `go mod init example.com/my-xrpl-app` first. See [Installation](/docs/installation) for more setup details.

Save this as `main.go` to query a Testnet server. No wallet or funds are needed.

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/xrpl/queries/server"
	"github.com/Peersyst/xrpl-go/xrpl/rpc"
)

func main() {
	cfg, err := rpc.NewClientConfig("https://s.altnet.rippletest.net:51234/")
	if err != nil {
		log.Fatal(err)
	}
	client := rpc.NewClient(cfg)

	info, err := client.GetServerInfo(&server.InfoRequest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Connected to Testnet. Server version: %s\n", info.Info.BuildVersion)
}
```

Resolve its dependencies and run it:

```bash
go mod tidy
go run .
```

With network access, it prints the server's software version. You have now made your first XRPL request.

## What next?

- **Query ledger data** with the [JSON-RPC client](/docs/xrpl/rpc), or subscribe to live updates with [WebSocket](/docs/xrpl/websocket).
- **Create a wallet and sign locally** with the [wallet guide](/docs/xrpl/wallet).
- **Build and send transactions** with [typed transaction models](/docs/xrpl/transaction) and the complete [Testnet payment example](https://github.com/XRPLF/xrpl-go/tree/main/examples/send-xrp/rpc).
- **Explore more use cases** in the [examples directory](https://github.com/XRPLF/xrpl-go/tree/main/examples), or look up types and methods in the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go).

## Optional confidential helpers

Most applications only need core. Add the separate confidential module when you need confidential MPT builders, encryption, balance decryption, or proofs. Its cryptographic operations require cgo and a supported native toolchain. Start with the [confidential guide](/docs/confidential) and [installation requirements](/docs/confidential/installation).

## Security

The signing functionality has **not been independently audited**. Test with non-production funds on Testnet or Devnet and review signing behavior before production use.

Never print, log, commit, or send real seeds, private keys, or mnemonics to telemetry. Protect confidential encryption private keys as well as wallet signing keys.
