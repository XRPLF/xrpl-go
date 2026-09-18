# XRPL-GO

A Go SDK for building applications on the [XRP Ledger](https://xrpl.org/). Query the ledger, manage wallets, and sign and submit transactions with typed Go models.

[![Go Reference](https://pkg.go.dev/badge/github.com/Peersyst/xrpl-go.svg)](https://pkg.go.dev/github.com/Peersyst/xrpl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Peersyst/xrpl-go)](https://goreportcard.com/report/github.com/Peersyst/xrpl-go)
[![Core release](https://img.shields.io/github/v/release/XRPLF/xrpl-go)](https://github.com/XRPLF/xrpl-go/releases/latest)

[Documentation](https://xrplf.github.io/xrpl-go/docs/installation) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go) · [Examples](examples) · [Contributing](CONTRIBUTING.md)

## Quick start

Requires **Go 1.25.13 or later**. The core SDK is pure Go and does not require a C/C++ toolchain.

From your application's Go module, install the SDK:

```bash
go get github.com/Peersyst/xrpl-go@latest
```

For a new application, run `go mod init example.com/my-xrpl-app` first.

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

Run it with:

```bash
go run .
```

With network access, it prints the server's software version. To send your first payment, follow the [Testnet payment example](examples/send-xrp/rpc).

## What you can build

- **Query ledger data** with [JSON-RPC](https://xrplf.github.io/xrpl-go/docs/xrpl/rpc), or subscribe to updates with [WebSocket](https://xrplf.github.io/xrpl-go/docs/xrpl/websocket).
- **Create and submit transactions** using [typed transaction models](https://xrplf.github.io/xrpl-go/docs/xrpl/transaction), autofill, and submission helpers that wait for validation.
- **Manage wallets and sign locally**, including multisigning and derivation from seeds or mnemonics. See the [wallet guide](https://xrplf.github.io/xrpl-go/docs/xrpl/wallet).
- **Work with XRPL encodings** through the [address codec](address-codec), [binary codec](binary-codec), and [keypair utilities](keypairs).

### Sending transactions

The usual transaction flow is:

1. Build a typed transaction and call `Flatten()`.
2. Call `client.Autofill()` to set network fields such as the fee and sequence.
3. Sign locally with `wallet.Sign()`.
4. Submit with `client.SubmitTxBlobAndWait()`.

For the combined flow, use `client.SubmitTxAndWait()` to autofill, sign, submit, and wait in one call. For offline signing, keep autofill and signing separate: autofill needs network access, signing does not.

See the [payment example](examples/send-xrp/rpc) for the complete flow and the [examples directory](examples) for more use cases.

## Optional confidential helpers

Most applications only need the core SDK. For confidential MPT encryption, balance decryption, commitments, proofs, and transaction builders, use the separate [`github.com/Peersyst/xrpl-go/confidential`](confidential/README.md) module.

Core already includes the transaction models and ledger fields. It can encode, sign, and submit confidential transactions when your application supplies the ciphertexts and proofs. The optional module generates these fields and decrypts balances.

The helpers require a native C/C++ toolchain and cgo for cryptographic operations on Linux or macOS with amd64 or arm64. Both modules require Go 1.25.13 or later. They are versioned independently, and core does not depend on the optional module.

See the [confidential README](confidential/README.md) for release availability, installation, migration, and platform requirements, or start with the [confidential examples](confidential/examples).

## Documentation and support

- [Guides](https://xrplf.github.io/xrpl-go/docs/installation) for setup, clients, wallets, and transactions.
- [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go) for package types and methods.
- [Core changelog](CHANGELOG.md) and [confidential changelog](confidential/CHANGELOG.md) for release changes.
- [XRPL documentation](https://xrpl.org/docs) for ledger concepts and protocol rules.

Found a bug or a documentation gap? [Open an issue](https://github.com/XRPLF/xrpl-go/issues).

## Security and audits

The signing functionality in this repository has **not been independently audited**. Test on Testnet or Devnet first, and review signing behavior before using it to control production funds.

Never print, log, commit, or send real seeds, private keys, or mnemonics to telemetry. Anyone with those values can control the account.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, tests, and pull request guidance. To work on both modules, follow the [confidential development setup](CONTRIBUTING.md#work-on-confidential-helpers). Root `go test ./...` only tests core.

## License

[MIT](LICENSE).
