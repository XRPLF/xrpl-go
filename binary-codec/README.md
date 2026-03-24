# Binary Codec

This package contains functions to encode/decode to/from the [ripple binary serialization format](https://xrpl.org/serialization.html).

## Overview

XRPL nodes communicate using a compact binary format rather than JSON. Before a transaction can be signed or submitted, it must be serialized into this format. The `binarycodec` package handles that transformation in both directions.

**Encoding** converts a JSON transaction object into a canonical binary blob:
- Fields are sorted by a protocol-defined ordinal (type code + field code).
- Each field is prefixed with a compact header identifying its type and position.
- Variable-length fields (blobs, account IDs) are preceded by a length prefix.
- The resulting hex string is what gets signed and submitted to the network.

**Decoding** reverses the process, turning a hex-encoded binary blob back into a JSON object.

There are several encoding variants depending on the use case:

| Function | Use case |
|---|---|
| `Encode` | Produce the full transaction blob for submission |
| `Decode` | Parse a binary blob back to JSON |
| `EncodeForSigning` | Produce the payload that the private key signs (excludes `TxnSignature`) |
| `EncodeForMultisigning` | Like `EncodeForSigning` but appends the signing account ID |
| `EncodeForSigningClaim` | For payment channel claims |
| `EncodeQuality` / `DecodeQuality` | Encode offer quality (exchange rate) values |
| `DecodeLedgerData` | Parse raw ledger state data |

## Package Structure

The codec is split into three sub-packages, each with a distinct responsibility:

### `definitions/`

The schema registry for the entire codec. At startup it embeds and parses `definitions.json` — the authoritative XRPL protocol document — into a singleton `Definitions` struct. Everything the serializer and parser need to know about a field lives here:

- **`Types`** — maps type names (e.g. `"UInt32"`, `"Amount"`) to their numeric type codes.
- **`Fields`** — maps field names (e.g. `"Fee"`, `"Destination"`) to a `FieldInstance`, which contains:
  - `FieldHeader` (`TypeCode` + `FieldCode`): the binary identity of the field written into the encoded stream.
  - `Ordinal` (`TypeCode<<16 | FieldCode`): used to sort fields into canonical order before encoding.
  - `IsVLEncoded`: whether the field value is preceded by a variable-length prefix.
  - `IsSerialized`: whether the field is included in the encoded blob at all.
  - `IsSigningField`: whether the field is included when computing the signing payload (`EncodeForSigning`). Fields like `TxnSignature` are excluded here.
- **`TransactionTypes`**, **`TransactionResults`**, **`LedgerEntryTypes`** — numeric code mappings for each category.
- **`DelegatablePermissions`** / **`GranularPermissions`** — permission value mappings used for account delegation features.

### `serdes/`

Contains the low-level binary read/write primitives:

- **`BinarySerializer`** — accumulates bytes into a sink. For each field it writes the field header (via `FieldIDCodec`), an optional variable-length prefix (for VL-encoded fields), and then the raw value bytes. Appends `0xE1` as the `STObject` end marker when needed.
- **`BinaryParser`** — the inverse: reads a hex-encoded stream, decodes field headers back to field names, and hands off to the appropriate type deserializer.
- **`FieldIDCodec`** — encodes/decodes the compact field header bytes that prefix every field in the binary format.

### `types/`

One file per XRPL serialization type. Each type implements the `SerializedType` interface:

```go
type SerializedType interface {
    FromJSON(json any) ([]byte, error)
    ToJSON(parser BinaryParser, opts ...int) (any, error)
}
```

| Type | Description |
|------|-------------|
| `UInt8`, `UInt16`, `UInt32`, `UInt64` | Fixed-width unsigned integers |
| `Int32` | Fixed-width signed integer |
| `Hash128`, `Hash160`, `Hash192`, `Hash256` | Fixed-length byte arrays for hashes and addresses |
| `AccountID` | 20-byte account address (Base58Check encoded in JSON, raw bytes in binary) |
| `Amount` | XRP drops (64-bit) or issued currency amounts (special 64-bit float encoding) |
| `Blob` | Variable-length byte array (VL-encoded) |
| `Currency` | 160-bit currency representation |
| `Issue` | Currency + issuer pair |
| `STObject` | Nested object; fields sorted by ordinal, terminated with `0xE1` |
| `STArray` | Array of `STObject` entries, terminated with `0xF1` |
| `PathSet` | Payment path data |
| `Vector256` | Array of `Hash256` values |
| `XChainBridge` | Cross-chain bridge descriptor |
| `Number` | Arbitrary-precision number (STNumber) |

`GetSerializedType(typeName string)` acts as the factory, returning the right implementation for a given type name from the definitions.

## API

### Encode

```go
encoded, err := binarycodec.Encode(jsonObject)
```

### Decode

```go
json, err := binarycodec.Decode(hexEncodedString)
```
### EncodeForMultisigning

```go
encoded, err := binarycodec.EncodeForMultisigning(jsonObject, xrpAccountID)
```

### EncodeForSigning

```go
encoded, err := binarycodec.EncodeForSigning(jsonObject)
```

### EncodeForSigningClaim

```go
encoded, err := binarycodec.EncodeForSigningClaim(jsonObject)
```

### EncodeQuality

```go
encoded, err := binarycodec.EncodeQuality(amountString)
```

### DecodeQuality

```go
decoded, err := binarycodec.DecodeQuality(encoded)
```

### DecodeLedgerData

```go
ledgerData, err := binarycodec.DecodeLedgerData(hexEncodedString)
```
