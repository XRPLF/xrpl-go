# Currency amounts

Use `xrpl/currency` for exact XRP/drop conversions and calculations. **1 XRP is 1,000,000 drops.** Use decimal strings rather than floating-point values when converting user input.

## Convert XRP and drops

This example runs offline:

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/xrpl/currency"
)

func main() {
	drops, err := currency.XrpToDrops("1.25")
	if err != nil {
		log.Fatal(err)
	}
	xrp, err := currency.DropsToXrp(drops)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Drops:", drops)
	fmt.Println("XRP:", xrp)
}
```

It prints `1250000` drops and `1.25` XRP. Both conversion functions validate the native XRP supply limit. XRP inputs must convert to a whole number of drops.

## Calculate fees without losing precision

The immutable `Drops` type supports exact arithmetic. Its zero value is zero drops. Intermediate calculations can contain a fractional drop, but `WholeString` and `XRPString` require a whole number of drops.

For example, in a function returning an error:

```go
base, err := currency.DropsFromXRP("0.000012")
if err != nil {
	return err
}
adjusted, err := base.MulDecimal("1.2")
if err != nil {
	return err
}
feeDrops, err := adjusted.Ceil().WholeString()
if err != nil {
	return err
}
fmt.Println(feeDrops) // 15
```

Choose rounding deliberately. `Ceil` rounds upward, while `RoundHalfUp` rounds to the nearest whole drop with halves rounded up. `Drops` also supports addition, integer/decimal/rational multiplication, comparison, and minimum selection.

`Drops` constructors do not enforce the protocol supply limit. Check final values against `currency.MaxNativeDrops` before protocol use. Do not confuse an intermediate calculation with an amount accepted by a transaction field.

## Other currencies

`ConvertStringToHex` and `ConvertHexToString` handle non-standard currency-code strings. They do not convert balances or exchange rates. See [XRPL currency formats](https://xrpl.org/docs/references/protocol/data-types/currency-formats) for XRP, issued-currency, and MPT representations.

For transaction amount structs, use [transaction/types](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/transaction/types). For the complete arithmetic API, use [currency](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/currency).
