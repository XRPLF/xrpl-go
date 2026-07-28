---
sidebar_position: 1
sectionTopLabel: Packages
---

# confidential

## Overview

The `confidential` packages add support for XLS-96 confidential MPT workflows in `xrpl-go`.

They cover three layers:

- `confidential/mptcrypto`: low-level CGo bindings to the XRPLF `mpt-crypto` library.
- `confidential/elgamal`, `confidential/commitment`, `confidential/proof`: Go-friendly hex-string APIs for encryption, commitments, context hashes, and zero-knowledge proofs.
- `confidential/builder`: high-level transaction builders for constructing confidential MPT transactions from either ledger state or explicit inputs.

## Build requirements

Confidential MPT support depends on CGo-enabled builds.

```bash
CGO_ENABLED=1 go test ./confidential/...
```

If CGo is disabled, `confidential/mptcrypto` returns `ErrCgoRequired`, which means the builder and proof helpers cannot perform the underlying cryptographic operations.

## Package map

### `confidential/elgamal`

Use this package when you need raw confidential amount encryption helpers.

- `GenerateKeypair()` creates a confidential holder, issuer, or auditor keypair.
- `GenerateBlindingFactor()` creates the shared randomness used across ciphertexts and commitments.
- `Encrypt()` encrypts a `uint64` amount to a compressed secp256k1 public key.
- `Decrypt()` decrypts a confidential balance ciphertext with the matching private key.

### `confidential/commitment`

Use this package to create Pedersen commitments for confidential amounts.

- `Create(amount, bf)` returns the compressed commitment used by confidential proofs and transaction fields such as `AmountCommitment` and `BalanceCommitment`.

### `confidential/proof`

Use this package if you want fine-grained control over proof generation or verification.

- Context-hash helpers bind proofs to a specific XRPL transaction: `ConvertContextHash`, `ConvertBackContextHash`, `SendContextHash`, `ClawbackContextHash`.
- Top-level proof helpers mirror the confidential transaction families: `GenerateConvertProof`, `GenerateConvertBackProof`, `GenerateSendProof`, `GenerateClawbackProof`.
- Verification helpers let you validate proofs before submission or in tests.

All APIs in this layer operate on hex strings and classic XRPL addresses, which makes them suitable for transaction assembly.

## Confidential transaction types

The `xrpl/transaction` package now includes five confidential MPT transaction types:

- `ConfidentialMPTConvert`: moves public MPT into confidential balance and optionally registers the holder encryption key on first use.
- `ConfidentialMPTSend`: sends confidential MPT between opted-in holders using encrypted amounts plus a composite proof.
- `ConfidentialMPTConvertBack`: converts confidential balance back into public balance with a proof of sufficient confidential funds.
- `ConfidentialMPTClawback`: lets the issuer reclaim a holder's confidential balance with an equality proof.
- `ConfidentialMPTMergeInbox`: merges a holder's confidential inbox balance into their spending balance.

Related XRPL types were extended as well:

- `MPTokenIssuanceCreate` and `MPTokenIssuanceSet` support confidential-transfer flags and issuer/auditor encryption keys.
- `MPToken` and `MPTokenIssuance` ledger-entry types expose confidential balance and encryption-key fields.

## When to use builders

Use [`builders`](/docs/confidential/builders) when you want the SDK to:

- fetch ledger state such as `Sequence`, registered encryption keys, and confidential balance fields;
- decrypt the holder's current confidential balance when required;
- generate ciphertexts, commitments, and ZK proofs with the correct context hash;
- return a ready-to-sign `xrpl/transaction` struct.

Drop down to `elgamal`, `commitment`, and `proof` when you need custom transaction assembly, explicit control over proof inputs, or standalone verification in tests.
