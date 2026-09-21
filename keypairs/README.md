# Keypairs

Generate seeds, derive keypairs and addresses, and sign or verify messages for XRP Ledger accounts. The package supports ED25519 and secp256k1.

[Installation](../README.md#quick-start) · [API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/keypairs) · [Wallet guide](../xrpl/README.md#wallets-and-multisigning)

Use [`xrpl/wallet`](../xrpl/wallet) for normal wallet management and transaction signing. Use `keypairs` when you need the lower-level operations directly.

## Generate a seed and address

This example generates a random ED25519 seed, derives its keypair, and prints only the public address. It runs offline and does not fund an account.

Save it as `main.go` in your application after installing the SDK, then run `go run .`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
	"github.com/Peersyst/xrpl-go/pkg/random"
)

func main() {
	seed, err := keypairs.GenerateSeed(nil, crypto.ED25519(), random.NewRandomizer())
	if err != nil {
		log.Fatal(err)
	}
	_, publicKey, err := keypairs.DeriveKeypair(seed, false)
	if err != nil {
		log.Fatal(err)
	}
	address, err := keypairs.DeriveClassicAddress(publicKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Classic address:", address)
}
```

The address changes on each run. The example discards the credentials when it exits, so **do not send funds to this address**.

## Choose an operation

| Task | Function |
| --- | --- |
| Generate a seed | `GenerateSeed` |
| Derive private and public keys from a seed | `DeriveKeypair` |
| Derive a classic address from a public key | `DeriveClassicAddress` |
| Derive a node address | `DeriveNodeAddress` |
| Sign a message | `Sign` |
| Verify a signature | `Validate` |

Choose `crypto.ED25519()` or `crypto.SECP256K1()` when generating a seed. `DeriveKeypair` detects the algorithm from the encoded seed. Pass `false` for its `validator` argument. Validator keypair derivation is not supported and returns an error. The result order is **private key, public key, error**.

`Sign` and `Validate` select the algorithm from the key encoding. For transactions, use the wallet signing methods rather than signing arbitrary JSON or a submission blob.

## Supply your own entropy

Prefer the random seed generation shown above. If you supply entropy, it must be exactly 16 raw bytes. The randomizer is not used in that case.

Do not pass a passphrase directly. Deterministic derivation belongs outside this function, and a derived seed is only as strong as its input and derivation method. Hashing a weak password does not make it a strong secret.

## Recover a legacy seed

Follow the [v0.2 migration guide](https://xrplf.github.io/xrpl-go/docs/upgrading-from-v0.1.x-to-v0.2.0#keypairs) to reproduce legacy string-entropy derivation. Recovery rules are not recommendations for new wallets.

## Security

Never print, log, commit, or send real seeds or private keys to telemetry. Use secure storage for credentials that must survive a process exit. Read the [security and audit notice](../README.md#security-and-audits) before using the SDK with production funds.

## Package internals

The [interfaces package](interfaces) defines the keypair, node-derivation, and randomizer contracts. Algorithm implementations live in [`pkg/crypto`](../pkg/crypto). `DeriveKeypair` checks the derived pair by signing and verifying a test message before returning it.
