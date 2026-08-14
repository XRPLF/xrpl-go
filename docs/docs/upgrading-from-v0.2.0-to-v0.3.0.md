---
sidebar_position: 4
---

# Upgrade from v0.2.0 to v0.3.0

This guide covers the source and behavior changes that are most likely to affect applications that upgrade from `v0.2.0` to `v0.3.0`. See the [v0.3.0 changelog](/changelog/v0.3.x/v0_3_0) for the complete list.

## Go version

`v0.3.0` requires Go `1.25.12` or later.

## Binary codec definitions

The protocol type definitions named `UInt384` and `UInt512` are now `Hash384` and `Hash512`. The obsolete `tecHOOK_REJECTED` and `tecNO_DELEGATE_PERMISSION` transaction result mappings were removed. Update code that reads these definition names or result mappings directly.

## Binary codec quality values

`EncodeQuality` now normalizes nonzero values to a 16-digit mantissa and accepts only normalized exponents from `-96` through `80`. Some extreme values that v0.2.0 accepted, such as `1e-85`, now return `ErrInvalidQuality`.

## Exact XRP amounts and fees

The old `binary-codec/types.MaxDrops` and `currency.DropsPerXrp` values were removed. Use `currency.MaxNativeDrops` and `currency.DropsPerXRP`.

The new `currency.Drops` type keeps native XRP calculations exact. It supports construction from drops or XRP strings, exact arithmetic, comparison, rounding, and formatting. `DropsFromString`, `DropsFromUint64`, and `DropsFromXRP` do not enforce `MaxNativeDrops`. Validate the final whole-drop value against `currency.MaxNativeDrops` before protocol encoding or submission:

```go
fee, err := currency.DropsFromXRP("0.000012")
if err != nil {
 return err
}

adjusted, err := fee.MulDecimal("1.2")
if err != nil {
 return err
}

feeDrops, err := adjusted.Ceil().WholeString()
```

RPC and WebSocket fee configuration changed as follows:

```go
rpc.WithFeeCushion(1.2)       // float64
rpc.WithMaxFeeXRP("2")        // decimal string

cfg := websocket.NewClientConfig().
 WithFeeCushion(1.2).
 WithMaxFeeXRP("2")
```

`common.DefaultFeeCushion` and `websocket.DefaultFeeCushion` are now `float64`. `common.DefaultMaxFeeXRP` and `websocket.DefaultMaxFeeXRP` are decimal strings.

## Key formats and crypto errors

`keypairs.DeriveClassicAddress` now accepts only Ed25519 and compressed secp256k1 public keys. Uncompressed secp256k1 and other unsupported public-key encodings return `ErrInvalidPublicKeyFormat`.

`SECP256K1CryptoAlgorithm.Sign` now returns `crypto.ErrInvalidPrivateKey` directly for malformed hexadecimal private keys. It no longer wraps the hexadecimal decode error. Ed25519 signing continues to preserve its wrapped decode error.

## Network identity

The RPC and WebSocket clients no longer expose a public `NetworkID` field. Use the concurrency-safe snapshot accessor:

```go
networkID, buildVersion := client.NetworkIdentity()
if networkID == nil {
 // Identity discovery has not completed.
}
```

The client gets network identity from `server_info` before identity-dependent autofill or signing. When the server omits `network_id`, the client uses the standard network ID `0`. For a trusted deployment, you can provide both values and bypass discovery:

```go
rpcCfg, err := rpc.NewClientConfig(
 serverURL,
 rpc.WithNetworkIdentity(0, "3.3.0"),
)

wsCfg := websocket.NewClientConfig().
 WithHost(serverURL).
 WithNetworkIdentity(0, "3.3.0")
```

A non-empty build version is required for a complete override. An empty build version makes the client perform discovery.

`websocket.Client.Connect` now requests `server_info` before it starts the background reader. Standard rippled and Clio servers support this request. A custom server, proxy, or test double must return `server_info`, or the client must use `WithNetworkIdentity` with trusted values.

Network identity policy applies to outer transactions and Batch inner transactions. Public network IDs from `0` through `1024` must omit the transaction `NetworkID`. IDs above `1024` require the exact `NetworkID` on rippled `1.11.0` or later.

## X-address autofill

