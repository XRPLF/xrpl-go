# Keypairs

This package handles cryptographic key pair generation, derivation, signing, and verification for XRP Ledger accounts. It supports both **ED25519** and **secp256k1** algorithms.

## Overview

An XRPL account is ultimately derived from a cryptographic key pair. The lifecycle looks like this:

```
entropy (random bytes)
    → seed (Base58Check encoded, algorithm-specific prefix)
    → private key + public key  (via DeriveKeypair)
    → account ID (SHA-256 → RIPEMD-160 of public key)
    → classic address (Base58Check encoded account ID)
```

The algorithm used (ED25519 or secp256k1) is embedded in the seed's prefix, so `DeriveKeypair` can automatically detect which one to use.

## Supported Algorithms

| Algorithm | Key prefix | Use case |
|---|---|---|
| **ED25519** | `0xED` | Default for new accounts; faster, smaller signatures |
| **secp256k1** | `0x00` | Bitcoin-compatible; required for some hardware wallets and validators |

The algorithm is detected automatically from the first byte of the private or public key hex string.

## API

### Generate a Seed

```go
import (
    addresscodec "github.com/Peersyst/xrpl-go/address-codec"
    "github.com/Peersyst/xrpl-go/keypairs"
    "github.com/Peersyst/xrpl-go/pkg/crypto"
    "github.com/Peersyst/xrpl-go/pkg/random"
)

// Random seed (recommended)
seed, err := keypairs.GenerateSeed(nil, crypto.ED25519(), random.NewRandomizer())
```

Caller-supplied entropy must be exactly 16 raw bytes. Do not pass passphrases directly. If you need deterministic passphrase-based generation, derive 16 bytes before calling this function, for example with SHA-512 and the first 16 bytes, HKDF, or a password KDF. The resulting seed is still limited by the real entropy of the input.

Migration only: older versions silently used the first 16 bytes of any non-empty string passed to `GenerateSeed`. If you need to recover the exact same seed from a legacy input, reproduce that truncation before calling this function:

```go
legacyEntropy := []byte("setPasswordOverLen16")
seed, err := keypairs.GenerateSeed(legacyEntropy[:addresscodec.FamilySeedLength], crypto.ED25519(), nil)
```

If your legacy input was shorter than 16 bytes the old `GenerateSeed` would have panicked, so there is no deterministic seed to recover.

Do not use this pattern for new wallets. New code should provide 16 bytes generated from a cryptographically secure random source, or 16 bytes derived deliberately outside this function.

### Derive a Key Pair

```go
// validator=false for regular accounts, true for validator nodes
privateKey, publicKey, err := keypairs.DeriveKeypair(seed, false)
```

After derivation, the pair is automatically verified by signing and validating a test message — `DeriveKeypair` returns an error if the pair is inconsistent.

### Derive an Address

```go
// Classic address from a public key
classicAddress, err := keypairs.DeriveClassicAddress(publicKey)

// Node/validator address from a node public key
nodeAddress, err := keypairs.DeriveNodeAddress(nodePublicKey, crypto.SECP256K1())
```

### Sign and Verify

```go
// Sign a hex-encoded message with a private key
signature, err := keypairs.Sign(messageHex, privateKey)

// Verify a signature
valid, err := keypairs.Validate(messageHex, publicKey, signature)
```

The algorithm is inferred automatically from the key prefix — no need to specify it explicitly.

## Interfaces

The package is built around two interfaces in `interfaces/`, enabling testing with mocks and supporting future algorithm additions:

- **`KeypairCryptoAlg`** — `DeriveKeypair`, `Sign`, `Validate`
- **`NodeDerivationCryptoAlg`** — `DerivePublicKeyFromPublicGenerator` (secp256k1 only, used for validator node address derivation)
- **`Randomizer`** — `GenerateBytes` (used by `GenerateSeed` for entropy)

Concrete implementations live in `pkg/crypto/` (`ed25519.go`, `secp256k1.go`).
