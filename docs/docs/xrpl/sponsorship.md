# Sponsorship

Use sponsorship when another account pays fees or provides reserves. This guide covers the Go SDK's preparation, signing, and preflight rules for both clients. See [XLS-68](https://xls.xrpl.org/xls/XLS-0068-sponsored-fees-and-reserves.html) for the protocol specification and check amendment availability on your target network.

## Choose a flow

| Flow | What you need | Before account signing |
| --- | --- | --- |
| Co-signed | A sponsor wallet that will authorize the transaction | Set `Sponsor` and `SponsorFlags`, then autofill |
| Pre-funded | An existing authorized `Sponsorship` ledger object | Call `AddPreFundedSponsor`, then autofill |

```text
sponsor fields -> autofill -> account signature(s) -> sponsor signature(s) -> submit final blob
```

Do not modify or autofill the transaction after signing. For pre-funded sponsorship, there is no sponsor-signing step.

## Sponsor signing

Sponsor helpers accept flattened transaction maps or encoded blobs. They return new maps without changing the inputs. Errors return no partial result.

Use `SignAsSponsor` for maps or `SignAsSponsorBlob` for encoded input. Use `CombineSponsorSigners` or `CombineSponsorSignersBlob` for independent multisigner fragments. `AddPreFundedSponsor` prepares the unsigned pre-funded flow. See the [wallet API reference](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/wallet) for signatures and option types.

Signing returns the transaction, final blob, and final hash. Combining returns the transaction and final blob. Use `hash.SignTxBlob(blob)` to calculate the combined hash.

### Co-signed sponsorship

Set `Sponsor` and `SponsorFlags` (`types.SpfSponsorFee`, `types.SpfSponsorReserve`, or both) **before account signing**. Autofill before signing too. Then sign the account and pass its signed map or blob to `SignAsSponsor` or `SignAsSponsorBlob`.

The helpers preserve account signatures and the top-level `SigningPubKey`. They add `SponsorSignature`. For Payments, `SignAsSponsor` also converts `DeliverMax` to `Amount` on the returned map before signing and encoding. Conflicting values return `wallet.ErrAmountAndDeliverMaxMustBeIdentical`. Account signatures must already be present. This signing order is an offline helper policy, not a claim that consensus requires that invocation order.

Use `SignAsSponsorOptions{Multisign: true}` for sponsor multisigning. Each signer signs the same account-signed transaction independently. Combine those fragments with `CombineSponsorSigners`. Do not chain sponsor signing calls. If the account also uses multisigning, combine its fragments before sponsor signing.

For a regular-key multisigner, set `MultisignAccount` to the signer account. A nonempty override enables multisigning and takes precedence over the wallet address. For single-signing with a regular key, construct the wallet with `FromSeed(seed, sponsorAccount)`.

The combiner requires canonical wire equivalence except for `SponsorSignature.Signers`. It compares account authorization and other sponsor fields, but not nonserialized metadata. Each input signer list must be sorted and unique by decoded AccountID. The result is sorted, and overlaps between fragments retain the **first input occurrence** by decoded identity. If duplicates have different signatures, the first is retained even if it is invalid. No signature or quorum verification is performed.

### Pre-funded sponsorship

Call `AddPreFundedSponsor` before signing. It sets sponsor fields without adding a sponsor signature. It permits an absent or empty `SigningPubKey`, but rejects existing account, counterparty, sponsor, or Batch authorization fields, including partial or null values.

This helper does not create or fund a ledger `Sponsorship` object. The caller needs an existing sponsorship with enough resources and with the applicable require-sign flags disabled. The helper does not check funds or ledger authorization.

For pre-funded fee or reserve sponsorship, xrpld looks up the `Sponsorship` object between `Sponsor` and `Account`. For fee sponsorship with a `Delegate`, it uses `Sponsor` and `Delegate` instead. Reserve sponsorship with `Delegate` is not supported.

A pre-funded object does not replace the sponsor signature for either account-level `SponsorshipTransfer` case:

- `tfSponsorshipCreate` with no `ObjectID`.
- `tfSponsorshipReassign` with no `ObjectID`.

Both require `SponsorSignature`. Use co-signed sponsorship for these cases.

### Fees and amendment compatibility

For fee sponsorship, if the matching `Sponsorship` object exists, xrpld deducts the fee from its `FeeAmount` pool even when `SponsorSignature` is present. A sponsor signature does not bypass the pool or its `MaxFee` limit. Co-signed fee sponsorship uses the sponsor's account balance only when no matching object exists.

Pass the **combined planned account and sponsor multisigner count** to the existing RPC or WebSocket `AutofillMultisigned` before account signing. A single sponsor signature adds no surcharge. A single account signature also adds no multisigner count. Leave Fee absent so autofill can calculate it. A supplied Fee is not overwritten.

