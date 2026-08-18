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

First-time detection reads the holder's `MPToken`, which `ConfidentialMPTConvert` debits, so the entry
must already exist. A holder that has not authorized the issuance gets `ErrMPTokenNotFound` instead of
being treated as a first-time opt-in, and a failed read reports `ErrLedgerQuery` rather than silently
taking the first-time path.

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
- Decrypts the sender's current confidential balance with the supplied private key and inclusive `BalanceRange`.
- Encrypts the transfer amount for sender, destination, issuer, and optional auditor.
- Builds both Pedersen commitments and the composite send proof.

This path requires the destination holder to already have a registered `HolderEncryptionKey`.

`DestinationTag` and `CredentialIDs` are optional and forwarded to the transaction unchanged. Set `DestinationTag` when the destination is a hosted account, and `CredentialIDs` when the destination sits behind a permissioned domain.

```go
tx, err := builder.BuildSend(client, builder.BuildSendParams{
    Account:       senderAddress,
    Destination:   receiverAddress,
    IssuanceID:    issuanceID,
    Amount:        25,
    SenderPrivKey: senderPrivKeyHex,
    SenderPubKey:  senderPubKeyHex,
    BalanceRange: elgamal.AmountRange{
        Low:  0,
        High: 1_000_000,
    },
})
```

### `BuildConvertBack` and `PrepareConvertBack`

Use these for `ConfidentialMPTConvertBack`.

- Resolves issuer and optional auditor keys.
- Reads and decrypts the holder's current confidential spending balance within the supplied inclusive `BalanceRange`.
- Uses `ConfidentialBalanceVersion` from ledger state.
- Builds the encrypted withdrawal amount, balance commitment, and convert-back proof.

```go
tx, err := builder.BuildConvertBack(client, builder.BuildConvertBackParams{
    Account:       holderAddress,
    IssuanceID:    issuanceID,
    Amount:        10,
    HolderPrivKey: holderPrivKeyHex,
    HolderPubKey:  holderPubKeyHex,
    BalanceRange: elgamal.AmountRange{
        Low:  0,
        High: 1_000_000,
    },
})
```

### Bounded balance decryption

`BuildSend` and `BuildConvertBack` decrypt the current on-ledger spending balance before constructing a transaction. Their `BalanceRange` is the expected range of that **current balance**, not the amount being sent or converted back.

The `Low` and `High` bounds are inclusive and must contain the plaintext balance. They must satisfy `Low <= High < math.MaxUint64`. Decryption searches the interval linearly, so use the narrowest practical range; unnecessarily large ranges can make transaction construction slow. Omitting `BalanceRange` produces `[0, 0]`, which only succeeds for a zero balance.

`PrepareSend` and `PrepareConvertBack` do not decrypt ledger state because their `CurrentBalance` is supplied explicitly.

### `BuildClawback` and `PrepareClawback`

Use these for `ConfidentialMPTClawback`.

- Resolves the issuer sequence and issuer encryption key.
- Reads the holder's `IssuerEncryptedBalance` from the ledger.
- Decrypts that ciphertext with `IssuerPrivKey` to derive the amount.
- Generates the equality proof that binds the clawback amount to the issuer-visible ciphertext.

A clawback always removes the holder's complete confidential balance, so `BuildClawback` derives
the amount rather than accepting one. The search is bounded by `BalanceRange` and additionally
capped at the issuance's `ConfidentialOutstandingAmount`, which no holder balance can exceed.
Supply the amount yourself only on the offline `PrepareClawback` path, via `ClawbackParams.Amount`.

