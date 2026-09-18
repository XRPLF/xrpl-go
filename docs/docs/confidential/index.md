---
sidebar_position: 1
sectionTopLabel: Packages
---

# confidential

## Overview

The optional `github.com/Peersyst/xrpl-go/confidential` module provides XLS-96 confidential MPT builders and cryptographic helpers. It depends on the core `xrpl-go` module, but core does not depend on it.

Use this module to generate encrypted transaction fields and proofs, read confidential spending balances, or build transactions from ledger state. Applications that supply ciphertexts and proofs from another implementation can use core alone.

See [installation and versions](/docs/confidential/installation) for the initial `v0.1.0` release, its core `v0.3.1` minimum dependency, and migration from the combined module. Import paths such as `github.com/Peersyst/xrpl-go/confidential/builder` remain unchanged.

## Build requirements

The cryptographic helpers require cgo and a supported native toolchain. Core transaction models, codecs, and wallet signing do not require the optional module.

To test the helpers from a repository checkout:

```bash
make workspace
make test-confidential
```

If cgo is disabled, `confidential/mptcrypto` returns `ErrCgoRequired` for cryptographic operations. See the [native build requirements](/docs/confidential/installation#native-build-requirements) for supported platforms and the fallback behavior.

## Package map

| Package | Purpose |
| --- | --- |
| [`confidential/builder`](/docs/confidential/builders) | Build transactions from ledger state with `Build*`, prepare them from explicit inputs with `Prepare*`, and read spending balances with `GetSpendingBalance` |
| `confidential/elgamal` | Generate encryption keys and encrypt or decrypt amounts |
| `confidential/commitment` | Create Pedersen commitments |
| `confidential/proof` | Generate and verify proofs and transaction context hashes |
| [`confidential/mptcrypto`](/docs/confidential/mptcrypto) | Low-level native API, data types, proof contracts, and errors |

### `confidential/elgamal`

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

### `confidential/commitment`

Use this package to create Pedersen commitments for confidential amounts.

- `Create(amount, bf)` returns the compressed commitment used by confidential proofs and transaction fields such as `AmountCommitment` and `BalanceCommitment`.

### `confidential/proof`

Use this package if you want fine-grained control over proof generation or verification.

- Context-hash helpers bind proofs to a specific XRPL transaction: `ConvertContextHash`, `ConvertBackContextHash`, `SendContextHash`, `ClawbackContextHash`.
- Top-level proof helpers mirror the confidential transaction families: `GenerateConvertProof`, `GenerateConvertBackProof`, `GenerateSendProof`, `GenerateClawbackProof`.
- Verification helpers let you validate proofs before submission or in tests.

All APIs in this layer operate on hex strings and XRPL addresses in either form, which makes them suitable for transaction assembly. Context hashes bind the decoded AccountID, so a classic address and its X-address form produce the same hash.

Every `Generate*Proof` helper verifies the proof it just produced before returning it, because the native
generator reports no error for a mismatched amount or key pair. That check costs one verification per
generation, measured at roughly 35% added wall time for a send proof. Callers that batch proof generation
should budget for it. The `Build*` and `Prepare*` helpers inherit the same cost, since they generate
through these functions.

## Confidential transaction types

The core `xrpl/transaction` package includes five confidential MPT transaction types. These models, their codecs, and normal wallet signing do not depend on the optional module:

- `ConfidentialMPTConvert`: moves public MPT into confidential balance and optionally registers the holder encryption key on first use.
- `ConfidentialMPTSend`: sends confidential MPT between opted-in holders using encrypted amounts plus a composite proof.
- `ConfidentialMPTConvertBack`: converts confidential balance back into public balance with a proof of sufficient confidential funds.
- `ConfidentialMPTClawback`: lets the issuer reclaim a holder's confidential balance with an equality proof.
- `ConfidentialMPTMergeInbox`: merges a holder's confidential inbox balance into their spending balance.

Related XRPL types were extended as well:

- `MPTokenIssuanceCreate` and `MPTokenIssuanceSet` carry the confidential-transfer capability flags. `MPTokenIssuanceSet` also carries `IssuerEncryptionKey` and the optional `AuditorEncryptionKey`, which is where an issuance registers its keys.
- `MPToken` and `MPTokenIssuance` ledger-entry types expose confidential balance and encryption-key fields.

### Transaction cost

Confidential MPT transactions cost ten network base fees, not one. rippled charges one base fee for the transaction itself plus an extra multiplier of nine. Multisigning adds the usual one base fee per signer on top, and the same multiplier applies to a confidential transaction nested inside a `Batch`.

RPC and WebSocket autofill apply the multiplier for you, but only when the transaction has no `Fee` field. A hand-set `Fee` is submitted unchanged, so a value sized for an ordinary transaction underpays and the submission fails with `telINSUF_FEE_P`.

## When to use builders

Use [`builders`](/docs/confidential/builders) when you want the SDK to:

- fetch ledger state such as `Sequence`, registered encryption keys, and confidential balance fields;
- decrypt the holder's current confidential balance within a caller-supplied inclusive `BalanceRange` when required;
- generate ciphertexts, commitments, and ZK proofs with the correct context hash;
- return a ready-to-sign `xrpl/transaction` struct.

Flatten the returned transaction, autofill the remaining network fields, sign with a core wallet, and submit through the RPC or WebSocket client. Proofs bind the transaction nonce, so do not change a prepared sequence or Ticket before submission. Core field validation does not verify the cryptographic validity of a ZK proof.

Use `elgamal`, `commitment`, and `proof` directly when you need custom transaction assembly, control over proof inputs, or standalone verification in tests.

## Examples

From the repository root, set up the workspace and run the offline example from the optional module:

```bash
make workspace
cd confidential
CGO_ENABLED=1 go run ./examples/offline
```

The example prepares an issuance, issuer-key registration, a holder conversion, and an inbox merge. It does not connect, sign, or submit transactions. Offline proof generation still needs the native toolchain.

The [`rpc`](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/rpc) and [`ws`](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/ws) examples run a full lifecycle against devnet. They fund test wallets, create an issuance, register keys, opt holders in, and submit confidential transactions. Read the examples before running them.

## Development and releases

Examples are in `confidential/examples/`, and integration tests are in `confidential/integration/`. Root `go test ./...` does not include this module. Use the [development workspace](/docs/confidential/installation#development-workspace) and separate test targets to check it against the local core checkout.

The module has its own [changelog](https://github.com/XRPLF/xrpl-go/blob/main/confidential/CHANGELOG.md) and uses Git tags such as `confidential/v0.1.0`. Core versions and releases are independent. See the [release guide](https://github.com/XRPLF/xrpl-go/blob/main/RELEASING.md) for release order and checks against published dependencies.

## Security

Protect both wallet signing secrets and confidential encryption private keys. Do not print, log, commit, or send them to telemetry. Test with non-production funds and read the [security and audit notice](https://github.com/XRPLF/xrpl-go#security-and-audits) before production use.
