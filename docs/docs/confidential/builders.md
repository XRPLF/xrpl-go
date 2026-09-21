# Confidential builders

The `confidential/builder` package is the high-level entry point for XLS-96 transaction construction. It is part of the [optional confidential module](/docs/confidential/installation), not the core module.

Each operation comes in two forms:

- `Build*`: queries live ledger state through a `LedgerQuerier`.
- `Prepare*`: builds the same transaction from explicit inputs, which is useful for offline signing or test fixtures.

## Before you build

1. [Install confidential helpers](/docs/confidential/installation) with the native toolchain.
2. Use a network that supports the required confidential MPT amendments.
3. Enable the issuance's confidential capability and register its issuer encryption key. Register an auditor key too if the issuance uses one. See [MPT operations](/docs/xrpl/mpt).
4. Ensure each holder has an `MPToken` entry, using `MPTokenAuthorize` as required. The first confidential convert does not create that entry.
5. Keep wallet signing keys separate from ElGamal encryption keys. Builders need encryption keys. Wallets sign the final transaction.

## Typical flow

```text
issuance setup -> holder authorization -> convert/opt in -> merge inbox -> send
                                                                         |
receive <- confidential inbox <-------------------------------------------+
   -> merge inbox -> spend or convert back
```

`BuildConvert` with `Amount: 0` registers a holder key without converting public MPT. A positive convert credits the inbox, so merge before spending those funds. The recipient must opt in before a send. Received funds also need an inbox merge before they become spendable.

## Build, sign, and submit

This fragment runs inside a function returning an error. `client` is a configured RPC client, `signingWallet` is the sender's core wallet, and `params` is a populated `builder.BuildSendParams` with the sender's encryption keys and a balance range. The imports are `confidential/builder` plus `fmt` and your client/wallet packages.

```go
tx, err := builder.BuildSend(client, params)
if err != nil {
	return err
}
flat := tx.Flatten()
if err := client.Autofill(&flat); err != nil {
	return err
}
blob, _, err := signingWallet.Sign(flat)
if err != nil {
	return err
}
response, err := client.SubmitTxBlobAndWait(blob, false)
if err != nil {
	return err
}
if response.Meta.TransactionResult != "tesSUCCESS" {
	return fmt.Errorf("transaction result: %s", response.Meta.TransactionResult)
}
return nil
```

Builders return transaction structs, not signed blobs. Keep the prepared sequence or Ticket unchanged: proofs bind that nonce. See [Submission and finality](/docs/xrpl/submission) for timeouts and uncertain outcomes.

