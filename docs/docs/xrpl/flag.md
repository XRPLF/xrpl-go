# flag

## Overview

The `flag` package provides utility functions for working with bitwise flags in XRPL.

- `Contains`: Checks whether a flag is fully set within a combined flag value. Useful for inspecting transaction or ledger object flags.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/flag"
```

## API

### Contains

```go
func Contains(currentFlag uint32, flag uint32) bool
```

Returns `true` if all bits of `flag` are set in `currentFlag` (`(currentFlag & flag) == flag`). Returns `false` if any bit of `flag` is missing in `currentFlag`, or if `flag` is `0`.

:::warning

The comparison is based on the flag value as a `uint32`. Different contexts may use the same numeric values (e.g. a transaction flag and a ledger-state flag), so a match only indicates the bit is set — not that it belongs to a specific context. Always pair `Contains` with the flag constant that matches the context you are checking.

:::

### Example

```go
package main

import (
	"fmt"

	"github.com/Peersyst/xrpl-go/xrpl/flag"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

func main() {
	offer := transaction.OfferCreate{
		BaseTx: transaction.BaseTx{
			Account: "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		},
		TakerGets: nil,
		TakerPays: nil,
	}

	// Set the sell flag on the offer
	offer.SetSellFlag()

	// Check whether the sell flag is set
	if flag.Contains(offer.Flags, transaction.TfSell) {
		fmt.Println("Offer is a sell offer")
	}

	// Check whether the fill-or-kill flag is set (it is not)
	if !flag.Contains(offer.Flags, transaction.TfFillOrKill) {
		fmt.Println("Offer is not fill-or-kill")
	}
}
```
