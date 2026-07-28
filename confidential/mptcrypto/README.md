# mptcrypto

Go bindings for the [XRPLF/mpt-crypto](https://github.com/xrplf/mpt-crypto) C library. This package is the **only** place in the codebase that imports `"C"` (CGo). Everything above this layer (elgamal/, proof/, commitment/) is pure Go.

## Build requirements

CGo is required to run confidential cryptographic operations (`CGO_ENABLED=1`). The vendored C libraries live in `confidential/deps/libs/<os-arch>/`. Without CGo, no confidential cryptographic operation is available and every `mptcrypto` function returns `ErrCgoRequired` without processing its inputs.

```bash
# normal build (CGo on by default)
go test ./confidential/mptcrypto/...

# force CGo off (exercise no-CGo fallbacks)
CGO_ENABLED=0 go test ./confidential/mptcrypto/...
```

## How this package is organized

```
mptcrypto/
  types.go                  # Size constants, defined value types, and proof types
  errors.go                 # Shared errors and range validation
  mptcrypto_cgo.go          # Real implementations (only built with CGo)
  mptcrypto_nocgo.go        # ErrCgoRequired stubs for builds without CGo
  mptcrypto_test.go         # CGo-backed tests
  mptcrypto_nocgo_test.go   # No-CGo availability contract
```

Every function uses **defined, fixed-size byte-array types** such as `PrivateKey`, `PublicKey`, and `Ciphertext`. These prevent callers from mixing semantically different values with the same underlying size. Hex encoding/decoding happens in the layers above (elgamal/, proof/, commitment/), never here.

---

## Types and constants

### Size constants

| Constant | Bytes | What it is |
|---|---|---|
| `PrivKeySize` | 32 | secp256k1 private key |
| `PubKeySize` | 33 | Compressed secp256k1 public key |
| `BlindingFactorSize` | 32 | Random scalar for encryption/commitment |
| `CiphertextSize` | 66 | ElGamal ciphertext (two compressed points: C1 &#124;&#124; C2) |
| `AccountIDSize` | 20 | XRPL account ID (decoded from classic address) |
| `IssuanceIDSize` | 24 | MPTokenIssuance ID |
| `HashOutputSize` | 32 | Context hash output (half-SHA) |
| `CommitmentSize` | 33 | Compressed Pedersen commitment point |
| `SchnorrProofSize` | 64 | Schnorr proof of knowledge |
| `SingleBulletproofSize` | 688 | Single bulletproof (range proof for 1 value) |
| `DoubleBulletproofSize` | 754 | Double bulletproof (range proof for 2 values) |
| `CompactClawbackProofSize` | 64 | Compact sigma proof for clawback |
| `CompactConvertBackProofSize` | 128 | Compact sigma proof for convert-back |
| `CompactSendProofSize` | 192 | Compact sigma proof for send |
| `ConvertBackProofSize` | 816 | Compact sigma + single bulletproof (128 + 688) |
| `SendProofSize` | 946 | Compact sigma + double bulletproof (192 + 754) |
| `MaxParticipants` | 255 | Max participants in a send (C API uses uint8_t) |

### Value types

```go
type PrivateKey [PrivKeySize]byte
type PublicKey [PubKeySize]byte
type BlindingFactor [BlindingFactorSize]byte
type Ciphertext [CiphertextSize]byte
type Commitment [CommitmentSize]byte
type ContextHash [HashOutputSize]byte
```

These are defined types rather than aliases, so Go rejects accidental substitutions such as passing a `BlindingFactor` where a `PrivateKey` is required. They enforce semantic separation, while the C library remains responsible for validating key and curve-point contents.

### Structs

```go
// A party in a confidential send (public key + their encrypted amount).
type Participant struct {
    PubKey     PublicKey
    Ciphertext Ciphertext
}

// Parameters for generating Pedersen linkage proofs.
type PedersenProofParams struct {
    Commitment     Commitment
    Amount         uint64
    Ciphertext     Ciphertext
    BlindingFactor BlindingFactor
}
```

---

## Function reference

### 1. ElGamal encryption

These handle key generation, encryption, and decryption for confidential amounts.

#### `GenerateKeypair() (privkey PrivateKey, pubkey PublicKey, err error)`

Creates a new secp256k1 ElGamal keypair.

```go
priv, pub, err := mptcrypto.GenerateKeypair()
// priv: 32-byte private key
// pub:  33-byte compressed public key (starts with 0x02 or 0x03)
```

#### `GenerateBlindingFactor() (bf BlindingFactor, err error)`

Generates a cryptographically random 32-byte scalar. Used as the randomness parameter (`r`) when encrypting amounts or creating Pedersen commitments.

```go
bf, err := mptcrypto.GenerateBlindingFactor()
```

#### `EncryptAmount(amount uint64, pubkey PublicKey, bf BlindingFactor) (ct Ciphertext, err error)`

Encrypts an amount using ElGamal. The ciphertext is 66 bytes: two compressed EC points concatenated (C1 || C2).

```go
ct, err := mptcrypto.EncryptAmount(1000, pubkey, blindingFactor)
// ct: 66-byte ciphertext
```

#### `DecryptAmount(ciphertext Ciphertext, privateKey PrivateKey, rangeLow, rangeHigh uint64) (uint64, error)`

Decrypts an ElGamal ciphertext by searching the inclusive `[rangeLow, rangeHigh]` interval. Bounds must satisfy `rangeLow <= rangeHigh < math.MaxUint64`. Search cost grows linearly with the interval size, so use the narrowest practical bounds.

```go
amount, err := mptcrypto.DecryptAmount(ciphertext, privateKey, 0, 10_000)
```

### 2. Context hashes

Every ZK proof is bound to a specific transaction via a **context hash**. This prevents proof reuse across transactions. Each transaction type has its own hash function because the inputs differ.

All context hash functions return a `ContextHash`.

#### `ConvertContextHash(account [20]byte, iss [24]byte, seq uint32) (ContextHash, error)`

For **ConfidentialMPTConvert** transactions (public amount -> confidential).

- `account`: the sender's 20-byte account ID
- `iss`: the 24-byte MPTokenIssuance ID
- `seq`: the transaction sequence number

#### `ConvertBackContextHash(account [20]byte, iss [24]byte, seq, ver uint32) (ContextHash, error)`

For **ConfidentialMPTConvertBack** transactions (confidential -> public amount).

Same as above plus `ver` (the version counter from the ledger object).

#### `SendContextHash(account [20]byte, iss [24]byte, seq uint32, dest [20]byte, ver uint32) (ContextHash, error)`

For **ConfidentialMPTSend** transactions (confidential transfer between accounts).

Adds `dest` (destination account ID) and `ver`.

#### `ClawbackContextHash(account [20]byte, iss [24]byte, seq uint32, holder [20]byte) (ContextHash, error)`

For **ConfidentialMPTClawback** transactions (issuer reclaims tokens from a holder).

Adds `holder` (the account being clawed back from).

### 3. Pedersen commitment

#### `PedersenCommitment(amount uint64, bf BlindingFactor) (commitment Commitment, err error)`

Computes `C = amount*G + bf*H` where G and H are generator points. The result is a 33-byte compressed point. Two commitments with the same amount and blinding factor always produce the same output (deterministic).

```go
commitment, err := mptcrypto.PedersenCommitment(1000, blindingFactor)
// commitment: 33-byte compressed point (starts with 0x02 or 0x03)
```

### 4. Proof generation

Each XRPL confidential transaction type requires a specific proof. The proof convinces validators that the transaction is valid without revealing the actual amounts.

#### `GenerateConvertProof(pubkey PublicKey, privkey PrivateKey, ctxHash ContextHash) ([64]byte, error)`

**Schnorr proof of knowledge.** Proves you own the private key for the public key being registered, bound to the transaction via ctxHash.

Used in: **ConfidentialMPTConvert** (registering a keypair on the ledger).

#### `GenerateConvertBackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, params PedersenProofParams) ([816]byte, error)`

**Compact AND-composed sigma proof + single Bulletproof range proof.** Proves:
1. Your encrypted balance matches the Pedersen commitment (sigma proof over balance witness)
2. After subtracting the convert-back amount, the remaining balance is non-negative (range proof over remainder commitment)

Used in: **ConfidentialMPTConvertBack**.

#### `GenerateClawbackProof(privkey PrivateKey, pubkey PublicKey, ctxHash ContextHash, amount uint64, ciphertext Ciphertext) ([64]byte, error)`

**Compact sigma proof.** Proves that the ciphertext decrypts to exactly the claimed amount, without revealing the private key.

Used in: **ConfidentialMPTClawback** (issuer proves the amount they're clawing back matches the encrypted balance).

#### `GenerateSendProof(privkey PrivateKey, pubkey PublicKey, amount uint64, participants []Participant, txBF BlindingFactor, ctxHash ContextHash, amountCommitment Commitment, balanceParams PedersenProofParams) ([]byte, error)`

**Compact AND-composed sigma proof + aggregated Bulletproof range proof** (the most complex one). Combines:
1. **Equality proof** - same amount encrypted for sender, receiver, issuer (and optionally auditor)
2. **Amount linkage** - ElGamal ciphertext matches amount commitment
3. **Balance linkage** - sender's encrypted balance matches balance commitment
4. **Range proof** - amount and remaining balance are both in [0, 2^64-1]

Returns a fixed-size byte slice of `SendProofSize` (946) bytes.

Used in: **ConfidentialMPTSend**.

### 5. Proof verification (top-level)

These are the four main verifiers, one per transaction type. Each returns `nil` on success or an error on failure.

#### `VerifyConvertProof(proof [64]byte, pubkey PublicKey, ctxHash ContextHash) error`

Verifies the Schnorr proof from a ConfidentialMPTConvert.

#### `VerifyConvertBackProof(proof [816]byte, pubkey PublicKey, ciphertext Ciphertext, balanceCommit Commitment, amount uint64, ctxHash ContextHash) error`

Verifies the compact sigma + range proof from a ConfidentialMPTConvertBack.

#### `VerifySendProof(proof []byte, participants []Participant, senderCt Ciphertext, amountCommit, balanceCommit Commitment, ctxHash ContextHash) error`

Verifies the compact sigma + range proof from a ConfidentialMPTSend.

#### `VerifyClawbackProof(proof [64]byte, amount uint64, pubkey PublicKey, ciphertext Ciphertext, ctxHash ContextHash) error`

Verifies the compact sigma proof from a ConfidentialMPTClawback.

### 6. Proof verification (internal components)

These verify individual pieces of a send proof. Useful for debugging or testing each component in isolation.

#### `VerifyRevealedAmount(amount uint64, bf BlindingFactor, holder, issuer Participant, auditor *Participant) error`

Verifies that a plaintext amount and blinding factor are consistent with the participants' ciphertexts. `auditor` can be `nil` if there's no auditor.

#### `VerifySendRangeProof(proof [754]byte, amountCommit, balanceCommitment Commitment, ctxHash ContextHash) error`

Verifies a double bulletproof: both the transfer amount and remaining balance are in [0, 2^64-1].

### 7. Utilities

#### `ComputeConvertBackRemainder(commitmentIn Commitment, amount uint64) (Commitment, error)`

Subtracts a transparent (public) amount from a hidden Pedersen commitment, producing a new commitment for the remaining balance. Used in convert-back to compute the post-transaction balance commitment.

```go
remainder, err := mptcrypto.ComputeConvertBackRemainder(balanceCommitment, 500)
```

---

## CGo patterns used in this package

If you need to modify or extend the bindings, here's how the CGo boundary works.

### The preamble

At the top of `mptcrypto_cgo.go`:

```go
/*
#cgo CFLAGS: -I${SRCDIR}/../deps/include -I${SRCDIR}/../deps/include/utility
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../deps/libs/linux-amd64 -lmpt-crypto -lsecp256k1 ...

#include "mpt_utility.h"
*/
import "C"
```

The comment block before `import "C"` is special: it's the **CGo preamble**. `#cgo` directives set compiler/linker flags per platform. `#include` pulls in the C header. `import "C"` must appear immediately after the comment (no blank line).

### Passing byte arrays to C with unsafe.Pointer

The C functions expect raw `uint8_t*` pointers. Go arrays live in Go-managed memory, so we take the address of the first element and cast:

```go
// Go side
var pubkey PublicKey

// Pass to C: "give me a *C.uint8_t pointing to pubkey[0]"
C.some_c_function((*C.uint8_t)(unsafe.Pointer(&pubkey[0])))
```

**What's happening step by step:**

1. `&pubkey[0]` - address of the first byte (type `*byte`)
2. `unsafe.Pointer(...)` - convert to an untyped pointer (required bridge between Go and C pointer types)
3. `(*C.uint8_t)(...)` - cast to the C type the function expects

This is safe because:
- Go arrays are contiguous in memory, just like C arrays
- The C function only reads/writes within the declared size
- The Go array stays alive for the duration of the C call (it's on the stack or referenced)

### Converting Go structs to C structs

For complex types (account IDs, participants, proof params), we use helper functions that copy field-by-field:

```go
func toParticipant(p Participant) C.mpt_confidential_participant {
    var c C.mpt_confidential_participant
    for i, b := range p.PubKey {
        c.pubkey[i] = C.uint8_t(b)
    }
    for i, b := range p.Ciphertext {
        c.ciphertext[i] = C.uint8_t(b)
    }
    return c
}
```

We copy byte-by-byte instead of using `unsafe` casts on structs because Go and C may have different struct layouts (padding, alignment). Byte-by-byte copy is always correct.

### Passing slices to C (variable-length data)

For `GenerateSendProof` and similar functions that take a variable number of participants:

```go
cParts := make([]C.mpt_confidential_participant, n)
for i, p := range participants {
    cParts[i] = toParticipant(p)
}

// Pass the slice's backing array to C
C.mpt_get_confidential_send_proof(
    // ...
    &cParts[0],         // pointer to first element
    C.size_t(n),        // length
    // ...
)
```

Go slices have a backing array that's contiguous, so `&cParts[0]` gives C a valid pointer to `n` consecutive structs.

### Optional (nullable) pointers

Some C functions accept `NULL` for optional parameters (e.g., auditor in `VerifyRevealedAmount`):

```go
var cAuditor *C.mpt_confidential_participant  // nil by default (maps to NULL)
if auditor != nil {
    a := toParticipant(*auditor)
    cAuditor = &a
}
C.mpt_verify_revealed_amount(..., cAuditor)
```

A nil Go pointer becomes `NULL` in C.

### Error handling

All C functions return `int`: 0 for success, -1 for failure. The Go wrappers turn non-zero returns into errors:

```go
ret := C.mpt_some_function(...)
if ret != 0 {
    return fmt.Errorf("mpt_some_function failed with code %d", ret)
}
```

### The no-CGo build

`mptcrypto_nocgo.go` provides identical function signatures without linking the C library. Every function immediately returns `ErrCgoRequired`; no argument validation or cryptographic work is performed. This lets applications handle unavailable confidential functionality as a normal error while the rest of the codebase continues to compile without CGo.
