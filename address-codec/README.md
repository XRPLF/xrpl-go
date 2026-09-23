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

## Next steps

See the [address codec guide](https://xrplf.github.io/xrpl-go/docs/address-codec) for function selection, tag handling, and validation limits. The [XRPL address documentation](https://xrpl.org/docs/concepts/accounts/addresses) defines the encoding formats. For wallet creation and signing, use [`xrpl/wallet`](../xrpl/wallet).

**Keep seeds private.** Never print, log, or commit real seeds. Encoding a seed does not encrypt it.
