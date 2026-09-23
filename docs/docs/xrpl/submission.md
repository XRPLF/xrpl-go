# Submission and finality

Use this guide after [constructing a transaction](/docs/xrpl/transaction). The RPC and WebSocket clients share the behavior below.

```text
prepare fields -> autofill -> sign -> submit -> wait for validation -> inspect result
```

**A successful request is not necessarily a successful transaction.** Immediate submission results are preliminary. A validated transaction can have a failed result, including a `tec` result that still consumes a fee.

## Choose a method

| Input and goal | Method |
| --- | --- |
| Signed blob, immediate result | `SubmitTxBlob` |
| Flat transaction, immediate result | `SubmitTx` |
| Signed blob, wait for finality | `SubmitTxBlobAndWait`, `SubmitTxBlobAndWaitContext` |
| Flat transaction, wait for finality | `SubmitTxAndWait`, `SubmitTxAndWaitContext` |
| Combined multisigned blob, immediate result | `SubmitMultisigned` |

For client-side signing, supply `SubmitOptions.Wallet`. Set `Autofill: true` explicitly if network fields are missing. For offline signing, autofill first and submit the resulting blob unchanged.

## Check the final result

In a function returning an error, after creating a client and signed `blob`:

```go
response, err := client.SubmitTxBlobAndWaitContext(ctx, blob, false)
if err != nil {
	return err
}
if response.Meta.TransactionResult != "tesSUCCESS" {
	return fmt.Errorf("transaction result: %s", response.Meta.TransactionResult)
}
```

