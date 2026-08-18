---
title: Unreleased
---

### Added

#### address-codec

- Added `DecodeAddress()`, which resolves a classic address or an X-address to the `AccountID` both forms encode, along with the classic spelling, whether an X-address carried a tag, and whether it was encoded for a test network. It is the decoding counterpart to `IsValidAddress()`, which already accepts either form, and lets callers compare accounts by decoded `AccountID` so a classic address and its X-address form are recognized as the same account. Adds `IsZeroAccountID()`, which reports ACCOUNT_ZERO only for a 20-byte identifier and false for any other length.

#### binary-codec

- Synced the embedded `definitions.json` with the rippled 3.3.0 protocol definitions.
- Added XLS-96 confidential MPT fields and transaction definitions for `ConfidentialMPTSend`, `ConfidentialMPTConvert`, `ConfidentialMPTConvertBack`, `ConfidentialMPTMergeInbox`, and `ConfidentialMPTClawback`, with `ConfidentialOutstandingAmount` serialized as a decimal string like the other MPT amount fields.
- Added the `tecBAD_PROOF`, `tecNO_SPONSOR_PERMISSION`, `temBAD_CIPHERTEXT`, `tefBAD_PATH_COUNT`, `tefNO_DST_PARTIAL`, and `terNO_PERMISSION` transaction results, which previously could not be encoded by name.

#### confidential

- Added CGo bindings and vendored native libraries for XRPLF `mpt-crypto`, with a `!cgo` fallback and maintainer tooling for dependency updates.
- Added hex-string APIs for ElGamal encryption, Pedersen commitments, context hashes, and zero-knowledge proof generation and verification.
- Proof context hashes bind the decoded AccountID, so a proof matches whether the caller supplied a classic address or its X-address form.
- Added `test-confidential` and `update-mpt-crypto` Makefile targets and an automated dependency-update workflow.

#### confidential/builder