```go
tx, err := builder.BuildClawback(client, builder.BuildClawbackParams{
    Account:       issuerAddress,
    Holder:        holderAddress,
    IssuanceID:    issuanceID,
    IssuerPrivKey: issuerPrivKeyHex,
    BalanceRange:  elgamal.AmountRange{Low: 0, High: 1_000},
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

Each proof commits to the transaction sequence, so a `Prepare*` helper that emits a proof rejects a
zero `Sequence` with `ErrMissingSequence` rather than produce a proof a later autofill would
invalidate. The two proof-free forms are exempt: `PrepareMergeInbox`, and `PrepareConvert` for a
holder whose encryption key is already registered. Both accept a zero `Sequence` and can be autofilled.

`Build*` also preflights the issuance capabilities the protocol requires, so a transaction the
network would reject never costs a fee and a sequence.

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

## Address forms

Every address field accepts either a classic address or an X-address. The builder resolves
both to the same account, so `Account` given as `rHb9…` and as its X-address form name the
same account for the self-send and self-clawback checks. Addresses are normalized to their
classic spelling before they reach the ledger queries and keylet computation, and the proof
layer binds the decoded account ID, so the address form never changes a proof.

A tagged X-address is accepted only where the transaction has a companion tag field:

- `Account` has `SourceTag`, so a tagged X-address is allowed.
- `Destination` has `DestinationTag`, so a tagged X-address is allowed unless you also set
  `BuildSendParams.DestinationTag`, which would name the tag twice.
- `Holder` has no tag field, because `ConfidentialMPTClawback` defines none, so a tagged
  X-address is rejected.

ACCOUNT_ZERO is rejected in every address field. It decodes cleanly in either form, but no
keypair can produce it, so it can never sign a transaction nor hold an `MPToken`.

## Common failure cases

Most builder errors are explicit and map to missing ledger state or invalid inputs:

- `ErrEncryptionKeyNotSet`: the issuance does not yet have the issuer encryption key configured.
- `ErrReceiverNotOptedIn`: the destination holder has no registered `HolderEncryptionKey`.
- `ErrMPTokenNotFound`: the account does not yet have the expected `MPToken` ledger entry.
- `ErrIssuanceNotFound`: the `MPTokenIssuance` ledger entry does not exist.
- `ErrInsufficientBalance`: the requested confidential send or convert-back amount exceeds the decrypted balance.
- `ErrMissingSequence`: a proof-bearing `Prepare*` helper was given a zero `Sequence`.
- `ErrKeyMismatch`: the supplied public key differs from the one registered on the ledger.
- `ErrInvalidCredentialIDs`: `BuildSendParams.CredentialIDs` is not a valid hexadecimal string
  array. It wraps `transaction.ErrInvalidCredentialIDs`, so `errors.Is` matches either sentinel.
- `elgamal.ErrInvalidAmountRange`: `BalanceRange` is inverted or its upper bound is `math.MaxUint64`.
- `ErrCryptoFailed`: a cryptographic primitive failed, or the current balance falls outside `BalanceRange`.

Address fields report the field that failed and wrap the reason:

- `ErrInvalidAccount`, `ErrInvalidDestination`, `ErrInvalidHolder`: the address is neither a
  classic address nor an X-address, or it decodes to ACCOUNT_ZERO. Match
  `transaction.ErrZeroAccountID` with `errors.Is` to tell the two apart.
- `ErrInvalidHolder` wrapping `transaction.ErrAccountIDTagNotAllowed`: a tagged X-address was
  used in `Holder`, which has no companion tag field.
- `ErrInvalidDestination` wrapping `transaction.ErrDuplicateXAddressTag`: `Destination` is a
  tagged X-address and `DestinationTag` is also set.
- `ErrInvalidAddress`: an address failed to decode inside the MPToken keylet helper, which
  serves `Account`, `Destination`, and `Holder` alike and so names no field. The builders
  validate their address fields first, so this reports against the field only in code that
  calls the helper directly.

The issuance capability checks mirror the conditions the network enforces:

- `ErrConfidentialDisabled`: the issuance does not have `lsfMPTCanHoldConfidentialBalance` set.
- `ErrTransferDisabled`: a confidential send needs `lsfMPTCanTransfer`, which the issuance does not have.
- `ErrTransferFeeSet`: the issuance charges a transfer fee, which confidential sends forbid.
- `ErrAmountExceedsOutstanding`: `BalanceRange.Low` is above the issuance `ConfidentialOutstandingAmount`.
