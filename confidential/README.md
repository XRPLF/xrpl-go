# Confidential MPT helpers

Build confidential MPT transactions, generate encrypted fields and proofs, and decrypt balances. These helpers are part of the optional `github.com/Peersyst/xrpl-go/confidential` module for XLS-96.

[Installation](#installation) · [Offline example](#run-an-offline-example) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential) · [Changelog](CHANGELOG.md)

## Do you need this module?

The [core SDK](../README.md) already provides transaction models, codecs, clients, and wallet signing. If your application supplies ciphertexts and proofs from another implementation, core can encode, sign, and submit the transactions without these helpers.

Install this module when you need the SDK to generate those fields, decrypt balances, or build transactions from ledger state. It depends on core, but core does not depend on it.

## Installation

Both modules require **Go 1.25.13 or later**. Native cryptographic operations also require:

- Linux or macOS on `amd64` or `arm64`.
- cgo enabled with `CGO_ENABLED=1`.
- A C/C++ compiler and linker. Use `build-essential` on Ubuntu or the Xcode command-line tools on macOS.
- The zlib development library on Linux, such as `zlib1g-dev` on Ubuntu.

The first independent helper release is `v0.1.0` and requires core `v0.3.1` or later.

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go selects the required core dependency automatically. Import packages such as `github.com/Peersyst/xrpl-go/confidential/builder`, then run `go mod tidy`.

Build your application with:

```bash
CGO_ENABLED=1 go build ./...
```

Headers and static libraries for all four supported targets are included in [`deps/`](deps/). You do not need Conan or a separate `mpt-crypto` installation to use the module. Cross-compilation requires a toolchain for the target platform.

**There is no pure-Go cryptographic fallback.** Unsupported targets and builds with `CGO_ENABLED=0` compile, but native operations return `mptcrypto.ErrCgoRequired`.

For an unreleased checkout, use the [development workspace](#development). See the [installation guide](https://xrplf.github.io/xrpl-go/docs/confidential/installation) for version selection and updates.

## Run an offline example

With the native toolchain installed, run these commands from the repository root:

```bash
make workspace
cd confidential
CGO_ENABLED=1 go run ./examples/offline
```

The example prepares an issuance, issuer-key registration, holder conversion, and inbox merge. It does not connect, sign, or submit transactions. Offline proof generation still requires the native toolchain.

The [`rpc`](examples/rpc/) and [`ws`](examples/ws/) examples fund test wallets and submit a full confidential lifecycle on Devnet. Read them before running them.

## Choose a package

| Package | Purpose |
| --- | --- |
| [`builder`](builder/) | Build transactions from ledger state with `Build*`, from explicit inputs with `Prepare*`, or read balances with `GetSpendingBalance` |
| [`elgamal`](elgamal/) | Generate encryption keys and encrypt or decrypt amounts |
| [`commitment`](commitment/) | Create Pedersen commitments |
| [`proof`](proof/) | Generate and verify proofs and transaction context hashes |
| [`mptcrypto`](mptcrypto/README.md) | Access the low-level native bindings |

Builders return core transaction types. Flatten the result, autofill the remaining network fields, sign with a core wallet, and submit through a client. **Do not change a prepared sequence or Ticket before submission.** Proofs bind the transaction nonce. See the [builder guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for the complete flow.

## Migration from the combined module

If you used confidential packages from `v0.3.1-mpt.0`, update core and add the new module. Keep the existing imports:

```bash
go get github.com/Peersyst/xrpl-go@v0.3.1 github.com/Peersyst/xrpl-go/confidential@v0.1.0
go mod tidy
```

Do not force an older combined core release with `replace`. It can provide the same package paths as the new module and cause ambiguous-import errors.

## Development

From the repository root:

```bash
make workspace
make test-confidential
make test-confidential-nocgo
make lint-confidential
```

The ignored workspace links both checked-out modules and replaces the required core version locally. This supports development before the core release is published. Keep replacements out of published `go.mod` files.

Root `go test ./...` skips this module. To run package-level commands, first enter `confidential/`. Keep confidential examples and integration tests in [`examples/`](examples/) and [`integration/`](integration/). Integration tests need a compatible ledger. See [CONTRIBUTING.md](../CONTRIBUTING.md) for network-specific test commands.

## Versions and releases

Core and confidential are versioned independently. This module uses tags such as `confidential/v0.1.0`, but `go get` takes `@v0.1.0`, without the directory prefix.

A confidential release needs a new core release only when it uses a new core API. Releases are checked against published dependencies with `GOWORK=off`. See [RELEASING.md](../RELEASING.md) for release order and the module picker. Conan is needed only when maintaining native bundles, not when using them.

## Security

Protect both wallet signing secrets and confidential encryption private keys. Never print, log, commit, or send them to telemetry. Test with non-production funds and read the core [security and audit notice](../README.md#security-and-audits) before production use.
