---
title: Unreleased
---

### Added

#### binary-codec

- Synced the embedded `definitions.json` with the rippled 3.3.0 protocol definitions.
- Added XLS-96 confidential MPT fields and transaction definitions for `ConfidentialMPTSend`, `ConfidentialMPTConvert`, `ConfidentialMPTConvertBack`, `ConfidentialMPTMergeInbox`, and `ConfidentialMPTClawback`, with `ConfidentialOutstandingAmount` serialized as a decimal string like the other MPT amount fields.
- Added the `tecBAD_PROOF`, `tecNO_SPONSOR_PERMISSION`, `temBAD_CIPHERTEXT`, `tefBAD_PATH_COUNT`, `tefNO_DST_PARTIAL`, and `terNO_PERMISSION` transaction results, which previously could not be encoded by name.

#### confidential

- Added CGo bindings and vendored native libraries for XRPLF `mpt-crypto`, with a `!cgo` fallback and maintainer tooling for dependency updates.
- Added hex-string APIs for ElGamal encryption, Pedersen commitments, context hashes, and zero-knowledge proof generation and verification.
- Added `test-confidential` and `update-mpt-crypto` Makefile targets and an automated dependency-update workflow.

#### confidential/builder

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions.

#### docs

- Added confidential MPT documentation covering the CGo requirement, package layout, transaction types, and high-level builders.

#### pkg/hexutil

- Added `DecodeFixedHex()` to decode hexadecimal values and enforce their decoded byte length.

#### xrpl

- Added confidential-transfer flags and encryption-key fields to MPT issuance transaction and ledger-entry models.
- Added confidential balance fields to `MPToken`, five confidential MPT transaction models, and supporting amount, encryption-key, hex-blob, blinding-factor, and proof validation helpers.

#### xrpl/hash

- Added `MPToken()` and `MPTokenIssuance()` helpers for computing MPT ledger-entry keylet indexes.

### Fixed

#### dependencies

- Raised the minimum Go version from 1.25.12 to 1.25.13 to fix the `net/url` quadratic path-resolution vulnerability (GO-2026-6218).
