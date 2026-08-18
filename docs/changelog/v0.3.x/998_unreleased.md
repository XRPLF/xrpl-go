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

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions. Parameter validation mirrors the transaction models: the issuance issuer is rejected from holder roles and from a send destination, required for clawback, and amounts are bounded by the protocol maximum, so an invalid request fails before any ledger query or proof generation. These helpers require classic addresses, because the keylet and proof paths they feed decode classic addresses only.

#### docs

- Added confidential MPT documentation covering the CGo requirement, package layout, transaction types, and high-level builders.

#### pkg/crypto

- Added `IsCompressedSECP256K1Point()` to report whether a hex string decodes to a compressed secp256k1 point that lies on the curve, and `CompressedSECP256K1PointByteLength` so callers can size a compressed point without importing the curve library.

#### pkg/hexutil

- Added `DecodeFixedHex()` to decode hexadecimal values and enforce their decoded byte length.

#### pkg/mptsizes

- Added `mptsizes`, a CGo-free package holding the XLS-96 confidential MPT wire sizes that the vendored `mpt-crypto` headers define. The transaction models and the CGo bindings both derive their lengths from it, so a proof-format bump cannot leave the two disagreeing.

#### xrpl

- Added confidential-transfer flags and encryption-key fields to MPT issuance transaction and ledger-entry models.
- Added confidential balance fields to `MPToken`, five confidential MPT transaction models, and supporting amount, encryption-key, blinding-factor, ciphertext, commitment, and proof validation helpers.
- Added `IsMPTokenIssuer()` to report whether an address is the issuer encoded in an MPT issuance ID. Confidential MPT self-send and self-clawback checks compare decoded AccountIDs, so an X-address naming the submitting account is rejected; a `Holder` X-address carrying a tag is rejected because `ConfidentialMPTClawback` has no tag field to hold it; and an `Account` X-address tag combined with an explicit `SourceTag` is rejected before encoding rather than during it. `ConfidentialMPTSend` carries an optional `DestinationTag`, matching rippled's transaction format, and rejects it when the `Destination` X-address already embeds a tag. `ConfidentialMPTConvert` is now listed as non-delegatable, matching the protocol, so `DelegateSet` no longer accepts a permission the network rejects.

#### xrpl/hash

- Added `MPToken()` and `MPTokenIssuance()` helpers for computing MPT ledger-entry keylet indexes.

### Fixed

#### dependencies

- Raised the minimum Go version from 1.25.12 to 1.25.13 to fix the `net/url` quadratic path-resolution vulnerability (GO-2026-6218).
