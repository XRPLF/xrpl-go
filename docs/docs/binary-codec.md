# Binary codec

Use `binarycodec` to convert XRP Ledger objects between field maps and the canonical binary format. The package also prepares signing payloads and decodes binary ledger data.

[Install the SDK](/docs/installation) · [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/binary-codec)

For normal transaction signing, use the [wallet package](/docs/xrpl/wallet). It prepares the signing payload and encodes the signed transaction for you.

## Encode and decode an object

This offline example encodes an object containing a sequence number, then decodes it. It demonstrates serialization, not a complete transaction.

After installing the SDK, save this as `main.go` and run `go run .`:

```go
package main

import (
	"fmt"
	"log"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
)

func main() {
	object := map[string]any{"Sequence": uint32(1)}

	blob, err := binarycodec.Encode(object)
	if err != nil {
		log.Fatal(err)
	}
	decoded, err := binarycodec.Decode(blob)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Encoded:", blob)
	fmt.Println("Sequence:", decoded["Sequence"])
}
```

Expected output:

```text
Encoded: 2400000001
Sequence: 1
```

`Encode` returns a hex string. `Decode` returns a `map[string]any`, not an automatically selected transaction or ledger-entry struct.

## Choose an encoding

| Task | Function |
| --- | --- |
| Serialize a full object or transaction | `Encode` |
| Decode a serialized object | `Decode` |
| Prepare a single-signing payload | `EncodeForSigning` |
| Prepare a multisigning payload | `EncodeForMultisigning` |
| Prepare a payment-channel claim payload | `EncodeForSigningClaim` |
| Convert offer-quality values | `EncodeQuality`, `DecodeQuality` |
| Decode binary ledger state | `DecodeLedgerData` |

See the [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/binary-codec) for all encoders, including sponsor, counterparty, and batch signing.

## Signing payloads and submission blobs

These are different representations:

```text
transaction fields -> signing encoder -> payload to sign
transaction fields + signature -> Encode() -> blob to submit
```

Do not use a full submission blob as the signing payload. Use the encoder for the signing mode, or let the wallet methods handle the complete process.

Successful encoding does not mean a transaction is valid for submission. The codec serializes fields. It does not establish that the transaction can succeed against the current ledger state.

## Binary responses

Normal JSON responses do not need binary decoding. If you request a binary ledger entry, decode its `node_binary` value with `Decode` to obtain a field map. See [ledger data](/docs/xrpl/ledger-entry-types) for the typed objects those fields describe.

## Protocol definitions

The codec uses embedded definitions to order fields, select types, and decide which fields belong in a signing payload. Encoding adds field headers and length prefixes where needed. Decoding reverses that process.

Definitions describe what a node build can parse, not which amendments its network has activated. A field appearing in the definitions does not establish that it can be used on Mainnet.

See the [serialization specification](https://xrpl.org/serialization.html) for format details. Maintainers can follow the [definition update instructions](https://github.com/XRPLF/xrpl-go/blob/main/binary-codec/README.md#update-protocol-definitions).
