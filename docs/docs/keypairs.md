# Keypairs

Use `keypairs` for low-level seed generation, key derivation, and message signatures. For account workflows and transaction signing, start with [Wallets and signing](/docs/xrpl/wallet).

[Install the SDK](/docs/installation) · [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/keypairs) · [XRPL cryptographic keys](https://xrpl.org/docs/concepts/accounts/cryptographic-keys)

## Generate keys and verify a message

This offline example generates temporary keys, derives an address, and verifies a signature. It prints only the public address and verification result. Save it as `main.go` and run `go run .` after installing the SDK.

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
	seed, err := keypairs.GenerateSeed(nil, crypto.SECP256K1(), random.NewRandomizer())
	if err != nil {
		log.Fatal(err)
	}
	private, public, err := keypairs.DeriveKeypair(seed, false)
	if err != nil {
		log.Fatal(err)
	}
	address, err := keypairs.DeriveClassicAddress(public)
	if err != nil {
		log.Fatal(err)
	}
	message := "Example message"
	signature, err := keypairs.Sign(message, private)
	if err != nil {
		log.Fatal(err)
	}
	valid, err := keypairs.Validate(message, public, signature)
	if err != nil {
		log.Fatal(err)
	}
	if !valid {
		log.Fatal("signature did not verify")
	}
	fmt.Println("Address:", address)
	fmt.Println("Signature verified:", valid)
}
```

Use `crypto.ED25519()` to generate an Ed25519 seed instead. Protect both the seed and the private key. Never print, log, commit, or send either to telemetry. Store recovery material securely before funding an address.

## Input formats

| Value | Format |
| --- | --- |
| Seed | XRPL Base58-encoded seed |
| Supplied seed entropy | Exactly 16 raw bytes, not a hex string or passphrase |
| Ed25519 public or private key | Hex, `ED` + 32 bytes |
| secp256k1 public key | Hex, 33 bytes starting `02` or `03` |
| secp256k1 private key | Hex, 32 raw bytes or `00` + 32 bytes |
| Message passed to `Sign` or `Validate` | Raw message bytes carried in a Go string |
| Signature | Hex string |

`Sign("4142", private)` signs the four literal bytes `4142`, not the two bytes represented by that hex string. The library does not hex-decode messages. Use the same raw message for verification. For transaction payload encoding and signing, use the wallet rather than assembling the signing input yourself.

## Seeds and derivation

`GenerateSeed` accepts either exactly 16 entropy bytes or nil/empty entropy with a non-nil randomizer. With explicit entropy, the randomizer is not used. A seed encodes entropy and an algorithm identifier. It is not an encrypted private key.

`DeriveKeypair(seed, false)` derives an account keypair. Its private and public results are 33-byte hex-encoded values in the SDK's algorithm-specific formats. Validator keypair derivation is not supported. Pass `false` for the `validator` argument, because `true` returns an error for both algorithms.

Do not use fixed sample entropy or passphrases for new funded wallets. The [v0.2 migration guide](/docs/upgrading-from-v0.1.x-to-v0.2.0#keypairs) documents legacy string-entropy recovery separately.

## Key and signature checks

- Account public-key operations accept Ed25519 and compressed secp256k1 keys. `DeriveClassicAddress` returns `ErrInvalidPublicKeyFormat` for uncompressed secp256k1 keys and invalid compressed points. `Validate` returns it for uncompressed or malformed keys, and returns `false` with a nil error for a compressed key that is not on the curve.
- secp256k1 signing accepts raw and `00`-prefixed private keys. Zero or out-of-range scalars return `crypto.ErrInvalidPrivateKey` from `pkg/crypto`.
- Invalid private-key formats return `ErrInvalidPrivateKeyFormat`. Both key-format errors also match `ErrInvalidCryptoImplementation` through `errors.Is`.
- secp256k1 verification rejects high-S signatures that are not fully canonical for XRPL.
- Check both the error and the boolean returned by `Validate`. A false result is not successful verification.

Message verification does not prove that the key is authorized for an account on the current ledger. Use [Address codec](/docs/address-codec) for address conversions and the [Go API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/keypairs) for all derivation and signing functions.