The request may have reached the server even if a timeout or connection failure prevents you from receiving a response. Keep the signed blob and its transaction hash. Check its status before deciding whether to submit again or build a replacement. The clients have no typed `tx` method, so send a `transactions.TxRequest` through a [generic request](/docs/xrpl/queries#api-versions-and-generic-requests). Do not assume cancellation undoes a submission.

| Outcome | Next action |
| --- | --- |
| Validated response | Inspect `Meta.TransactionResult` before treating the operation as successful |
| `ErrPreliminaryResult` | Inspect the engine result and correct the transaction before preparing a new one |
| `ErrFinalityTransport` or context cancellation | Resolve the original transaction hash before taking another action |
| `ErrTransactionExpired` | Check the reported expiry and available ledger history before preparing a replacement |

Use `errors.Is` for SDK sentinel errors. The [XRPL reliable submission guide](https://xrpl.org/docs/concepts/transactions/reliable-transaction-submission) explains protocol finality and recovery. The sections below document this SDK's additional behavior.

## SDK behavior

### Autofill missing fields

The `Autofill` method fills missing fields in a flat transaction, such as `Sequence`, `LastLedgerSequence`, and `Fee`, and applies the network `NetworkID` policy. It checks selected input rules and returns errors from preparation or network requests. It does not perform full transaction validation. Call the typed transaction's `Validate()` method to check local field rules. Neither validation nor autofill guarantees that the transaction will succeed on the ledger. The `AutofillMultisigned` method provides the same behavior for multisigned transactions. Its `nSigners` argument is the number of signatures the transaction will carry, and the fee grows with it.

Both methods support `Batch` transactions and fill the inner `RawTransactions` and the outer `Batch` transaction. They convert X-addresses in `Account`, `Destination`, `Authorize`, `Unauthorize`, `Owner`, `RegularKey`, `Delegate`, `Sponsor`, `Sponsee`, `CounterpartySponsor`, `NFTokenMinter`, `Subject`, `Issuer`, and `Holder` to classic addresses. Embedded tags in `Account` and `Destination` populate `SourceTag` and `DestinationTag`. A conflicting explicit tag returns `ErrMismatchedTag`. A tagged X-address in any other listed field returns `ErrAccountIDTagNotAllowed`.

Autofill sets `Fee` only when the transaction has none. Special transaction costs are included automatically. A supplied fee is not overwritten and can underpay. See [confidential transaction costs](/docs/confidential#transaction-cost) for those transactions.

### Submit without waiting

The `SubmitTx` and `SubmitTxBlob` methods submit a transaction to the XRPL network. They return a `SubmitResponse` with the immediate submission result. `SubmitTxBlob` requires a signed transaction blob. `SubmitTx` accepts a signed flat transaction, or it can sign an unsigned transaction when `SubmitOptions.Wallet` is set. It enables autofill only when `SubmitOptions.Autofill` is true.

The trailing `bool` of the blob methods and `SubmitOptions.FailHard` request `fail_hard` submission, which stops the server from retrying or relaying a transaction that fails when the server applies it locally. The clients' submit methods always send `AccountDelete` with `fail_hard`, whatever the caller passes.

Submission rejects incomplete, empty, or mixed signing fields before it sends the request. `SubmitMultisigned` requires a structurally complete multisigned transaction blob.

### Wait for validation

The reliable-submission methods require `LastLedgerSequence` before they send the `submit` request. The Go SDK does not enable autofill by default. Provide `LastLedgerSequence` directly or set `Autofill: true` when you submit a transaction that the client can sign.

A missing `engine_result` or a preliminary `tem` result returns `ErrPreliminaryResult` immediately. The error message includes the engine result and its message. The client monitors `tes`, `ter`, `tec`, `tef`, `tel`, and non-empty unknown preliminary results. An exact `txnNotFound` response is inconclusive and the client retries it.

Each polling round waits for the configured interval, requests the latest validated ledger, and then looks up the transaction. The transaction expires only when the validated ledger is strictly greater than `LastLedgerSequence` and the final transaction lookup does not return a validated result. Validation exactly at `LastLedgerSequence` is accepted. The final lookup reduces a race with lagging read backends, but without `searched_all` it does not prove absence from history that the endpoint does not provide.

`WithMaxRetries` limits consecutive incomplete monitoring rounds caused by query or transport errors. A complete round resets the count. Successful pending rounds do not consume this limit. `WithRetryDelay` sets the polling interval, which can be zero but not negative.

Every validated transaction response returns with a nil error, including validated `tec` results. Inspect `TxResponse.Meta.TransactionResult` to determine the validated engine result. `ErrTransactionExpired` reports the preliminary engine result and ledger expiry details. `ErrFinalityTransport` reports repeated query or transport failure and wraps the last failure. Context-aware methods propagate caller cancellation through transaction preparation queries, submission, and finality monitoring, and return `ctx.Err()` directly on cancellation or deadline.

The client verifies that each validated-ledger response is marked as validated and contains a ledger index. A negative polling interval returns `ErrInvalidPollInterval` before submission. A zero or negative maximum retry value returns `ErrInvalidMaxRetries` before submission. A zero `LastLedgerSequence` returns `ErrInvalidLastLedgerSequence` before submission.

### Simulate

`Simulate` runs an XLS-69 dry run against the current open-ledger state. It sends JSON transaction input or a blob without local request preflight or additional network-identity discovery. WebSocket still discovers identity during `Connect`. The server validates the input. Nil-request protection and response validation remain enabled. It returns either decoded or binary transaction and metadata output. Supply an unsigned transaction, and do not send signed transactions to an untrusted node. A simulation does not guarantee the result of a later submission.

See the [XRPL simulate method](https://xrpl.org/docs/references/http-websocket-apis/public-api-methods/transaction-methods/simulate) for accepted fields and the [request type guide](/docs/xrpl/queries#simulate-before-submitting) for the Go input forms.

## Related guides

- [Sponsorship](/docs/xrpl/sponsorship) for preflight and sponsor signing.
- [JSON-RPC](/docs/xrpl/rpc) and [WebSocket](/docs/xrpl/websocket) for configuration.
- [Go RPC API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/rpc) and [Go WebSocket API](https://pkg.go.dev/github.com/Peersyst/xrpl-go/xrpl/websocket) for complete signatures. Use `xrpl/rpc/types` for RPC submission options and `xrpl/websocket/types` for WebSocket options.
