# Address Codec

This package provides encoding and decoding of XRP Ledger addresses and keys using Base58Check — the same scheme used by Bitcoin but with XRPL's own alphabet and version prefixes.

## Overview

Every XRPL account has a 20-byte account ID derived from the public key (`SHA-256` → `RIPEMD-160`). The address codec turns that raw byte representation into human-readable strings and back.

The package supports two address formats:

- **Classic addresses** — the standard `r...` format (e.g. `rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh`). A 1-byte version prefix (`0x00`) is prepended to the 20-byte account ID, then Base58Check encoded.
- **X-addresses** — a newer format that encodes an account ID + destination tag + network flag into a single string (e.g. `X7AcgcsBL4L51nv2theWeVeD8HoecNqNKq31GV6rsPPgFm3`). Avoids the need to pass destination tags separately.

Both formats embed a 4-byte checksum (double SHA-256) to catch typos.

## How Base58Check Works

```
account ID (20 bytes)
    → prepend version prefix (e.g. 0x00 for classic address)
    → append 4-byte checksum (SHA-256(SHA-256(prefix + payload))[:4])
    → Base58-encode using XRPL's alphabet
    → human-readable address string
```

Decoding reverses the process and validates the checksum before returning the payload.

## Address Types and Prefixes

| Type | Prefix | Length |
|---|---|---|
| Classic address (`r...`) | `0x00` | 20 bytes |
| Account public key | `0x23` | 33 bytes |
| Family seed (secp256k1) | `0x21` | 16 bytes |
| Family seed (ed25519) | `0x01 0xE1 0x4B` | 16 bytes |
| Node/validator public key | `0x1C` | 33 bytes |
| X-address (mainnet) | `0x05 0x44` | 35 bytes |
| X-address (testnet) | `0x04 0x93` | 35 bytes |

## API

### Classic Addresses

```go
// Derive classic address from a public key hex string
address, err := addresscodec.EncodeClassicAddressFromPublicKeyHex(pubKeyHex)

// Decode a classic address back to its account ID bytes
typePrefix, accountID, err := addresscodec.DecodeClassicAddressToAccountID(address)

// Encode a raw 20-byte account ID to a classic address
address, err := addresscodec.EncodeAccountIDToClassicAddress(accountID)
```

### X-Addresses

```go
// Encode an account ID + tag into an X-address
xAddress, err := addresscodec.EncodeXAddress(accountID, tag, hasTag, isTestnet)

// Decode an X-address back to account ID, tag, and network flag
accountID, tag, isTestnet, err := addresscodec.DecodeXAddress(xAddress)

// Convert between formats
xAddress, err := addresscodec.ClassicAddressToXAddress(classicAddress, tag, hasTag, isTestnet)
classicAddress, tag, isTestnet, err := addresscodec.XAddressToClassicAddress(xAddress)
```

### Seeds and Keys

```go
// Encode a 16-byte seed (specify ed25519 or secp256k1 algorithm)
seed, err := addresscodec.EncodeSeed(entropy, crypto.ED25519())
seed, err := addresscodec.EncodeSeed(entropy, crypto.SECP256K1())

// Decode a seed string back to raw bytes and its algorithm
entropy, algorithm, err := addresscodec.DecodeSeed(seed)

// Encode/decode public keys
encoded, err := addresscodec.EncodeAccountPublicKey(pubKeyBytes)
encoded, err := addresscodec.EncodeNodePublicKey(pubKeyBytes)
decoded, err := addresscodec.DecodeAccountPublicKey(encoded)
decoded, err := addresscodec.DecodeNodePublicKey(encoded)
```

### Validation

```go
addresscodec.IsValidClassicAddress(address) // bool
addresscodec.IsValidXAddress(address)       // bool
addresscodec.IsValidAddress(address)        // bool — accepts either format
```
