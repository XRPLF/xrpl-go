# mptcrypto

Low-level Go bindings for the [XRPLF/mpt-crypto](https://github.com/XRPLF/mpt-crypto) C library. This package provides encryption, commitments, context hashes, and proofs for XLS-96 confidential MPT transactions.

[Installation](../README.md#installation) · [Native API guide](../../docs/docs/confidential/mptcrypto.md) · [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/mptcrypto)

## Choose the right level

Most applications should use the [builders and higher-level helpers](../README.md#choose-a-package). Use `mptcrypto` when you need direct access to native operations and fixed-size byte types.

Only this package imports `"C"`. The higher-level helpers handle hex encoding, address decoding, and domain-specific errors in Go, but their cryptographic operations still require this native backend.

## Native requirements

Native operations require:

- Go 1.25.13 or later and cgo enabled with `CGO_ENABLED=1`.
- Linux or macOS on `amd64` or `arm64`.
- A C/C++ compiler and linker, plus the zlib development library on Linux.
- A build that does not target `js`, `wasip1`, TinyGo, or go-fuzz.

Headers and static libraries for the supported platforms are included in [`../deps/`](../deps/). You do not need a separate `mpt-crypto` installation. See the [module installation guide](../README.md#installation) for toolchain setup.

**There is no pure-Go cryptographic fallback.** With cgo disabled or on an unsupported target, the package still compiles, but every operation immediately returns `ErrCgoRequired` before validating or processing inputs.

## Read the native API guide

The [detailed guide](../../docs/docs/confidential/mptcrypto.md) covers the contracts you need when calling these low-level functions:

- [Data types and sizes](../../docs/docs/confidential/mptcrypto.md#data-model), including participant ordering.
- [Function behavior](../../docs/docs/confidential/mptcrypto.md#function-reference), including proof inputs, context hashes, and verification requirements.
- [Errors](../../docs/docs/confidential/mptcrypto.md#error-behavior) and how to match them.
- [Contributor notes](../CONTRIBUTING.md#native-package-layout) on package layout and [the cgo boundary](../CONTRIBUTING.md#maintaining-the-cgo-boundary).

Use the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/mptcrypto) for complete function signatures.

## Test the package

From the repository root, with the native toolchain installed:

```bash
make workspace
cd confidential

# Test the native implementation on a supported host.
CGO_ENABLED=1 go test ./mptcrypto

# Test the unavailable-backend behavior without cgo.
CGO_ENABLED=0 go test ./mptcrypto
```

The second command tests the fallback contract, not cryptographic operations.
