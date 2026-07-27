---
sidebar_position: 3
---

# Upgrade from v0.1.x to v0.2.0

This guide covers the source changes most likely to affect applications upgrading from `v0.1.x` to `v0.2.0`.

## Keypairs

`keypairs.GenerateSeed` now accepts raw entropy bytes instead of a string:

```go
seed, err := keypairs.GenerateSeed(nil, crypto.SECP256K1(), random.NewRandomizer())
```

When entropy is provided, it must be exactly 16 raw bytes:

```go
entropy := []byte{
	0x00, 0x01, 0x02, 0x03,
	0x04, 0x05, 0x06, 0x07,
	0x08, 0x09, 0x0A, 0x0B,
	0x0C, 0x0D, 0x0E, 0x0F,
}

seed, err := keypairs.GenerateSeed(entropy, crypto.ED25519(), nil)
```

Do not pass passphrases directly. For deterministic passphrase-based seed generation, derive exactly 16 bytes outside `GenerateSeed`:

```go
sum := sha512.Sum512([]byte(passphrase))
seed, err := keypairs.GenerateSeed(sum[:addresscodec.FamilySeedLength], crypto.ED25519(), nil)
```

Migration only: older versions used the first 16 bytes of any non-empty entropy string. If you must recover a legacy seed, reproduce that truncation before calling `GenerateSeed`:

```go
legacyEntropy := []byte(oldEntropy)
if len(legacyEntropy) < addresscodec.FamilySeedLength {
	return errors.New("legacy entropy was shorter than 16 bytes")
}

seed, err := keypairs.GenerateSeed(legacyEntropy[:addresscodec.FamilySeedLength], crypto.ED25519(), nil)
```

## X-address Tags

`DecodeXAddress` and `XAddressToClassicAddress` now return `hasTag` so callers can distinguish no tag from explicit tag `0`:

```go
accountID, tag, hasTag, testnet, err := addresscodec.DecodeXAddress(xAddress)
classic, tag, hasTag, testnet, err := addresscodec.XAddressToClassicAddress(xAddress)
```

When encoding or signing a transaction, do not provide both an embedded X-address tag and a separate `SourceTag` or `DestinationTag`. The binary codec now rejects duplicate tag data, including the case where both values are `0`.

`AccountID.FromJSON` now rejects any X-address that carries an embedded tag with `ErrAccountIDTagNotAllowed`, including non-`Account`/`Destination` fields such as `SignerEntry.Account` and inputs to `EncodeForMultisigning`. Previously the tag was silently dropped. Pass the classic address (without an embedded tag) for these fields.

`wallet.ErrAddressTagNotZero` was renamed to `wallet.ErrAddressHasTag`. The new sentinel applies to any embedded X-address tag, including explicit tag `0`.

## Amounts

`binary-codec` no longer accepts `float64` values for `Amount` serialization. Use strings, `json.Number`, or exact amount types instead:

```go
amount := map[string]any{
	"currency": "USD",
	"issuer":   issuer,
	"value":    "12.5",
}
```

If you decode JSON before passing values to the codec, preserve numbers with `UseNumber`:

```go
decoder := json.NewDecoder(reader)
decoder.UseNumber()
```

Native XRP amount serialization now validates drops with exact integer bounds. `types.MaxDrops` is a typed `uint64`, and `types.MinXRP` was removed because serialization validates drops, not XRP-denominated decimal values.

## UInt64 Fields

`UInt64.FromJSON` now accepts only 1 to 16 character hex strings. Decimal-looking strings are still interpreted as hex:

```go
// Encodes the value 16, not decimal 10.
raw := "10"
```

Code that previously passed decimal strings should convert to hex first. `ErrUInt64OutOfRange` was removed, and invalid inputs now return `ErrInvalidUInt64String`.

## Signers and Multisigning

`xrpl.SortSigners` now returns an error:

```go
if err := xrpl.SortSigners(signers); err != nil {
	return err
}
```

Signer ordering now uses decoded account ID bytes instead of classic address string ordering. This matches canonical XRPL signer ordering and affects `Multisign`, `CombineLoanSetCounterpartySigners`, and `CombineBatchSigners`.

## Transactions

Loan transaction `Flatten()` methods now return `transaction.FlatTransaction`, matching the rest of the transaction package:

```go
flatTx := loanSet.Flatten()
```

