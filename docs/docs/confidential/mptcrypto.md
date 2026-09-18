---
sidebar_position: 4
sidebar_label: Native API reference
---

# Native cryptography reference

The `confidential/mptcrypto` package provides low-level Go bindings for the [XRPLF/mpt-crypto](https://github.com/XRPLF/mpt-crypto) C library, used by XLS-96 confidential MPT transactions. It exposes encryption, commitments, transaction context hashes, and proof generation and verification.

[Installation](/docs/confidential/installation) · [Go API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/mptcrypto) · [Package README](https://github.com/XRPLF/xrpl-go/blob/main/confidential/mptcrypto/README.md)

## Choose the right level

Most applications should start with the [builders](/docs/confidential/builders) and [higher-level helpers](/docs/confidential). Use `mptcrypto` when you need direct access to native operations and fixed-size byte types.

Only this package imports `"C"`. The higher-level helpers handle hex encoding, address decoding, and domain-specific errors in Go, but their cryptographic operations still require this native backend.

This guide documents input relationships and limits that matter when calling the native library. Use the API reference for complete function signatures.

## Native backend availability

The native implementation is selected only when all of the following are true:

- cgo is enabled (`CGO_ENABLED=1`)
- the target OS is Linux or macOS (`darwin`)
- the target architecture is `amd64` or `arm64`
- the build is not targeting `js`, `wasip1`, TinyGo, or go-fuzz

Building the native implementation also requires the C/C++ compiler and linker toolchain used by cgo (for example, `gcc`/`g++` on Linux or the Xcode command-line tools on macOS). Linux also needs the zlib development library. Enabling cgo alone does not install that toolchain. See [native build requirements](/docs/confidential/installation#native-build-requirements) for setup.

Vendored headers and static libraries live under `confidential/deps/`:

| Target | Library directory |
| --- | --- |
| Linux amd64 | `confidential/deps/libs/linux-amd64/` |
| Linux arm64 | `confidential/deps/libs/linux-arm64/` |
| macOS amd64 | `confidential/deps/libs/darwin-amd64/` |
| macOS arm64 | `confidential/deps/libs/darwin-arm64/` |

All other builds select `mptcrypto_nocgo.go`. The package still compiles and exposes the same API, but every operation immediately returns `ErrCgoRequired` without validating or processing its inputs. This includes builds with cgo enabled on an unsupported OS or architecture.

## Data model

### Size constants

All sizes are in bytes. On native builds, `sizes_cgo.go` checks the shared `pkg/mptsizes` constants against the vendored native headers at compile time.

| Constant | Bytes | Meaning |
| --- | ---: | --- |
| `PrivKeySize` | 32 | ElGamal private key |
| `PubKeySize` | 33 | Compressed secp256k1 ElGamal public key |
| `BlindingFactorSize` | 32 | ElGamal randomness / Pedersen blinding scalar |
| `CiphertextSize` | 66 | Two compressed EC points (`C1 &#124;&#124; C2`) |
| `AccountIDSize` | 20 | Decoded XRPL account ID |
| `IssuanceIDSize` | 24 | MPToken issuance ID |
| `HashOutputSize` | 32 | Transaction context hash |
| `CommitmentSize` | 33 | Compressed Pedersen commitment |
| `SchnorrProofSize` | 64 | Convert Schnorr proof |
| `SingleBulletproofSize` | 688 | Range proof for one value |
| `DoubleBulletproofSize` | 754 | Aggregated range proof for two values |
| `CompactClawbackProofSize` | 64 | Clawback compact sigma proof |
| `CompactConvertBackProofSize` | 128 | Convert-back compact sigma proof |
| `CompactSendProofSize` | 192 | Send compact sigma proof |
| `ConvertBackProofSize` | 816 | `128 + 688` bytes |
| `SendProofSize` | 946 | `192 + 754` bytes |
| `MaxParticipants` | 255 | Maximum representable participant count in the verification C API |

### Defined byte-array types

The main cryptographic values use distinct fixed-size types:

```go
type PrivateKey [PrivKeySize]byte
type PublicKey [PubKeySize]byte
type BlindingFactor [BlindingFactorSize]byte
type Ciphertext [CiphertextSize]byte
type Commitment [CommitmentSize]byte
type ContextHash [HashOutputSize]byte
```

These types prevent accidental substitutions between same-sized values. Account IDs and issuance IDs are accepted as `[AccountIDSize]byte` and `[IssuanceIDSize]byte`. Proof parameters use fixed-size arrays except for a full send proof, which crosses the API as `[]byte` and is length-checked by `VerifySendProof`.

The Go types enforce byte lengths, not cryptographic validity. The native library validates private scalars, public keys, curve points, ciphertexts, commitments, and proofs when performing an operation.

### Compound inputs

```go
// One encrypted copy of a confidential send amount.
type Participant struct {
    PubKey     PublicKey
    Ciphertext Ciphertext
}

// A value represented by both an ElGamal ciphertext and a Pedersen commitment.
type PedersenProofParams struct {
    Commitment     Commitment
    Amount         uint64
    Ciphertext     Ciphertext
    BlindingFactor BlindingFactor
}
```

For send proofs, XLS-96 uses three participants, or four when an auditor is configured. Their order is part of the native proof contract:

1. sender
2. destination
3. issuer
4. optional auditor

Each participant ciphertext must encrypt the transfer amount under that participant's public key using the same transaction blinding factor. The Go wrapper rejects an empty list or more than `MaxParticipants`. The underlying native routine defines the valid XLS-96 count as three or four.

## Function reference

- [ElGamal](#elgamal): keys, blinding factors, encryption, and decryption.
- [Ciphertext arithmetic](#ciphertext-arithmetic) and [canonical encrypted zero](#canonical-encrypted-zero).
- [Transaction context hashes](#transaction-context-hashes) and [Pedersen commitments](#pedersen-commitments).
- [Proof generation](#proof-generation), [top-level verification](#top-level-verification), and [auxiliary verification](#auxiliary-verification).

### ElGamal

#### `GenerateKeypair`

Generates a secp256k1 ElGamal keypair. The public key is a 33-byte compressed point.

#### `GenerateBlindingFactor`

Generates a random scalar suitable for ElGamal encryption and Pedersen commitments.

#### `EncryptAmount`

Encrypts `amount` under `pubkey` using `bf`. The result is the concatenation of two compressed EC points.

#### `DecryptAmount`

Searches for the plaintext in the inclusive interval `[rangeLow, rangeHigh]`. On a native build, the range must satisfy:

```text
rangeLow <= rangeHigh < math.MaxUint64
```

Invalid ranges wrap `ErrInvalidAmountRange`. Decryption cost grows linearly with the interval width, so callers should use the narrowest practical range. If the native backend is unavailable, `ErrCgoRequired` is returned before range validation.

### Ciphertext arithmetic

Both operations require ciphertexts encrypted under the same public key. Invalid ciphertext encodings and results that cannot be serialized wrap `ErrInvalidCiphertext`. Other native-operation failures return descriptive errors.

#### `AddCiphertexts`

Adds two ElGamal ciphertexts and returns an encryption of the sum of their plaintexts.

#### `SubtractCiphertexts`

Subtracts `b` from `a` and returns an encryption of the difference of their plaintexts. Subtracting a ciphertext from itself returns an error, not an encrypted zero.

### Canonical encrypted zero

#### `CanonicalEncryptedZero`

Returns a deterministic encryption of zero derived from the public key, decoded holder AccountID, and issuance ID. Use it to reproduce a confidential balance that xrpld initializes or resets. No caller-supplied blinding factor is needed.

An invalid public key returns `ErrInvalidPublicKey`. A result that cannot be serialized wraps `ErrInvalidCiphertext`.

### Transaction context hashes

Context hashes bind proofs to transaction-specific fields. All helpers return `ContextHash`.

```go
func ConvertContextHash(
    account [AccountIDSize]byte,
    iss [IssuanceIDSize]byte,
    seq uint32,
) (ContextHash, error)

func ConvertBackContextHash(
    account [AccountIDSize]byte,
    iss [IssuanceIDSize]byte,
    seq, ver uint32,
) (ContextHash, error)

func SendContextHash(
    account [AccountIDSize]byte,
    iss [IssuanceIDSize]byte,
    seq uint32,
    dest [AccountIDSize]byte,
    ver uint32,
) (ContextHash, error)

func ClawbackContextHash(
    account [AccountIDSize]byte,
    iss [IssuanceIDSize]byte,
    seq uint32,
    holder [AccountIDSize]byte,
) (ContextHash, error)
```

The fields correspond to the relevant XLS-96 transaction:

- convert: holder account, issuance ID, and transaction sequence
- convert back: the same fields plus the holder's confidential balance version
- send: sender, destination, issuance ID, sequence, and sender balance version
- clawback: issuer account, target holder, issuance ID, and sequence

### Pedersen commitments

#### `PedersenCommitment`

Computes a compressed Pedersen commitment to `amount` using `bf`. The operation is deterministic for the same amount and blinding factor.

#### `ComputeConvertBackRemainder`

Subtracts the transparent amount from a balance commitment and returns the commitment to the convert-back remainder.

### Proof generation

#### `GenerateConvertProof`

Generates the Schnorr proof of private-key knowledge used when a `ConfidentialMPTConvert` transaction registers a holder encryption key.

#### `GenerateConvertBackProof`

Generates an 816-byte proof containing:

- a 128-byte compact sigma proof binding the holder key, encrypted spending balance, and balance commitment
- a 688-byte range proof showing that the balance remaining after `amount` is subtracted is non-negative

`params` describes the holder's original spending balance, ciphertext, commitment, and commitment blinding factor.

#### `GenerateClawbackProof`

Generates the 64-byte compact sigma proof used by `ConfidentialMPTClawback`. It proves that the issuer-encrypted balance ciphertext contains the revealed clawback amount without exposing the issuer private key.

#### `GenerateSendProof`

Generates the 946-byte `ConfidentialMPTSend` proof:

- 192-byte compact sigma proof for ciphertext consistency, amount linkage, balance linkage, and sender key ownership
- 754-byte aggregated range proof for the transfer amount and post-send balance

The inputs have the following relationships:

- `privkey` and `pubkey` are the sender's keypair.
- `participants` follows the sender, destination, issuer, optional-auditor order described above.
- Every participant ciphertext encrypts `amount` with `txBF`.
- `amountCommitment` must be the commitment returned by `PedersenCommitment(amount, txBF)`. It intentionally reuses the ElGamal randomness.
- `balanceParams` describes the sender's original spending balance and its commitment witness.

The successful native call currently writes `SendProofSize` bytes. The slice return type mirrors the C API's output buffer plus output-length contract.

### Top-level verification

```go
func VerifyConvertProof(
    proof [SchnorrProofSize]byte,
    pubkey PublicKey,
    ctxHash ContextHash,
) error

func VerifyConvertBackProof(
    proof [ConvertBackProofSize]byte,
    pubkey PublicKey,
    ciphertext Ciphertext,
    balanceCommit Commitment,
    amount uint64,
    ctxHash ContextHash,
) error

func VerifySendProof(
    proof []byte,
    participants []Participant,
    senderCt Ciphertext,
    amountCommit, balanceCommit Commitment,
    ctxHash ContextHash,
) error

func VerifyClawbackProof(
    proof [CompactClawbackProofSize]byte,
    amount uint64,
    pubkey PublicKey,
    ciphertext Ciphertext,
    ctxHash ContextHash,
) error
```

Each verifier returns `nil` only when the native proof check succeeds. Additional input contracts:

- `VerifyConvertBackProof` expects the original balance commitment. The native library subtracts `amount` before verifying the remainder range proof. Do not pass a precomputed remainder commitment.
- `VerifySendProof` requires exactly `SendProofSize` proof bytes and the same ordered participant list used for generation. `senderCt` is the sender's original on-ledger spending-balance ciphertext, while the participant ciphertexts encrypt the transfer amount.
- `VerifySendProof` expects the original amount and balance commitments used to generate the proof.

### Auxiliary verification

#### `VerifyRevealedAmount`

Checks deterministically that the holder, issuer, and optional auditor ciphertexts all encrypt the revealed `amount` using `bf`. This is a direct plaintext/ciphertext consistency check, not a ZK-proof verifier. Pass `nil` when no auditor ciphertext is required.

#### `VerifySendRangeProof`

Verifies the 754-byte aggregated range-proof component from a send proof. `balanceCommitment` must be the sender's original balance commitment. The native library derives the post-send remainder from it and `amountCommit`. Do not pass a precomputed remainder commitment.

## Error behavior

The package exposes four sentinel errors:

- `ErrCgoRequired`: the native backend is unavailable for the current build.
- `ErrInvalidAmountRange`: a native `DecryptAmount` call received invalid search bounds.
- `ErrInvalidCiphertext`: ciphertext arithmetic received an invalid ciphertext, or an arithmetic or canonical-zero result could not be serialized.
- `ErrInvalidPublicKey`: a native `CanonicalEncryptedZero` call received a public key that does not decode to a curve point.

Use `errors.Is` to match these sentinels because errors can include input or operation details:

```go
amount, err := mptcrypto.DecryptAmount(ciphertext, privateKey, low, high)
if errors.Is(err, mptcrypto.ErrCgoRequired) {
    // Confidential cryptography is unavailable in this build.
}
if errors.Is(err, mptcrypto.ErrInvalidAmountRange) {
    // Fix the caller-supplied range.
}
```

Other validation and native-library failures are returned as descriptive errors.

## Contributor notes

### Package layout

```text
mptcrypto/
  types.go                 # Package documentation, sizes, and value types
  errors.go                # Shared sentinel errors
  sizes_cgo.go             # Compile-time checks against native size constants
  mptcrypto_cgo.go          # Native bindings and native-only validation
  mptcrypto_nocgo.go        # Unavailable-backend stubs
  mptcrypto_test.go         # Native cryptographic tests
  mptcrypto_nocgo_test.go    # Fallback availability contract
```

See the [package test commands](https://github.com/XRPLF/xrpl-go/blob/main/confidential/mptcrypto/README.md#test-the-package) to check both native and fallback builds. Use the [development workspace](/docs/confidential/installation#development-workspace) to test against the checked-out core module.

### Maintaining the cgo boundary

`mptcrypto_cgo.go` contains the build constraint and per-platform linker flags. It passes fixed-size byte arrays to C through pointers to their first elements and copies Go compound values into their corresponding C structs field by field. Variable participant lists are copied into a contiguous slice of `C.mpt_confidential_participant` values before the native call.

The native routines use these pointers only for the duration of the call. They must not retain Go memory after returning. Keep all `import "C"`, `unsafe`, C layout conversion, and native linker changes inside this package so the higher-level confidential packages remain portable pure Go code.
