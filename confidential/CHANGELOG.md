# Changelog

This file records changes to the optional `github.com/Peersyst/xrpl-go/confidential` module.
Core transaction models, codecs, clients, and wallet signing are covered in the [core changelog](../CHANGELOG.md).

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this module follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0]

First independent release of the confidential module, tagged `confidential/v0.1.0`. Requires core `github.com/Peersyst/xrpl-go v0.3.1` or later. Existing package import paths are unchanged.

### Added

#### module

- Added a separate Go module for confidential builders and cryptography. Core does not depend on it, and core module downloads exclude its native bundles. Installing this module also selects its required core dependency.
- Added confidential examples in `confidential/examples/` and integration tests in `confidential/integration/`. See the [installation and migration guide](https://xrplf.github.io/xrpl-go/docs/confidential/installation) for application and checkout setup.

#### mptcrypto

- Added CGo bindings and vendored XRPLF `mpt-crypto` 1.0.5 libraries for Linux and macOS on amd64 and arm64. Native operations require a C/C++ toolchain and, on Linux, the zlib development library. Unsupported targets and builds with cgo disabled compile but return `ErrCgoRequired` for native operations.
- Added byte-array APIs for ElGamal encryption, ciphertext arithmetic, canonical encrypted zero, Pedersen commitments, context hashes, and zero-knowledge proof generation and verification.

#### elgamal

- Added hex-string APIs for key and blinding-factor generation, encryption, and decryption with caller-supplied search bounds.
- Added `Add()` and `Subtract()` for homomorphic ciphertext arithmetic. Subtracting a ciphertext from itself reports `ErrCiphertextArithmetic` because the result cannot be represented as a ciphertext.
- Added `EncryptCanonicalZero()` to reproduce the deterministic encrypted zero used when confidential MPT transactions initialize or reset balances.

#### commitment, proof

- Added hex-string APIs for Pedersen commitments, proof context hashes, and proof generation and verification. Context hashes bind the decoded AccountID, so equivalent classic and X-addresses produce the same context.
- Proof helpers validate participant inputs and proof sizes. Each `Generate*Proof` helper verifies its output before returning it to detect mismatched amounts or key pairs.

#### builder

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions. Inputs are validated before ledger queries or proof generation, and prepared transactions are validated before return.
- Online builders check issuance capabilities and holder state using one validated ledger snapshot. `BuildSend` and `BuildConvertBack` detect open-ledger balance-version changes with `ErrStaleBalanceVersion`.
- Added `TxOptions` for sequences, Tickets, and delegation, with proofs bound to the transaction's nonce. Builders normalize classic and X-addresses and reject invalid holder roles. Send parameters support destination tags and credential IDs.
- Added `BuildClawback()` to derive the clawback amount from the holder's issuer-encrypted balance. The decryption search is bounded by `BalanceRange` and the issuance's confidential outstanding amount. Offline `PrepareClawback()` accepts a caller-supplied amount.
- Added `BuildBatch()` to assemble two to eight ordered operations using predicted balances and versions from one validated ledger snapshot. It resolves inner nonces before generating proofs and supports only `tfAllOrNothing`. Outer fee autofill and signing remain with the caller.
- Added `ConvertOp`, `ConvertBackOp`, `SendOp`, `MergeInboxOp`, `ClawbackOp`, and `TransactionOp` for Batch assembly. `IsSupportedInnerTransactionType()` identifies supported ordinary inner transactions. See the [builders guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for nonce rules and restrictions.
- Added `GetSpendingBalance()` to read and decrypt a holder's spendable confidential balance through either client. It excludes the inbox, returns zero when no spending ciphertext is present, and bounds decryption by `BalanceRange` and the issuance's confidential outstanding amount.

#### development

- Added workspace support, separate confidential test and lint targets, and no-cgo fallback checks. Root `./...` commands do not include this module.
- Added confidential CI checks, localnet integration coverage, and native tests on Linux and macOS for amd64 and arm64. Native dependency updates use a separate `mpt-crypto` workflow.
- Added independent releases using `confidential/vX.Y.Z` tags and this changelog. Release checks use published core dependencies with `GOWORK=off`.

#### docs

- Added confidential installation and API guides, plus offline, RPC, and WebSocket examples. The builders guide covers standalone operations, Batch assembly, balance reading, autofill, and signing.
