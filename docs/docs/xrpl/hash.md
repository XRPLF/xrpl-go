# hash

## Overview

The `hash` package contains functions for hashing XRPL transactions and for computing ledger-entry indexes.

- `SignTxBlob`: Hashes a signed transaction blob. It accepts a signed transaction blob as input and returns the transaction's hash. This is mainly used for verifying transaction integrity, including multisigned transactions.

- `SignTx`: Hashes a signed transaction provided as a decoded map object. Primarily used internally for batch transactions within the wallet.

- Ledger-entry index helpers: `MPToken`, `MPTokenIssuance`, `MPTID`, `Vault`, `LoanBroker`, `Loan`, and `PaymentChannel`. Each computes the index of a ledger entry so it can be read with a `ledger_entry` request.

## Usage

To import the package, you can use the following code:

```go
import "github.com/Peersyst/xrpl-go/xrpl/hash"
```

## API

### SignTxBlob

```go
func SignTxBlob(txBlob string) (string, error)
```

Hashes a signed transaction blob and returns the transaction hash as an uppercase hexadecimal string, or an error if the blob is invalid.

The transaction must use one complete signing form. A single-signed transaction requires `SigningPubKey` and `TxnSignature`. A multisigned transaction requires `Signers` and an explicitly empty top-level `SigningPubKey`. Partial, empty, or mixed signing structures return an error.

Inner Batch transactions are hashable only in their canonical unsigned form with an explicitly empty `SigningPubKey` and no `TxnSignature` or `Signers`. Consensus-generated `EnableAmendment`, `SetFee`, and `UNLModify` pseudo-transactions can be hashed without account signatures.

### SignTx

```go
func SignTx(tx map[string]any) (string, error)
```

Hashes a signed transaction provided as a decoded map and returns the transaction hash or an error if the transaction object is invalid. It applies the same canonical signing-form checks as `SignTxBlob` and does not modify the input map.

## Ledger entry indexes

These helpers compute the index of a ledger entry from the values that identify it, so an entry can be read without first searching for it.

### MPTokenIssuance

```go
func MPTokenIssuance(issuanceIDHex string) (string, error)
```

Computes the index of an `MPTokenIssuance` entry from its 48-character hexadecimal issuance ID. Use `MPTID` to derive that ID from the issuer address and the sequence of the `MPTokenIssuanceCreate` that produced it.

### MPToken

```go
func MPToken(issuanceIDHex string, holder string) (string, error)
```

Computes the index of the `MPToken` entry a holder owns for an issuance. This is the entry that carries a holder's confidential balance fields, so it is what the [confidential builders](/docs/confidential/builders) read.

```go
index, err := hash.MPToken(issuanceID, holderAddress)
if err != nil {
 return err
}

response, err := client.GetLedgerEntry(&ledger.EntryRequest{
 Index:       index,
 LedgerIndex: common.Validated,
})
```
