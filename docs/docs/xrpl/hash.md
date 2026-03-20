# hash

## Overview

The `hash` package contains functions for hashing XRPL transactions.

- `SignTxBlob`: Hashes a signed transaction blob. It accepts a signed transaction blob as input and returns the transaction's hash. This is mainly used for verifying transaction integrity, including multisigned transactions.

- `SignTx`: Hashes a signed transaction provided as a decoded map object. Primarily used internally for batch transactions within the wallet.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/hash"
```

## API

### SignTxBlob

```go
func SignTxBlob(txBlob string) ([]byte, error)
```

Hashes a signed transaction blob and returns the transaction hash or an error if the blob is invalid.

### SignTx

```go
func SignTx(tx map[string]any) (string, error)
```

Hashes a signed transaction provided as a decoded map and returns the transaction hash or an error if the transaction object is invalid.
