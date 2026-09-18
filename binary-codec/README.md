# Binary codec

Encode and decode XRP Ledger objects in the canonical binary format. Use this package for transaction serialization, signing payloads, and binary ledger data.

[Installation](../README.md#quick-start) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/binary-codec) · [Serialization format](https://xrpl.org/serialization.html)

For normal transaction signing and submission, start with [`xrpl/wallet` and the clients](../xrpl/README.md#build-and-send-transactions). They call the codec for you.

## Encode and decode an object

This offline example encodes an object containing a sequence number, then decodes it. It demonstrates serialization, not a complete transaction.

Save it as `main.go` in your application after installing the SDK, then run `go run .`:

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

A signing payload is not the final transaction blob. Use the appropriate signing encoder to prepare the payload, then `Encode` to serialize the transaction with its signature fields. Successful encoding alone does not mean a transaction is valid for submission.

See the [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/binary-codec) for all encoders, including sponsor, counterparty, and batch signing.

## How the codec is organized

Encoding sorts fields by their protocol-defined order, adds field headers and length prefixes where required, and serializes each value. Decoding reverses that process.

| Package | Responsibility |
| --- | --- |
| [`definitions`](definitions) | Field and type codes, canonical ordering, and serialization flags |
| [`serdes`](serdes) | Binary parsing, serialization, and field headers |
| [`types`](types) | Conversion between JSON values and individual XRPL binary types |

The definitions include transaction types, result codes, ledger entry types, and delegation permissions. Field metadata determines whether a field is serialized, used for signing, or length-prefixed. The serializer and parser use this metadata to select the appropriate type implementation.

## Update protocol definitions

For maintainers: the embedded [`definitions.json`](definitions/definitions.json) is a snapshot of a node's `server_definitions` response. Do not edit individual entries by hand.

From the repository root:

```bash
make update-definitions                                  # Mainnet
make update-definitions NODE_URL=http://127.0.0.1:5005/    # Localnet
```

The target fetches the response, removes the JSON-RPC envelope and request-scoped `status`, and writes sorted keys with two-space indentation. It reports the node version, and the response's `hash` identifies the snapshot. Fetching the same definitions again should produce no diff.

Definitions describe what the node's build can parse, not which amendments its network has activated. A field appearing here does not establish that it can be used on Mainnet.