Sponsor signing requires `Sponsor` and `fixCleanup3_4_0` on the target network. These helpers always use sponsor prefixes `0x53504E00` and `0x53504D00`. They do not silently fall back to ordinary signing prefixes. Pre-funded use requires the `Sponsor` amendment but no sponsor signature prefix.

All checks are offline structural checks. They do not establish cryptographic validity of supplied signatures, ledger authorization, quorum, sponsorship balance, or amendment activation. `transaction.InspectSponsorFields` applies the shared raw sponsorship rules, not full transaction validation, and returns an independent `*types.SponsorSignature` (nil when absent). Discard the returned signature when only validation is needed. It retains field presence, including an explicitly empty `SigningPubKey`, and returns no signature on error.

Submit the final blob without modifying it or autofilling again.

See the runnable sponsor examples for [RPC](https://github.com/XRPLF/xrpl-go/tree/main/examples/sponsor-signing/rpc) and [WebSocket](https://github.com/XRPLF/xrpl-go/tree/main/examples/sponsor-signing/ws) for complete co-signed, multisigned, and pre-funded flows.

## Sponsorship preflight

`ValidateSponsorship` checks a sponsored transaction (XLS-68) against the `Sponsorship` ledger entry between its sponsor and its sponsee, reading the current ledger. The check is explicit and optional: autofill and submission never query sponsorship on their own.

The transaction must carry `Sponsor` and a valid `SponsorFlags`, and either a `Fee` or a nonempty `estimatedFee` in drops. Before any lookup, the sponsorship fields are checked with the same rules `BaseTx.Validate` applies, and a failure is returned as the matching `xrpl/transaction` error: for example `ErrSponsorFieldsMissing`, `ErrInvalidSponsorFlags`, `ErrSponsorAccountConflict`, `ErrSponsorDelegateConflict` for reserve sponsorship on a delegated transaction, `ErrReserveSponsorshipNotAllowed`, `ErrInnerBatchFeeSponsorship`, or `ErrInvalidSponsorSignature`. A present `SponsorSignature` must hold `SigningPubKey` with `TxnSignature` or a `Signers` array, except on an inner Batch transaction (`tfInnerBatchTxn`), where it may hold at most an empty `SigningPubKey` because the sponsor signs the outer `Batch`.

The sponsee is the transaction's `Delegate` when present and its `Account` otherwise. Only the sponsor and resolved sponsee addresses are converted to classic addresses for the lookup, because xrpld accepts only classic addresses in the `sponsorship` selector. Unrelated addresses, explicit tags, and Batch inner transactions are outside this lookup's validation scope. The caller's transaction is not modified.

Without a `Sponsorship` entry, only a sponsor co-signature (`SponsorSignature`) authorizes the sponsorship. With an entry, its budget always applies, even to a co-signed transaction, because rippled prefers the pre-funded fee payer whenever the entry exists: a sponsored fee must fit within `FeeAmount` and any `MaxFee` cap. Pre-funded use is additionally rejected when the entry sets `lsfSponsorshipRequireSignForFee` or `lsfSponsorshipRequireSignForReserve` for the requested sponsorship type. A zero fee draws nothing from the entry.

A nil error means the preflight completed; read `SponsorshipValidation.Valid` and `SponsorshipValidation.Reason`, which wraps an `ErrSponsorship*` sentinel. A non-nil error means the preflight could not run, because the transaction inputs were unusable or the `ledger_entry` lookup failed. Only `entryNotFound` counts as an absent entry; transport, permission, and decoding failures, and an entry whose `Owner` or `Sponsee` differs from the requested pair (`ErrSponsorshipEntryMismatch`), are returned as errors.

The check does not confirm that the sponsor account exists, which rippled requires even for a co-signed transaction, or that a co-signing sponsor's balance minus its reserve covers the fee when no entry pays it. It does not check that the fee includes the extra base fee rippled charges for each `SponsorSignature.Signers` entry, nor, for a co-signed inner Batch transaction, that the sponsor also signs the outer `Batch` through `BatchSigners`, which the inner transaction alone cannot show. It does not check `RemainingOwnerCount` either, because the client cannot know how many reserved objects the transaction creates; the entry is returned in `SponsorshipValidation.Sponsorship` for callers that do. rippled remains authoritative.

Use `ValidateSponsorshipContext` when the lookup needs caller cancellation. See the [RPC preflight API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/rpc#Client.ValidateSponsorship) or [WebSocket preflight API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/websocket#Client.ValidateSponsorship) for the result type.

See [submission and finality](/docs/xrpl/submission) for interpreting the validated result.