For complete application setup, use the [RPC example](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/rpc) or [WebSocket example](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/ws). For offline input assembly, use the [offline example](https://github.com/XRPLF/xrpl-go/tree/main/confidential/examples/offline).

## Choose an operation

| Goal | Helpers |
| --- | --- |
| Opt in or convert public MPT | `BuildConvert`, `PrepareConvert` |
| Send confidential MPT | `BuildSend`, `PrepareSend` |
| Convert back to public MPT | `BuildConvertBack`, `PrepareConvertBack` |
| Reclaim a holder's confidential balance | `BuildClawback`, `PrepareClawback` |
| Make inbox funds spendable | `BuildMergeInbox`, `PrepareMergeInbox` |
| Read a spendable balance | `GetSpendingBalance` |
| Combine ordered operations | [Confidential batches](/docs/confidential/batch) |

The operation examples below are fragments. Supply the named addresses, keys, issuance ID, and client. Check each returned error before using the result. Use the [Go builder reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/confidential/builder) for full parameter types.

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

This path requires the destination holder to already be initialized to receive: a registered
`HolderEncryptionKey`, a `ConfidentialBalanceInbox`, and the mirror balances the issuance
implies. A destination missing any of them, or with no `MPToken` at all, reports
`ErrReceiverNotOptedIn`.

`DestinationTag` and `CredentialIDs` are optional and forwarded to the transaction unchanged. Set `DestinationTag` when the destination is a hosted account, and `CredentialIDs` when the destination requires credential-based deposit authorization.

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

`BuildClawback` and `GetSpendingBalance` bound their searches the same way, additionally capping `High` at the issuance's `ConfidentialOutstandingAmount`, which no single holder balance can exceed.

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

- Resolves the account sequence.
- Reads the `MPTokenIssuance` to confirm it allows confidential balances and is not locked.
- Reads the holder `MPToken` to confirm it carries both confidential balances and the holder
  encryption key, and that the holder is neither locked nor unauthorized.
- Does not require `IssuerEncryptionKey`, which `ConfidentialMPTMergeInbox` never reads.
- Performs no cryptographic work.
- Lets a holder move confidential inbox balance into spending balance.

```go
tx, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
	Account:    holderAddress,
	IssuanceID: issuanceID,
})
```

## Reading a spending balance

`GetSpendingBalance` is the read-only counterpart to the builders: it resolves a holder's
`ConfidentialBalanceSpending` from the ledger and decrypts it with that holder's own ElGamal
private key. Nothing is built or submitted, and no account sequence is read, because a balance
read spends none.

```go
balance, err := builder.GetSpendingBalance(client, builder.SpendingBalanceParams{
	Holder:        holderAddress,
	IssuanceID:    issuanceID,
	HolderPrivKey: holderPrivKeyHex,
	BalanceRange:  elgamal.AmountRange{Low: 0, High: 1_000},
})
```

- Reads the holder `MPToken` and the `MPTokenIssuance` from one validated ledger, pinned by
  hash after the first read, so the supply that bounds the search can never predate the
  balance it must cover.
- Returns `ErrMPTokenNotFound` when the holder holds no `MPToken` for the issuance.
- Returns `0` when the `MPToken` exists but carries no spending ciphertext, which is a holder
  that has never converted: the first `ConfidentialMPTConvert` writes an encrypted zero spending
  balance alongside the inbox credit. The issuance is not read in that case, and no
  cryptographic work is done.
- Excludes `ConfidentialBalanceInbox`. An inbox is not spendable until a
  `ConfidentialMPTMergeInbox` moves it, so counting it would report a balance the holder
  cannot send or convert back.
- Bounds the search by `BalanceRange`, capped at the issuance `ConfidentialOutstandingAmount`,
  exactly as `BuildClawback` does. If omitted, `BalanceRange` is `[0, 0]`, so callers must set
  a range that contains any nonzero balance. The SDK does not default the range to the
  issuance's whole confidential supply, because that range can be as large as the issuance.

Like every other decryption in this package, it needs a CGo-enabled build. The zero-balance
case above is the one answer it can give without one.

## `Build*` vs `Prepare*`

Choose `Build*` when you have access to a live ledger connection and want the SDK to resolve:

- account sequence numbers;
- issuer and auditor encryption keys;
- holder `MPToken` fields such as `HolderEncryptionKey`, `ConfidentialBalanceSpending`, `IssuerEncryptedBalance`, and `ConfidentialBalanceVersion`.

Choose `Prepare*` when you already have those values and want offline transaction assembly. Offline does not mean deterministic: proof and ciphertext construction can use fresh randomness.

Each proof commits to the nonce the transaction spends, so a `Prepare*` helper that emits a proof
rejects options carrying neither `Sequence` nor `TicketSequence` with `ErrMissingSequence` rather
than produce a proof a later autofill would invalidate. The two proof-free forms are exempt:
`PrepareMergeInbox`, and `PrepareConvert` for a holder whose encryption key is already registered.
Both accept a zero nonce and can be autofilled.

`Build*` checks issuance capabilities and required confidential state before submission. These checks catch some failures early, but they do not guarantee success: ledger state can change between construction and validation.

Some conditions the network enforces are left to it. A destination that requires a destination
tag (`tecDST_TAG_NEEDED`), a destination behind deposit authorization (`tecNO_PERMISSION`), and
an issuance that authorizes through a permissioned domain all depend on account state or
credentials the builder does not read. Preflight covers the issuance capabilities and the
confidential state each transactor requires, not the destination's own access policy.

## Transaction options

Every `Build*Params` embeds `TxOptions`, which carries the fields that are about the transaction
rather than the confidential operation: which nonce authorizes it, and who submits it.

```go
type TxOptions struct {
	Sequence       uint32
	TicketSequence uint32
	Delegate       string
}
```

Set either `Sequence` or `TicketSequence`, never both: XRPL requires `Sequence` to be 0 whenever a
transaction spends a Ticket, so setting both returns `ErrConflictingNonce`.

Every helper accepts a `TicketSequence`. xrpld hashes the sequence proxy into every confidential
context hash, and the sequence proxy is the ticket whenever the transaction spends one, so a proof
built here commits to the ticket sequence rather than to an account sequence. That covers a
first-time `PrepareConvert`, `PrepareClawback`, `PrepareSend`, and `PrepareConvertBack`.

There is one case a Ticket cannot rescue, and no builder can detect it, so it is documented rather
than refused. `ConfidentialMPTSend` and `ConfidentialMPTConvertBack` commit their proofs to the
submitter's own `ConfidentialBalanceVersion`, which a send, a convert-back, a merge-inbox, or a
clawback against that holder bumps. Of several such transactions built against a single reading of
one `MPToken`, the first to land bumps the version and the rest fail with `tecBAD_PROOF`, each
paying a fee and destroying its Ticket. Submit them one at a time on that issuance and wait for
validation.

The collision is per `MPToken`, because the version lives on the `(issuance, holder)` entry, so
plenty still runs in parallel on Tickets:

- Clawbacks against different holders. `PrepareClawback` binds the target holder's
  `IssuerEncryptedBalance` rather than any state of the submitting issuer, so they bind disjoint
  entries. Two against the same holder are redundant, because a clawback removes that holder's
  balance in full.
- Converts. A convert credits the inbox rather than the spending balance, so it never bumps the
  version, and a repeat convert carries no proof at all.
- Merges and sends across different issuances, which never touch the same `MPToken`.

A merge carries no proof, so nothing of its own can go stale, but it bumps the version regardless.
A ticketed merge can therefore land out of order and invalidate a pending send or convert-back on
the same issuance, so land a merge and wait for validation before preparing either.

A `Build*` helper reads the account sequence only when both are zero, so supplying either one keeps
the build off the account query entirely:

```go
tx, err := builder.BuildMergeInbox(client, builder.BuildMergeInboxParams{
	TxOptions: builder.TxOptions{
		TicketSequence: ticketSequence,
		Delegate:       delegateAddress,
	},
	Account:    holderAddress,
	IssuanceID: issuanceID,
})
```

The Ticket is spent from the transaction `Account`'s account root, so it must be one that account
created with `TicketCreate`. Only the `Account`'s sequence is consumed, never the `Delegate`'s, so
a Ticket the delegate owns is rejected on-ledger with `tefNO_TICKET`.

`Delegate` names the account submitting on the transaction account's behalf, per XLS-75. It is
rejected when it is not a valid address, decodes to ACCOUNT_ZERO, carries an X-address tag, or
names the transaction account itself, and the sentinels are the ones `BaseTx.Validate` reports.
`ConfidentialMPTConvert` is marked non-delegable by xrpld, so `BuildConvert` and `PrepareConvert`
reject any delegate with `ErrDelegateNotAllowed`.

Each `Prepare*Params` reads the options from its embedded `Build*Params`, so a composite literal
sets them there:

```go
params := builder.MergeInboxParams{
	BuildMergeInboxParams: builder.BuildMergeInboxParams{
		TxOptions:  builder.TxOptions{Sequence: sequence},
		Account:    holderAddress,
		IssuanceID: issuanceID,
	},
}
```

The promoted selectors, such as `params.Sequence` and `params.Delegate`, stay available after
construction and reach that same value.

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

| Cause | Recovery |
| --- | --- |
| Missing issuance, holder entry, or registered key | Complete issuance setup, holder authorization, or opt-in before rebuilding |
| Insufficient spendable funds | Check the spending balance separately from the inbox, then merge if needed |
| Invalid or too-narrow balance range | Supply bounds that contain the current balance, not just the transfer amount |
| Stale balance version | Wait for the earlier transaction to validate, read fresh state, and rebuild the proof |
| Invalid nonce or address | Correct the input before rebuilding, do not patch a signed/proven transaction |
| Cryptographic failure | Check key/state consistency and the decryption range before retrying |

Match sentinel errors with `errors.Is`, not error-message text. The following details describe the SDK preflight boundary, not a complete list of server result codes.

Most builder errors are explicit and map to missing ledger state or invalid inputs:

- `ErrEncryptionKeyNotSet`: the issuance does not yet have the issuer encryption key configured.
- `ErrReceiverNotOptedIn`: the destination holder is not initialized to receive. It has no
  `MPToken`, no registered `HolderEncryptionKey`, no `ConfidentialBalanceInbox`, or is missing
  a mirror balance the issuance implies.
- `ErrMPTokenNotFound`: the account does not yet have the expected `MPToken` ledger entry.
- `ErrMissingSenderState`: an `MPToken` exists but lacks confidential state the transaction
  needs, such as a spending balance, a mirror balance, or the holder encryption key. A
  clawback reports a holder that never opted in this way too.
- `ErrIssuanceNotFound`: the `MPTokenIssuance` ledger entry does not exist.
- `ErrInsufficientBalance`: the requested confidential send or convert-back amount exceeds the
  decrypted balance, or a convert amount exceeds the holder's public `MPTAmount`.
- `ErrMissingSequence`: a proof-bearing `Prepare*` helper was given neither a `Sequence` nor a
  `TicketSequence`.
- `ErrConflictingNonce`: both `Sequence` and `TicketSequence` were set.
- `ErrDelegateNotAllowed`: a `Delegate` was set on a type `NonDelegatableTransactionsMap` lists,
  which among the confidential types is `ConfidentialMPTConvert`.
- `ErrKeyMismatch`: the supplied public key differs from the one registered on the ledger.
- `ErrInvalidCredentialIDs`: a nonempty `BuildSendParams.CredentialIDs` list must contain
  1 to 8 distinct, nonzero, 256-bit hexadecimal IDs. Hex letter case does not affect uniqueness.
  It wraps `transaction.ErrInvalidCredentialIDs`, so `errors.Is` matches either sentinel.
- `ErrStaleBalanceVersion`: a confidential transaction of the holder's own is still in flight and
  has already moved `ConfidentialBalanceVersion`, so a proof built against the validated ledger
  would be rejected. Rebuild once it validates.
- `ErrInvalidLedgerState`: a ledger response was missing, malformed, or did not come from the
  validated ledger the build selected.
- `ErrInvalidTransaction`: the assembled transaction failed its own `Validate()`.
- `elgamal.ErrInvalidAmountRange`: `BalanceRange` is inverted, its upper bound is
  `math.MaxUint64`, or its `Low` is above the issuance `ConfidentialOutstandingAmount`
  that `BuildClawback` and `GetSpendingBalance` cap the search at, which puts every
  possible balance outside the range.
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
- `ErrClawbackDisabled`: a clawback needs `lsfMPTCanClawback`, which the issuance does not have.
- `ErrIssuanceLocked`: the issuance has `lsfMPTLocked`, so every balance of it is locked.
- `ErrHolderLocked`: the holder's `MPToken` has `lsfMPTLocked`. A clawback is exempt, because
  an issuer must be able to claw back from a holder it has locked. A send checks both
  participants and prefixes the error with `sender` or `destination` to name the side that
  blocked it.
- `ErrHolderNotAuthorized`: the issuance has `lsfMPTRequireAuth` and the holder's `MPToken`
  lacks `lsfMPTAuthorized`. A send names the participant the same way `ErrHolderLocked` does. An issuance that authorizes through a permissioned domain is left
  to the network, because the credentials that path accepts are not read here.
- `ErrAmountExceedsOutstanding`: a convert-back `Amount` exceeds the issuance
  `ConfidentialOutstandingAmount`.

## Custom ledger access

The `LedgerQuerier` interface is intentionally small, and both `rpc.Client` and `websocket.Client` satisfy it:

```go
type LedgerQuerier interface {
	GetAccountInfo(req *account.InfoRequest) (*account.InfoResponse, error)
	GetLedgerEntry(req *ledger.EntryRequest) (*ledger.EntryResponse, error)
}
```
