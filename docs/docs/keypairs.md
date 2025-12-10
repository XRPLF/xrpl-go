---
sidebar_position: 5
pagination_next: xrpl/currency
---

# keypairs

## Introduction 

The keypairs package provides a set of functions for generating and managing cryptographic keypairs. It includes functionality for creating new keypairs, deriving public keys from private keys, and verifying signatures.

This package is used internally by the `xrpl` package to expose a [`Wallet`](/docs/xrpl/wallet) interface for easier wallet management. Nevertheless, it can be used independently of the `xrpl` package for cryptographic operations.

## Key components

This package works with the following key components from the XRP Ledger:

- **Seed**: A base58-encoded string that represents a keypair.
- **Keypair**: A pair of a private and public key.
- **Address**: A base58-encoded string that represents an account.

To learn more about these components, you can check the [official documentation](https://xrpl.org/docs/concepts/accounts/cryptographic-keys).

## Supported algorithms

Cryptographic algorithms supported by this package are:

- ed25519
- secp256k1

 Every function in the package that requires a cryptographic algorithm will accept any type that satisfies the `KeypairCryptoAlg` interface. So, if desired, you can implement your own algorithm and use it in this package.

 However, the library already exports both algorithm getters that satisfy the `KeypairCryptoAlg` and `NodeDerivationCryptoAlg` interfaces. They're available under the package `github.com/Peersyst/xrpl-go/pkg/crypto`, which exports both algorithm getters that satisfy the `KeypairCryptoAlg`, `NodeDerivationCryptoAlg` interfaces.

### crypto package

The `crypto` package exports the following algorithm getters that satisfy the `KeypairCryptoAlg`, `NodeDerivationCryptoAlg` interfaces:

- `ED25519()`
- `SECP256K1()`

You can use them to generate a seed or derive a keypair as the following example shows:

```go
seed, err := keypairs.GenerateSeed("", crypto.SECP256K1(), random.NewRandomizer())
```

## API

These are the functions available in this package:

```go
// Key generation
func GenerateSeed(entropy string, alg interfaces.KeypairCryptoAlg, r interfaces.Randomizer) (string, error)
func DeriveKeypair(seed string, validator bool) (private, public string, err error)
func DeriveClassicAddress(pubKey string) (string, error)
func DeriveNodeAddress(pubKey string, alg interfaces.NodeDerivationCryptoAlg) (string, error)

// Signing
func Sign(msg, privKey string) (string, error)
func Validate(msg, pubKey, sig string) (bool, error)
```

They can be split into two groups:

- Key generation: Functions that generate seeds and addresses.
- Signing: Functions that sign and validate messages.

### Key generation

#### GenerateSeed

```go
func GenerateSeed(entropy string, alg interfaces.KeypairCryptoAlg, r interfaces.Randomizer) (string, error)
```

Generate a seed that can be used to generate keypairs. You can specify the entropy of the seed or let the function generate a random one (by passing an empty string as entropy and providing a randomizer) and use one of the supported algorithms to generate the seed. The result is a base58-encoded seed, which starts with the character `s`. 

:::info

A randomizer satisfies the `Randomizer` interface. The `random` package exports a `NewRandomizer` function that returns a new randomizer.

:::

#### DeriveKeypair

```go
func DeriveKeypair(seed string, validator bool) (private, public string, err error)
```

Derives a keypair (private and public keys) from a seed. If the `validator` parameter is `true`, the keypair will be a validator keypair; otherwise, it will be a user keypair. The result for both the private and public keys is a 33-byte hexadecimal string.


#### DeriveClassicAddress

```go
func DeriveClassicAddress(pubKey string) (string, error)
```

After deriving a keypair, you can derive the classic address from the public key. The result is a base58 encoded address, which starts with the character `r`. If you're interested in X-Address derivation, the `address-codec` package contains functions to encode and decode X-Addresses from and to classic addresses.

#### DeriveNodeAddress

```go
func DeriveNodeAddress(pubKey string, alg interfaces.NodeDerivationCryptoAlg) (string, error)
```

Derives a node address from a public key. The result is a base58-encoded address, which starts with the character `n`.

### Signing

#### Sign

```go
func Sign(msg, privKey string) (string, error)
```

Signs the provided message with the provided private key. To be able to sign a message, the private key must be a valid keypair and the message must be hex-encoded. The result is a hexadecimal string that represents the signature of the message. To verify the signature, you can use the `Validate` function.

#### Validate

```go
func Validate(msg, pubKey, sig string) (bool, error)
```

Verifies a signature of a message. To be able to verify a signature, the public key must be valid, and the message and the signature must be hex-encoded. The result is a boolean value that indicates if the signature is valid or not.

## Guides

### How to generate a new random keypair

This example generates a new keypair using the `SECP256K1` algorithm and a random entropy. It then derives a keypair from the seed and derives the classic address from the public key.

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
	seed, err := keypairs.GenerateSeed("", crypto.SECP256K1(), random.NewRandomizer())
	if err != nil {
		log.Fatal(err)
	}

	privK, pubK, err := keypairs.DeriveKeypair(seed, false)
	if err != nil {
		log.Fatal(err)
	}

	addr, err := keypairs.DeriveClassicAddress(pubK)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Seed: ", seed)
	fmt.Println("Private Key: ", privK)
	fmt.Println("Public Key: ", pubK)
	fmt.Println("Address: ", addr)
}
```


### How to generate a new keypair from entropy

This example generates a new keypair using the `ED25519` algorithm and a provided entropy. Then, it derives the keypair and the address as the previous example.

```go
package main

import (
	"fmt"
	"log"

	"github.com/Peersyst/xrpl-go/keypairs"
	"github.com/Peersyst/xrpl-go/pkg/crypto"
)

func main() {
	seed, err := keypairs.GenerateSeed("ThisIsMyCustomEntropy", crypto.ED25519(), nil)
	if err != nil {
		log.Fatal(err)
	}

	privK, pubK, err := keypairs.DeriveKeypair(seed, false)
	if err != nil {
		log.Fatal(err)
	}

	addr, err := keypairs.DeriveClassicAddress(pubK)

	fmt.Println("Seed: ", seed)
	fmt.Println("Private Key: ", privK)
	fmt.Println("Public Key: ", pubK)
	fmt.Println("Address: ", addr)
}
```