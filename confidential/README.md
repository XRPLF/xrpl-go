# Confidential MPT helpers

`github.com/Peersyst/xrpl-go/confidential` is the optional Go module for XLS-96 confidential MPT builders and cryptography. It depends on [core `xrpl-go`](../README.md), but core does not depend on it.

Use this module to generate encrypted transaction fields and proofs, read confidential spending balances, or build transactions from ledger state. Transaction models, codecs, RPC and WebSocket clients, and normal wallet signing stay in core. Applications that supply ciphertexts and proofs from another implementation do not need this module.

## Installation

The first independent release is `v0.1.0`, with core `v0.3.1` as its minimum dependency. Once both releases are published, run this in your application module:

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go selects the required core dependency automatically. Package import paths remain unchanged, for example `github.com/Peersyst/xrpl-go/confidential/builder`. After adding imports, run `go mod tidy`.

For an unreleased repository checkout, use the [development workspace](#development) instead.

### Migration from the combined module

Applications using confidential packages from `v0.3.1-mpt.0` must update core and add the new module. Keep the existing imports:

```bash
go get github.com/Peersyst/xrpl-go@v0.3.1 github.com/Peersyst/xrpl-go/confidential@v0.1.0
go mod tidy
```

Do not force an older combined core release with `replace`. It can provide the same package paths as the new module and cause ambiguous-import errors.

See the [installation guide](https://xrplf.github.io/xrpl-go/docs/confidential/installation) for version selection and updates.

## Native build requirements

Both modules require Go `1.25.13` or later. Native cryptographic operations also require:

- `CGO_ENABLED=1`.
- Linux or macOS, on `amd64` or `arm64`.
- A C/C++ compiler and linker. Linux also needs the system zlib development library.

On Ubuntu, install `build-essential` and `zlib1g-dev`. On macOS, install the Xcode command-line tools. Cross-compilation requires a toolchain for the target platform.

Headers and static libraries for all four targets are included under [`deps/`](deps/). Conan and a separate `mpt-crypto` installation are not needed to use the module. Conan is used only to maintain the native bundles.

Build your application with:

```bash
CGO_ENABLED=1 go build ./...
```

Unsupported targets and builds with `CGO_ENABLED=0` still compile, but native operations return `mptcrypto.ErrCgoRequired`. There is no pure-Go cryptographic fallback.

## Packages

| Package | Purpose |
| --- | --- |
| [`builder`](builder/) | Build transactions from ledger state with `Build*`, prepare them from explicit inputs with `Prepare*`, and read spending balances with `GetSpendingBalance` |
| [`elgamal`](elgamal/) | Generate encryption keys and encrypt or decrypt amounts |
| [`commitment`](commitment/) | Create Pedersen commitments |
| [`proof`](proof/) | Generate and verify proofs and transaction context hashes |
| [`mptcrypto`](mptcrypto/README.md) | Low-level cgo bindings to `XRPLF/mpt-crypto` |

Builders return core transaction types. Flatten the result, autofill the remaining network fields, sign with a core wallet, and submit through the RPC or WebSocket client. Proofs bind the transaction nonce, so do not change a prepared sequence or Ticket before submission. See the [builder guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for the full flow.

## Run an offline example

From the repository root, set up the workspace and run the example from this module:

```bash
make workspace
cd confidential
CGO_ENABLED=1 go run ./examples/offline
```

The example prepares an issuance, issuer-key registration, a holder conversion, and an inbox merge. It does not connect, sign, or submit transactions. Offline proof generation still needs the native toolchain.

The [`rpc`](examples/rpc/) and [`ws`](examples/ws/) examples run a full confidential lifecycle against devnet. They fund test wallets and submit transactions. Read the examples before running them.

## Development

From the repository root:

```bash
make workspace
make test-confidential
make test-confidential-nocgo
make lint-confidential
```

The ignored workspace uses both checked-out modules and includes a local replacement for the required core version. This lets development work before that core version is published. Do not add the replacement to a published `go.mod` file.

Root `go test ./...` does not include this module. After workspace setup, run commands from `confidential/` to address its packages. Examples are in [`examples/`](examples/) and integration tests are in [`integration/`](integration/). Integration tests need a compatible ledger. See [CONTRIBUTING.md](../CONTRIBUTING.md) for the separate localnet and devnet targets.

## Versions and releases

This module has its own [changelog](CHANGELOG.md) and uses Git tags such as `confidential/v0.1.0`. Use `@v0.1.0`, without the directory prefix, in `go get` commands.

Core and confidential versions are independent. A confidential release needs a new core release only when it uses a new core API. Releases are checked against published dependencies with `GOWORK=off`. See [RELEASING.md](../RELEASING.md) for release order and the module picker.

## Security

Protect both wallet signing secrets and confidential encryption private keys. Do not print, log, commit, or send them to telemetry. Test with non-production funds and read the core [security and audit notice](../README.md#security-and-audits) before production use.
