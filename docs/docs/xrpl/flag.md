# Flags

Use `flag.Contains(value, mask)` to check that all bits in a flag mask are set. Use transaction setters to add flags where the transaction type provides them.

## Set and inspect a flag

This offline example demonstrates flag handling only. The offer omits amounts and is not ready for submission.

```go
package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

func main() {
	var offer transaction.OfferCreate
	offer.SetSellFlag()

	fmt.Println("Sell:", flag.Contains(offer.Flags, transaction.TfSell))
	fmt.Println("Fill or kill:", flag.Contains(offer.Flags, transaction.TfFillOrKill))
}
```

It prints `Sell: true` and `Fill or kill: false`.

## Match the context

`Contains` checks `(value & mask) == mask`, but returns false for a zero mask. A combined mask matches only when **all** its bits are present. Use `flag.ContainsAny(value, mask)` to match when any bit of the mask is set, and `flag.ContainsOnly(value, allowed)` to check that no bit outside the allowed mask is set.

The function cannot tell whether a numeric bit belongs to a transaction or a ledger entry. Different contexts can reuse the same numbers. Always pair the value with the constants for that exact transaction or ledger-entry type.

See the [Go flag API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/flag) for the helper and the [XRPL transaction reference](https://xrpl.org/docs/references/protocol/transactions/types) for each transaction's flag meanings.
