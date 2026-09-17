---
sidebar_position: 2
---

# Install confidential helpers

Starting with core `v0.3.1`, `xrpl-go` has two Go modules in one repository. The optional confidential module starts at `v0.1.0`. Their versions and releases are independent.

The commands below require those releases to be published. To work on an unreleased checkout, use the [development workspace](#development-workspace).

## Choose a module

| You need | Module |
| --- | --- |
| Transactions, ledger types, codecs, clients, and wallet signing | `github.com/Peersyst/xrpl-go` |
| Confidential builders, encryption, balance decryption, commitments, or proofs | `github.com/Peersyst/xrpl-go/confidential` |

The core module retains confidential transaction models and ledger fields. It can encode, decode, sign, and submit a confidential transaction when the required ciphertexts and proofs are supplied by your application. Transaction validation checks field formats and applicable local rules, not the cryptographic validity of a ZK proof.

Normal wallet signing does not require the confidential module. The optional builders generate encrypted fields and proofs, then return a transaction for the normal signing and submission flow.

## Core only

Run this in your application module:

```bash
go get github.com/Peersyst/xrpl-go@latest
```

The core module has no dependency on the confidential module. Its Go module archive excludes the `confidential/` directory, including the native headers and static libraries. Build tags alone would not provide this download separation.

This applies to Go module downloads. A repository clone or a GitHub source archive still contains both modules. Older module releases and existing caches are not changed by the split.

## Add confidential helpers

```bash
go get github.com/Peersyst/xrpl-go/confidential@v0.1.0
```

Go also selects the required core module. You do not need a separate core installation command.

Package import paths have not changed:

```go
import (
    "github.com/Peersyst/xrpl-go/confidential/builder"
    "github.com/Peersyst/xrpl-go/xrpl/transaction"
)
```

After adding imports, run `go mod tidy` to record the dependencies and checksums used by your application.

### Native build requirements

The helpers use cgo to call `XRPLF/mpt-crypto`. Native operations require:

- Go `1.25.13` or later.
- `CGO_ENABLED=1`.
- Linux or macOS, on `amd64` or `arm64`.
- A C/C++ compiler and linker toolchain. On Linux, the linker also needs the system zlib development library.

For example, Ubuntu uses `build-essential` and `zlib1g-dev`. On macOS, install the Xcode command-line tools.

```bash
CGO_ENABLED=1 go build ./...
```

The module includes headers and static libraries for all four supported targets. Users do not need Conan or a separate `mpt-crypto` installation. `go get` does not install a compiler. Cross-compilation requires a matching target toolchain.

With cgo disabled, or on an unsupported target, the packages still build. Native cryptographic operations return `mptcrypto.ErrCgoRequired`. This fallback does not provide a pure-Go implementation of the cryptography.

## Versions and updates

The first confidential release requires core `v0.3.1` or later. Go selects one version of each module for the application:

| Existing core requirement | After adding confidential `v0.1.0` |
| --- | --- |
| None | Adds core `v0.3.1` |
| `v0.3.0` | Upgrades core to `v0.3.1` |
| A newer compatible core version | Keeps the newer version |

A minimum dependency is not an exact version lock or a guarantee that every future version is compatible. Check release notes before upgrading.

Update the modules separately:

```bash
go get github.com/Peersyst/xrpl-go@latest
go get github.com/Peersyst/xrpl-go/confidential@latest
```

A confidential update upgrades core only when its dependency requirements need that upgrade. A core update does not automatically update confidential.

The Go command uses a version without the directory prefix. The Git tag includes the prefix:

| Module | `go get` version | Git tag |
| --- | --- | --- |
| Core | `@v0.3.1` | `v0.3.1` |
| Confidential | `@v0.1.0` | `confidential/v0.1.0` |

### Migration from the combined module

Update applications that used confidential packages from `v0.3.1-mpt.0` to core `v0.3.1` and add the confidential module. Keep the same imports. The new confidential module's minimum core requirement performs the core upgrade automatically.

Do not use a local `replace` directive to force an older combined core release. Both modules would then contain the same helper package paths, which can cause ambiguous-import errors.

## Development workspace

From a repository checkout:

```bash
make workspace
make test-confidential
make test-confidential-nocgo
```

`make workspace` creates an ignored `go.work` file that uses both checked-out modules. It also maps the confidential module's required core version to the local core checkout, so development works before the first core release is published. This local replacement is not part of either published module.

Root `go test ./...` does not test the nested module. To run an example after workspace setup:

```bash
cd confidential
go run ./examples/offline
```

The RPC and WebSocket examples are in `confidential/examples/rpc` and `confidential/examples/ws`. They connect to devnet and submit transactions.

For contributor checks, see [CONTRIBUTING.md](https://github.com/XRPLF/xrpl-go/blob/main/CONTRIBUTING.md). For release order and the GitHub Actions picker, see [RELEASING.md](https://github.com/XRPLF/xrpl-go/blob/main/RELEASING.md).
