---
pagination_prev: keypairs
---

# currency

## Overview

`currency` is a package that provides utility functions to handle XRPL ledger currency types. For **native currency**, it provides XRP and drops conversions. For **IOUs**, it provides utility functions to convert non-standard currency codes (you can learn more about it in the [official documentation](https://xrpl.org/docs/references/protocol/data-types/currency-formats#nonstandard-currency-codes)).

## XRP/Drops conversions

The package provides the following functions to convert XRP to drops and vice versa:

```go
func XrpToDrops(value string) (string, error)
func DropsToXrp(value string) (string, error)
```

Both functions return the converted value as a string and an error if the value is not a valid number.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/currency"
```

## API

```go
// XRP <-> Drops conversions
func XrpToDrops(value string) (string, error)
func DropsToXrp(value string) (string, error)

// Non-standard currency codes conversions
func ConvertStringToHex(input string) string
func ConvertHexToString(input string) (string, error)