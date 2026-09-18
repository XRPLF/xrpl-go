# Installation

Install the core SDK for transactions, queries, and wallet signing. Add the optional confidential module only when you need its builders or cryptographic helpers.

## Prerequisites

Before installing, you need to have Go installed on your machine. You can download it from the [official website](https://go.dev/doc/install).

Both modules require the following Go version:

| Software | Version |
|------------|---------|
| Go | >= 1.25.13 |

## Install the core module

Run this command in your application directory, which must contain a `go.mod` file. For a new application, create one with `go mod init <your-module-path>` first.

```bash
go get github.com/Peersyst/xrpl-go@latest
```

This adds the core SDK to your application's module dependencies. Core includes transactions, codecs, clients, ledger types, and wallet signing. It does not require cgo.

Starting with core `v0.3.1`, confidential cryptographic helpers are a separate, optional Go module. Core-only module downloads do not contain their native headers or libraries. A repository clone or GitHub source archive still contains both modules.

Confidential transaction models and ledger fields remain in core. Your application can supply ciphertexts and proofs from another implementation, then use core to encode, sign, and submit the transaction. Core validates field formats and local transaction rules, not the cryptographic validity of a ZK proof.

## Optional confidential helpers

The first independent confidential release is `v0.1.0` and requires core `v0.3.1` or later.

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go also selects the required core dependency. Package import paths remain unchanged, such as `github.com/Peersyst/xrpl-go/confidential/builder`.

Native operations require cgo and a C/C++ toolchain on Linux or macOS with amd64 or arm64. Linux also needs the zlib development library. With cgo disabled or on unsupported targets, packages compile but native operations return `mptcrypto.ErrCgoRequired`.

See [Install confidential helpers](/docs/confidential/installation) for build requirements, version selection, updates, and migration from the combined module.

## Import and start using the SDK

Import the packages provided by the module you selected. After adding imports, run `go mod tidy` to record the dependencies and checksums your application uses.

The following example uses core packages to create a WebSocket client for XRPL testnet:

```go
package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/websocket"
)

func main() {
	client := websocket.NewClient(
		websocket.NewClientConfig().
			WithHost("wss://s.altnet.rippletest.net:51233").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	defer client.Disconnect()

	if err := client.Connect(); err != nil {
		fmt.Println(err)
		return
	}
}
```

## Next steps

The SDK is installed. Choose what you want to build:

- **Read ledger data:** use the [RPC client](/docs/xrpl/rpc) for queries. No wallet is needed.
- **Receive live updates:** use the [WebSocket client](/docs/xrpl/websocket) for subscriptions and event handlers.
- **Send transactions:** start with [wallets and signing](/docs/xrpl/wallet), then [build transactions](/docs/xrpl/transaction) to submit through your client.
- **Build confidential transfers:** if you installed the optional helpers, continue to the [builder guide](/docs/confidential/builders).
