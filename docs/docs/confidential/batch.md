# Confidential batches

Use ordered batches when several confidential operations must use the state produced by earlier operations in the same transaction. Start with the [builders guide](/docs/confidential/builders) for standalone operations.

## Build a batch

`BuildBatch` assembles several confidential operations into one XLS-56 `Batch` that the
ledger applies in order. It exists because calling the standalone builders in a row cannot
produce one: each of them reads the ledger, and inside a `Batch` the ledger does not yet show
what an earlier inner leaves behind, so every proof after the first would bind a balance and
a version the transaction will no longer find when it applies.

```go
batch, err := builder.BuildBatch(client, builder.BuildBatchParams{
	Account: senderAddress,
	Operations: []builder.BatchOperation{
		builder.SendOp{BuildSendParams: builder.BuildSendParams{
			Account:       senderAddress,
			Destination:   receiverAddress,
			IssuanceID:    issuanceID,
			Amount:        30,
			SenderPrivKey: senderKey.PrivKeyHex,
			SenderPubKey:  senderKey.PubKeyHex,
			BalanceRange:  elgamal.AmountRange{Low: 0, High: 1_000},
		}},
		builder.MergeInboxOp{BuildMergeInboxParams: builder.BuildMergeInboxParams{
			Account:    receiverAddress,
			IssuanceID: issuanceID,
		}},
	},
})
```

Each of the five confidential operations wraps the parameters of the standalone builder it
mirrors, so an inner reads the same as the call it replaces: `ConvertOp`, `ConvertBackOp`,
`SendOp`, `MergeInboxOp`, and `ClawbackOp`. `TransactionOp` carries a ready-made ordinary
transaction, which the assembler only shapes as an inner. Each operation can be passed as a
value or as a non-nil pointer.

The assembler owns five things:

- **Up-front validation.** Every operation's inputs are checked before the first ledger query,
  by the same validator and with the same sentinels as the standalone builder it mirrors,
  including the `TxOptions` rules. An invalid later operation costs no ledger read and no
  proof for the operations before it.
- **One validated ledger.** Every `MPToken` and `MPTokenIssuance` the `Batch` touches is read
  from a single snapshot, pinned by hash after the first read, so no inner's proof mixes state
  from two ledgers.
- **Predicted state.** A map keyed by the decoded holder `AccountID` and the issuance ID
  carries the spending and inbox ciphertexts, the issuer and auditor mirror balances, the
  holder keys, the balance versions, and the public amounts. After each inner it is advanced
  by exactly what the transactor does, including the re-randomization the network applies to a
  send's credited ciphertexts and the canonical encrypted zero it writes when it resets a
  balance: a merge resets the inbox, a holder's first convert starts its spending balance at
  zero, and a clawback resets every balance of its holder. `elgamal.EncryptCanonicalZero`
  derives that ciphertext from the key, the holder account, and the issuance the same way the
  network does, so a later inner can spend from a reset balance within the same `Batch`. As in
  the standalone builders, an open-ledger version change rejects the build with
  `ErrStaleBalanceVersion` only for a holder whose version a send or convert-back proof binds.
- **Final nonces.** Each inner's `Sequence`, or the `TicketSequence` it spends instead, is
  resolved before any proof is generated, because a confidential context hash commits to the
  nonce and no later autofill can repair a proof. An account's inners take consecutive
  sequences; the outer `Batch` account's start one past the sequence the `Batch` itself spends,
  or at its current sequence when the `Batch` spends a `Ticket`. A `TicketCreate` inner moves
  its account's later sequences past every `Ticket` it creates.
- **Inner shape.** Every inner carries `tfInnerBatchTxn`, a zero `Fee`, an empty
  `SigningPubKey`, and no signature of its own.

`Fee` and `LastLedgerSequence` are left unset, so the returned `Batch` goes through the
client's own autofill, which prices a `Batch` by summing its inners and charges each
confidential inner the multiplier the network applies. Autofill cannot disturb a proof: every
nonce the proofs bind is already set, and autofill assigns only nonces that are missing.
Signing stays with the caller. Each participating account signs with
`wallet.SignMultiBatch`, several signatures are merged with `wallet.CombineBatchSigners`, and
the outer account signs the `Batch` itself:

```go
flat := batch.Flatten()
if err := client.AutofillMultisigned(&flat, 1); err != nil {
	return err
}
if err := wallet.SignMultiBatch(receiverWallet, &flat, nil); err != nil {
	return err
}
response, err := client.SubmitTxAndWait(flat, &rpctypes.SubmitOptions{Wallet: &senderWallet})
if err != nil {
	return err
}
if response.Meta.TransactionResult != "tesSUCCESS" {
	return fmt.Errorf("transaction result: %s", response.Meta.TransactionResult)
}
```

This fragment uses `rpctypes` for `xrpl/rpc/types`. For WebSocket, use `xrpl/websocket/types` instead. `senderWallet` and `receiverWallet` are the account-signing wallets, not the encryption keypairs.

## Batch limits

The assembler refuses to emit a proof it can already tell the network will reject. Each
refusal has its own sentinel:

- `ErrBatchOperationCount`: a `Batch` holds between two and eight inners. The check runs
  before any ledger read, so an impossible size costs nothing.
- `ErrBatchModeNotSupported`: only `tfAllOrNothing`, the default, is supported. Under any
  other mode an inner can be skipped or fail while later inners still apply, and every
  prediction after it would describe a ledger that never happened.
- `ErrBatchMissingOperation`: an operation is nil, as an interface or as a pointer, or a
  `TransactionOp` carries no transaction.
- `ErrBatchInnerNotSupported`: a `TransactionOp` of a type the assembler does not accept.
  `IsSupportedInnerTransactionType` reports the allowlist: `AccountSet`, `SetRegularKey`,
  `SignerListSet`, `TicketCreate`, `TrustSet`, `DepositPreauth`, `DelegateSet`,
  `CredentialCreate`, `CredentialAccept`, and `CredentialDelete`. Anything that could change a
  confidential balance, an `MPToken`'s existence or authorization, or an issuance is kept out,
  because the assembler would have to predict its effect to keep the later proofs valid:
  `MPTokenAuthorize` creates and deletes the `MPToken` the predictions are keyed by,
  `MPTokenIssuanceSet` can lock an issuance or change its keys, and `Payment` and `Clawback`
  can move the public MPT a convert is funded from. Submit those before or after the `Batch`.
- `ErrBatchInnerSequenceSet`: a confidential operation set its own `Sequence`. The assembler
  derives every inner sequence from the operation's position, so a caller-set one describes an
  order it cannot honor. A `TicketSequence` is accepted, and the proof binds it in place of
  the sequence.
- `ErrBatchInnerSequenceMismatch`: a `TransactionOp` carries a `Sequence` that is not the one
  its position in the `Batch` requires for its account, such as the outer `Batch`'s own
  sequence, one past an allocated inner, or one a `TicketCreate` earlier in the `Batch` turned
  into a `Ticket`. This holds for every account, so a caller-set sequence of an account other
  than the outer one is checked against that account's current sequence.
- `ErrConflictingNonce`: an inner, confidential or ready-made, sets both `Sequence` and
  `TicketSequence`. The network requires exactly one.
- `ErrBatchDuplicateNonce`: two inners of one account spend the same sequence or `Ticket`, or
  an inner spends the `Ticket` the outer `Batch` itself spends. The network rejects an
  all-or-nothing `Batch` that repeats a nonce.

Everything the standalone builders reject, a `Batch` inner rejects too, with the same
sentinel: the issuance capability checks, the locked and authorized preflights, the key
mismatches, and the balance bounds. The bounds are checked against the running state rather
than the pre-batch ledger, so a convert earlier in the same `Batch` funds a later convert-back
and widens the decryption bound a later spend searches under.

For protocol fields and authorization, see [Batch](https://xrpl.org/docs/references/protocol/transactions/types/batch). For nonce and proof-state restrictions, see [transaction options](/docs/confidential/builders#transaction-options).
