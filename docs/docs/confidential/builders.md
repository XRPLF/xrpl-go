---
sidebar_position: 2
---

# builders

## Overview

The `confidential/builder` package is the high-level entry point for XLS-96 transaction construction.

Each operation comes in two forms:

- `Build*`: queries live ledger state through a `LedgerQuerier`.
- `Prepare*`: builds the same transaction from explicit inputs, which is useful for offline signing or test fixtures.

The `LedgerQuerier` interface is intentionally small, and both `rpc.Client` and `websocket.Client` satisfy it:

```go
type LedgerQuerier interface {
    GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error)
    GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error)
}
```

## Builder families

### `BuildConvert` and `PrepareConvert`

Use these for `ConfidentialMPTConvert`.

- Queries or accepts the account sequence.
- Resolves issuer and optional auditor encryption keys from the `MPTokenIssuance`.
- Detects whether the holder is opting in for the first time.
- Encrypts the converted amount for the holder, issuer, and optional auditor.
- On first use, adds `HolderEncryptionKey` and generates the Schnorr proof required to register it.

`Amount == 0` is allowed here because zero-amount convert is the opt-in path for registering a holder key.

```go
tx, err := builder.BuildConvert(client, builder.BuildConvertParams{
    Account:       holderAddress,
    IssuanceID:    issuanceID,
    Amount:        100,
    HolderPrivKey: holderPrivKeyHex,
    HolderPubKey:  holderPubKeyHex,
})
```

### `BuildSend` and `PrepareSend`

Use these for `ConfidentialMPTSend`.

- Resolves issuer, auditor, sender, and destination encryption keys.
- Reads the sender `MPToken` state, including `ConfidentialBalanceSpending` and `ConfidentialBalanceVersion`.
- Decrypts the sender's current confidential balance with the supplied private key.
- Encrypts the transfer amount for sender, destination, issuer, and optional auditor.
- Builds both Pedersen commitments and the composite send proof.

This path requires the destination holder to already have a registered `HolderEncryptionKey`.

```go
tx, err := builder.BuildSend(client, builder.BuildSendParams{
    Account:       senderAddress,
    Destination:   receiverAddress,
    IssuanceID:    issuanceID,
    Amount:        25,
    SenderPrivKey: senderPrivKeyHex,
    SenderPubKey:  senderPubKeyHex,
})
```

### `BuildConvertBack` and `PrepareConvertBack`

Use these for `ConfidentialMPTConvertBack`.

- Resolves issuer and optional auditor keys.
- Reads and decrypts the holder's current confidential spending balance.
- Uses `ConfidentialBalanceVersion` from ledger state.
- Builds the encrypted withdrawal amount, balance commitment, and convert-back proof.

```go
tx, err := builder.BuildConvertBack(client, builder.BuildConvertBackParams{
    Account:       holderAddress,
    IssuanceID:    issuanceID,
    Amount:        10,
    HolderPrivKey: holderPrivKeyHex,
    HolderPubKey:  holderPubKeyHex,
})
```

### `BuildClawback` and `PrepareClawback`

Use these for `ConfidentialMPTClawback`.

- Resolves the issuer sequence and issuer encryption key.
- Reads the holder's `IssuerEncryptedBalance` from the ledger.
- Generates the equality proof that binds the clawback amount to the issuer-visible ciphertext.

```go
tx, err := builder.BuildClawback(client, builder.BuildClawbackParams{
    Account:       issuerAddress,
    Holder:        holderAddress,
    IssuanceID:    issuanceID,
    Amount:        50,
    IssuerPrivKey: issuerPrivKeyHex,
})
```

### `BuildMergeInbox` and `PrepareMergeInbox`

Use these for `ConfidentialMPTMergeInbox`.

- Only needs the account, issuance ID, and sequence.
- Performs no cryptographic work.
- Lets a holder move confidential inbox balance into spending balance.

```go
tx, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
    Account:    holderAddress,
    IssuanceID: issuanceID,
})
```

## `Build*` vs `Prepare*`

Choose `Build*` when you have access to a live ledger connection and want the SDK to resolve:

- account sequence numbers;
- issuer and auditor encryption keys;
- holder `MPToken` fields such as `HolderEncryptionKey`, `ConfidentialBalanceSpending`, `IssuerEncryptedBalance`, and `ConfidentialBalanceVersion`.

Choose `Prepare*` when you already have those values and want deterministic, offline transaction assembly.

## Typical flow

1. Enable confidential transfers on the issuance with `MPTokenIssuanceCreate` or `MPTokenIssuanceSet`, including `IssuerEncryptionKey` and optionally `AuditorEncryptionKey`.
2. Generate a holder keypair with `confidential/elgamal.GenerateKeypair()`.
3. Opt the holder in with `BuildConvert` or `PrepareConvert`, optionally with `Amount: 0` for key registration only.
4. Use `BuildSend` for confidential transfers between opted-in holders.
5. Use `BuildMergeInbox` after receiving confidential transfers, if the holder wants to spend the received balance.
6. Use `BuildConvertBack` to move confidential balance back into public MPT balance.

## Signing and submission

Builders return concrete transaction structs from `xrpl/transaction`, so the rest of the flow is the same as other XRPL transactions: autofill any remaining fields if needed, sign with a wallet, then submit through RPC or WebSocket.

```go
tx, err := builder.BuildSend(client, params)
if err != nil {
    return err
}

signed, err := wallet.Sign(tx)
if err != nil {
    return err
}

_, err = client.SubmitTx(signed, nil)
return err
```

## Common failure cases

Most builder errors are explicit and map to missing ledger state or invalid inputs:

- `ErrEncryptionKeyNotSet`: the issuance does not yet have the issuer encryption key configured.
- `ErrReceiverNotOptedIn`: the destination holder has no registered `HolderEncryptionKey`.
- `ErrMPTokenNotFound`: the account does not yet have the expected `MPToken` ledger entry.
- `ErrInsufficientBalance`: the requested confidential send or convert-back amount exceeds the decrypted balance.
- `ErrCryptoFailed`: a cryptographic primitive failed or the provided private key does not match ledger state.
