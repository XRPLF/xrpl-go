# Installation

Install the core SDK for queries, transactions, wallets, and codecs. Add confidential helpers only if you need to generate encrypted fields or proofs.

## Requirements

Install [Go 1.25.13 or later](https://go.dev/doc/install). Core is pure Go and does not require cgo or a C/C++ toolchain.

## Add the SDK

Run these commands in a new application directory:

```bash
go mod init example.com/my-xrpl-app
go get github.com/Peersyst/xrpl-go@latest
```

For an existing application, skip `go mod init`. The module path is `github.com/Peersyst/xrpl-go`, even though the repository is hosted under XRPLF.

Import the packages your application uses, then run:

```bash
go mod tidy
go build ./...
```

To check the selected version:

```bash
go list -m github.com/Peersyst/xrpl-go
```

## Optional confidential helpers

The confidential module provides builders, encryption, balance decryption, commitments, and proofs. Core alone can encode, sign, and submit confidential transactions if your application supplies the encrypted fields and proofs.

Follow [Install confidential helpers](/docs/confidential/installation) for the module version, supported platforms, and native toolchain requirements. Native cryptographic operations have no pure-Go fallback.

## Next steps

- **Read ledger data:** use the [JSON-RPC client](/docs/xrpl/rpc). No wallet is needed.
- **Receive live updates:** use the [WebSocket client](/docs/xrpl/websocket).
- **Send transactions:** start with [wallets and signing](/docs/xrpl/wallet), then [Transactions](/docs/xrpl/transaction).
- **Build confidential transfers:** continue to the [builder guide](/docs/confidential/builders).
