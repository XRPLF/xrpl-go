# mptcrypto

Go bindings for the [XRPLF/mpt-crypto](https://github.com/xrplf/mpt-crypto) C library. This package is the **only** place in the codebase that imports `"C"` (CGo). Everything above this layer (elgamal/, proofs/, commitment/) is pure Go.

## Build requirements

CGo must be enabled (`CGO_ENABLED=1`). The vendored C libraries live in `confidential/deps/libs/<os-arch>/`. If you build without CGo, every function returns `ErrCgoRequired`.

```bash
# normal build (CGo on by default)
go test ./confidential/mptcrypto/...

# force CGo off (all functions return ErrCgoRequired)
CGO_ENABLED=0 go test ./confidential/mptcrypto/...
```

## How this package is organized

```
mptcrypto/
  types.go             # Size constants, Participant, PedersenProofParams
  mptcrypto_cgo.go     # Real implementations (only built with CGo)
  mptcrypto_nocgo.go   # Stubs that return ErrCgoRequired (built without CGo)
  mptcrypto_test.go    # Tests (build tag: cgo)
```

Every function works with **fixed-size byte arrays** (`[32]byte`, `[33]byte`, `[66]byte`, etc.). Hex encoding/decoding happens in the layers above (elgamal/, proofs/, commitment/), never here.

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
| `SchnorrProofSize` | 65 | Schnorr proof of knowledge |
| `EqualityProofSize` | 98 | Equality proof (same value encrypted under different keys) |
| `PedersenLinkSize` | 195 | Pedersen linkage proof |
| `SingleBulletproofSize` | 688 | Single bulletproof (range proof for 1 value) |
| `DoubleBulletproofSize` | 754 | Double bulletproof (range proof for 2 values) |
| `ConvertBackProofSize` | 883 | Linkage + range proof (195 + 688) |
| `MaxParticipants` | 255 | Max participants in a send (C API uses uint8_t) |

### Structs

```go
// A party in a confidential send (public key + their encrypted amount).
type Participant struct {
    PubKey     [PubKeySize]byte     // 33 bytes
    Ciphertext [CiphertextSize]byte // 66 bytes
}

// Parameters for generating Pedersen linkage proofs.
type PedersenProofParams struct {
    Commitment     [CommitmentSize]byte     // 33 bytes
    Amount         uint64
    Ciphertext     [CiphertextSize]byte     // 66 bytes
    BlindingFactor [BlindingFactorSize]byte // 32 bytes
}
```

---

## Function reference

### 1. ElGamal encryption

These handle key generation, encryption, and decryption for confidential amounts.

#### `GenerateKeypair() (privkey [32]byte, pubkey [33]byte, err error)`

Creates a new secp256k1 ElGamal keypair.

```go
priv, pub, err := mptcrypto.GenerateKeypair()
// priv: 32-byte private key
// pub:  33-byte compressed public key (starts with 0x02 or 0x03)
```

#### `GenerateBlindingFactor() (bf [32]byte, err error)`

Generates a cryptographically random 32-byte scalar. Used as the randomness parameter (`r`) when encrypting amounts or creating Pedersen commitments.

```go
bf, err := mptcrypto.GenerateBlindingFactor()
```

#### `EncryptAmount(amount uint64, pubkey [33]byte, bf [32]byte) (ct [66]byte, err error)`

Encrypts an amount using ElGamal. The ciphertext is 66 bytes: two compressed EC points concatenated (C1 || C2).

```go
ct, err := mptcrypto.EncryptAmount(1000, pubkey, blindingFactor)
// ct: 66-byte ciphertext
```

#### `DecryptAmount(ciphertext [66]byte, privkey [32]byte) (uint64, error)`

Decrypts an ElGamal ciphertext back to the original amount. Uses a baby-step giant-step (BSGS) lookup table internally, so very large values (close to `math.MaxUint64`) may fail to decrypt.

```go
amount, err := mptcrypto.DecryptAmount(ciphertext, privkey)
```

### 2. Context hashes

Every ZK proof is bound to a specific transaction via a **context hash**. This prevents proof reuse across transactions. Each transaction type has its own hash function because the inputs differ.

All context hash functions return a `[32]byte` hash.

#### `ConvertContextHash(account [20]byte, iss [24]byte, seq uint32) ([32]byte, error)`

For **ConfidentialMPTConvert** transactions (public amount -> confidential).

- `account`: the sender's 20-byte account ID
- `iss`: the 24-byte MPTokenIssuance ID
- `seq`: the transaction sequence number

#### `ConvertBackContextHash(account [20]byte, iss [24]byte, seq, ver uint32) ([32]byte, error)`

For **ConfidentialMPTConvertBack** transactions (confidential -> public amount).

Same as above plus `ver` (the version counter from the ledger object).

#### `SendContextHash(account [20]byte, iss [24]byte, seq uint32, dest [20]byte, ver uint32) ([32]byte, error)`

For **ConfidentialMPTSend** transactions (confidential transfer between accounts).

Adds `dest` (destination account ID) and `ver`.

#### `ClawbackContextHash(account [20]byte, iss [24]byte, seq uint32, holder [20]byte) ([32]byte, error)`

For **ConfidentialMPTClawback** transactions (issuer reclaims tokens from a holder).

Adds `holder` (the account being clawed back from).

### 3. Pedersen commitment

#### `PedersenCommitment(amount uint64, bf [32]byte) (commitment [33]byte, err error)`

Computes `C = amount*G + bf*H` where G and H are generator points. The result is a 33-byte compressed point. Two commitments with the same amount and blinding factor always produce the same output (deterministic).

```go
commitment, err := mptcrypto.PedersenCommitment(1000, blindingFactor)
// commitment: 33-byte compressed point (starts with 0x02 or 0x03)
```

### 4. Proof generation

Each XRPL confidential transaction type requires a specific proof. The proof convinces validators that the transaction is valid without revealing the actual amounts.

#### `GenerateConvertProof(pubkey [33]byte, privkey [32]byte, ctxHash [32]byte) ([65]byte, error)`

**Schnorr proof of knowledge.** Proves you own the private key for the public key being registered, bound to the transaction via ctxHash.

Used in: **ConfidentialMPTConvert** (registering a keypair on the ledger).

#### `GenerateConvertBackProof(privkey [32]byte, pubkey [33]byte, ctxHash [32]byte, amount uint64, params PedersenProofParams) ([883]byte, error)`

**Linkage + range proof.** Proves:
1. Your encrypted balance matches the Pedersen commitment (linkage)
2. After subtracting the convert-back amount, the remaining balance is non-negative (range proof, via bulletproof)

Used in: **ConfidentialMPTConvertBack**.

#### `GenerateClawbackProof(privkey [32]byte, pubkey [33]byte, ctxHash [32]byte, amount uint64, ciphertext [66]byte) ([98]byte, error)`

**Equality proof.** Proves that the ciphertext decrypts to exactly the claimed amount, without revealing the private key.

Used in: **ConfidentialMPTClawback** (issuer proves the amount they're clawing back matches the encrypted balance).

#### `GenerateSendProof(privkey [32]byte, amount uint64, participants []Participant, txBF [32]byte, ctxHash [32]byte, amountParams, balanceParams PedersenProofParams) ([]byte, error)`

**Full send proof** (the most complex one). Combines:
1. **Equality proof** - same amount encrypted for sender, receiver, issuer (and optionally auditor)
2. **Amount linkage** - ElGamal ciphertext matches amount commitment
3. **Balance linkage** - sender's encrypted balance matches balance commitment
4. **Range proof** - amount and remaining balance are both in [0, 2^64-1]

Returns a variable-length byte slice (size depends on number of participants). Use `GetSendProofSize(n)` to compute the expected size.

Used in: **ConfidentialMPTSend**.

#### `GenerateAmountLinkageProof(pubkey [33]byte, bf [32]byte, ctxHash [32]byte, params PedersenProofParams) ([195]byte, error)`

**Standalone linkage proof** between an ElGamal ciphertext and a Pedersen commitment for the transaction amount. This is a building block used internally by `GenerateSendProof`, but exposed separately for testing.

#### `GenerateBalanceLinkageProof(privkey [32]byte, pubkey [33]byte, ctxHash [32]byte, params PedersenProofParams) ([195]byte, error)`

**Standalone linkage proof** for the sender's balance. Same idea as amount linkage, but uses the private key (because the sender's balance ciphertext was created with their key). Also a building block exposed for testing.

### 5. Proof verification (top-level)

These are the four main verifiers, one per transaction type. Each returns `nil` on success or an error on failure.

#### `VerifyConvertProof(proof [65]byte, pubkey [33]byte, ctxHash [32]byte) error`

Verifies the Schnorr proof from a ConfidentialMPTConvert.

#### `VerifyConvertBackProof(proof [883]byte, pubkey [33]byte, ciphertext [66]byte, balanceCommit [33]byte, amount uint64, ctxHash [32]byte) error`

Verifies the linkage + range proof from a ConfidentialMPTConvertBack.

#### `VerifySendProof(proof []byte, participants []Participant, senderCt [66]byte, amountCommit, balanceCommit [33]byte, ctxHash [32]byte) error`

Verifies the full send proof. `proof` is variable-length (depends on participant count).

#### `VerifyClawbackProof(proof [98]byte, amount uint64, pubkey [33]byte, ciphertext [66]byte, ctxHash [32]byte) error`

Verifies the equality proof from a ConfidentialMPTClawback.

### 6. Proof verification (internal components)

These verify individual pieces of a send proof. Useful for debugging or testing each component in isolation.

#### `VerifyRevealedAmount(amount uint64, bf [32]byte, holder, issuer Participant, auditor *Participant) error`

Verifies that a plaintext amount and blinding factor are consistent with the participants' ciphertexts. `auditor` can be `nil` if there's no auditor.

#### `VerifyAmountLinkage(proof [195]byte, ciphertext [66]byte, pubkey [33]byte, commitment [33]byte, ctxHash [32]byte) error`

Verifies that the ElGamal ciphertext and Pedersen commitment encode the same amount.

#### `VerifyBalanceLinkage(proof [195]byte, ciphertext [66]byte, pubkey [33]byte, commitment [33]byte, ctxHash [32]byte) error`

Same as amount linkage but for the sender's balance. Note: unlike `VerifyAmountLinkage`, this does NOT require a `secp256k1_context` internally (asymmetry in the C API, not a bug).

#### `VerifyEqualityProof(proof []byte, participants []Participant, ctxHash [32]byte) error`

Verifies that all participants' ciphertexts encrypt the same value.

#### `VerifySendRangeProof(proof [754]byte, amountCommit, remainderCommit [33]byte, ctxHash [32]byte) error`

Verifies a double bulletproof: both the transfer amount and remaining balance are in [0, 2^64-1].

### 7. Utilities

#### `GetSendProofSize(nRecipients int) int`

Returns the expected proof size in bytes for a send with `nRecipients` participants. Use this to pre-allocate or validate proof buffers.

```go
size := mptcrypto.GetSendProofSize(3) // 3 participants: sender, dest, issuer
```

#### `ComputeConvertBackRemainder(commitmentIn [33]byte, amount uint64) ([33]byte, error)`

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
var pubkey [33]byte

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

### secp256k1 context

Some C verification functions need a `secp256k1_context*`. The Go wrappers get it internally:

```go
ctx := C.mpt_secp256k1_context()   // returns a shared global context
C.mpt_verify_amount_linkage(ctx, ...)
```

This is never exposed to callers. The context is managed by the C library.

### Error handling

All C functions return `int`: 0 for success, -1 for failure. The Go wrappers turn non-zero returns into errors:

```go
ret := C.mpt_some_function(...)
if ret != 0 {
    return fmt.Errorf("mpt_some_function failed with code %d", ret)
}
```

### The no-CGo build

`mptcrypto_nocgo.go` has the build tag `//go:build !cgo`. It provides identical function signatures but every function returns `ErrCgoRequired`. This lets the rest of the codebase compile and run tests for pure-Go packages even without the C library.
