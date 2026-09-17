---
sidebar_position: 2
---

# Installation

On this page, you'll learn how to set up `xrpl-go` SDK in your project, to start interacting with the XRP Ledger.

## Prerequisites

Before installing, you need to have Go installed on your machine. You can download it from the [official website](https://go.dev/doc/install).

The minimum version of Go required to use `xrpl-go` is:

| Software | Version |
|------------|---------|
| Go | >= 1.25.13 |

## Install the core module

Run this command in your application directory, which must contain a `go.mod` file. For a new application, create one with `go mod init <your-module-path>` first.

```bash
go get github.com/Peersyst/xrpl-go@latest
```

This adds the core SDK to your application's module dependencies. Core includes transactions, codecs, clients, ledger types, and wallet signing.

Starting with core `v0.3.1`, confidential cryptographic helpers are a separate, optional Go module. Core-only module downloads do not contain their native headers or libraries. Confidential transaction models and normal transaction signing remain in core.

## Optional confidential helpers

After confidential `v0.1.0` is published, install it with:

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go also selects its required core dependency. Use this module for encryption, proof generation, confidential balance decryption, and transaction builders. It requires cgo and a supported native toolchain for cryptographic operations.

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

Now that you have the `xrpl-go` package downloaded and imported in your project, you can start interacting with the XRP Ledger.

To learn more about the `xrpl-go` packages, you can find the documentation for each package:

- [confidential](/docs/confidential)
- [keypairs](/docs/keypairs)
- [xrpl](/docs/xrpl/currency)
