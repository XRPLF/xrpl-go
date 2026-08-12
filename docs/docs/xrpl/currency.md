---
pagination_prev: keypairs
---

# currency

## Overview

`currency` is a package that provides utility functions to handle XRPL ledger currency types. For **native currency**, it provides XRP and drops conversions. For **IOUs**, it provides utility functions to convert non-standard currency codes (you can learn more about it in the [official documentation](https://xrpl.org/docs/references/protocol/data-types/currency-formats#nonstandard-currency-codes)).

## XRP and drops

The package provides exact string conversions between XRP and drops:

```go
func XrpToDrops(value string) (string, error)
func DropsToXrp(value string) (string, error)
```

Both functions validate the native XRP supply limit. XRP values can have at most six decimal places because one XRP equals `DropsPerXRP`, or 1,000,000 drops.

For calculations, use the exact and immutable `Drops` type. Its zero value is zero drops. Intermediate values can contain a fractional drop, but `WholeString` and `XRPString` require a whole number of drops.

```go
base, err := currency.DropsFromXRP("0.000012")
if err != nil {
 return err
}

adjusted, err := base.MulDecimal("1.2")
if err != nil {
 return err
}

fee := adjusted.Ceil()
feeDrops, err := fee.WholeString()
```

`Drops` supports exact addition, integer and decimal multiplication, rational multiplication, comparison, minimum selection, half-up rounding, and ceiling. Use `MaxNativeDrops` for the maximum native XRP amount in drops.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/currency"
```

## API

```go
const DropsPerXRP = 1_000_000
const MaxNativeDrops uint64 = 100_000_000_000_000_000

// XRP and drops conversions
func XrpToDrops(value string) (string, error)
func DropsToXrp(value string) (string, error)

// Exact drops values
func DropsFromString(value string) (Drops, error)
func DropsFromUint64(value uint64) Drops
func DropsFromXRP(value string) (Drops, error)
func (d Drops) Add(other Drops) Drops
func (d Drops) Mul(multiplier uint64) Drops
func (d Drops) MulDecimal(multiplier string) (Drops, error)
func (d Drops) MulRat(numerator, denominator uint64) (Drops, error)
func (d Drops) Min(other Drops) Drops
func (d Drops) Cmp(other Drops) int
func (d Drops) Ceil() Drops
func (d Drops) RoundHalfUp() Drops
func (d Drops) IsWhole() bool
func (d Drops) WholeString() (string, error)
func (d Drops) XRPString() (string, error)

// Non-standard currency code conversions
func ConvertStringToHex(input string) string
func ConvertHexToString(input string) (string, error)
```
