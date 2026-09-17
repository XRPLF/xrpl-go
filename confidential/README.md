# Confidential MPT helpers

`github.com/Peersyst/xrpl-go/confidential` is the optional Go module for XLS-96 confidential MPT builders and cryptography. It depends on core `github.com/Peersyst/xrpl-go`, but core does not depend on it.

The first independent release is `v0.1.0`, with core `v0.3.1` as its minimum dependency. After these versions are published, install the helpers in your application with:

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go selects the required core version automatically. Import paths remain unchanged, for example `github.com/Peersyst/xrpl-go/confidential/builder`.

## Packages

| Package | Purpose |
| --- | --- |
| `builder` | Build transactions from ledger state or explicit inputs, and read confidential spending balances |
| `elgamal` | Generate encryption keys and encrypt or decrypt amounts |
| `commitment` | Create Pedersen commitments |
| `proof` | Generate and verify proofs and transaction context hashes |
| `mptcrypto` | Low-level cgo bindings to `XRPLF/mpt-crypto` |

Transaction models, codecs, RPC and WebSocket clients, and normal wallet signing stay in core. Core-only users can supply ciphertexts and proofs from another implementation without installing these helpers.

## Requirements

Native operations require Go `1.25.13` or later, cgo, and a C/C++ compiler and linker on Linux or macOS with amd64 or arm64. Linux also needs the system zlib development library. Headers and static libraries for all four targets are included under `deps/`. Conan is only needed to maintain those bundles, not to use the module.

Unsupported targets and builds with `CGO_ENABLED=0` still compile, but native cryptographic operations return `mptcrypto.ErrCgoRequired`.

See the [installation guide](https://xrplf.github.io/xrpl-go/docs/confidential/installation) for version selection, migration, and build requirements, and the [builder guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for usage.

## Development

From the repository root:

```bash
make workspace
make test-confidential
make test-confidential-nocgo
make lint-confidential
```

Examples are under `examples/`, and integration tests are under `integration/`, relative to this module. Root `go test ./...` does not include this module.

The local workspace makes unreleased core changes available to the helpers. Releases are independently tagged as `confidential/vX.Y.Z` and are checked against the published core dependency with `GOWORK=off`. See [RELEASING.md](https://github.com/XRPLF/xrpl-go/blob/main/RELEASING.md).