If your code explicitly used `map[string]any`, update the type:

```go
var flatTx transaction.FlatTransaction = loanSet.Flatten()
```

The exported `DomainIDLength` and `SHA512HalfLength` constants were removed. Use `Hex256Length`, `IsHex256`, `IsDomainID`, or `IsLedgerEntryID` depending on whether the code needs raw 256-bit hex validation or semantic ledger-entry validation.

## Balance Changes

`transaction.GetBalanceChanges` no longer returns an error when an affected `AccountRoot` or `RippleState` node carries no balance change. Previously a single balance-neutral affected node aborted the whole computation, those nodes are now skipped, and affected nodes whose net balance delta is zero are skipped instead of being emitted as `0`-value entries.

```go
changes, err := transaction.GetBalanceChanges(meta)
```

If your code relied on that error to detect balance-neutral metadata, it will no longer fire, inspect the returned slice instead. Genuinely malformed balance values still return an error.

## Clients

RPC responses are capped at 64 MiB by default, and WebSocket inbound messages are capped at 16 MiB by default. Set the max response size to `0` only when you deliberately need to disable the cap:

```go
rpcCfg, err := rpc.NewClientConfig(url, rpc.WithMaxResponseSize(0))
wsCfg := websocket.NewClientConfig().WithMaxResponseSize(0)
```

Remote non-TLS client URLs now emit SDK warnings with userinfo redacted. Use `rpc.SetLogger` or `websocket.SetLogger` to override or silence those warnings.

WebSocket stream handlers now run under lifecycle-bound handler goroutines. Do not call `Connect` synchronously from stream or error handlers.

## Flags After Autofill

After `Autofill` (and any direct call to `FlatTransaction.NormalizeFlags`), the `Flags` entry in a `FlatTransaction` is always stored as `uint32`. A missing `Flags` field is defaulted to `uint32(0)`, and any present integer, whole-number float, or `json.Number` value is coerced to `uint32` when it fits in `[0, 4294967295]`.

Callers that previously relied on the original Go type of a present `Flags` value (for example, asserting `flatTx["Flags"].(int)` after autofill) must update their assertions to `uint32`:

```go
flags, ok := flatTx["Flags"].(uint32)
```

Values that fall outside the `uint32` range now return the new `transaction.ErrInvalidFlagsValue` sentinel from `NormalizeFlags`.

## Error Sentinels

`ErrTransactionTypeMissing` no longer lives in the `rpc` or `websocket` packages. Use the canonical sentinel from the `transaction` package:

```go
import "github.com/Peersyst/xrpl-go/xrpl/transaction"

if errors.Is(err, transaction.ErrTransactionTypeMissing) {
    // ...
}
```

`binary-codec` removed the exported `ErrInvalidJSONNumber`. `PermissionValue.FromJSON` now returns `ErrPermissionValueOutOfRange` for any `json.Number` input that cannot be coerced to a `uint32` in the `[0, 4294967295]` range (including malformed, fractional, or negative values that previously surfaced as `ErrInvalidJSONNumber`):

```go
if errors.Is(err, types.ErrPermissionValueOutOfRange) {
    // ...
}
```

## Issued-Currency Values

`binary-codec` and `xrpl/transaction` now validate issued-currency `value` strings as XRPL String Numbers at both encode time (`VerifyIOUValue`, `SerializeIssuedCurrencyValue`) and validation time (`IsIssuedCurrency`). Inputs that previously slipped through are now rejected:

- `NaN`, `Inf`, `+Inf`, `-Inf`
- Hex floats such as `0x1p10`
- Prefixed or suffixed strings
- Leading-zero mantissas such as `-000.2345` or `00.5`
- Incomplete exponents like `1e`, `1e+`, `1e-`
- Out-of-range exponents such as `1e1000`

Canonical zero forms (`"0"`, `"0.0"`, `"-0"`, `"0e5"`, etc.) are accepted as valid token amounts, `SerializeIssuedCurrencyValue` emits the XRPL zero encoding (`0x8000000000000000`) for them. Negative amounts remain rejected.

`IsIssuedCurrency` now returns errors that wrap both `ErrInvalidTokenValue` and the underlying binary-codec cause via `errors.Is`, so existing checks against `ErrInvalidTokenValue` keep matching:

```go
if errors.Is(err, transaction.ErrInvalidTokenValue) {
    // ...
}
```
