---
sidebar_position: 1
sectionTopLabel: Introduction
---

# Getting Started

[`xrpl-go`](https://github.com/XRPLF/xrpl-go) is a Go SDK for the [XRP Ledger](https://xrpl.org/). It provides address codecs, key management, binary serialization, typed transaction and ledger models, RPC and WebSocket clients, and local wallet signing.

## Choose a module

The repository contains two independently versioned Go modules:

| Module | Provides | Native toolchain |
| --- | --- | --- |
| `github.com/Peersyst/xrpl-go` | Transaction and ledger models, codecs, clients, key management, and wallet signing | Not required |
| [`github.com/Peersyst/xrpl-go/confidential`](/docs/confidential) | Confidential MPT builders, encryption, balance decryption, commitments, and proofs | Required for cryptographic operations |

Core includes the confidential transaction models and ledger fields. It can encode, sign, and submit these transactions when your application supplies the ciphertexts and proofs. Install the optional module when you want the SDK to generate those fields or decrypt balances.

The dependency goes from confidential to core only. Starting with core `v0.3.1`, core Go module downloads exclude the confidential module and its native bundles. Repository clones and GitHub source archives still contain both modules.

Both modules require Go `1.25.13` or later. See [Installation](/docs/installation) for core setup, or [Install confidential helpers](/docs/confidential/installation) for the native toolchain, module versions, and migration steps.

## Core packages

| Package | Use it for |
| --- | --- |
| `address-codec` | Encode and decode XRPL classic addresses and X-addresses |
| `binary-codec` | Encode and decode XRPL objects and transactions in canonical binary format |
| [`keypairs`](/docs/keypairs) | Generate seeds, derive keypairs, sign payloads, and verify signatures |
| [`xrpl/rpc`](/docs/xrpl/rpc) | Send JSON-RPC requests, autofill transactions, submit transactions, and fund test wallets |
| [`xrpl/websocket`](/docs/xrpl/websocket) | Connect to WebSocket servers, make requests, submit transactions, and subscribe to ledger streams |
| [`xrpl/transaction`](/docs/xrpl/transaction) | Build typed XRPL transaction models |
| [`xrpl/ledger-entry-types`](/docs/xrpl/ledger-entry-types) | Read typed ledger state |
| [`xrpl/wallet`](/docs/xrpl/wallet) | Create wallets, sign transactions locally, multisign transactions, and authorize payment channels |

The optional module has a separate [package guide](/docs/confidential#package-map) and [builder guide](/docs/confidential/builders).

## Transaction lifecycle

The usual write path is:

1. Build a typed transaction, such as `transaction.Payment`.
2. Call `Flatten()` to get a `transaction.FlatTransaction`.
3. Call `client.Autofill()` to add network fields such as `Fee`, `Sequence`, and `LastLedgerSequence`.
4. Sign locally with `wallet.Sign()`.
5. Submit with `client.SubmitTxBlobAndWait()`, or use `client.SubmitTxAndWait()` to autofill, sign, submit, and wait in one call.

Autofill requires network access. Wallet signing only needs the transaction data and wallet credentials, and does not require the confidential module.

A transaction is a command you submit. Ledger entries are the state you query after validation. Transaction metadata describes the ledger entries that changed.

For confidential transactions, the optional builders generate the encrypted fields and proofs before this normal signing and submission flow. Core field validation does not verify the cryptographic validity of a ZK proof.

## Examples and development

- [Core quickstart](https://github.com/XRPLF/xrpl-go#quickstart) and [core examples](https://github.com/XRPLF/xrpl-go/tree/main/examples).
- [Confidential examples](/docs/confidential#examples), including an offline walkthrough and RPC/WebSocket devnet programs.
- [Contributor guide](https://github.com/XRPLF/xrpl-go/blob/main/CONTRIBUTING.md) for tests and documentation development.

Root `go test ./...` only tests core. Use the [development workspace](/docs/confidential/installation#development-workspace) and separate module test targets when working on both modules.

## Versions and release notes

Core releases use tags such as `v0.3.1`. Confidential releases use tags such as `confidential/v0.1.0`. A core update does not automatically update the optional module.

See the [core changelog](https://github.com/XRPLF/xrpl-go/blob/main/CHANGELOG.md), [confidential changelog](https://github.com/XRPLF/xrpl-go/blob/main/confidential/CHANGELOG.md), and [release guide](https://github.com/XRPLF/xrpl-go/blob/main/RELEASING.md).

## Security

Never print, log, commit, or send real seeds, private keys, or mnemonics to telemetry. Protect confidential encryption private keys as well as wallet signing keys.

The signing functionality has not been independently audited. Test with non-production funds on a compatible test network and review signing behavior before production use.
