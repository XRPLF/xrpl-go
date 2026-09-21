# Cryptographic helpers

Most applications should use [builders](/docs/confidential/builders). Use these higher-level helpers for custom transaction assembly and tests. They accept hex strings and addresses rather than the fixed-size byte arrays used by the [native API](/docs/confidential/mptcrypto). They still require the [native toolchain](/docs/confidential/installation#native-build-requirements).

## `confidential/elgamal`

Use this package when you need raw confidential amount encryption helpers.

- `GenerateKeypair()` creates a confidential holder, issuer, or auditor keypair.
- `GenerateBlindingFactor()` creates the shared randomness used across ciphertexts and commitments.
- `Encrypt(amount, pubKeyHex, bfHex)` encrypts a `uint64` amount to a compressed secp256k1 public key under a given blinding factor. Reusing one blinding factor across the ciphertexts of a single transaction is what lets a proof tie them together.
- `Decrypt(ciphertextHex, privateKeyHex, amountRange)` decrypts a confidential balance ciphertext with the matching private key by searching an inclusive `AmountRange`.

Decryption requires bounds that contain the plaintext amount and satisfy `Low <= High < math.MaxUint64`. Search cost grows linearly with the interval size, so use the narrowest practical range:

```go
amount, err := elgamal.Decrypt(ciphertextHex, privateKeyHex, elgamal.AmountRange{
	Low:  0,
	High: 1_000_000,
})
```

## `confidential/commitment`

Use this package to create Pedersen commitments for confidential amounts.

- `Create(amount, bf)` returns the compressed commitment used by confidential proofs and transaction fields such as `AmountCommitment` and `BalanceCommitment`.

## `confidential/proof`

Use this package if you want fine-grained control over proof generation or verification.

- Context-hash helpers bind proofs to a specific XRPL transaction: `ConvertContextHash`, `ConvertBackContextHash`, `SendContextHash`, `ClawbackContextHash`.
- Top-level proof helpers mirror the confidential transaction families: `GenerateConvertProof`, `GenerateConvertBackProof`, `GenerateSendProof`, `GenerateClawbackProof`.
- Verification helpers let you validate proofs before submission or in tests.

All APIs in this layer operate on hex strings and XRPL addresses in either form, which makes them suitable for transaction assembly. Context hashes bind the decoded AccountID, so a classic address and its X-address form produce the same hash.

Every `Generate*Proof` helper verifies the proof it just produced before returning it, because the native
generator reports no error for a mismatched amount or key pair. That check costs one verification per
generation. Include that verification cost when measuring your application. Callers that batch proof generation
should budget for it. The `Build*` and `Prepare*` helpers inherit the same cost, since they generate
through these functions.

## API references

- [ElGamal](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/elgamal)
- [Commitments](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/commitment)
- [Proofs](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/proof)

Protect encryption private keys and use fresh randomness for each independent operation. Shared randomness within one proof is an input relationship, not permission to reuse it across unrelated transactions.
