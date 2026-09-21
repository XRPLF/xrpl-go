# Wallets and signing

Use `xrpl/wallet` to create or recover signing keys and sign transactions locally. Creating a wallet creates keys and an address. It does **not** create or fund an account on the ledger.

[Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/wallet) · [XRPL cryptographic keys](https://xrpl.org/docs/concepts/accounts/cryptographic-keys)

## Create or recover a wallet

| Source | Constructor |
| --- | --- |
| Fresh random keys | `wallet.New(crypto.ED25519())` or `wallet.New(crypto.SECP256K1())` |
| Existing family seed | `wallet.FromSeed(seed, "")` |
| Existing family seed, short form | `wallet.FromSecret(seed)` |
| Mnemonic | `wallet.FromMnemonic(mnemonic)` |
| Regular key for another account | `wallet.FromSeed(seed, masterAddress)` |

Import `crypto` from `github.com/Peersyst/xrpl-go/pkg/crypto`. The SDK supports Ed25519 and secp256k1 seeds. `FromMnemonic` accepts a BIP39 mnemonic, always derives a secp256k1 key at `m/44'/144'/0'/0/0`, returns a `*Wallet` rather than a value, and leaves `Seed` empty. Use the constructor that matches your credential format rather than treating a seed, mnemonic, and private key as interchangeable strings.

For regular-key signing, the account must authorize that key on the ledger. Setting `masterAddress` does not grant permission. See [regular key pairs](https://xrpl.org/docs/concepts/accounts/cryptographic-keys#regular-key-pair).

:::warning

Seeds, private keys, and mnemonics control the account. Never print, log, commit, or send them to telemetry. Store recovery material securely before funding a wallet. Signing functionality has not been independently audited.

:::

## Sign offline

This example creates a temporary wallet and signs a demonstration Payment with explicit network fields. It runs offline and does not submit anything. The sequence, fee, and expiry are illustrative, not values to reuse on a live ledger.

Save it as `main.go` in an application with the SDK installed, then run `go run .`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		log.Fatal(err)
	}
	payment := transaction.Payment{
		BaseTx: transaction.BaseTx{
			Account:            w.GetAddress(),
			Sequence:           1,
			Fee:                types.XRPCurrencyAmount(12),
			LastLedgerSequence: 100,
		},
		Destination: "rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe",
		Amount:      types.XRPCurrencyAmount(1_000_000),
	}
	blob, hash, err := w.Sign(payment.Flatten())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Address:", w.GetAddress())
	fmt.Println("Transaction hash:", hash)
	fmt.Println("Signed blob length:", len(blob))
}
```

`Sign` takes a flat map and returns the **signed transaction blob and transaction hash**, not a standalone signature. It signs an internal copy without changing the input map. A successful signature does not establish ledger authorization or transaction validity.

For a real transaction, prepare network fields before offline signing, or use a client's autofill first. See [Transactions](/docs/xrpl/transaction) for a complete funded Testnet flow. Submit the final blob unchanged.

## Multisigning

Each signer calls `w.Multisign(flat)` on the same prepared transaction. It returns that signer's blob and transaction hash. Combine the independent blobs with `xrpl.Multisign()` before submission. Do not successively modify the transaction between signatures.

Configure the account's signer list and quorum on the ledger first. Autofill the fee for the planned signer count before collecting signatures. See the [complete Go multisigning example](https://github.com/XRPLF/xrpl-go/tree/main/examples/multisigning/rpc) and the [XRPL multisigning guide](https://xrpl.org/docs/concepts/accounts/multi-signing).

## Sponsor and Batch signatures

- [Sponsorship](/docs/xrpl/sponsorship) documents co-signed and pre-funded flows. Sponsor fields must be prepared before account signing.
- `wallet.SignMultiBatch` signs Batch authorization, not each inner transaction as an ordinary signed transaction. Combine fragments with `wallet.CombineBatchSigners`, then sign the outer transaction. See the [Go API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/wallet#SignMultiBatch), [Batch protocol](https://xrpl.org/docs/references/protocol/transactions/types/batch), and [confidential batch guide](/docs/confidential/batch).

## Authorize a payment-channel claim

`wallet.AuthorizeChannel(channelID, amount, w)` returns a claim signature. `amount` is in drops. This does not create a channel, submit a claim, or sign a Payment transaction.

This offline demonstration uses a temporary wallet and an illustrative channel ID. A real claim requires the signing key authorized for that channel:

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/xrpl/wallet"
)

func main() {
	w, err := wallet.New(crypto.ED25519())
	if err != nil {
		log.Fatal(err)
	}
	channelID := "5DB01B7FFED6B67E6B0414DED11E051D2EE2B7619CE0EAA6286D67A3A4D5BDB3"
	signature, err := wallet.AuthorizeChannel(channelID, "1000000", w)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Claim signature:", signature)
}
```

For settlement and channel rules, see [XRPL payment channels](https://xrpl.org/docs/concepts/payment-types/payment-channels).