RPC and WebSocket autofill now convert X-addresses in `Account`, `Destination`, `Authorize`, `Unauthorize`, `Owner`, `RegularKey`, `Delegate`, `NFTokenMinter`, `Subject`, `Issuer`, and `Holder`, including Batch inner transactions.

Embedded `Account` and `Destination` tags populate `SourceTag` and `DestinationTag`. A conflicting explicit tag now matches an error sentinel:

```go
if errors.Is(err, rpc.ErrMismatchedTag) {
 // The explicit tag conflicts with the X-address tag.
}
```

`rpc.ErrMismatchedTag` was an exported struct in `v0.2.0`. It is now a sentinel. Replace struct literals and `errors.As` checks with `errors.Is`. A tagged X-address in a field without a tag counterpart returns `ErrAccountIDTagNotAllowed`.

## Signing and submission

`hash.SignTx` and `hash.SignTxBlob` now require a complete, canonical signing form. A single-signed transaction must contain both `SigningPubKey` and `TxnSignature`. A multisigned transaction must contain `Signers` and an explicitly empty top-level `SigningPubKey`. Mixed or partial signing fields are rejected.

RPC and WebSocket submit helpers apply the same checks before submission. `SubmitMultisigned` returns `ErrTransactionNotMultisigned` for another signing form. `ErrSignerDataIsEmpty` remains as a deprecated compatibility alias. `hash.ErrMissingSignature` is now a deprecated alias of `hash.ErrNonSignedTransaction`, so `errors.Is` matches either name.

`SubmitTx` does not enable autofill by default. Set `SubmitOptions.Autofill` to `true` when the client must fill an unsigned transaction. Client-side signing discovers and validates network identity even when autofill is disabled. Use `wallet.Sign` for fully offline signing.

`DeliverMax` normalization is now limited to Payment transactions.

## Reliable submission

`SubmitTxAndWait` and `SubmitTxBlobAndWait` now use validated-ledger finality. They require a positive `LastLedgerSequence` and reject non-positive retry limits or negative polling intervals before submission. `WithMaxRetries` now limits consecutive incomplete rounds caused by query or transport failures. Successful pending rounds do not consume this limit.

An exact `txnNotFound` response is treated as inconclusive until validation, expiry, repeated transport failure, or cancellation. The removed `ErrTransactionNotFound` sentinel is no longer part of the RPC or WebSocket API.

Every validated transaction response returns with a nil error, including a validated `tec` result. Check the validated result explicitly:

```go
response, err := client.SubmitTxBlobAndWait(blob, false)
if err != nil {
 return err
}

result := response.Meta.TransactionResult
```

Use `SubmitTxAndWaitContext` or `SubmitTxBlobAndWaitContext` when caller cancellation must cover preparation, submission, and finality monitoring.

## Dynamic MPT

`v0.3.0` implements the Dynamic MPT contract from rippled `3.3.0`. The unreleased `MutableFlags` model present in `v0.2.0` is replaced by `ImmutableFlags`.

### Issuance creation

Replace `types.MutableFlags`, `TmfMPTCanMutate*`, and `SetMPTCanMutate*` with `types.ImmutableFlags`, `TifMPT*`, and the immutable flag setters. A set immutable bit means that the related capability or field can no longer change.

```go
maximumAmount := types.MPTAmount(1000000)
immutableFlags := types.ImmutableFlags(
 transaction.TifMPTCanLock |
  transaction.TifMPTMetadata,
)

create := transaction.MPTokenIssuanceCreate{
 MaximumAmount: &maximumAmount,
 ImmutableFlags: immutableFlags,
}
```

`MPTokenIssuanceCreate.MaximumAmount` changed from `*types.XRPCurrencyAmount` to `*types.MPTAmount`. It is serialized as a quoted base-10 value and must be from `1` through `2^63-1`.

### Issuance updates

Capability enablement moved to the normal transaction `Flags` field. For example, use `TfMPTSetRequireAuth`, not field `53`, to enable authorization:

```go
set := transaction.MPTokenIssuanceSet{
 BaseTx: transaction.BaseTx{
  Flags: transaction.TfMPTSetRequireAuth,
 },
 MPTokenIssuanceID: issuanceID,
}
```

Use `ImmutableFlags` only to make capabilities or fields permanently immutable. A `Holder`-only `MPTokenIssuanceSet` is a no-op and now fails local validation. Pair `Holder` with `TfMPTLock` or `TfMPTUnlock`.

