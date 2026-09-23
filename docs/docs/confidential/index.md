# Confidential transfers

Use the optional confidential module to build encrypted MPT transactions and their proofs. Most applications should start with [builders](/docs/confidential/builders), not the native cryptography API.

You need this module if your application generates ciphertexts or proofs, or decrypts confidential balances. If another implementation supplies those values, core alone can encode, sign, and submit the transactions. Core field validation does not verify a ZK proof.

## Before you start

- Follow [installation](/docs/confidential/installation) for the module version and native toolchain.
- Use a ledger with the required amendments enabled. SDK support and a final specification do not establish Mainnet availability.
- Keep **wallet signing keys** separate from **confidential encryption keys**. Wallets authorize transactions. ElGamal keys encrypt amounts, decrypt balances, and support proofs.

## How the flow fits together

```text
configure issuance -> register issuer/auditor encryption keys
                   -> authorize holder -> opt in with holder encryption key
public MPT -> convert -> confidential inbox -> merge -> spending balance
                                                      -> send -> recipient inbox
                                                      -> convert back -> public MPT
```

Builders read the needed ledger state, prepare ciphertexts and proofs, and return typed transactions. You still autofill remaining network fields, sign with a core wallet, submit, and check the validated result. Do not change a sequence or Ticket after preparing a proof that binds it.

## Choose a guide

| Goal | Guide |
| --- | --- |
| Build, sign, and submit a confidential operation | [Builders](/docs/confidential/builders) |
| Read a spendable confidential balance | [Balance reads](/docs/confidential/builders#reading-a-spending-balance) |
| Combine dependent operations atomically | [Confidential batches](/docs/confidential/batch) |
| Assemble encrypted fields and proofs yourself | [Cryptographic helpers](/docs/confidential/primitives) |
| Work with fixed-size bytes and native proof contracts | [Native API reference](/docs/confidential/mptcrypto) |

## Operations

| Transaction | Purpose |
| --- | --- |
| `ConfidentialMPTConvert` | Convert public MPT and optionally register a holder key on first use |
| `ConfidentialMPTMergeInbox` | Make inbox funds available in the spending balance |
| `ConfidentialMPTSend` | Transfer between opted-in holders |
| `ConfidentialMPTConvertBack` | Return confidential funds to public balance |
| `ConfidentialMPTClawback` | Let the issuer reclaim a holder's confidential balance |

These transaction models and their ledger fields belong to core. See [MPT operations](/docs/xrpl/mpt) for issuance capabilities and key registration. The [XLS-96 specification](https://github.com/XRPLF/XRPL-Standards/tree/master/XLS-0096-confidential-mpt) defines the protocol.

## Transaction cost

Confidential MPT transactions cost ten network base fees: one base fee plus an extra multiplier of nine. Multisigning adds the usual one base fee per signer. The multiplier also applies to a confidential transaction inside a Batch.

RPC and WebSocket autofill apply this cost only when `Fee` is absent. A manually supplied fee is not overwritten. A fee sized for an ordinary transaction can underpay and fail with `telINSUF_FEE_P`. See [Submission and finality](/docs/xrpl/submission) for the SDK's autofill behavior.

## Examples

- [Offline example](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/offline): prepares an issuance, key registration, a holder conversion, and an inbox merge. It does not connect, sign, or submit. Proof generation still requires the native toolchain.
- [RPC lifecycle](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/rpc) and [WebSocket lifecycle](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/ws): fund test wallets, configure an issuance and holders, and submit transactions on Devnet.

Read the examples before running them. For a repository checkout, follow the [workspace instructions](https://github.com/XRPLF/xrpl-go/blob/main/CONTRIBUTING.md#work-on-confidential-helpers). For version changes, use the [confidential changelog](/changelog/confidential/v0.1.x/changelog).

## Security

Protect signing secrets and encryption private keys. Do not print, log, commit, or send them to telemetry. Test with non-production funds and read the [security and audit notice](https://github.com/XRPLF/xrpl-go#security-and-audits) before production use.
