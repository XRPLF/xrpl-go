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

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions. See the [confidential builders guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for the full parameter and error reference.
- Builder parameters are validated before any ledger query or proof work: addresses must be classic, the issuance issuer is rejected from holder roles and from a send destination, private keys must be usable secp256k1 scalars, and amounts are bounded by the protocol maximum.
- `Build*` helpers preflight the issuance capabilities the network enforces, so a doomed transaction never costs a fee and a sequence. Adds `ErrConfidentialDisabled`, `ErrTransferDisabled`, `ErrTransferFeeSet`, `ErrIssuanceNotFound`, `ErrKeyMismatch`, and `ErrAmountExceedsOutstanding`.
- `BuildSendParams` carries an optional `DestinationTag`, and `CredentialIDs` are validated before use, reported through `ErrInvalidCredentialIDs`, which wraps `transaction.ErrInvalidCredentialIDs` so a caller matching the builder error set does not have to import `xrpl/transaction`.
- `BuildClawback` derives the clawback amount by decrypting the holder's `IssuerEncryptedBalance`, bounded by `BalanceRange` and capped at the issuance `ConfidentialOutstandingAmount`. `Amount` moved from `BuildClawbackParams` to `ClawbackParams`, so it is supplied only on the offline `PrepareClawback` path.
- `BuildConvert` reports a missing holder `MPToken` as `ErrMPTokenNotFound` rather than treating it as a first-time opt-in, matching `ConfidentialMPTConvert`, which debits the entry and so requires it to exist. A failed read reports `ErrLedgerQuery`, so a transport error can no longer be mistaken for first-time state.
- Proof-bearing `Prepare*` helpers reject a zero `Sequence` with `ErrMissingSequence`, because each proof binds the sequence and a later autofill would invalidate it. `PrepareMergeInbox` and a repeat `PrepareConvert` carry no proof and still accept a zero `Sequence`.

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
- Added confidential balance fields to `MPToken`, five confidential MPT transaction models, and supporting amount, encryption-key, blinding-factor, ciphertext, commitment, and proof validation helpers. RPC and WebSocket autofill apply the required 10x base fee to confidential MPT transactions, including inner Batch transactions, plus the normal per-signer surcharge.
- Added `IsMPTokenIssuer()` to report whether an address is the issuer encoded in an MPT issuance ID. Confidential MPT self-send and self-clawback checks compare decoded AccountIDs, so an X-address naming the submitting account is rejected; a `Holder` X-address carrying a tag is rejected because `ConfidentialMPTClawback` has no tag field to hold it; and an `Account` X-address tag combined with an explicit `SourceTag` is rejected before encoding rather than during it. `ConfidentialMPTSend` carries an optional `DestinationTag`, matching rippled's transaction format, and rejects it when the `Destination` X-address already embeds a tag. `ConfidentialMPTConvert` is now listed as non-delegatable, matching the protocol, so `DelegateSet` no longer accepts a permission the network rejects.

#### xrpl/hash

- Added `MPToken()` and `MPTokenIssuance()` helpers for computing MPT ledger-entry keylet indexes.

### Fixed

#### confidential

- Tightened participant and fixed-size proof validation before invoking the native verifier, and aligned proof helper naming with the current `mpt-crypto` contract. Send proofs require exactly the sender, destination, and issuer, followed by the optional auditor, and any other count is rejected with `ErrInvalidParticipantCount`.
- Every `Generate*Proof` helper now verifies its own output before returning it. The native generator reports no error for a mismatched amount or key pair, so a bad proof previously surfaced only when the network rejected the transaction. The check costs one verification per generation, roughly 35% added wall time for a send proof, which the confidential documentation records for callers that batch proof generation.

#### dependencies

- Raised the minimum Go version from 1.25.12 to 1.25.13 to fix the `net/url` quadratic path-resolution vulnerability (GO-2026-6218).

#### xrpl/rpc

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first.

#### xrpl/transaction

- Corrected confidential transaction key encoding and proof-size validation.

#### xrpl/websocket

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first.