`MPTokenIssuanceSet`, `MPTokenAuthorize`, and `MPTokenIssuanceDestroy` require an exact 192-bit hexadecimal issuance ID.

### Ledger entries

`MPTokenIssuance.MutableFlags` and `LsmfMPT*` constants were removed. Use `MPTokenIssuance.ImmutableFlags` and `LsifMPT*` constants.

MPT ledger amounts are now quoted base-10 strings. `MPToken.OwnerNode` and `MPTokenIssuance.OwnerNode` are hexadecimal strings.

## Server and Clio response types

Several response fields changed to preserve protocol precision and field presence:

- `server/types.Info.NetworkID` changed from `uint` to `*uint32`.
- The normalized load-factor fields on `server/types.Info` changed from `uint` to `float64`.
- `server/types.Info.LoadFactorFeeEscelation` was renamed to `server/types.Info.LoadFactorFeeEscalation`. These changes do not apply to `server/types.State`, whose load-factor fields remain `uint`.
- `ClosedLedger.BaseFeeXRP` changed from `float32` to `*float64`.
- `ClosedLedgerState.BaseFee` and `ReserveBase` changed from `float32` to `uint64`.
- `ClosedLedgerState.ReserveInc` changed from `float32` to `*uint64`.
- `LedgerState.BaseFee` and `ReserveBase` changed from `uint` to `uint64`.
- `LedgerState.ReserveInc` changed from `uint` to `*uint64`.
- Clio `LedgerInfo.BaseFeeXRP` and `ReserveIncXRP` changed from `float32` to `*float64`.

Check pointer fields before dereferencing them. A nil value means that the response omitted the value or returned null.

## Ledger model fields

The following ledger model changes can require application updates:

- `Escrow.IssuerNode` and `Oracle.OwnerNode` are hexadecimal strings instead of `uint64`.
- `PriceData.AssetPrice` is `*uint64`, so absent and explicit zero values are distinct. Use `ledger.AssetPrice(value)` to create the pointer.
- Oracle `Scale` values through `20` are valid.
- `Oracle` now includes `LedgerEntryType` and `Flags`.
- `MPToken.MPTAmount`, `MPToken.LockedAmount`, and MPT issuance amount fields are quoted base-10 strings.

For example:

```go
priceData := ledger.PriceData{
 BaseAsset:  "XRP",
 QuoteAsset: "USD",
 AssetPrice: ledger.AssetPrice(740),
}
```

## Batch transactions

A Batch transaction now requires from 2 through 8 inner transactions. `ErrBatchRawTransactionsEmpty` remains a compatibility alias for `ErrBatchRawTransactionsCount`.

Inner transaction maps must contain an explicitly empty `SigningPubKey`. `TxnSignature`, `Signers`, and `LastLedgerSequence` must be absent. Null values do not replace required omission or an empty string.

Batch signing now uses the `BatchV1_1` payload. Code that calls `binarycodec.EncodeForSigningBatch` directly must add the outer account and effective sequence:

```go
payload, err := binarycodec.EncodeForSigningBatch(map[string]any{
 "account":  outerAccount,
 "sequence": effectiveSequence,
 "flags":    uint32(flags),
 "txIDs":    txIDs,
})
```

For a ticketed Batch, `effectiveSequence` is the ticket sequence. `batchAccount` binds an outer Batch signer. `signerAccount` binds a nested multisigner and is valid only when `batchAccount` is also present.

The public `wallet/types.BatchSignable` struct now includes `Account`, `Sequence`, `BatchAccount`, and `SignerAccount`. Its `Equals` method now compares the outer account and effective sequence in addition to flags and transaction IDs. The new constructor sentinels are `ErrAccountFieldIsNotAString`, `ErrSequenceFieldIsNotAnUint32`, `ErrTicketSequenceFieldIsNotAnUint32`, `ErrBatchSequenceAndTicket`, and `ErrBatchSequenceNotSet`. Old unkeyed struct literals no longer compile. Old keyed literals that omit `Account` or `Sequence` compile but cannot create a valid `BatchV1_1` payload.

`wallet.SignMultiBatch` requires the outer `Account` and exactly one nonzero `Sequence` or `TicketSequence` before signing. Create all Batch signature fragments again after the upgrade. Do not combine fragments created by v0.2.0 with fragments created by v0.3.0.

