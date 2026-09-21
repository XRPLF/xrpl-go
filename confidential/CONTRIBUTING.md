# Contributing to confidential helpers

See the [repository contributor guide](../CONTRIBUTING.md) for the general workflow. This guide covers the optional module and native boundary.

## Development workspace

Follow [Work on confidential helpers](../CONTRIBUTING.md#work-on-confidential-helpers) for workspace setup, local dependency selection, and module test commands. Root `go test ./...` does not test the nested module.

After workspace setup, run the offline example from the repository root:

```bash
cd confidential
CGO_ENABLED=1 go run ./examples/offline
```

The offline example generates proofs without connecting, signing, or submitting. It still requires the native toolchain. The RPC and WebSocket examples are in `confidential/examples/rpc` and `confidential/examples/ws`. They connect to devnet, fund test wallets, and submit transactions.

Confidential integration tests live in `confidential/integration/` and require a compatible ledger. Keep examples and tests that import the optional helpers inside this module, so root dependency management stays independent.

For release order and the GitHub Actions picker, see [RELEASING.md](../RELEASING.md).

## Native package layout

```text
mptcrypto/
  types.go                 # Package documentation, MaxParticipants, and value types
  errors.go                # Shared sentinel errors
  sizes_cgo.go             # Compile-time checks against native size constants
  mptcrypto_cgo.go          # Native bindings and native-only validation
  mptcrypto_nocgo.go        # Unavailable-backend stubs
  mptcrypto_test.go         # Native cryptographic tests
  mptcrypto_nocgo_test.go    # Fallback availability contract
```

See the [package test commands](mptcrypto/README.md#test-the-package) to check both native and fallback builds. Use the [development workspace](#development-workspace) to test against the checked-out core module.

## Maintaining the cgo boundary

Vendored headers are under `confidential/deps/include/`. Static libraries are selected from the matching directory under `confidential/deps/libs/`:

| Target | Directory |
| --- | --- |
| Linux amd64 | `linux-amd64/` |
| Linux arm64 | `linux-arm64/` |
| macOS amd64 | `darwin-amd64/` |
| macOS arm64 | `darwin-arm64/` |

`mptcrypto_cgo.go` contains the build constraint and per-platform linker flags. It passes fixed-size byte arrays to C through pointers to their first elements and copies Go compound values into their corresponding C structs field by field. Variable participant lists are copied into a contiguous slice of `C.mpt_confidential_participant` values before the native call.

The native routines use these pointers only for the duration of the call. They must not retain Go memory after returning. Keep all `import "C"`, `unsafe`, C layout conversion, and native linker changes inside this package so the higher-level confidential packages remain portable pure Go code.
