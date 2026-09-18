# Address codec

Encode, decode, and validate XRP Ledger addresses, seeds, and public keys. Use this package when you need to convert between classic addresses and X-addresses or work with their raw bytes.

[Installation](../README.md#quick-start) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/address-codec) · [Wallet guide](../xrpl/README.md#wallets-and-multisigning)

## Convert an address

This example adds a destination tag to a classic address and encodes it as a Testnet X-address. It runs offline.

Save it as `main.go` in your application after installing the SDK, then run `go run .`:

```go
package main

import (
	"fmt"
	"log"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
)

func main() {
	classic := "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh"
	const tag uint32 = 123

	xAddress, err := addresscodec.ClassicAddressToXAddress(classic, tag, true, true)
	if err != nil {
		log.Fatal(err)
	}

	address, decodedTag, hasTag, isTestnet, err := addresscodec.XAddressToClassicAddress(xAddress)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("X-address:", xAddress)
	fmt.Printf("Classic: %s, tag: %d, has tag: %t, testnet: %t\n",
		address, decodedTag, hasTag, isTestnet)
}
```

The final line contains the original classic address, tag `123`, and two `true` values. In `ClassicAddressToXAddress`, the last two arguments specify tag presence and the test-network flag. Tag presence is separate from the tag value, so a tag of zero is not the same as no tag.

## Choose a function

| Task | Functions |
| --- | --- |
| Convert address formats | `ClassicAddressToXAddress`, `XAddressToClassicAddress` |
| Work with raw account IDs | `EncodeAccountIDToClassicAddress`, `DecodeClassicAddressToAccountID` |
| Derive an address from a public key | `EncodeClassicAddressFromPublicKeyHex` |
| Encode or decode an X-address payload | `EncodeXAddress`, `DecodeXAddress` |
| Encode or decode a seed | `EncodeSeed`, `DecodeSeed` |
| Encode or decode public keys | `EncodeAccountPublicKey`, `DecodeAccountPublicKey`, `EncodeNodePublicKey`, `DecodeNodePublicKey` |
| Check an address format | `IsValidClassicAddress`, `IsValidXAddress`, `IsValidAddress` |

Validation checks the address encoding. It does not query the ledger or establish that an account exists. For wallet creation and signing, use [`xrpl/wallet`](../xrpl/wallet) instead.

## How encoding works

A classic address represents a 20-byte account ID. An X-address also includes a destination tag and a network flag. Both use Base58Check with the XRPL alphabet and a checksum to detect invalid encodings.

For a classic address:

```text
account ID -> version prefix + account ID -> append checksum -> Base58 string
```

Decoding reverses this process and checks the checksum. See the [XRPL address documentation](https://xrpl.org/docs/concepts/accounts/addresses) for format details, and the [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/address-codec) for function signatures and errors.

**Keep seeds private.** Never print, log, or commit real seeds. Encoding a seed does not encrypt it.
