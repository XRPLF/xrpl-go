# Address codec

Use `addresscodec` to encode, decode, and validate XRP Ledger addresses, seeds, and public keys. These operations run offline. For wallet creation and signing, use the [wallet package](/docs/xrpl/wallet) instead.

[Install the SDK](/docs/installation) · [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/address-codec)

## Classic addresses and X-addresses

A classic address identifies an account. An X-address combines the account ID, an optional destination tag, and a network flag into one string.

Use an X-address when you need to carry the destination tag with the address. When converting it back, keep the tag and network information rather than discarding them.

## Convert an address

After installing the SDK, save this as `main.go` and run `go run .`. It converts a classic address to a Testnet X-address with tag `123`, then decodes it:

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

The final line contains the original classic address, tag `123`, and two `true` values.

The last two arguments to `ClassicAddressToXAddress` mean:

| Argument | Meaning |
| --- | --- |
| `hasTag` | Whether the destination tag is present |
| `isTestnet` | Whether to use the test-network address format |

A tag of zero and an absent tag are different. Use `hasTag` to distinguish them.

## Choose a function

| Task | Functions |
| --- | --- |
| Convert address formats | `ClassicAddressToXAddress`, `XAddressToClassicAddress` |
| Encode or decode raw account IDs | `EncodeAccountIDToClassicAddress`, `DecodeClassicAddressToAccountID` |
| Derive an address from a public key | `EncodeClassicAddressFromPublicKeyHex` |
| Encode or decode X-address payloads | `EncodeXAddress`, `DecodeXAddress` |
| Encode or decode seeds | `EncodeSeed`, `DecodeSeed` |
| Encode or decode account public keys | `EncodeAccountPublicKey`, `DecodeAccountPublicKey` |
| Encode or decode node public keys | `EncodeNodePublicKey`, `DecodeNodePublicKey` |
| Validate address encodings | `IsValidClassicAddress`, `IsValidXAddress`, `IsValidAddress` |

## What validation tells you

Validation checks the address encoding and checksum. It does not query the ledger or establish that an account exists, is funded, or belongs to the intended recipient. Use a [client query](/docs/xrpl/rpc) to inspect ledger state.

Classic addresses represent 20-byte account IDs. Both address formats use Base58Check with the XRPL alphabet. See the [XRPL address documentation](https://xrpl.org/docs/concepts/accounts/addresses) for format details.

## Seeds are secrets

Encoding a seed does not encrypt it. Never print, log, commit, or send real seeds to telemetry. Use [keypairs](/docs/keypairs) for lower-level key generation, or [wallets](/docs/xrpl/wallet) for normal account and signing workflows.