## WebSocket lifecycle

Pending requests now return `ErrDisconnected` when their connection closes or a write fails. Requests are not replayed after reconnect. `Connect` no longer replaces an active or in-progress connection and returns `ErrAlreadyConnected` instead.

Stream handlers run in lifecycle-bound goroutines. A handler is serialized with itself, including across disconnect and reconnect lifecycles. Different handlers can run concurrently. Registering a new handler replaces the previous handler for that stream. Do not call `Connect` synchronously from a stream or error handler. Automatic reconnect does not replay subscriptions, so applications must resubscribe.

`ErrMaxReconnectionAttemptsReached` has a new exported `Err` field and unwraps the last connection failure. Replace unkeyed literals:

```go
err := websocket.ErrMaxReconnectionAttemptsReached{
 Attempts: 3,
}
```

## RPC authorization

RPC configuration now rejects authorization over plaintext transport and authenticated HTTPS-to-HTTP redirects. A nil custom HTTP client returns `ErrNilHTTPClient`. Diagnostic errors redact URL user information and authorization credentials.

Custom `HTTPClient` implementations are supported for HTTPS endpoints, but they control their own redirects and credentials.

## Integration test clients

The `xrpl/testutil/integration.Client` interface adds `NetworkIdentity`, `GetServerDefinitions`, and `Simulate`. External test clients and generated mocks that implement this interface must add these methods.

`NewRunner` now replaces zero `WalletCount` and `MaxRetries` values with defaults and modifies those fields in the supplied `RunnerConfig`. `WithWallets(0)` and `WithMaxRetries(0)` no longer preserve explicit zero values.

## DelegateSet

A present empty `DelegateSet.Permissions` list now deletes the Delegate object. An absent list remains invalid. `Batch` permissions and Vault and Loan transaction permissions are rejected as non-delegable.

## Error sentinels

Several errors are new or changed in `v0.3.0`.

Update references to these removed errors:

- Replace `binarycodec.ErrBatchTxIDNotString` with `binarycodec.ErrBatchTxIDsNotArray`.
- Replace `transaction.ErrMPTIssuanceCreateMutableFlagsZero` with `transaction.ErrMPTIssuanceCreateImmutableFlagsZero`.
- Replace `transaction.ErrMPTIssuanceSetMutableFlagsZero` with `transaction.ErrMPTIssuanceSetImmutableFlagsZero`.
- Remove checks for `transaction.ErrMPTIssuanceSetMutableFlagsConflict` and `transaction.ErrMPTIssuanceSetTransferFeeWithClearCanTransfer`. The old set and clear mutable-flag model no longer exists.

The following errors are new or have new behavior:

- `ErrInvalidPrivateKeyFormat` and `ErrInvalidPublicKeyFormat` report key format failures and preserve `ErrInvalidCryptoImplementation` matching.
- `ErrInvalidFeeValue` and `ErrFeeHasTooManyDecimals` report invalid fee configuration.
- `ErrNetworkIDUnavailable`, `ErrBuildVersionUnavailable`, `ErrInvalidBuildVersion`, `ErrNetworkIDFieldUnexpected`, and `ErrNetworkIDOverrideMismatch` report network identity failures.
- `ErrPreliminaryResult`, `ErrTransactionExpired`, `ErrFinalityTransport`, `ErrInvalidPollInterval`, `ErrInvalidMaxRetries`, and `ErrInvalidLastLedgerSequence` report reliable-submission failures.
- `ErrNilTransaction` reports nil client transaction inputs.
- `ErrLastLedgerSequenceFieldMustBeAbsent` reports an invalid Batch inner transaction.

Shared Batch structure and NetworkID validation errors now use the same sentinel identity in the RPC and WebSocket packages. An `errors.Is` check can match the corresponding sentinel from either package.

## Protocol and query additions

`v0.3.0` adds typed `server_definitions`, XLS-69 `simulate`, expanded `ledger_entry` selectors and response forms, query field coverage, MPT Clawback support, and typed `bookChanges` fields. These are additive, but applications should use the new typed request validation instead of constructing unvalidated payloads when possible.

Account lines add the `ignore_default` request field and `limit` response field. AMM info adds account lookup, ledger selection, frozen flags, and auction `time_interval`. NFT offer queries add `limit` and `marker`. Vault and v1 account NFT responses add current-ledger metadata.
