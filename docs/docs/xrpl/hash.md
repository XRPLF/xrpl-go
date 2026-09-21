# Hashes

Use `xrpl/hash` for two tasks: calculating transaction hashes and deriving ledger-entry indexes. These operations run offline.

**The `SignTx` and `SignTxBlob` names refer to hashing signed transactions. They do not create or verify signatures.** Use [wallets](/docs/xrpl/wallet) for signing.

## Transaction hashes

`hash.SignTxBlob(blob)` returns the transaction hash as an uppercase hex string. `hash.SignTx(flat)` accepts a decoded transaction map and does not modify it. Wallet signing already returns the hash, so you usually do not need to calculate it again.

Both functions check the signing structure:

- Single-signed transactions need `SigningPubKey` and `TxnSignature`.
- Multisigned transactions need `Signers` and an explicitly empty top-level `SigningPubKey`.
- Partial, empty, or mixed signing structures return an error.
- Inner Batch transactions are hashable in canonical unsigned form: empty `SigningPubKey`, no `TxnSignature`, and no `Signers`.
- Consensus-generated `EnableAmendment`, `SetFee`, and `UNLModify` pseudo-transactions can be hashed without account signatures.

These checks do not establish cryptographic validity, signer authorization, or quorum.

## Derive a ledger-entry index

This offline example derives an MPT issuance ID from an issuer and creation sequence, then derives the index of a holder's MPToken entry. The values are illustrative. Computing an index does not prove the entry exists.

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/xrpl/hash"
)

func main() {
	issuanceID, err := hash.MPTID(1, "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh")
	if err != nil {
		log.Fatal(err)
	}
	index, err := hash.MPToken(issuanceID, "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Issuance ID:", issuanceID)
	fmt.Println("Holder entry index:", index)
}
```

Pass `index` as `ledger.EntryRequest.Index` to `GetLedgerEntry`. See [Ledger data](/docs/xrpl/ledger-entry-types) for the complete read and decode flow.

`MPTID` produces a 48-character issuance ID, not a ledger-entry index. Use `MPTokenIssuance(issuanceID)` for the issuance entry index or `MPToken(issuanceID, holder)` for the holder entry index. That holder entry carries the confidential balance fields that the [confidential builders](/docs/confidential/builders) read. Other helpers include `Vault`, `LoanBroker`, `Loan`, and `PaymentChannel`.

Use the [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/hash) for the helper arguments and the [XRPL ledger-entry reference](https://xrpl.org/docs/references/protocol/ledger-data/ledger-entry-types) for protocol definitions.
