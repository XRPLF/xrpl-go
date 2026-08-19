---
title: Unreleased
---

### Added

#### address-codec

- Added `DecodeAddress()`, the decoding counterpart to `IsValidAddress()`. It resolves a classic address or an X-address to the `AccountID` both forms encode, so callers can compare accounts without knowing which form they were given.
- Added `IsZeroAccountID()`, which reports ACCOUNT_ZERO for a 20-byte identifier and false for any other length.

#### binary-codec

- Synced the embedded `definitions.json` with the rippled 3.3.0 protocol definitions.
- Added XLS-96 confidential MPT fields and transaction definitions for `ConfidentialMPTSend`, `ConfidentialMPTConvert`, `ConfidentialMPTConvertBack`, `ConfidentialMPTMergeInbox`, and `ConfidentialMPTClawback`. `ConfidentialOutstandingAmount` serializes as a decimal string like the other MPT amount fields.
- Added the `tecBAD_PROOF`, `tecNO_SPONSOR_PERMISSION`, `temBAD_CIPHERTEXT`, `tefBAD_PATH_COUNT`, `tefNO_DST_PARTIAL`, and `terNO_PERMISSION` transaction results, which previously could not be encoded by name.

#### confidential

- Added CGo bindings and vendored native libraries for XRPLF `mpt-crypto`, with a `!cgo` fallback that reports `ErrCgoRequired`. See the [confidential guide](https://xrplf.github.io/xrpl-go/docs/confidential) for the build requirements.
- Added hex-string APIs for ElGamal encryption, Pedersen commitments, context hashes, and zero-knowledge proof generation and verification.
- Proof context hashes bind the decoded AccountID, so a proof matches whether the caller supplied a classic address or its X-address form.
- Added the `test-confidential`, `update-mpt-crypto`, `test-integration-confidential-localnet`, and `test-integration-confidential-devnet` Makefile targets, plus an automated `mpt-crypto` dependency-update workflow. The confidential suites are excluded from the standard targets because they require CGo.
- Added a CGo matrix workflow that runs the unit suites, the confidential ones included, on every platform the build tags accept: linux and darwin on amd64 and arm64.

#### confidential/builder

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions. See the [confidential builders guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for the full parameter and error reference.
- Builder parameters are validated before any ledger query or proof work. Addresses may be classic or X-addresses and are normalized before use, ACCOUNT_ZERO is rejected, the issuance issuer is rejected from holder roles and from a send destination, and amounts are bounded by the protocol maximum.
- `Build*` helpers preflight the issuance capabilities and the ledger state each confidential MPT transactor requires, so a doomed transaction never costs a fee and a sequence.
- Online builders read every ledger entry from one validated ledger, pinned by hash after the first read, and take the account sequence from the open ledger. `BuildSend` and `BuildConvertBack` report `ErrStaleBalanceVersion` when a transaction still in flight has already moved the holder's `ConfidentialBalanceVersion`, rather than paying a fee for a `tecBAD_PROOF`.
- `Prepare*` helpers validate the transaction they assemble, so a field set wrongly is reported as `ErrInvalidTransaction` instead of surfacing at submission.
- Every `Build*Params` embeds `TxOptions`, which carries `Sequence`, `TicketSequence`, and `Delegate`. Each proof binds the nonce the transaction spends, so a proof-bearing `Prepare*` helper requires one of the two.
- `BuildSendParams` carries an optional `DestinationTag`, and its `CredentialIDs` are validated before use.
- `BuildClawback` derives the clawback amount by decrypting the holder's `IssuerEncryptedBalance`, bounded by `BalanceRange` and capped at the issuance `ConfidentialOutstandingAmount`. `Amount` is supplied by the caller only on the offline `PrepareClawback` path.

#### docs

- Added confidential MPT documentation covering the CGo requirement, package layout, transaction types, transaction cost, and high-level builders.
- Added `examples/confidential`, with an offline walkthrough that assembles an opt-in and an inbox merge without connecting, and RPC and WebSocket examples that run a full confidential lifecycle against devnet.

#### pkg/crypto

- Added `IsCompressedSECP256K1Point()` to report whether a hex string decodes to a compressed secp256k1 point on the curve, and `CompressedSECP256K1PointByteLength` so callers can size a point without importing the curve library.

#### pkg/hexutil

- Added `DecodeFixedHex()` to decode hexadecimal values and enforce their decoded byte length.

#### pkg/mptsizes

- Added `mptsizes`, a CGo-free package holding the XLS-96 confidential MPT wire sizes that the vendored `mpt-crypto` headers define. The transaction models and the CGo bindings both derive their lengths from it, and `confidential/mptcrypto` pins every constant to the `mpt-crypto` define it mirrors with a compile-time assertion.

#### xrpl

- Added the five XLS-96 confidential MPT transaction models, along with supporting amount, encryption-key, blinding-factor, hex-blob, ciphertext, commitment, and proof validation helpers.
- Added `IssuerEncryptionKey` and `AuditorEncryptionKey` to `MPTokenIssuanceSet`. An auditor key requires an issuer key in the same transaction, and neither can be combined with `Holder`.
- Added confidential balance and encryption-key fields to the `MPToken` and `MPTokenIssuance` ledger entries.
- RPC and WebSocket autofill apply the required 10x base fee to confidential MPT transactions, including inner `Batch` transactions, plus the normal per-signer surcharge.
- Added `IsMPTokenIssuer()` to report whether an address is the issuer encoded in an MPT issuance ID.
- Added `ErrZeroAccountID`, `ErrAccountZero`, `ErrDelegateZero`, `ErrDelegateTagNotAllowed`, `ErrSignerAccountZero`, and `ErrSignerAccountTagNotAllowed` for the address conditions `BaseTx.Validate()` now reports. Each names its field and wraps its condition, so `errors.Is` matches either one.
- Added `ErrAccountIDTagNotAllowed` and `ErrDuplicateXAddressTag`, which alias the `binary-codec/types` sentinels so preflight and encoding report one error identity for these conditions. `ErrClawbackHolderTagNotAllowed` now wraps `ErrAccountIDTagNotAllowed` for the same reason, which appends the wrapped reason to its message text.
- `ConfidentialMPTSend` carries an optional `DestinationTag`, matching rippled's transaction format, and rejects it when the `Destination` X-address already embeds a tag.
- `ConfidentialMPTConvert` is listed as non-delegatable, matching the protocol, so `DelegateSet` no longer accepts a permission the network rejects.

#### xrpl/hash

- Added `MPToken()` and `MPTokenIssuance()` helpers for computing MPT ledger-entry keylet indexes.

### Changed

#### xrpl/transaction

- `BaseTx.Validate()` rejects an `Account` or `Delegate` that decodes to ACCOUNT_ZERO, and a `Delegate` given as a tagged X-address, which has no companion tag field. Both already failed later, on encode or on the ledger, and now fail in preflight. Consensus-generated pseudo-transactions are exempt from the ACCOUNT_ZERO rule, because the binary codec requires their `Account` to be ACCOUNT_ZERO.

#### xrpl/rpc

- `ErrRawTransactionsFieldMissing`, `ErrRawTransactionFieldMissing`, `ErrCouldNotGetBaseFeeXrp`, `ErrCouldNotFetchOwnerReserve`, `ErrLoanBrokerIDRequired`, and `ErrCouldNotFetchLoanBrokerOwner` now share one value with their `xrpl/websocket` counterparts, so `errors.Is` matches an error raised by either client.
- Deprecated `ErrFeeFieldMissing`, `ErrCounterpartyRequired`, and `ErrFailedToParseFee`. Fee calculation no longer returns them and they will be removed in a future version.
- Autofill fetches the network fee once per transaction instead of once per inner `Batch` transaction, so an eight-transaction `Batch` issues one `server_info` request instead of nine.

#### xrpl/websocket

- `ErrRawTransactionsFieldMissing`, `ErrRawTransactionFieldMissing`, `ErrCouldNotGetBaseFeeXrp`, `ErrCouldNotFetchOwnerReserve`, `ErrLoanBrokerIDRequired`, and `ErrCouldNotFetchLoanBrokerOwner` now share one value with their `xrpl/rpc` counterparts, so `errors.Is` matches an error raised by either client.
- Deprecated `ErrFeeFieldMissing`, `ErrCounterpartyRequired`, and `ErrFailedToParseFee`. Fee calculation no longer returns them and they will be removed in a future version.
- Autofill fetches the network fee once per transaction instead of once per inner `Batch` transaction, so an eight-transaction `Batch` issues one `server_info` request instead of nine.

### Fixed

#### confidential

- Tightened participant and fixed-size proof validation before invoking the native verifier, and aligned proof helper naming with the current `mpt-crypto` contract. A send proof requires exactly the sender, destination, and issuer, followed by the optional auditor.
- Every `Generate*Proof` helper now verifies its own output before returning it, because the native generator reports no error for a mismatched amount or key pair. The check costs one verification per generation, roughly 35% added wall time for a send proof.

#### dependencies

- Raised the minimum Go version from 1.25.12 to 1.25.13 to fix the `net/url` quadratic path-resolution vulnerability (GO-2026-6218).

#### xrpl/rpc

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first. A `Batch` sums its inner fees at the same exact precision.

#### xrpl/transaction

- Corrected confidential transaction key encoding and proof-size validation.
- A `Signers` entry given as a tagged X-address is now rejected with `ErrSignerAccountTagNotAllowed`. The binary codec routes an embedded tag by field name alone, so such an entry previously encoded into a `Signer` object carrying a `SourceTag` the transaction format does not define. A `Signers` entry naming ACCOUNT_ZERO is rejected with `ErrSignerAccountZero`.
- `BaseTx.Validate()` compares `Delegate` to `Account` by decoded `AccountID`, so the same account named once as a classic address and once as an X-address is now recognized as a conflict instead of failing on the ledger.

#### xrpl/websocket

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first. A `Batch` sums its inner fees at the same exact precision.

