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
| Go | >= 1.24.3 |

## Download package

Once you have Go installed and you have a Go workspace, you can download the package with the following command:

```bash
go get github.com/Peersyst/xrpl-go
```

By running this command, the latest version of the `xrpl-go` package will be downloaded and added to your Go workspace.

## Import and start using the SDK

Once you have the package downloaded, you can import any `xrpl-go` package in your project and start working with it. 
The following example shows how to import the `xrpl` package and create a new WebSocket client to connect to the XRPL testnet chain:

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

- [keypairs](/docs/keypairs)
- [xrpl](/docs/xrpl/currency)