- Added online `Build*` and offline `Prepare*` helpers for confidential MPT send, convert, convert-back, clawback, and inbox-merge transactions. See the [confidential builders guide](https://xrplf.github.io/xrpl-go/docs/confidential/builders) for the full parameter and error reference.
- Builder parameters are validated before any ledger query or proof work: addresses may be classic or X-addresses and are normalized to their classic form for ledger lookups and keylet computation, ACCOUNT_ZERO is rejected because it can never sign, the issuance issuer is rejected from holder roles and from a send destination, private keys must be usable secp256k1 scalars, and amounts are bounded by the protocol maximum.
- `Build*` helpers preflight what the network enforces, so a doomed transaction never costs a fee and a sequence: the issuance capabilities, and the ledger state each confidential MPT transactor requires. `BuildClawback` rejects an issuance without `lsfMPTCanClawback` and requires the holder encryption key alongside the issuer mirror balance. `BuildConvertBack` rejects an amount above the confidential supply and requires the issuer mirror balance on the holder, and `BuildSend` requires it on both participants plus a registered key and an inbox on the destination. Both require the auditor mirror balance on every account they read when the issuance registers an auditor key. `BuildConvert` rejects an amount above the holder's public `MPTAmount`, which `ConfidentialMPTConvert` debits. `BuildMergeInbox` requires the holder encryption key but not the issuer one, which `ConfidentialMPTMergeInbox` does not use. Every builder except `BuildClawback` also rejects a locked issuance, a locked MPToken, and a holder an issuance with `lsfMPTRequireAuth` has not authorized, mirroring the `checkFrozen` and `requireAuth` checks each transactor ends its preclaim with. The clawback runs neither, so a locked holder stays clawable, and an issuance authorizing through a permissioned domain is left to the network, because the credentials it accepts are not read here. Adds `ErrConfidentialDisabled`, `ErrTransferDisabled`, `ErrTransferFeeSet`, `ErrClawbackDisabled`, `ErrIssuanceNotFound`, `ErrKeyMismatch`, `ErrIssuanceLocked`, `ErrHolderLocked`, `ErrHolderNotAuthorized`, and `ErrAmountExceedsOutstanding`.
- Online builders read every ledger entry from one validated ledger, pinned by hash after the first read, and reject a malformed or inconsistent snapshot response. The account sequence is read from the open ledger instead, so a build that follows a submission the network has not validated yet does not reuse a sequence. Because the proofs that consume a balance version bind it into their context hash, `BuildSend` and `BuildConvertBack` reread the holder's `ConfidentialBalanceVersion` from the open ledger and fail with `ErrStaleBalanceVersion` when a transaction still in flight has already moved it, rather than paying a fee for a `tecBAD_PROOF`. Adds `ErrInvalidLedgerState` and `ErrStaleBalanceVersion`.
- `Prepare*` helpers validate the transaction they assemble before returning it, so a field set wrongly is reported as `ErrInvalidTransaction` instead of surfacing at submission.
- `BuildSendParams` carries an optional `DestinationTag`, and `CredentialIDs` use `types.CredentialIDs` and are validated before use, reported through `ErrInvalidCredentialIDs`, which wraps `transaction.ErrInvalidCredentialIDs` so a caller matching the builder error set does not have to import `xrpl/transaction`.
- A tagged X-address is accepted only where the transaction has a companion tag field. `Destination` accepts one unless an explicit `DestinationTag` is also set, and `Holder` rejects one because `ConfidentialMPTClawback` has no tag field to carry it. Both wrap the field sentinel around the `xrpl/transaction` condition sentinel (`ErrAccountIDTagNotAllowed`, `ErrDuplicateXAddressTag`), so `errors.Is` matches the field and the condition, and the builder shares one condition identity with the encoder.
- `ErrInvalidAccount`, `ErrInvalidDestination`, and `ErrInvalidHolder` wrap the reason the address was rejected, so the `address-codec` decode failure or `transaction.ErrZeroAccountID` stays matchable with `errors.Is`. Adds `ErrInvalidAddress` for the keylet helper, which resolves an MPToken for an `Account`, a `Destination`, or a `Holder` depending on the caller and so names no field.
- `BuildClawback` derives the clawback amount by decrypting the holder's `IssuerEncryptedBalance`, bounded by `BalanceRange` and capped at the issuance `ConfidentialOutstandingAmount`. `Amount` moved from `BuildClawbackParams` to `ClawbackParams`, so it is supplied only on the offline `PrepareClawback` path.
- `BuildConvert` reports a missing holder `MPToken` as `ErrMPTokenNotFound` rather than treating it as a first-time opt-in, matching `ConfidentialMPTConvert`, which debits the entry and so requires it to exist. A failed read reports `ErrLedgerQuery`, so a transport error can no longer be mistaken for first-time state.
- Proof-bearing `Prepare*` helpers reject options that carry neither a `Sequence` nor a `TicketSequence` with `ErrMissingSequence`, because each proof binds the nonce and a later autofill would invalidate it. `PrepareMergeInbox` and a repeat `PrepareConvert` carry no proof and still accept a zero nonce.
- Every `Build*Params` embeds `TxOptions`, which carries `Sequence`, `TicketSequence`, and `Delegate`. A build reads the account sequence only when both nonce fields are zero. Adds `ErrConflictingNonce` for setting both, and `ErrDelegateNotAllowed` for a delegate on a type `NonDelegatableTransactionsMap` lists, which among the confidential types is `ConfidentialMPTConvert`. Every helper accepts a ticket, and because xrpld hashes the sequence proxy, a proof commits to the ticket the transaction spends rather than to an account sequence: a first-time `PrepareConvert`, `PrepareClawback`, `PrepareSend`, and `PrepareConvertBack` all bind it. The send and convert-back proofs also commit to the submitter's own `ConfidentialBalanceVersion`, which a send, a convert-back, a merge-inbox, or a clawback against that holder bumps, so of several built against a single reading of one `MPToken` the first to land strands the rest on stale proofs, each paying a fee and a destroyed ticket for a `tecBAD_PROOF`. That collision is per `MPToken` and no single build can observe it, so it is documented rather than refused.

#### docs

- Added confidential MPT documentation covering the CGo requirement, package layout, transaction types, transaction cost, and high-level builders.
- Added `examples/confidential`, which assembles a confidential MPT opt-in and inbox merge offline, without connecting, signing, or submitting.

#### pkg/crypto

- Added `IsCompressedSECP256K1Point()` to report whether a hex string decodes to a compressed secp256k1 point that lies on the curve, and `CompressedSECP256K1PointByteLength` so callers can size a compressed point without importing the curve library.

#### pkg/hexutil

- Added `DecodeFixedHex()` to decode hexadecimal values and enforce their decoded byte length.

#### pkg/mptsizes

- Added `mptsizes`, a CGo-free package holding the XLS-96 confidential MPT wire sizes that the vendored `mpt-crypto` headers define. The transaction models and the CGo bindings both derive their lengths from it, so a proof-format bump cannot leave the two disagreeing.

#### xrpl

- Added `ErrZeroAccountID`, `ErrAccountZero`, `ErrDelegateZero`, `ErrDelegateTagNotAllowed`, `ErrSignerAccountZero`, and `ErrSignerAccountTagNotAllowed` for the address conditions `BaseTx.Validate()` now reports. Each is a fixed sentinel that names its field and wraps its condition, so `errors.Is` matches either one and a direct comparison against the returned value still works.
- Added `ErrAccountIDTagNotAllowed` and `ErrDuplicateXAddressTag`, which alias the `binary-codec/types` sentinels so preflight and encoding report one error identity for these conditions. `ErrClawbackHolderTagNotAllowed` and `ErrConfidentialClawbackHolderTagNotAllowed` now wrap `ErrAccountIDTagNotAllowed` for the same reason, which appends the wrapped reason to their message text.
- Added confidential-transfer flags and encryption-key fields to MPT issuance transaction and ledger-entry models.
- Added confidential balance fields to `MPToken`, five confidential MPT transaction models, and supporting amount, encryption-key, blinding-factor, hex-blob, ciphertext, commitment, and proof validation helpers.
- RPC and WebSocket autofill apply the required 10x base fee to confidential MPT transactions, including inner `Batch` transactions, plus the normal per-signer surcharge.
- Added `IsMPTokenIssuer()` to report whether an address is the issuer encoded in an MPT issuance ID. Confidential MPT self-send and self-clawback checks compare decoded `AccountID` values, so an X-address naming the submitting account is rejected; a `Holder` X-address carrying a tag is rejected because `ConfidentialMPTClawback` has no tag field to hold it; a `Destination` or `Holder` naming ACCOUNT_ZERO is rejected because it can never hold the MPToken, with the field sentinel wrapping `ErrZeroAccountID`; and an `Account` X-address tag combined with an explicit `SourceTag` is rejected before encoding rather than during it.
- `ConfidentialMPTSend` now carries an optional `DestinationTag`, matching rippled's transaction format, and rejects it when the `Destination` X-address already embeds a tag.
- `ConfidentialMPTConvert` is now listed as non-delegatable, matching the protocol, so `DelegateSet` no longer accepts a permission the network rejects.

#### xrpl/hash

- Added `MPToken()` and `MPTokenIssuance()` helpers for computing MPT ledger-entry keylet indexes.

### Changed

#### xrpl/transaction

- `BaseTx.Validate()` rejects an `Account` or `Delegate` that decodes to ACCOUNT_ZERO, and a `Delegate` given as a tagged X-address, which has no companion tag field to carry the tag. Both already failed later, on encode or on the ledger, and now fail in preflight. Consensus-generated pseudo-transactions (`EnableAmendment`, `SetFee`, `UNLModify`) are exempt from the ACCOUNT_ZERO rule, because the binary codec requires their `Account` to be ACCOUNT_ZERO. An `Account` pairing a tagged X-address with an explicit `SourceTag` is deliberately still accepted, because client autofill rewrites the address to its classic form and carries the tag across. `Clawback` and the confidential MPT transactions, which resolve the tag themselves, continue to reject it.

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

- Tightened participant and fixed-size proof validation before invoking the native verifier, and aligned proof helper naming with the current `mpt-crypto` contract. Send proofs require exactly the sender, destination, and issuer, followed by the optional auditor, and any other count is rejected with `ErrInvalidParticipantCount`.
- Every `Generate*Proof` helper now verifies its own output before returning it. The native generator reports no error for a mismatched amount or key pair, so a bad proof previously surfaced only when the network rejected the transaction. The check costs one verification per generation, roughly 35% added wall time for a send proof, which the confidential documentation records for callers that batch proof generation.

#### dependencies

- Raised the minimum Go version from 1.25.12 to 1.25.13 to fix the `net/url` quadratic path-resolution vulnerability (GO-2026-6218).

#### xrpl/rpc

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first. A `Batch` sums its inner fees at the same exact precision, so a fractional drop is no longer discarded per inner transaction.

#### xrpl/transaction

- Corrected confidential transaction key encoding and proof-size validation.
- A `Signers` entry given as a tagged X-address is now rejected with `ErrSignerAccountTagNotAllowed`. The binary codec routes an embedded tag by field name alone, so such an entry previously encoded without error into a `Signer` object carrying a `SourceTag` the transaction format does not define. A `Signers` entry naming ACCOUNT_ZERO is rejected with `ErrSignerAccountZero`, because no keypair can produce it and the entry can never be signed.
- `BaseTx.Validate()` compares `Delegate` to `Account` by decoded `AccountID`, so the same account named once as a classic address and once as an X-address is now recognized as a conflict instead of passing preflight and failing on the ledger.

#### xrpl/websocket

- Autofill no longer underpays transactions whose fee is a multiple of the base fee. The multiplier now applies to the exact load-adjusted network fee and the total is rounded once, instead of rounding the network fee to whole drops first. A `Batch` sums its inner fees at the same exact precision, so a fractional drop is no longer discarded per inner transaction.
